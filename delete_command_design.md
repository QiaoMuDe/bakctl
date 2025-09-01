# Delete 子命令实现方案

## 命令功能概述

delete 子命令用于删除备份任务，支持单个删除和批量删除，删除过程包括：
1. 删除备份文件（存储目录中的实际文件）
2. 删除备份记录（数据库中的记录）
3. 删除备份任务（数据库中的任务配置）

## 命令参数设计

### 支持的参数
- `-id <任务ID>`：删除指定ID的单个备份任务
- `-ids <任务ID列表>`：批量删除多个备份任务（逗号分隔）
- `-force`：强制删除，跳过确认提示（可选）
- `-keep-files`：只删除数据库记录，保留备份文件（可选）

### 参数互斥性
- `-id` 和 `-ids` 互斥，只能使用其中一个
- 必须指定 `-id` 或 `-ids` 中的一个

## 删除流程设计

### 主要步骤
1. **参数验证**：检查参数有效性和互斥性
2. **任务查询**：根据ID获取要删除的任务信息
3. **用户确认**：显示删除信息，等待用户确认（除非使用-force）
4. **执行删除**：按顺序删除文件、记录、任务
5. **结果报告**：显示删除结果统计

### 详细删除逻辑
对于每个任务：
1. 获取备份任务信息
2. 获取该任务的所有备份记录
3. 删除备份文件（如果存在且未指定-keep-files）
4. 删除备份记录（数据库中）
5. 删除备份任务（数据库中）

## 代码结构设计

### 文件结构
```
cmd/subcmd/delete/
├── delete.go          # 主要逻辑
├── flags.go           # 命令行参数定义
└── helpers.go         # 辅助函数
```

### 核心函数设计

#### 1. 主函数
```go
// DeleteCmdMain delete命令的主函数
func DeleteCmdMain(db *sqlx.DB) error
```

#### 2. 参数验证
```go
// validateFlags 验证命令行参数
func validateFlags() error
```

#### 3. 任务选择
```go
// selectTasksToDelete 选择要删除的任务
func selectTasksToDelete(db *sqlx.DB) ([]baktypes.BackupTask, error)
```

#### 4. 用户确认
```go
// confirmDeletion 显示删除信息并确认
func confirmDeletion(tasks []baktypes.BackupTask) (bool, error)
```

#### 5. 删除执行
```go
// deleteTasks 批量删除任务
func deleteTasks(db *sqlx.DB, tasks []baktypes.BackupTask) error

// deleteTask 删除单个任务
func deleteTask(db *sqlx.DB, task baktypes.BackupTask) error
```

#### 6. 辅助函数
```go
// deleteBackupFiles 删除备份文件
func deleteBackupFiles(records []baktypes.BackupRecord) error

// deleteBackupRecords 删除备份记录
func deleteBackupRecords(db *sqlx.DB, taskID int) error

// deleteBackupTask 删除备份任务
func deleteBackupTask(db *sqlx.DB, taskID int) error

// getBackupRecords 获取任务的备份记录
func getBackupRecords(db *sqlx.DB, taskID int) ([]baktypes.BackupRecord, error)
```

## 删除结果结构

```go
// DeleteResult 删除结果
type DeleteResult struct {
    TaskID          int
    TaskName        string
    FilesDeleted    int     // 删除的文件数量
    FilesSkipped    int     // 跳过的文件数量（文件不存在等）
    RecordsDeleted  int     // 删除的记录数量
    Success         bool    // 是否成功
    ErrorMsg        string  // 错误信息
}

// DeleteSummary 删除汇总
type DeleteSummary struct {
    TotalTasks      int
    SuccessTasks    int
    FailedTasks     int
    TotalFiles      int
    TotalRecords    int
    Results         []DeleteResult
}
```

## 错误处理策略

### 错误类型
1. **参数错误**：立即返回，不执行删除
2. **任务不存在**：跳过该任务，继续处理其他任务
3. **文件删除失败**：记录错误，继续删除记录和任务
4. **数据库操作失败**：记录错误，但不影响其他任务

### 事务处理
- 每个任务的删除操作使用数据库事务
- 文件删除失败不回滚数据库操作
- 数据库操作失败则回滚该任务的所有数据库更改

## 用户交互设计

### 确认提示示例
```
即将删除以下备份任务：

任务ID: 1, 名称: "项目备份", 备份记录: 5个, 预计删除文件: 5个
任务ID: 2, 名称: "文档备份", 备份记录: 3个, 预计删除文件: 3个

总计: 2个任务, 8个备份记录, 8个备份文件

警告: 此操作不可逆！备份文件将被永久删除。

确认删除? (y/N): 
```

### 进度显示
```
[1/2] 正在删除任务: 项目备份 (ID: 1)
  ✓ 删除备份文件: 5个
  ✓ 删除备份记录: 5个  
  ✓ 删除任务配置
  
[2/2] 正在删除任务: 文档备份 (ID: 2)
  ✓ 删除备份文件: 3个
  ✓ 删除备份记录: 3个
  ✓ 删除任务配置

删除完成！成功: 2个任务, 失败: 0个任务
```

## 安全考虑

### 防误删机制
1. **默认需要确认**：除非使用-force参数
2. **详细信息显示**：显示将要删除的具体内容
3. **分步骤删除**：先删文件，再删记录，最后删任务
4. **错误隔离**：单个任务失败不影响其他任务

### 权限检查
1. **文件权限**：检查备份文件的删除权限
2. **目录权限**：检查备份目录的访问权限
3. **数据库权限**：确保有删除记录的权限

## 实现优先级

### 第一阶段（核心功能）
1. 基本的单个任务删除
2. 文件和数据库记录删除
3. 基本的错误处理

### 第二阶段（增强功能）
1. 批量删除支持
2. 用户确认机制
3. 详细的进度显示

### 第三阶段（高级功能）
1. -keep-files 参数支持
2. -force 参数支持
3. 完善的错误恢复机制

## 测试用例设计

### 正常场景
1. 删除存在的单个任务
2. 批量删除多个任务
3. 删除没有备份记录的任务
4. 删除备份文件不存在的任务

### 异常场景
1. 删除不存在的任务ID
2. 文件删除权限不足
3. 数据库连接失败
4. 部分文件删除失败

### 边界场景
1. 删除大量任务
2. 删除大文件备份
3. 并发删除操作
4. 磁盘空间不足

## 配置选项

### 可配置参数
```go
type DeleteConfig struct {
    MaxConcurrentDeletes int    // 最大并发删除数
    ConfirmTimeout      int     // 确认超时时间（秒）
    RetryAttempts       int     // 删除重试次数
    ChunkSize           int     // 批量处理大小
}
```

这个设计方案提供了完整的delete子命令实现思路，包括参数设计、流程控制、错误处理和用户体验等各个方面。