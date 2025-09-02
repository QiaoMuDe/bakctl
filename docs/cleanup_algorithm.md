# 备份文件清理算法设计文档

## 概述

本文档描述了 bakctl 项目中备份文件清理算法的设计思路、实现方案和使用方法。

## 设计目标

1. **自动清理**: 根据配置的保留策略自动清理过期的备份文件
2. **灵活策略**: 支持按数量和按时间两种保留策略
3. **安全可靠**: 确保不会误删重要文件，提供详细的操作日志
4. **高效执行**: 优化文件扫描和删除操作的性能
5. **易于集成**: 提供简洁的API接口，方便集成到现有流程中

## 保留策略

### 1. 按数量保留 (retain_count)
- 保留最新的 N 个备份文件
- 按时间戳降序排序，保留前 N 个文件
- 设置为 0 表示不限制数量

### 2. 按时间保留 (retain_days)
- 保留最近 N 天内的备份文件
- 基于文件名中的时间戳判断文件创建时间
- 设置为 0 表示不限制时间

### 3. 策略组合
- 两个策略可以同时生效，取并集（保留满足任一条件的文件）
- 两个策略都为 0 时，不进行任何清理
- 优先保证数据安全，宁可多保留也不误删

## 文件识别规则

### 备份文件格式
```
{任务名}_{时间戳}.zip
```

示例：
- `database_backup_1672531200.zip`
- `web_files_1672617600.zip`

### 匹配规则
- 使用正则表达式精确匹配文件名格式
- 只处理指定任务名的备份文件
- 忽略其他格式的文件和目录

## 算法流程

```
1. 参数验证
   ├── 检查存储目录是否有效
   ├── 检查任务名称是否有效
   └── 检查保留策略参数是否合法

2. 文件收集
   ├── 扫描存储目录
   ├── 使用正则表达式匹配备份文件
   ├── 解析文件名中的时间戳
   └── 收集文件信息（路径、大小、创建时间等）

3. 策略应用
   ├── 按时间戳降序排序文件列表
   ├── 应用数量保留策略
   ├── 应用时间保留策略
   └── 计算需要删除的文件列表

4. 执行删除
   ├── 逐个删除标记的文件
   ├── 记录删除成功和失败的文件
   └── 统计清理结果

5. 结果报告
   ├── 生成清理统计信息
   ├── 记录操作日志
   └── 返回详细结果
```

## 核心函数说明

### CleanupBackupFiles
主要的清理函数，执行完整的清理流程。

```go
func CleanupBackupFiles(storageDir, taskName string, retainCount, retainDays int) (CleanupResult, error)
```

### collectBackupFiles
收集指定目录下的备份文件信息。

```go
func collectBackupFiles(storageDir, taskName string) ([]BackupFileInfo, error)
```

### determineFilesToDelete
根据保留策略确定需要删除的文件。

```go
func determineFilesToDelete(backupFiles []BackupFileInfo, retainCount, retainDays int) []BackupFileInfo
```

## 集成方式

### 在备份执行后自动清理

在 `cmd/subcmd/run/run.go` 的 `executeTask` 函数中集成：

```go
// 9. 清理历史备份
cl.White("  → 清理历史备份...")
if err := utils.CleanupBackupFilesWithLogging(task, cl); err != nil {
    // 清理失败不影响备份成功，只记录警告
    cl.Yellowf("  → 清理警告: %v", err)
}
```

### 独立的清理命令

可以创建独立的清理命令用于手动或定期清理：

```bash
bakctl cleanup --task-id 1
bakctl cleanup --all
bakctl cleanup --preview  # 预览模式，不实际删除
```

## 安全措施

1. **路径验证**: 确保只在指定的备份目录内操作
2. **文件格式验证**: 严格匹配备份文件格式，避免误删其他文件
3. **错误处理**: 删除失败时记录错误但不中断整个流程
4. **预览模式**: 提供预览功能，用户可以在实际删除前查看将要删除的文件
5. **详细日志**: 记录所有操作的详细信息，便于问题排查

## 性能优化

1. **单次目录扫描**: 一次性读取目录内容，避免重复I/O操作
2. **内存排序**: 在内存中对文件列表进行排序，避免多次文件系统访问
3. **批量操作**: 支持批量清理多个任务的备份文件
4. **错误恢复**: 单个文件删除失败不影响其他文件的处理

## 测试场景

### 基本场景
1. 只按数量保留（retain_count > 0, retain_days = 0）
2. 只按时间保留（retain_count = 0, retain_days > 0）
3. 数量和时间都限制（retain_count > 0, retain_days > 0）
4. 不清理（retain_count = 0, retain_days = 0）

### 边界场景
1. 空目录（无备份文件）
2. 目录不存在
3. 所有文件都在保留范围内
4. 所有文件都需要删除
5. 文件删除权限不足

### 异常场景
1. 文件名格式不匹配
2. 时间戳解析失败
3. 文件信息获取失败
4. 磁盘空间不足

## 使用示例

```go
// 基本使用
result, err := utils.CleanupBackupFiles("/backup/dir", "my_task", 5, 30)
if err != nil {
    log.Printf("清理失败: %v", err)
    return
}
fmt.Printf("清理完成: %s", utils.FormatCleanupResult(result))

// 带日志的清理
err := utils.CleanupBackupFilesWithLogging(task, colorLib)
if err != nil {
    log.Printf("清理失败: %v", err)
}

// 预览模式
filesToDelete, err := utils.GetCleanupPreview(task)
if err != nil {
    log.Printf("预览失败: %v", err)
    return
}
fmt.Printf("将删除 %d 个文件", len(filesToDelete))
```

## 配置建议

### 生产环境
- retain_count: 10-30（根据备份频率调整）
- retain_days: 30-90（根据业务需求调整）

### 开发环境
- retain_count: 3-5
- retain_days: 7-14

### 测试环境
- retain_count: 2-3
- retain_days: 3-7

## 监控和告警

建议监控以下指标：
1. 清理执行频率和耗时
2. 删除文件数量和释放空间大小
3. 清理失败的文件数量
4. 存储目录的总文件数和占用空间

当出现以下情况时应该告警：
1. 清理失败率过高（>10%）
2. 存储空间增长过快
3. 备份文件数量异常增长
4. 清理操作耗时过长