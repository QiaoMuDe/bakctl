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
8. **edit**：编辑任务

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

## 模板sql语句

以下是适配你的表结构和索引设计的 **`backup_tasks` 任务表** 和 **`backup_records` 记录表** 插入语句模板，包含「必填字段强制填写」「可选字段灵活赋值」「默认值自动生效」的逻辑，同时贴合 SQLite 语法特性：


### 一、备份任务表（`backup_tasks`）插入模板
#### 核心逻辑：
- **必填字段**：`name`、`backup_dir`、`storage_dir`（表结构中 `NOT NULL` 约束，必须赋值）；
- **可选字段**：`retain_count`（默认3）、`retain_days`（默认7）、`compress_strategy` 等（不赋值时自动使用表定义的默认值）；
- **特殊字段**：`created_at`/`updated_at`（表定义 `DEFAULT CURRENT_TIMESTAMP`，不赋值时自动生成当前时间）。

```sql
-- 备份任务表插入模板
-- 说明：1. 【】内为需替换的业务值，2. 可选字段可删除（删除后用默认值），3. 包含/排除规则需传入JSON字符串
INSERT INTO backup_tasks (
    name,                -- 必选：任务名称（唯一，不可重复）
    retain_count,        -- 可选：保留备份数量（默认3，删除则用默认值）
    retain_days,         -- 可选：保留天数（默认7，删除则用默认值）
    backup_dir,          -- 必选：备份源目录（多路径用逗号分隔，如 "D:/data1,D:/data2"）
    storage_dir,         -- 必选：备份存储目录（单路径，如 "/mnt/backup"）
    compress_strategy,   -- 可选：压缩策略（none/fast/balanced/best，不填则为NULL）
    include_rules,       -- 可选：包含规则（JSON数组字符串，如 '["*.sql", "*.conf"]'）
    exclude_rules,       -- 可选：排除规则（JSON数组字符串，如 '["*.tmp", "*.log.bak"]'）
    max_file_size,       -- 可选：最大文件大小（字节，如 1073741824 表示1GB，不填则为NULL）
    min_file_size        -- 可选：最小文件大小（字节，如 1024 表示1KB，不填则为NULL）
    -- created_at/updated_at：不写，自动用当前时间（表定义默认值）
) VALUES (
    '【任务名称，如：数据库每日备份】',  -- 示例：'MySQL_Prod_Daily_Backup'
    【保留数量，如：5】,               -- 示例：5（覆盖默认3，不填则删除此值和对应字段）
    【保留天数，如：15】,              -- 示例：15（覆盖默认7，不填则删除此值和对应字段）
    '【备份源目录，如：/var/lib/mysql】',-- 示例：'/var/lib/mysql,/etc/nginx'（多路径逗号分隔）
    '【存储目录，如：/mnt/backup/mysql】',-- 示例：'/mnt/backup/mysql'
    '【压缩策略，如：balanced】',       -- 示例：'fast'（不填则删除此值和对应字段）
    '["【包含规则1，如：*.sql】", "【包含规则2，如：/var/lib/mysql/*】"]',-- 示例：'["*.sql", "/var/lib/mysql/*"]'
    '["【排除规则1，如：*.tmp】", "【排除规则2，如：/var/lib/mysql/binlog/*】"]',-- 示例：'["*.tmp", "/var/lib/mysql/binlog/*"]'
    【最大文件大小，如：1073741824】,  -- 示例：1073741824（1GB，不填则删除此值和对应字段）
    【最小文件大小，如：1024】         -- 示例：1024（1KB，不填则删除此值和对应字段）
);
```

#### 实际使用示例（填充真实值）：
```sql
INSERT INTO backup_tasks (
    name,
    retain_count,
    backup_dir,
    storage_dir,
    compress_strategy,
    include_rules,
    exclude_rules
) VALUES (
    'MySQL_Prod_Daily_Backup',
    5,
    '/var/lib/mysql,/etc/mysql/conf.d',
    '/mnt/backup/mysql_prod',
    'fast',
    '["*.sql", "*.cnf", "/var/lib/mysql/*"]',
    '["*.tmp", "*.log.bak", "/var/lib/mysql/binlog/*"]'
);
```


### 二、备份记录表（`backup_records`）插入模板
#### 核心逻辑：
- **必填字段**：`task_id`、`task_name`、`version_id`、`backup_filename`、`backup_size`、`status`、`storage_path`（表结构 `NOT NULL` 约束，必须赋值）；
- **可选字段**：`failure_message`（成功时填 `NULL` 或空字符串）、`checksum`（不填则为 `NULL`）；
- **特殊字段**：`created_at`（表定义 `DEFAULT CURRENT_TIMESTAMP`，不赋值时自动生成备份时间）。

```sql
-- 备份记录表插入模板
-- 说明：1. 【】内为需替换的业务值，2. status仅支持 'true'/'false'（表结构注释约束），3. 失败信息仅在status为false时填写
INSERT INTO backup_records (
    task_id,             -- 必选：关联的任务ID（需存在于backup_tasks表，外键逻辑）
    task_name,           -- 必选：关联的任务名称（冗余存储，与task_id对应，如 "MySQL_Prod_Daily_Backup"）
    version_id,          -- 必选：备份版本ID（唯一，如 UUID/时间戳，示例：'BACKUP-1-20240826153000'）
    backup_filename,     -- 必选：备份文件名（含后缀，如 "mysql_prod_20240826153000.sql.gz"）
    backup_size,         -- 必选：备份文件大小（字节，如 52428800 表示50MB）
    status,              -- 必选：备份状态（仅支持 'true' 或 'false'）
    failure_message,     -- 可选：失败信息（status为false时填写，如 "连接数据库超时"；true时填 NULL 或 ''）
    checksum,            -- 可选：校验码（如 MD5/SHA256，格式建议带算法标识，如 "MD5:abc123def456"）
    storage_path         -- 必选：备份文件完整路径（如 "/mnt/backup/mysql_prod/mysql_prod_20240826153000.sql.gz"）
    -- created_at：不写，自动用当前时间（表定义默认值）
) VALUES (
    【关联任务ID，如：1】,                  -- 示例：1（需确保backup_tasks表有ID=1的任务）
    '【关联任务名称，如：MySQL_Prod_Daily_Backup】',-- 示例：'MySQL_Prod_Daily_Backup'（与task_id对应）
    '【备份版本ID，如：BACKUP-1-20240826153000】',-- 示例：'BACKUP-1-20240826153000'（确保唯一）
    '【备份文件名，如：mysql_prod_20240826153000.sql.gz】',-- 示例：'mysql_prod_20240826153000.sql.gz'
    【备份文件大小（字节），如：52428800】,  -- 示例：52428800（50MB）
    '【备份状态，true/false】',             -- 示例：'true'（成功）或 'false'（失败）
    【失败信息，如：'连接数据库超时'】,       -- 示例：NULL（成功时）或 '连接数据库超时'（失败时，需加引号）
    '【校验码，如：MD5:abc123def456】',     -- 示例：'MD5:e80b5017098950fc58aad83c8c14978e'（不填则删除此值和对应字段）
    '【完整存储路径，如：/mnt/backup/mysql_prod/mysql_prod_20240826153000.sql.gz】'-- 示例：'/mnt/backup/mysql_prod/mysql_prod_20240826153000.sql.gz'
);
```

#### 实际使用示例（分「成功」和「失败」场景）：
##### 场景1：备份成功
```sql
INSERT INTO backup_records (
    task_id,
    task_name,
    version_id,
    backup_filename,
    backup_size,
    status,
    checksum,
    storage_path
) VALUES (
    1,
    'MySQL_Prod_Daily_Backup',
    'BACKUP-1-20240826153000',
    'mysql_prod_20240826153000.sql.gz',
    52428800,
    'true',
    'MD5:e80b5017098950fc58aad83c8c14978e',
    '/mnt/backup/mysql_prod/mysql_prod_20240826153000.sql.gz'
);
```

##### 场景2：备份失败
```sql
INSERT INTO backup_records (
    task_id,
    task_name,
    version_id,
    backup_filename,
    backup_size,
    status,
    failure_message,
    storage_path
) VALUES (
    1,
    'MySQL_Prod_Daily_Backup',
    'BACKUP-1-20240826160000',
    'mysql_prod_20240826160000.sql.gz',
    0,  -- 失败时文件大小可填0
    'false',
    '连接数据库超时：无法访问 192.168.1.100:3306',
    '/mnt/backup/mysql_prod/mysql_prod_20240826160000.sql.gz'  -- 即使失败也可记录预期路径
);
```


### 三、插入前的关键校验（避免报错）
1. **`backup_tasks.name` 唯一性**：插入任务前，先校验名称是否已存在（避免唯一键冲突）：
   ```sql
   SELECT COUNT(1) FROM backup_tasks WHERE name = '【待插入的任务名称】';
   ```
2. **`backup_records.version_id` 唯一性**：插入记录前，校验版本ID是否重复：
   ```sql
   SELECT COUNT(1) FROM backup_records WHERE version_id = '【待插入的版本ID】';
   ```
3. **`task_id` 合法性**：插入记录前，确保 `task_id` 在 `backup_tasks` 表中存在（避免无效关联）：
   ```sql
   SELECT COUNT(1) FROM backup_tasks WHERE ID = 【待插入的task_id】;
   ```
4. **`status` 取值合法性**：严格按 `'true'`/`'false'` 赋值（与表结构注释约束一致，避免业务逻辑混乱）。

## 结构体方法

> **备份任务表**

该表需要实现插入方法(内部封装插入语句)，查询验证方法，查询方法

> **备份记录表**

该表需要实现插入方法(内部封装插入语句)，查询验证方法，查询方法

## 初始化数据库

需要实现一个结构体，用于存储初始化数据库相关的信息

```go
package db

import (
	"fmt"
	"path/filepath"
)

// DBInitConfig 数据库初始化配置结构体（极简版）
// 仅存储 SQLite 初始化必需的核心信息：数据库文件名 + 数据目录路径
type DBInitConfig struct {
	DBFilename  string `json:"db_filename" yaml:"db_filename"`  // 数据库文件名（含后缀），例："backup_system.db"
	DataDirPath string `json:"data_dir_path" yaml:"data_dir_path"` // 数据目录路径（数据库文件所在目录），例："/var/backup/db" 或 "D:\\backup\\db"
}

// GetDBFullPath 生成数据库文件的完整路径
// 自动适配 Windows/Linux 跨平台路径分隔符（无需手动处理 \ 或 /）
func (c *DBInitConfig) GetDBFullPath() string {
	return filepath.Join(c.DataDirPath, c.DBFilename)
}

// Validate 校验核心配置的合法性（确保文件名和目录路径非空）
// 初始化前调用，提前拦截无效配置，避免数据库连接失败
func (c *DBInitConfig) Validate() error {
	// 校验数据库文件名非空（必须含后缀，如 .db）
	if c.DBFilename == "" {
		return fmt.Errorf("数据库文件名不能为空，请指定完整文件名（例：backup_system.db）")
	}
	// 校验数据目录路径非空
	if c.DataDirPath == "" {
		return fmt.Errorf("数据目录路径不能为空，请指定数据库文件所在目录（例：/var/backup/db）")
	}
	return nil
}
```

```go
package main

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"your-project-path/db" // 导入上面定义的 DBInitConfig
)

// InitSQLite 基于 DBInitConfig 初始化 SQLite 数据库
func InitSQLite(config *db.DBInitConfig) (*sqlx.DB, error) {
	// 1. 先校验配置合法性
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置校验失败：%w", err)
	}

	// 2. 确保数据目录存在（目录不存在则自动创建，权限 0755）
	if err := os.MkdirAll(config.DataDirPath, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败：%w", err)
	}

	// 3. 获取完整路径，连接数据库
	dbFullPath := config.GetDBFullPath()
	sqlDB, err := sqlx.Connect("sqlite3", dbFullPath)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败（路径：%s）：%w", dbFullPath, err)
	}

	// 4. 验证连接可用性
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库 ping 失败：%w", err)
	}

	fmt.Printf("数据库初始化成功！完整路径：%s\n", dbFullPath)
	return sqlDB, nil
}

func main() {
	// 1. 构造极简配置（仅需指定文件名和目录路径）
	config := &db.DBInitConfig{
		DBFilename:  "backup_system.db",   // 数据库文件名（含 .db 后缀）
		DataDirPath: "/var/backup/db",     // Linux 数据目录示例（Windows 可改为 "D:\\backup\\db"）
	}

	// 2. 初始化数据库
	sqlDB, err := InitSQLite(config)
	if err != nil {
		fmt.Printf("初始化失败：%v\n", err)
		return
	}
	defer sqlDB.Close()

	// 3. 后续操作（如执行建表语句、插入数据等）
	// 示例：执行之前设计的建表语句
	createTableSQL := `
	-- 这里粘贴你的 backup_tasks 和 backup_records 建表语句
	CREATE TABLE IF NOT EXISTS backup_tasks (
	    ID INTEGER PRIMARY KEY AUTOINCREMENT,
	    name TEXT NOT NULL UNIQUE,
	    retain_count INTEGER DEFAULT 3,
	    retain_days INTEGER DEFAULT 7,
	    backup_dir TEXT NOT NULL,
	    storage_dir TEXT NOT NULL,
	    compress_strategy TEXT,
	    include_rules TEXT,
	    exclude_rules TEXT,
	    max_file_size INTEGER,
	    min_file_size INTEGER,
	    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
	    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS backup_records (
	    ID INTEGER PRIMARY KEY AUTOINCREMENT,
	    task_id INTEGER NOT NULL,
	    task_name TEXT NOT NULL,
	    version_id TEXT NOT NULL UNIQUE,
	    backup_filename TEXT NOT NULL,
	    backup_size INTEGER NOT NULL,
	    status TEXT NOT NULL,
	    failure_message TEXT,
	    checksum TEXT,
	    storage_path TEXT NOT NULL,
	    created_at TEXT DEFAULT CURRENT_TIMESTAMP
	);
	-- 索引创建语句
	CREATE INDEX IF NOT EXISTS idx_backup_tasks_name ON backup_tasks (name);
	CREATE INDEX IF NOT EXISTS idx_backup_records_created_at ON backup_records(created_at);
	CREATE INDEX IF NOT EXISTS idx_backup_records_task_id ON backup_records (task_id);
	CREATE INDEX IF NOT EXISTS idx_backup_records_task_name ON backup_records (task_name);
	`

	_, err = sqlDB.Exec(createTableSQL)
	if err != nil {
		fmt.Printf("执行建表语句失败：%v\n", err)
		return
	}
	fmt.Println("建表完成！")
}
```

