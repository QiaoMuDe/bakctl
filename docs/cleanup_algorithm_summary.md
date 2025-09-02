# 备份文件清理算法实现总结

## 实现概述

我已经为 bakctl 项目设计并实现了一个完整的备份文件清理算法，该算法可以根据配置的保留策略自动清理历史备份文件。

## 核心特性

### 1. 双重保留策略
- **数量保留** (`retain_count`): 保留最新的 N 个备份文件
- **时间保留** (`retain_days`): 保留最近 N 天内的备份文件
- **策略组合**: 两个策略同时生效时取并集，确保数据安全
- **灵活配置**: 任一策略设为 0 表示不限制该维度

### 2. 安全可靠的文件识别
- **精确匹配**: 使用正则表达式严格匹配 `{任务名}_{时间戳}.zip` 格式
- **任务隔离**: 只处理指定任务的备份文件，避免误删其他文件
- **类型过滤**: 自动忽略目录和非备份文件

### 3. 完善的错误处理
- **参数验证**: 清理前验证所有输入参数的合法性
- **容错机制**: 单个文件删除失败不影响其他文件的处理
- **详细日志**: 记录所有操作的详细信息和错误原因

## 文件结构

```
internal/cleanup/
├── cleanup.go              # 核心清理算法实现
├── integration.go          # 集成辅助函数和接口定义
├── adapter.go              # 任务适配器，解决循环依赖
├── example_integration.go  # 集成示例代码
└── usage_example.go        # 完整的使用演示

docs/
├── cleanup_algorithm.md         # 详细设计文档
├── cleanup_integration_guide.md # 集成指南
└── cleanup_algorithm_summary.md # 实现总结
```

## 核心算法流程

```
1. 参数验证
   ├── 验证存储目录路径
   ├── 验证任务名称
   └── 验证保留策略参数

2. 文件收集
   ├── 扫描存储目录
   ├── 正则表达式匹配备份文件
   ├── 解析文件名中的时间戳
   └── 收集文件元信息

3. 策略应用
   ├── 按时间戳降序排序
   ├── 应用数量保留策略
   ├── 应用时间保留策略
   └── 计算需要删除的文件列表

4. 执行删除
   ├── 逐个删除标记的文件
   ├── 记录成功和失败的操作
   └── 统计清理结果

5. 结果报告
   ├── 生成详细的统计信息
   ├── 格式化输出清理结果
   └── 记录操作日志
```

## 关键函数说明

### 主要清理函数
```go
func CleanupBackupFiles(storageDir, taskName string, retainCount, retainDays int, backupFileExt string) (CleanupResult, error)
```
- 执行完整的清理流程
- 返回详细的清理结果统计

### 集成辅助函数
```go
func CleanupBackupFilesWithLogging(task BackupTask, backupFileExt string, cl *colorlib.ColorLib) error
```
- 带彩色日志输出的清理函数
- 适合在命令行程序中使用

### 预览功能
```go
func GetCleanupPreview(task BackupTask, backupFileExt string) ([]BackupFileInfo, error)
```
- 预览将要删除的文件
- 不实际执行删除操作

## 集成方式

### 在备份执行后自动清理

在 `cmd/subcmd/run/run.go` 的 `executeTask` 函数中添加：

```go
// 9. 清理历史备份
cl.White("  → 清理历史备份...")

taskAdapter := cleanup.NewBackupTaskAdapter(
    task.ID, task.Name, task.StorageDir,
    task.RetainCount, task.RetainDays,
)

if err := cleanup.CleanupBackupFilesWithLogging(taskAdapter, baktypes.BackupFileExt, cl); err != nil {
    cl.Yellowf("  → 清理警告: %v", err)
}
```

### 独立的清理命令

可以创建 `bakctl cleanup` 命令用于手动清理：

```bash
bakctl cleanup --task-id 1      # 清理指定任务
bakctl cleanup --all             # 清理所有任务
bakctl cleanup --preview         # 预览模式
```

## 配置示例

### 生产环境配置
```toml
[AddTaskConfig]
name = "database_backup"
retain_count = 10    # 保留最新10个备份
retain_days = 30     # 保留30天内的备份
# ... 其他配置
```

### 开发环境配置
```toml
[AddTaskConfig]
name = "dev_backup"
retain_count = 3     # 保留最新3个备份
retain_days = 7      # 保留7天内的备份
# ... 其他配置
```

## 安全保障措施

1. **路径验证**: 确保只在指定的备份目录内操作
2. **格式验证**: 严格匹配备份文件格式，避免误删其他文件
3. **权限检查**: 删除前检查文件访问权限
4. **错误恢复**: 单个文件操作失败不影响整体流程
5. **详细日志**: 记录所有操作的详细信息

## 性能特点

1. **高效扫描**: 单次目录遍历收集所有文件信息
2. **内存排序**: 在内存中对文件列表进行排序
3. **批量处理**: 支持同时处理多个任务的清理
4. **资源友好**: 合理的内存使用和文件句柄管理

## 测试验证

### 自动化测试
- 提供了完整的演示程序 `DemoCleanupAlgorithm()`
- 包含多种清理场景的测试用例
- 支持创建测试环境和自动清理

### 手动测试
- 可以通过修改保留策略参数测试不同场景
- 支持预览模式，安全验证清理逻辑
- 提供详细的操作日志便于调试

## 使用建议

### 保留策略建议
- **高频备份** (每小时): retain_count=24, retain_days=7
- **日常备份** (每天): retain_count=7, retain_days=30
- **周期备份** (每周): retain_count=4, retain_days=90

### 监控建议
- 监控清理操作的执行频率和耗时
- 跟踪删除的文件数量和释放的空间大小
- 设置清理失败率的告警阈值

### 维护建议
- 定期检查存储目录的总文件数和占用空间
- 根据业务需求调整保留策略参数
- 在生产环境部署前充分测试清理逻辑

## 扩展性

该清理算法设计具有良好的扩展性：

1. **策略扩展**: 可以轻松添加新的保留策略（如按文件大小）
2. **格式支持**: 支持不同的备份文件格式和命名规则
3. **并发处理**: 可以扩展为支持并发清理多个任务
4. **存储后端**: 可以扩展支持云存储等不同的存储后端

## 总结

这个备份文件清理算法实现了以下目标：

✅ **功能完整**: 支持双重保留策略和灵活配置  
✅ **安全可靠**: 严格的文件识别和完善的错误处理  
✅ **易于集成**: 提供简洁的API和详细的集成指南  
✅ **性能优良**: 高效的算法实现和合理的资源使用  
✅ **可维护性**: 清晰的代码结构和完善的文档  

该算法已经准备好集成到现有的 bakctl 项目中，可以有效解决历史备份文件管理的问题，提高存储空间的利用效率。