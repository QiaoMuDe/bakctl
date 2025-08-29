package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/bakctl/internal/utils"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// validate 校验核心配置的合法性（确保文件名和目录路径非空）
// 初始化前调用，提前拦截无效配置，避免数据库连接失败
//
// 参数:
//   - dbFilename: 数据库文件名（含后缀）
//
// 返回值:
//   - error: 配置校验失败时返回错误信息，否则返回 nil
func validate(dbFilename string) error {
	// 校验数据库文件名非空（必须含后缀，如 .db）
	if dbFilename == "" {
		return fmt.Errorf("数据库文件名不能为空, 请指定完整文件名(例: backup_system.db)")
	}

	// 校验数据库文件名含后缀.db或.db3
	if filepath.Ext(dbFilename) != ".db" && filepath.Ext(dbFilename) != ".db3" {
		return fmt.Errorf("数据库文件名必须以 .db 或 .db3 结尾")
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
	if err := validate(dbFilename); err != nil {
		return nil, fmt.Errorf("配置校验失败：%w", err)
	}

	// 如果数据目录为空，则使用默认的数据目录
	if dataDirPath == "" {
		dataDirPath = types.DataDirPath
	}

	// 确保数据目录存在（目录不存在则自动创建，权限 0755）
	if err := os.MkdirAll(dataDirPath, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败：%w", err)
	}

	// 确保备份目录存在（目录不存在则自动创建，权限 0755）
	if err := os.MkdirAll(types.BackupDirPath, 0755); err != nil {
		return nil, fmt.Errorf("创建备份目录失败：%w", err)
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
    compress BOOLEAN DEFAULT FALSE,      -- 是否压缩 (true/false)
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
    status BOOLEAN NOT NULL,                  -- 备份状态(true/false)，非空
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

// GetTaskByID 根据任务ID从数据库中获取完整的任务信息。
//
// 参数：
//   - db：数据库连接对象
//   - taskID：任务ID
//
// 返回值：
//   - *types.BackupTask：任务信息
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetTaskByID(db *sqlx.DB, taskID int64) (*types.BackupTask, error) {
	var task types.BackupTask
	query := `SELECT ID, name, retain_count, retain_days, backup_dir, storage_dir, compress, include_rules, exclude_rules, max_file_size, min_file_size FROM backup_tasks WHERE ID = ?`
	err := db.Get(&task, query, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到ID为 %d 的任务", taskID)
		}
		return nil, fmt.Errorf("根据任务ID %d 获取任务信息失败: %w", taskID, err)
	}
	return &task, nil
}

// InsertAddTaskConfig 将 AddTaskConfig 结构体的数据插入到 backup_tasks 表中。
// 它将 []string 类型的规则字段转换为 JSON 字符串进行存储。
//
// 参数：
//   - db：数据库连接对象
//   - cfg：要插入的 AddTaskConfig 结构体
//
// 返回值：
//   - error：如果插入过程中发生错误, 则返回非 nil 错误信息
func InsertAddTaskConfig(db *sqlx.DB, cfg *types.AddTaskConfig) error {
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 处理包含规则
	includeRulesJSON, err := utils.MarshalRules(cfg.IncludeRules)
	if err != nil {
		return fmt.Errorf("编码包含规则失败: %w", err)
	}

	// 处理排除规则
	excludeRulesJSON, err := utils.MarshalRules(cfg.ExcludeRules)
	if err != nil {
		return fmt.Errorf("编码排除规则失败: %w", err)
	}

	// 构建实际存储目录：存储目录 + 备份源目录的目录名
	actualStorageDir := filepath.Join(cfg.StorageDir, filepath.Base(cfg.BackupDir))

	// 将 AddTaskConfig 转换为 BackupTask, 处理规则字段的 JSON 编码
	backupTask := types.BackupTask{
		Name:         cfg.Name,         // 任务名称
		RetainCount:  cfg.RetainCount,  // 保留备份数量
		RetainDays:   cfg.RetainDays,   // 保留天数
		BackupDir:    cfg.BackupDir,    // 备份源目录
		StorageDir:   actualStorageDir, // 存储目录（存储目录 + 备份源目录名）
		Compress:     cfg.Compress,     // 是否压缩
		IncludeRules: includeRulesJSON, // 包含规则
		ExcludeRules: excludeRulesJSON, // 排除规则
		MaxFileSize:  cfg.MaxFileSize,  // 最大文件大小
		MinFileSize:  cfg.MinFileSize,  // 最小文件大小
	}

	// 执行插入操作
	_, err = db.NamedExec(insertBackupTaskQuery, backupTask)
	if err != nil {
		return fmt.Errorf("插入备份任务失败: %w", err)
	}

	return nil
}

// SQL INSERT 语句，使用命名参数
// 注意：created_at 和 updated_at 字段在表定义中通常有 DEFAULT CURRENT_TIMESTAMP，
// 所以这里不需要显式插入它们，数据库会自动处理。
const insertBackupTaskQuery = `
	INSERT INTO backup_tasks (
		name,
		retain_count,
		retain_days,
		backup_dir,
		storage_dir,
		compress,
		include_rules,
		exclude_rules,
		max_file_size,
		min_file_size
	) VALUES (
		:name,
		:retain_count,
		:retain_days,
		:backup_dir,
		:storage_dir,
		:compress,
		:include_rules,
		:exclude_rules,
		:max_file_size,
		:min_file_size
	)`

// SQL INSERT 语句，用于 backup_records 表
const insertBackupRecordQuery = `
	INSERT INTO backup_records (
		task_id,
		task_name,
		version_id,
		backup_filename,
		backup_size,
		status,
		failure_message,
		checksum,
		storage_path
	) VALUES (
		:task_id,
		:task_name,
		:version_id,
		:backup_filename,
		:backup_size,
		:status,
		:failure_message,
		:checksum,
		:storage_path
	)`

// InsertBackupRecord 将 BackupRecord 结构体的数据插入到 backup_records 表中。
//
// 参数：
//   - db：数据库连接对象
//   - rec：要插入的 BackupRecord 结构体
//
// 返回值：
//   - error：如果插入过程中发生错误，则返回非 nil 错误信息
func InsertBackupRecord(db *sqlx.DB, rec *types.BackupRecord) error {
	// 执行插入操作
	// sqlx 的 NamedExec 方法会根据结构体的 db 标签自动映射字段
	_, err := db.NamedExec(insertBackupRecordQuery, rec)
	if err != nil {
		return fmt.Errorf("插入备份记录失败: %w", err)
	}

	return nil
}

// queryGetAllBackupTasks SQL SELECT 语句，用于查询所有备份任务
const queryGetAllBackupTasks = `
	SELECT
		ID,
		name,
		retain_count,
		retain_days,
		backup_dir,
		storage_dir,
		compress,
		include_rules,
		exclude_rules,
		max_file_size,
		min_file_size
	FROM backup_tasks
`

// GetAllBackupTasks 从数据库中获取所有备份任务。
//
// 参数：
//   - db：数据库连接对象
//
// 返回值：
//   - []types.BackupTask：所有备份任务的切片
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetAllBackupTasks(db *sqlx.DB) ([]types.BackupTask, error) {
	var tasks []types.BackupTask
	err := db.Select(&tasks, queryGetAllBackupTasks)
	if err != nil {
		return nil, fmt.Errorf("获取所有备份任务失败: %w", err)
	}
	return tasks, nil
}

// queryGetAllBackupRecords SQL SELECT 语句，用于查询所有备份记录
const queryGetAllBackupRecords = `
	SELECT
		ID,
		task_id,
		task_name,
		version_id,
		backup_filename,
		backup_size,
		status,
		failure_message,
		checksum,
		storage_path,
		created_at
	FROM backup_records
`

// GetAllBackupRecords 从数据库中获取所有备份记录。
//
// 参数：
//   - db：数据库连接对象
//
// 返回值：
//   - []types.BackupRecord：所有备份记录的切片
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetAllBackupRecords(db *sqlx.DB) ([]types.BackupRecord, error) {
	var records []types.BackupRecord
	err := db.Select(&records, queryGetAllBackupRecords)
	if err != nil {
		return nil, fmt.Errorf("获取所有备份记录失败: %w", err)
	}
	return records, nil
}
