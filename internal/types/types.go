package types

const (
	// 数据库文件名
	DBFilename = "bakctl.db3"

	// 数据库文件所在目录
	DataDirPath = "./test"

	// 备份状态常量（避免硬编码，统一业务规范）
	BackupStatusSuccess = "true"  // 备份成功
	BackupStatusFailure = "false" // 备份失败

)

// RootConfig 根配置结构体，用于解析TOML配置文件
type RootConfig struct {
	AddTask AddTaskConfig `toml:"AddTaskConfig"`
}

// AddTaskConfig 表示添加备份任务的配置结构
// 对应TOML配置文件中的[AddTaskConfig]部分
type AddTaskConfig struct {
	Name             string   `toml:"name" comment:"任务名称（必填，唯一，不可重复）"`                                         // 任务名称
	BackupDir        string   `toml:"backup_dir" comment:"备份源目录（必填，单个路径，支持Windows和Linux路径）"`                   // 备份源目录
	StorageDir       string   `toml:"storage_dir" comment:"备份存储目录（必填，单个路径，备份文件最终存放位置）"`                        // 备份存储目录
	RetainCount      int      `toml:"retain_count" comment:"保留备份文件的数量（可选，默认3个；设置为0表示不按数量限制）"`                  // 保留备份文件的数量
	RetainDays       int      `toml:"retain_days" comment:"保留备份文件的天数（可选，默认7天；设置为0表示不按天数限制）"`                   // 保留备份文件的天数
	CompressStrategy string   `toml:"compress_strategy" comment:"压缩策略（可选，默认\"fast\"；支持\"fast\"、\"full\"两种模式）"` // 压缩策略
	IncludeRules     []string `toml:"include_rules" comment:"包含规则（可选，仅备份符合规则的文件；空数组表示备份所有文件）"`                 // 包含规则
	ExcludeRules     []string `toml:"exclude_rules" comment:"排除规则（可选，不备份符合规则的文件；优先级高于包含规则，即\"先包含后排除\"）"`       // 排除规则
	MaxFileSize      int64    `toml:"max_file_size" comment:"最大文件大小（可选，超过此尺寸的文件不备份；示例：1073741824 = 1GB）"`      // 最大文件大小
	MinFileSize      int64    `toml:"min_file_size" comment:"最小文件大小（可选，小于此尺寸的文件不备份；示例：1024 = 1KB）"`            // 最小文件大小
}

// BackupTask 表示数据库中的备份任务记录
// 与AddTaskConfig的区别在于：
// 1. 增加了ID字段（数据库自增主键）
// 2. IncludeRules和ExcludeRules为字符串类型（存储JSON数组格式）
type BackupTask struct {
	ID               int64  `json:"id"`                // 任务唯一标识（自增主键）
	Name             string `json:"name"`              // 任务名称
	RetainCount      int    `json:"retain_count"`      // 保留备份数量
	RetainDays       int    `json:"retain_days"`       // 保留天数
	BackupDir        string `json:"backup_dir"`        // 备份源目录
	StorageDir       string `json:"storage_dir"`       // 存储目录
	CompressStrategy string `json:"compress_strategy"` // 压缩策略
	IncludeRules     string `json:"include_rules"`     // 包含规则（JSON格式字符串）
	ExcludeRules     string `json:"exclude_rules"`     // 排除规则（JSON格式字符串）
	MaxFileSize      int64  `json:"max_file_size"`     // 最大文件大小（字节）
	MinFileSize      int64  `json:"min_file_size"`     // 最小文件大小（字节）
}

// BackupRecord 对应 backup_records 表的结构体（适配 sqlx + SQLite）
// 字段标签说明：
// - db:"列名"：sqlx用于映射SQLite表列，确保与表字段名完全一致
// - json:"字段名"：可选，用于API返回或日志打印（按需保留）
type BackupRecord struct {
	// 1. 主键与外键关联（对应表中自增主键和任务关联字段）
	ID       int64  `db:"ID" json:"id"`               // 记录唯一标识（SQLite自增主键，INTEGER类型）
	TaskID   int64  `db:"task_id" json:"task_id"`     // 关联的任务ID（外键，非空，关联backup_tasks.ID）
	TaskName string `db:"task_name" json:"task_name"` // 关联的任务名称（冗余存储，非空，便于查询）

	// 2. 备份版本与文件核心信息（确保唯一性和可定位）
	VersionID      string `db:"version_id" json:"version_id"`           // 备份版本唯一标识（非空+唯一，如UUID/时间戳）
	BackupFilename string `db:"backup_filename" json:"backup_filename"` // 备份文件名（非空，如"db_20240520.sql.gz"）
	BackupSize     int64  `db:"backup_size" json:"backup_size"`         // 备份文件大小（非空，单位：字节，用int64支持大文件）
	StoragePath    string `db:"storage_path" json:"storage_path"`       // 备份文件完整路径（非空，如"/mnt/backup/db_20240520.sql.gz"）

	// 3. 备份状态与校验信息（区分成功/失败场景）
	Status         string  `db:"status" json:"status"`                             // 备份状态（非空，仅支持true/false）
	FailureMessage *string `db:"failure_message" json:"failure_message,omitempty"` // 失败信息（可空，成功时存NULL，用指针接收NULL值）
	Checksum       *string `db:"checksum" json:"checksum,omitempty"`               // 校验码（可空，如"MD5:abc123"，用指针接收NULL值）

	// 4. 时间字段（适配SQLite CURRENT_TIMESTAMP，用string接收避免类型转换问题）
	CreatedAt string `db:"created_at" json:"created_at"` // 备份时间（默认SQLite自动生成，ISO8601格式字符串，如"2024-05-20T15:30:00Z"）
}
