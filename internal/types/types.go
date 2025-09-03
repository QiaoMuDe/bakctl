// Package types 定义了 bakctl 工具的核心数据类型和配置结构。
// 包含备份任务配置、数据库记录结构以及相关的验证和转换功能。
package types

import (
	"path/filepath"

	"gitee.com/MM-Q/bakctl/internal/utils"
	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	// 数据库文件名
	DBFilename = "bakctl.db3"

	// 备份任务配置文件名
	AddTaskFilename = "add_task.toml"

	// 默认数据目录名字
	DataDirName = ".bakctl"

	// 默认备份目录名字
	BackupDirName = "bak"
)

var (
	// 数据目录路径（默认为用户主目录下的.bakctl）
	DataDirPath = filepath.Join(utils.GetUserHomeDir(), DataDirName)

	// 备份目录路径（默认为数据目录下的bak）
	BackupDirPath = filepath.Join(DataDirPath, BackupDirName)
)

// BackupTask 表示数据库中的备份任务记录
// 与AddTaskConfig的区别在于：
// 1. 增加了ID字段（数据库自增主键）
// 2. IncludeRules和ExcludeRules为字符串类型（存储JSON数组格式）
type BackupTask struct {
	ID           int64  `db:"ID" json:"id"`                       // 任务唯一标识（自增主键）
	Name         string `db:"name" json:"name"`                   // 任务名称
	RetainCount  int    `db:"retain_count" json:"retain_count"`   // 保留备份数量
	RetainDays   int    `db:"retain_days" json:"retain_days"`     // 保留天数
	BackupDir    string `db:"backup_dir" json:"backup_dir"`       // 备份源目录
	StorageDir   string `db:"storage_dir" json:"storage_dir"`     // 存储目录
	Compress     bool   `db:"compress" json:"compress"`           // 是否压缩
	IncludeRules string `db:"include_rules" json:"include_rules"` // 包含规则（JSON格式字符串）
	ExcludeRules string `db:"exclude_rules" json:"exclude_rules"` // 排除规则（JSON格式字符串）
	MaxFileSize  int64  `db:"max_file_size" json:"max_file_size"` // 最大文件大小（字节）
	MinFileSize  int64  `db:"min_file_size" json:"min_file_size"` // 最小文件大小（字节）
}

// UpdateTaskParams 封装了更新任务所需的参数
type UpdateTaskParams struct {
	ID           int64  `json:"id"`            // 任务唯一标识（自增主键）
	RetainCount  int    `json:"retain_count"`  // 保留备份数量
	RetainDays   int    `json:"retain_days"`   // 保留天数
	Compress     bool   `json:"compress"`      // 是否压缩
	IncludeRules string `json:"include_rules"` // 包含规则（JSON格式字符串）
	ExcludeRules string `json:"exclude_rules"` // 排除规则（JSON格式字符串）
	MaxFileSize  int64  `json:"max_file_size"` // 最大文件大小（字节）
	MinFileSize  int64  `json:"min_file_size"` // 最小文件大小（字节）
}

// BackupRecord 对应 backup_records 表的结构体（适配 sqlx + SQLite）
// 字段标签说明：
// - db:"列名"：sqlx用于映射SQLite表列，确保与表字段名完全一致
// - json:"字段名"：可选，用于API返回或日志打印（按需保留）
type BackupRecord struct {
	ID             int64  `db:"ID" json:"id"`                                     // 主键（自增）
	TaskID         int64  `db:"task_id" json:"task_id"`                           // 关联的任务ID（外键，非空，关联backup_tasks.ID）
	TaskName       string `db:"task_name" json:"task_name"`                       // 关联的任务名称（冗余存储，非空，便于查询）
	VersionID      string `db:"version_id" json:"version_id"`                     // 备份版本唯一标识（非空+唯一，如UUID/时间戳）
	BackupFilename string `db:"backup_filename" json:"backup_filename"`           // 备份文件名（非空，如"db_20240520.sql.gz"）
	BackupSize     int64  `db:"backup_size" json:"backup_size"`                   // 备份文件大小（非空，单位：字节，用int64支持大文件）
	StoragePath    string `db:"storage_path" json:"storage_path"`                 // 备份文件完整路径（非空，如"/mnt/backup/db_20240520.sql.gz"）
	Status         bool   `db:"status" json:"status"`                             // 备份状态（非空，仅支持true/false）
	FailureMessage string `db:"failure_message" json:"failure_message,omitempty"` // 失败信息（可空，成功时存NULL，用指针接收NULL值）
	Checksum       string `db:"checksum" json:"checksum,omitempty"`               // 校验码（可空，如"MD5:abc123"，用指针接收NULL值）
	CreatedAt      string `db:"created_at" json:"created_at"`                     // 备份时间（默认SQLite自动生成，ISO8601格式字符串，如"2024-05-20T15:30:00Z"）
}

// BackupResult 备份执行结果
type BackupResult struct {
	Success    bool   // 是否成功
	ErrorMsg   string // 错误信息
	BackupPath string // 备份文件路径
	FileSize   int64  // 文件大小
	Checksum   string // 校验码
}

// 定义存放表格样式的MAP
var (
	TableStyle = map[string]table.Style{
		"df":   table.StyleDefault,       // 默认样式
		"bd":   table.StyleBold,          // 加粗样式
		"cb":   table.StyleColoredBright, // 亮色样式
		"cd":   table.StyleColoredDark,   // 暗色样式
		"de":   table.StyleDouble,        // 双边框样式
		"lt":   table.StyleLight,         // 浅色样式
		"ro":   table.StyleRounded,       // 圆角样式
		"none": StyleNone,                // 禁用样式
	}
)

// 定义禁用样式
var StyleNone = table.Style{
	Box: table.BoxStyle{
		PaddingLeft:      " ", // 左边框
		PaddingRight:     " ", // 右边框
		MiddleHorizontal: " ", // 水平线
		MiddleVertical:   " ", // 垂直线
		TopLeft:          " ", // 左上角
		TopRight:         " ", // 右上角
		BottomLeft:       " ", // 左下角
		BottomRight:      " ", // 右下角
	},
}

// 定义存放表格样式的切片
var TableStyleList = []string{"df", "bd", "cb", "cd", "de", "lt", "ro", "none"}

// 常量定义
const (
	HashAlgorithm = "sha1"
	BackupFileExt = ".zip"
)
