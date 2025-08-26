package utils

// initDbScript 数据库初始化脚本
import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// validate 校验核心配置的合法性（确保文件名和目录路径非空）
// 初始化前调用，提前拦截无效配置，避免数据库连接失败
//
// 参数:
//   - dbFilename: 数据库文件名（含后缀）
//   - dataDirPath: 数据库文件所在目录路径
//
// 返回值:
//   - error: 配置校验失败时返回错误信息，否则返回 nil
func validate(dbFilename string, dataDirPath string) error {
	// 校验数据库文件名非空（必须含后缀，如 .db）
	if dbFilename == "" {
		return fmt.Errorf("数据库文件名不能为空, 请指定完整文件名(例: backup_system.db)")
	}

	// 校验数据库文件名含后缀.db或.db3
	if filepath.Ext(dbFilename) != ".db" && filepath.Ext(dbFilename) != ".db3" {
		return fmt.Errorf("数据库文件名必须以 .db 或 .db3 结尾")
	}

	// 校验数据目录路径非空
	if dataDirPath == "" {
		return fmt.Errorf("数据目录路径不能为空, 请指定数据库文件所在目录(例: /var/backup/db)")
	}
	return nil
}

// InitSQLite 基于 DBInitConfig 初始化 SQLite 数据库
//
// 参数:
//   - dbFilename: 待初始化的数据库文件名（含后缀）
//   - dataDirPath: 数据库文件所在目录路径
//
// 返回值:
//   - *sqlx.DB: 初始化成功的数据库连接对象
//   - error: 初始化失败时返回错误信息，否则返回 nil
func InitSQLite(dbFilename string, dataDirPath string) (*sqlx.DB, error) {
	// 先校验配置合法性
	if err := validate(dbFilename, dataDirPath); err != nil {
		return nil, fmt.Errorf("配置校验失败：%w", err)
	}

	// 确保数据目录存在（目录不存在则自动创建，权限 0755）
	if err := os.MkdirAll(dataDirPath, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败：%w", err)
	}

	// 获取完整路径
	dbFullPath := filepath.Join(dataDirPath, dbFilename)

	// 检查数据库文件是否存在
	dbExists := true
	if _, err := os.Stat(dbFullPath); os.IsNotExist(err) {
		dbExists = false
	}

	// 连接数据库
	sqlDB, err := sqlx.Connect("sqlite3", dbFullPath)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败 (路径：%s) :%w", dbFullPath, err)
	}

	// 验证连接可用性
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库 ping 失败：%w", err)
	}

	// 如果数据库文件不存在，则执行初始化脚本
	if !dbExists {
		fmt.Printf("数据库文件不存在，正在创建并执行初始化脚本：%s\n", dbFullPath)
		if _, err := sqlDB.Exec(initDbScript); err != nil {
			return nil, fmt.Errorf("执行数据库初始化脚本失败：%w", err)
		}
		fmt.Printf("数据库文件创建并初始化成功！完整路径：%s\n", dbFullPath)
	} else {
		fmt.Printf("数据库已存在，连接成功！完整路径：%s\n", dbFullPath)
	}

	return sqlDB, nil
}

// 初始化建库脚本
const initDbScript = `-- SQLite 数据库初始化脚本
CREATE TABLE IF NOT EXISTS backup_tasks (
    ID INTEGER PRIMARY KEY AUTOINCREMENT, -- 任务唯一标识，自增主键
    name TEXT NOT NULL UNIQUE,           -- 任务名称，唯一且非空
    retain_count INTEGER DEFAULT 3,      -- 保留备份数量(默认保留3个/每天保留3个)
    retain_days INTEGER DEFAULT 7,       -- 保留天数(默认保留7天)
    backup_dir TEXT NOT NULL,            -- 备份源目录，非空
    storage_dir TEXT NOT NULL,           -- 存储目录，非空
    compress_strategy TEXT,              -- 压缩策略 (none/fast/balanced/best)
    include_rules TEXT,                  -- 包含规则 (JSON数组字符串)
    exclude_rules TEXT,                  -- 排除规则 (JSON数组字符串)
    max_file_size INTEGER,               -- 最大文件大小 (字节)
    min_file_size INTEGER,               -- 最小文件大小 (字节)
    created_at TEXT DEFAULT CURRENT_TIMESTAMP, -- 任务创建时间 (ISO8601格式)
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP  -- 任务最后更新时间 (ISO8601格式)
);

CREATE TABLE IF NOT EXISTS backup_records (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,     -- 记录唯一标识，自增主键
    task_id INTEGER NOT NULL,                 -- 关联的备份任务ID
    task_name TEXT NOT NULL,                  -- 关联的备份任务名称
    version_id TEXT NOT NULL UNIQUE,          -- 备份版本ID，唯一且非空
    backup_filename TEXT NOT NULL,            -- 备份文件名，非空
    backup_size INTEGER NOT NULL,             -- 备份文件大小 (字节)，非空
    status TEXT NOT NULL,                     -- 备份状态(true/false)，非空
    failure_message TEXT,                     -- 备份失败时的错误信息
    checksum TEXT,                            -- 备份文件校验码
    storage_path TEXT NOT NULL,               -- 备份文件存放路径，非空
    created_at TEXT DEFAULT CURRENT_TIMESTAMP -- 备份完成时间 (ISO8601格式)
);

-- backup_tasks 表索引(显式)
CREATE INDEX IF NOT EXISTS idx_backup_tasks_name ON backup_tasks (name);

-- backup_records 表索引
CREATE INDEX IF NOT EXISTS idx_backup_records_created_at ON backup_records(created_at);
CREATE INDEX IF NOT EXISTS idx_backup_records_task_id ON backup_records (task_id); 
CREATE INDEX IF NOT EXISTS idx_backup_records_task_name ON backup_records (task_name);
`

// GetTaskIDByName 根据任务名称从数据库中获取任务ID。
// 如果找到任务，返回任务ID和nil错误；如果未找到，返回0和sql.ErrNoRows错误。
//
// 参数：
//   - db：数据库连接对象
//   - taskName：任务名称
//
// 返回值：
//   - int64：任务ID
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetTaskIDByName(db *sqlx.DB, taskName string) (int64, error) {
	var taskID int64
	query := `SELECT ID FROM backup_tasks WHERE name = ?`
	err := db.Get(&taskID, query, taskName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("未找到名称为 '%s' 的任务", taskName)
		}
		return 0, fmt.Errorf("根据任务名称 '%s' 获取ID失败: %w", taskName, err)
	}
	return taskID, nil
}

// GetTaskNameByID 根据任务ID从数据库中获取任务名称。
// 如果找到任务，返回任务名称和nil错误；如果未找到，返回空字符串和更具体的错误信息。
//
// 参数：
//   - db：数据库连接对象
//   - taskID：任务ID
//
// 返回值：
//   - string：任务名称
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetTaskNameByID(db *sqlx.DB, taskID int64) (string, error) {
	var taskName string
	query := `SELECT name FROM backup_tasks WHERE ID = ?`
	err := db.Get(&taskName, query, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("未找到ID为 %d 的任务", taskID)
		}
		return "", fmt.Errorf("根据任务ID %d 获取名称失败: %w", taskID, err)
	}
	return taskName, nil
}
