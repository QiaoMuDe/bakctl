# CBK命令行备份管理工具重构方案

[TOC]

## 重构后的名字和介绍

`cbk` --> `bakctl`

`bakctl`如何是一个专注命令行备份和备份恢复的的命令行工具，支持对于备份任务的增删改查。

## 子命令和功能

1. **add**：添加备份任务(√)
2. **run**：运行备份任务
3. **export**：导出备份任务
4. **list**：查看备份任务(√)
5. **log**：查看备份记录(√)
6. **restore**：恢复备份
7. **delete**：删除任务
8. **edit**：编辑任务(√)

### add 添加备份任务

add 子命令，主要用于根据添加文件或者命令行参数添加备份任务

| 长标志 | 短标志 | 描述 |
| ------ | ------ | ---- |
| `--config` | `-c` | 指定包含备份任务配置的文件路径 |
| `--generate-template` | `-g` | 生成添加备份任务的模板配置文件（默认输出到当前目录，文件名可自定义） |



## 数据库设计

### 备份任务表 backup_tasks

| 字段名 | SQLite类型 | 描述 | 备注 |
|--------|------------|------|------|
| ID | INTEGER PRIMARY KEY AUTOINCREMENT | 任务唯一标识 | 自增主键 |
| name | TEXT NOT NULL UNIQUE | 任务名称 | 唯一约束 |
| retain_count | INTEGER DEFAULT 5 | 保留备份数量 | 默认值5 |
| retain_days | INTEGER DEFAULT 30 | 保留天数 | 默认值30 |
| backup_dir | TEXT NOT NULL | 备份源目录 | 支持多个路径用分隔符 |
| storage_dir | TEXT NOT NULL | 存储目录 | 备份文件存放位置 |
| compress_strategy | TEXT DEFAULT 'balanced' | 压缩策略 | none/fast/balanced/best |
| include_rules | TEXT | 包含规则 | JSON格式存储多个规则 |
| exclude_rules | TEXT | 排除规则 | JSON格式存储多个规则 |
| max_file_size | INTEGER | 最大文件大小 | 单位：字节 |
| min_file_size | INTEGER | 最小文件大小 | 单位：字节 |

### 备份记录表 backup_records

| 字段名 | SQLite类型 | 描述 | 备注 |
|---|---|---|---|
| ID | INTEGER PRIMARY KEY AUTOINCREMENT | 记录唯一标识 | 自增主键 |
| task_id | INTEGER NOT NULL | 关联的任务ID | 外键，关联 `backup_tasks` 表的 `ID` 字段 |
| task_name | TEXT NOT NULL | 任务名称 | 便于查询，可冗余存储 |
| version_id | TEXT NOT NULL UNIQUE | 备份版本ID | 每次备份的唯一标识，例如时间戳或UUID |
| backup_filename | TEXT NOT NULL | 备份文件名 | 生成的备份文件名 |
| backup_size | INTEGER NOT NULL | 备份文件大小 | 单位：字节 |
| status | TEXT NOT NULL | 备份状态 | 'success' 或 'failure' |
| failure_message | TEXT | 失败信息 | 备份失败时的具体错误信息，成功时为空或 '-' |
| checksum | TEXT | 备份文件校验码 | 例如 MD5, SHA256 等 |
| storage_path | TEXT NOT NULL | 备份文件存放路径 | 备份文件在存储目录下的相对路径或完整路径 |
| created_at | TEXT DEFAULT CURRENT_TIMESTAMP | 备份时间 | 记录备份完成的时间，ISO8601格式 |

### 创建语句

```sql
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
```

## 添加任务配置设计

```toml
# ==============================
# 备份任务添加配置文件
# 说明：1. 带 "required" 标记的字段为必填项；2. 数组类型字段支持多元素配置；3. 单位未特别说明时，尺寸单位为「字节」
# ==============================
[addtask]
# 1. 任务基础信息
# 任务名称（必填，唯一，不可重复）
name = "每日数据库备份"  # required
# 备份源目录（必填，支持多路径，用英文逗号分隔；示例：Windows路径"D:/data1,D:/data2"，Linux路径"/home/data1,/var/log"）
backup_dir = "/var/lib/mysql,/etc/nginx/conf"  # required
# 备份存储目录（必填，单个路径，备份文件最终存放位置）
storage_dir = "/mnt/backup/database"  # required


# 2. 备份保留策略（二选一或同时生效，满足任一条件即清理旧备份）
# 保留备份文件的数量（可选，默认5个；设置为0表示不按数量限制）
retain_count = 7
# 保留备份文件的天数（可选，默认30天；设置为0表示不按天数限制）
retain_days = 90


# 3. 压缩策略（可选，默认"balanced"；支持值：none(不压缩)/fast(快速压缩)/balanced(平衡压缩)/best(高压缩比)）
compress_strategy = "fast"


# 4. 文件过滤规则（数组类型，支持多规则配置；规则语法：支持通配符*（匹配任意字符）、?（匹配单个字符），示例："*.log"匹配所有日志文件）
# 包含规则（可选，仅备份符合规则的文件；空数组表示备份所有文件）
include_rules = [
  "*.sql",          # 包含所有SQL文件
  "*.conf",         # 包含所有配置文件
  "/var/lib/mysql/*"# 包含指定目录下的所有文件（补充路径维度过滤）
]
# 排除规则（可选，不备份符合规则的文件；优先级高于包含规则，即"先包含后排除"）
exclude_rules = [
  "*.tmp",          # 排除所有临时文件
  "*.log.bak",      # 排除日志备份文件
  "/var/lib/mysql/binlog/*"  # 排除二进制日志目录（避免冗余）
]


# 5. 文件大小过滤（可选，单位：字节；0表示不限制）
# 最大文件大小（可选，超过此尺寸的文件不备份；示例：1073741824 = 1GB）
max_file_size = 1073741824
# 最小文件大小（可选，小于此尺寸的文件不备份；示例：1024 = 1KB）
min_file_size = 1024
```

## 结构体设计

> **备份任务表**

```go
package backup

// AddTaskConfig 表示添加备份任务的配置结构
// 对应TOML配置文件中的[addtask]部分
type AddTaskConfig struct {
	Name             string   `toml:"name" json:"name"`                      // 任务名称（必填，唯一）
	BackupDir        string   `toml:"backup_dir" json:"backup_dir"`          // 备份源目录（必填，多路径用逗号分隔）
	StorageDir       string   `toml:"storage_dir" json:"storage_dir"`        // 备份存储目录（必填）
	RetainCount      int      `toml:"retain_count" json:"retain_count"`      // 保留备份数量（默认5，0表示不限制）
	RetainDays       int      `toml:"retain_days" json:"retain_days"`        // 保留备份天数（默认30，0表示不限制）
	CompressStrategy string   `toml:"compress_strategy" json:"compress_strategy"` // 压缩策略（默认"balanced"）
	IncludeRules     []string `toml:"include_rules" json:"include_rules"`    // 包含规则（数组，空数组表示备份所有文件）
	ExcludeRules     []string `toml:"exclude_rules" json:"exclude_rules"`    // 排除规则（数组，优先级高于包含规则）
	MaxFileSize      int64    `toml:"max_file_size" json:"max_file_size"`    // 最大文件大小（字节，0表示不限制）
	MinFileSize      int64    `toml:"min_file_size" json:"min_file_size"`    // 最小文件大小（字节，0表示不限制）
}

// BackupTask 表示数据库中的备份任务记录
// 与AddTaskConfig的区别在于：
// 1. 增加了ID字段（数据库自增主键）
// 2. IncludeRules和ExcludeRules为字符串类型（存储JSON格式）
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

```

> **备份记录表**

```go
package backup

import "time"

// BackupRecord 对应 backup_records 表的结构体（适配 sqlx + SQLite）
// 字段标签说明：
// - db:"列名"：sqlx用于映射SQLite表列，确保与表字段名完全一致
// - json:"字段名"：可选，用于API返回或日志打印（按需保留）
type BackupRecord struct {
	// 1. 主键与外键关联（对应表中自增主键和任务关联字段）
	ID              int64     `db:"ID" json:"id"`              // 记录唯一标识（SQLite自增主键，INTEGER类型）
	TaskID          int64     `db:"task_id" json:"task_id"`    // 关联的任务ID（外键，非空，关联backup_tasks.ID）
	TaskName        string    `db:"task_name" json:"task_name"`// 关联的任务名称（冗余存储，非空，便于查询）

	// 2. 备份版本与文件核心信息（确保唯一性和可定位）
	VersionID       string    `db:"version_id" json:"version_id"`       // 备份版本唯一标识（非空+唯一，如UUID/时间戳）
	BackupFilename  string    `db:"backup_filename" json:"backup_filename"` // 备份文件名（非空，如"db_20240520.sql.gz"）
	BackupSize      int64     `db:"backup_size" json:"backup_size"`     // 备份文件大小（非空，单位：字节，用int64支持大文件）
	StoragePath     string    `db:"storage_path" json:"storage_path"`   // 备份文件完整路径（非空，如"/mnt/backup/db_20240520.sql.gz"）

	// 3. 备份状态与校验信息（区分成功/失败场景）
	Status          string    `db:"status" json:"status"`                // 备份状态（非空，仅支持true/false）
	FailureMessage  *string   `db:"failure_message" json:"failure_message,omitempty"` // 失败信息（可空，成功时存NULL，用指针接收NULL值）
	Checksum        *string   `db:"checksum" json:"checksum,omitempty"`  // 校验码（可空，如"MD5:abc123"，用指针接收NULL值）

	// 4. 时间字段（适配SQLite CURRENT_TIMESTAMP，用string接收避免类型转换问题）
	CreatedAt       string    `db:"created_at" json:"created_at"` // 备份时间（默认SQLite自动生成，ISO8601格式字符串，如"2024-05-20T15:30:00Z"）
}

// 备份状态常量（避免硬编码，统一业务规范）
const (
	BackupStatusSuccess = "true" // 备份成功
	BackupStatusFailure = "false" // 备份失败
)
```

