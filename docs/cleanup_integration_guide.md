# 备份文件清理算法集成指南

## 概述

本指南详细说明如何将备份文件清理算法集成到现有的 bakctl 项目中，实现自动清理历史备份文件的功能。

## 文件结构

```
internal/cleanup/
├── cleanup.go              # 核心清理算法实现
├── integration.go          # 集成辅助函数
├── adapter.go              # 任务适配器
├── example_integration.go  # 集成示例
└── usage_example.go        # 使用演示
```

## 核心功能

### 1. 清理策略

- **按数量保留** (`retain_count`): 保留最新的 N 个备份文件
- **按时间保留** (`retain_days`): 保留最近 N 天内的备份文件
- **策略组合**: 两个策略可以同时生效，保留满足任一条件的文件
- **无清理模式**: 两个策略都为 0 时不进行清理

### 2. 文件识别

- 严格按照文件名格式匹配: `{任务名}_{时间戳}.zip`
- 只处理指定任务的备份文件
- 忽略其他格式的文件和目录

### 3. 安全保障

- 参数验证确保输入合法性
- 路径验证防止误删系统文件
- 错误处理不影响主要备份流程
- 详细日志记录所有操作

## 集成步骤

### 步骤 1: 在 run 命令中集成

修改 `cmd/subcmd/run/run.go` 文件：

```go
// 在文件顶部添加导入
import (
    // ... 现有导入 ...
    "gitee.com/MM-Q/bakctl/internal/cleanup"
)

// 修改 executeTask 函数
func executeTask(task baktypes.BackupTask, db *sqlx.DB, cl *colorlib.ColorLib) error {
    // ... 现有的备份逻辑 ...

    // 8. 设置成功结果
    result.Success = true
    result.FileSize = size
    result.Checksum = checksum

    // 9. 清理历史备份
    cl.White("  → 清理历史备份...")
    
    // 创建任务适配器
    taskAdapter := cleanup.NewBackupTaskAdapter(
        task.ID,
        task.Name,
        task.StorageDir,
        task.RetainCount,
        task.RetainDays,
    )
    
    // 执行清理
    if err := cleanup.CleanupBackupFilesWithLogging(taskAdapter, baktypes.BackupFileExt, cl); err != nil {
        // 清理失败不影响备份成功，只记录警告
        cl.Yellowf("  → 清理警告: %v", err)
    }

    return nil
}
```

### 步骤 2: 创建独立的清理命令 (可选)

创建 `cmd/subcmd/cleanup/` 目录和相关文件：

```go
// cmd/subcmd/cleanup/cleanup.go
package cleanup

import (
    "fmt"
    
    DB "gitee.com/MM-Q/bakctl/internal/db"
    "gitee.com/MM-Q/bakctl/internal/cleanup"
    "gitee.com/MM-Q/bakctl/internal/types"
    "gitee.com/MM-Q/colorlib"
    "github.com/jmoiron/sqlx"
)

func CleanupCmdMain(db *sqlx.DB, cl *colorlib.ColorLib) error {
    // 获取所有任务
    tasks, err := DB.GetAllTasks(db)
    if err != nil {
        return fmt.Errorf("获取任务列表失败: %w", err)
    }

    if len(tasks) == 0 {
        cl.Yellow("没有找到任何备份任务")
        return nil
    }

    cl.Bluef("开始清理 %d 个任务的历史备份...\n", len(tasks))

    successCount := 0
    for i, task := range tasks {
        cl.Whitef("[%d/%d] 清理任务: %s (ID: %d)", i+1, len(tasks), task.Name, task.ID)
        
        // 创建任务适配器
        taskAdapter := cleanup.NewBackupTaskAdapter(
            task.ID,
            task.Name,
            task.StorageDir,
            task.RetainCount,
            task.RetainDays,
        )
        
        // 执行清理
        if err := cleanup.CleanupBackupFilesWithLogging(taskAdapter, types.BackupFileExt, cl); err != nil {
            cl.Redf("任务 %s 清理失败: %v", task.Name, err)
            continue
        }
        
        successCount++
    }

    cl.Greenf("清理完成！成功: %d, 失败: %d", successCount, len(tasks)-successCount)
    return nil
}
```

### 步骤 3: 添加清理命令到主程序 (可选)

修改 `cmd/bakctl/main.go`：

```go
import (
    // ... 现有导入 ...
    "gitee.com/MM-Q/bakctl/cmd/subcmd/cleanup"
)

func main() {
    // ... 现有代码 ...
    
    // 获取cleanup命令
    cleanupCmd := cleanup.InitCleanupCmd()
    
    // 注册子命令
    if err := qflag.AddSubCmd(addCmd, editCmd, listCmd, logCmd, runCmd, deleteCmd, exportCmd, restoreCmd, cleanupCmd); err != nil {
        // ... 错误处理 ...
    }
    
    // ... 现有代码 ...
    
    // 在路由命令部分添加
    case cleanupCmd.LongName(), cleanupCmd.ShortName(): // cleanup 命令
        if err := cleanup.CleanupCmdMain(db, CL); err != nil {
            CL.PrintError(err)
            os.Exit(1)
        }
        return
}
```

## 使用示例

### 基本使用

```go
// 创建任务适配器
task := cleanup.NewBackupTaskAdapter(
    1,                    // 任务ID
    "database_backup",    // 任务名称
    "/backup/database",   // 存储目录
    5,                    // 保留最新5个备份
    30,                   // 保留30天内的备份
)

// 执行清理
result, err := cleanup.CleanupBackupFiles(
    task.GetStorageDir(),
    task.GetName(),
    task.GetRetainCount(),
    task.GetRetainDays(),
    ".zip",
)

if err != nil {
    log.Printf("清理失败: %v", err)
    return
}

fmt.Printf("清理结果: %s", cleanup.FormatCleanupResult(result))
```

### 带日志的清理

```go
cl := colorlib.New()
if err := cleanup.CleanupBackupFilesWithLogging(task, ".zip", cl); err != nil {
    log.Printf("清理失败: %v", err)
}
```

### 预览模式

```go
filesToDelete, err := cleanup.GetCleanupPreview(task, ".zip")
if err != nil {
    log.Printf("预览失败: %v", err)
    return
}

fmt.Printf("将删除 %d 个文件:\n", len(filesToDelete))
for _, file := range filesToDelete {
    fmt.Printf("  - %s\n", file.FileName)
}
```

## 配置建议

### 生产环境推荐配置

```toml
[AddTaskConfig]
retain_count = 10    # 保留最新10个备份
retain_days = 30     # 保留30天内的备份
```

### 开发环境推荐配置

```toml
[AddTaskConfig]
retain_count = 3     # 保留最新3个备份
retain_days = 7      # 保留7天内的备份
```

### 测试环境推荐配置

```toml
[AddTaskConfig]
retain_count = 2     # 保留最新2个备份
retain_days = 3      # 保留3天内的备份
```

## 测试验证

### 运行演示程序

```go
// 在代码中调用演示函数
cleanup.DemoCleanupAlgorithm()
```

### 手动测试步骤

1. 创建测试任务并执行多次备份
2. 检查生成的备份文件
3. 修改保留策略参数
4. 执行清理操作
5. 验证清理结果是否符合预期

### 验证清理逻辑

```bash
# 创建测试备份文件
mkdir -p /tmp/test_backup
cd /tmp/test_backup

# 创建不同时间的备份文件
touch "test_task_$(date -d '10 days ago' +%s).zip"
touch "test_task_$(date -d '5 days ago' +%s).zip"
touch "test_task_$(date -d '2 days ago' +%s).zip"
touch "test_task_$(date +%s).zip"

# 运行清理测试
go run your_test_program.go
```

## 注意事项

1. **备份安全**: 清理操作不可逆，建议先在测试环境验证
2. **权限检查**: 确保程序有删除备份文件的权限
3. **磁盘空间**: 清理前检查磁盘空间，避免清理过程中空间不足
4. **并发安全**: 避免在备份执行过程中同时进行清理操作
5. **日志记录**: 重要的清理操作应该记录到日志文件中

## 故障排除

### 常见问题

1. **文件删除失败**: 检查文件权限和磁盘空间
2. **文件格式不匹配**: 确认备份文件名格式正确
3. **参数验证失败**: 检查保留策略参数是否合法
4. **目录不存在**: 确认存储目录路径正确

### 调试方法

1. 使用预览模式查看将要删除的文件
2. 检查日志输出中的详细错误信息
3. 验证文件名格式和时间戳解析
4. 测试不同的保留策略组合

## 性能优化

1. **批量操作**: 对于大量文件，考虑批量删除
2. **并发处理**: 可以并发处理多个任务的清理
3. **缓存优化**: 避免重复扫描同一目录
4. **内存管理**: 处理大量文件时注意内存使用

通过以上集成步骤，您可以成功将备份文件清理算法集成到 bakctl 项目中，实现自动化的历史备份文件管理。