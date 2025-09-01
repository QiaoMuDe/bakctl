# Restore 子命令实现方案

## 命令功能概述

restore 子命令用于恢复备份文件，支持根据任务ID和版本ID恢复指定的备份到指定目录。恢复过程包括：
1. 验证参数有效性
2. 查询备份记录信息
3. 检查备份文件完整性
4. 解压缩备份文件（如果需要）
5. 恢复文件到目标目录
6. 验证恢复结果

## 命令参数设计

### 支持的参数
- `-id <任务ID>`：指定要恢复的备份任务ID（必需）
- `-vid <版本ID>`：指定要恢复的备份版本ID（必需）
- `-d <目录路径>`：指定恢复到的目标目录（可选，默认为当前目录）

### 参数验证规则
- `-id` 和 `-vid` 为必需参数
- `-d` 目录必须存在且可写
- 如果目标目录有同名文件，将提示用户确认是否覆盖

## 恢复流程设计

### 主要步骤
1. **参数验证**：检查参数有效性和目录权限
2. **备份查询**：根据任务ID和版本ID查询备份记录
3. **文件检查**：验证备份文件是否存在且完整
4. **目录准备**：检查目标目录，处理文件冲突
5. **执行恢复**：解压缩并恢复文件到目标目录
6. **结果报告**：显示恢复结果统计

### 详细恢复逻辑
1. 验证任务ID和版本ID的有效性
2. 查询备份记录，获取备份文件路径和元信息
3. 查询备份任务，获取原始备份配置（过滤器、压缩设置等）
4. 检查备份文件是否存在，验证文件大小
5. 检查目标目录空间是否足够
6. 处理文件名冲突（询问用户是否覆盖）
7. 使用备份任务配置和 comprx.UnpackOptions 执行文件恢复操作

## 代码结构设计

### 文件结构
```
cmd/subcmd/restore/
├── restore.go         # 主要逻辑
├── flags.go           # 命令行参数定义
└── helpers.go         # 辅助函数
```

### 核心函数设计

#### 1. 主函数
```go
// RestoreCmdMain restore命令的主函数
func RestoreCmdMain(db *sqlx.DB) error
```

#### 2. 参数验证
```go
// validateFlags 验证命令行参数
func validateFlags() error
```

#### 3. 备份查询
```go
// getBackupRecord 根据任务ID和版本ID获取备份记录
func getBackupRecord(db *sqlx.DB, taskID int, versionID string) (*types.BackupRecord, error)

// getBackupTask 获取备份任务信息（用于获取原始配置）
func getBackupTask(db *sqlx.DB, taskID int) (*types.BackupTask, error)
```

#### 4. 文件检查
```go
// verifyBackupFile 验证备份文件完整性
func verifyBackupFile(record *types.BackupRecord) error
```

#### 5. 恢复执行
```go
// restoreBackup 执行备份恢复
func restoreBackup(record *types.BackupRecord, task *types.BackupTask, targetDir string) (*RestoreResult, error)

// extractBackupFile 使用comprx库解压缩备份文件
func extractBackupFile(backupPath, targetDir string, task *types.BackupTask) error {
    // 使用备份任务的配置构建解压选项
    opts := buildRestoreOptions(task)
    return comprx.UnpackOptions(backupPath, targetDir, opts)
}
```
```

#### 6. 辅助函数
```go
// checkTargetDirectory 检查目标目录
func checkTargetDirectory(targetDir string) error

// checkDiskSpace 检查磁盘空间
func checkDiskSpace(targetDir string, requiredSize int64) error

// handleFileConflicts 处理文件冲突
func handleFileConflicts(targetDir string) error
```

## 恢复结果结构

```go
// RestoreResult 恢复结果
type RestoreResult struct {
    TaskID           int       `json:"task_id"`
    TaskName         string    `json:"task_name"`
    VersionID        string    `json:"version_id"`
    TargetDirectory  string    `json:"target_directory"`
    FilesRestored    int       `json:"files_restored"`
    FilesSkipped     int       `json:"files_skipped"`
    FilesOverwritten int       `json:"files_overwritten"`
    TotalSize        int64     `json:"total_size"`
    Duration         string    `json:"duration"`
    Success          bool      `json:"success"`
    ErrorMsg         string    `json:"error_msg,omitempty"`
}

// RestoreOptions 恢复选项
type RestoreOptions struct {
    TaskID      int    `json:"task_id"`
    VersionID   string `json:"version_id"`
    TargetDir   string `json:"target_dir"`
}
```

## 错误处理策略

### 错误类型
1. **参数错误**：立即返回，不执行恢复
2. **备份不存在**：任务ID或版本ID不存在
3. **文件损坏**：备份文件不存在或校验失败
4. **权限错误**：目标目录无写权限
5. **空间不足**：目标目录磁盘空间不够
6. **解压失败**：备份文件解压缩失败

### 恢复策略
- 参数错误：显示帮助信息并退出
- 文件冲突：询问用户是否覆盖已存在的文件
- 部分失败：记录失败文件，继续恢复其他文件
- 严重错误：停止恢复并清理已恢复的文件

## 用户交互设计

### 恢复确认提示
```
即将恢复以下备份：

任务ID: 1
任务名称: "项目备份"
版本ID: "20240901_143022"
备份时间: 2024-09-01 14:30:22
备份大小: 256.8 MB
目标目录: /home/user/restore

预计恢复文件: 1,234个
预计占用空间: 512.5 MB

确认恢复? (y/N): 
```

### 进度显示
```
正在恢复备份: 项目备份 (版本: 20240901_143022)
  ✓ 检查备份文件存在性
  ✓ 检查目标目录权限
  ✓ 检查磁盘空间: 512.5 MB 可用
  → 正在解压缩备份文件...
  → 正在恢复文件... [████████████████████] 100% (1,234/1,234)
  ✓ 恢复完成: 1,234个文件

恢复成功！
- 恢复文件: 1,234个
- 覆盖文件: 0个
- 跳过文件: 0个
- 总大小: 512.5 MB
- 耗时: 2分35秒
```

## 安全考虑

### 安全机制
1. **路径验证**：防止路径遍历攻击
2. **权限检查**：验证目录读写权限
3. **空间检查**：防止磁盘空间耗尽
4. **文件检查**：检查备份文件存在性
5. **冲突处理**：安全处理文件名冲突

### 权限检查
1. **目标目录权限**：检查目录的写权限
2. **备份文件权限**：检查备份文件的读权限
3. **磁盘空间**：确保有足够的空间进行恢复

## 压缩格式支持

### 使用现有的 comprx 库
项目中已有 `comprx.UnpackOptions` 函数，支持多种压缩格式的自动检测和解压：

```go
// 执行解压缩恢复
func restoreBackupFile(record *types.BackupRecord, task *types.BackupTask, targetDir string) error {
    // 构建基于任务配置的恢复选项
    opts := buildRestoreOptions(task)
    
    // 执行解压缩，使用备份任务的配置
    return comprx.UnpackOptions(record.StoragePath, targetDir, opts)
}
```

### 支持的压缩格式
comprx 库自动支持常见的压缩格式，无需手动检测：
- **tar.gz**、**tar.bz2**、**tar.xz** 等 tar 格式
- **zip** 格式
- 其他 comprx 库支持的格式

## 实现优先级

### 第一阶段（核心功能）
1. 基本的单个备份恢复
2. 使用 comprx.UnpackOptions 进行解压缩
3. 基本的错误处理和进度显示

### 第二阶段（增强功能）
1. 更好的冲突处理机制
2. 恢复选项优化（覆盖策略等）
3. 恢复进度和用户体验优化

### 第三阶段（高级功能）
1. 增量恢复支持
2. 并发恢复优化
3. 恢复进度断点续传

## 测试用例设计

### 正常场景
1. 恢复存在的备份到空目录
2. 恢复压缩备份文件
3. 恢复到指定目录
4. 处理文件冲突的用户交互

### 异常场景
1. 任务ID或版本ID不存在
2. 备份文件损坏或丢失
3. 目标目录无写权限
4. 磁盘空间不足
5. 解压缩失败

### 边界场景
1. 恢复大文件备份
2. 恢复包含特殊字符的文件名
3. 恢复到根目录
4. 同时恢复多个备份

## 配置选项

### 基于备份任务配置的恢复选项
```go
// buildRestoreOptions 根据备份任务配置构建恢复选项
func buildRestoreOptions(task *types.BackupTask) comprx.Options {
    // 构建过滤器（使用备份任务的配置）
    filters := types.FilterOptions{
        Include: task.IncludePatterns, // 包含规则
        Exclude: task.ExcludePatterns, // 排除规则
        MinSize: task.MinFileSize,     // 最小文件大小
        MaxSize: task.MaxFileSize,     // 最大文件大小
    }

    return comprx.Options{
        OverwriteExisting:     true,                     // 覆盖已存在文件
        ProgressEnabled:       true,                     // 显示进度条
        ProgressStyle:         types.ProgressStyleASCII, // ASCII进度条
        DisablePathValidation: false,                    // 启用路径验证
        Filter:                filters,                  // 使用备份任务的过滤器
    }
}

// RestoreContext 恢复上下文
type RestoreContext struct {
    Record    *types.BackupRecord // 备份记录
    Task      *types.BackupTask   // 备份任务（包含原始配置）
    TargetDir string              // 目标目录
    Options   comprx.Options      // 恢复选项
}
```

## 命令行使用示例

```bash
# 基本恢复（恢复到当前目录）
bakctl restore -id 1 -vid "20240901_143022"

# 恢复到指定目录
bakctl restore -id 1 -vid "20240901_143022" -d /home/user/restore
```

这个设计方案提供了完整的 restore 子命令实现思路，包括参数设计、流程控制、错误处理、安全考虑和用户体验等各个方面。