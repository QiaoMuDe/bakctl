# Run 命令实现方案

## 概述

基于现有的备份库和代码架构，`run` 子命令主要负责任务选择、执行流程控制、备份库集成和结果记录。核心的文件操作、压缩、过滤等功能由现有备份库处理。

## 需要实现的核心功能

### 1. 命令行接口层 (`cmd/subcmd/run/`)

#### `flags.go` - 命令行参数定义
```go
package run

import "gitee.com/MM-Q/qflag"

var (
    // 任务选择参数
    taskIDFlag   *qflag.IntFlag    // --id, -i: 指定任务ID
    taskNameFlag *qflag.StringFlag // --name, -n: 指定任务名称
    allTasksFlag *qflag.BoolFlag   // --all, -a: 运行所有任务
    
    // 执行控制参数
    dryRunFlag   *qflag.BoolFlag   // --dry-run: 模拟运行
    verboseFlag  *qflag.BoolFlag   // --verbose, -v: 详细输出
    forceFlag    *qflag.BoolFlag   // --force, -f: 强制执行
    
    // 并发控制
    parallelFlag *qflag.IntFlag    // --parallel, -p: 并发数量
)

func InitRunCmd() *qflag.SubCmd {
    runCmd := qflag.NewSubCmd("run", "r", "运行备份任务")
    
    taskIDFlag = runCmd.Int("id", "i", 0, "指定要运行的任务ID")
    taskNameFlag = runCmd.String("name", "n", "", "指定要运行的任务名称")
    allTasksFlag = runCmd.Bool("all", "a", false, "运行所有任务")
    
    dryRunFlag = runCmd.Bool("dry-run", "", false, "模拟运行，不实际执行备份")
    verboseFlag = runCmd.Bool("verbose", "v", false, "显示详细执行信息")
    forceFlag = runCmd.Bool("force", "f", false, "强制执行，忽略警告")
    
    parallelFlag = runCmd.Int("parallel", "p", 1, "并发执行任务数量")
    
    return runCmd
}
```

#### `run.go` - 主要逻辑
```go
package run

import (
    "fmt"
    "sync"
    "time"
    
    DB "gitee.com/MM-Q/bakctl/internal/db"
    "gitee.com/MM-Q/bakctl/internal/types"
    "github.com/jmoiron/sqlx"
)

func RunCmdMain(db *sqlx.DB) error {
    // 1. 参数验证和任务选择
    tasks, err := selectTasks(db)
    if err != nil {
        return err
    }
    
    if len(tasks) == 0 {
        return fmt.Errorf("没有找到要执行的任务")
    }
    
    // 2. 显示执行计划
    if err := showExecutionPlan(tasks); err != nil {
        return err
    }
    
    // 3. 执行备份任务
    return executeTasks(db, tasks)
}

func selectTasks(db *sqlx.DB) ([]types.BackupTask, error) {
    // 参数互斥检查
    paramCount := 0
    if taskIDFlag.Get() > 0 { paramCount++ }
    if taskNameFlag.Get() != "" { paramCount++ }
    if allTasksFlag.Get() { paramCount++ }
    
    if paramCount == 0 {
        return nil, fmt.Errorf("请指定要运行的任务: --id, --name 或 --all")
    }
    if paramCount > 1 {
        return nil, fmt.Errorf("不能同时指定多个任务选择参数")
    }
    
    // 按优先级选择任务
    if taskIDFlag.Get() > 0 {
        return getTasksByID(db, taskIDFlag.Get())
    }
    
    if taskNameFlag.Get() != "" {
        return getTasksByName(db, taskNameFlag.Get())
    }
    
    if allTasksFlag.Get() {
        return DB.GetAllBackupTasks(db)
    }
    
    return nil, fmt.Errorf("未知错误")
}
```

### 2. 任务执行引擎

#### 核心执行逻辑
```go
func executeTasks(db *sqlx.DB, tasks []types.BackupTask) error {
    parallel := parallelFlag.Get()
    if parallel <= 0 {
        parallel = 1
    }
    
    // 单任务执行
    if len(tasks) == 1 || parallel == 1 {
        return executeTasksSequentially(db, tasks)
    }
    
    // 并发执行
    return executeTasksConcurrently(db, tasks, parallel)
}

func executeTasksSequentially(db *sqlx.DB, tasks []types.BackupTask) error {
    successCount := 0
    
    for i, task := range tasks {
        fmt.Printf("\n[%d/%d] 执行任务: %s (ID: %d)\n", i+1, len(tasks), task.Name, task.ID)
        
        if err := executeTask(db, &task); err != nil {
            fmt.Printf("❌ 任务执行失败: %v\n", err)
        } else {
            successCount++
            fmt.Printf("✅ 任务执行成功\n")
        }
    }
    
    fmt.Printf("\n执行完成: %d/%d 个任务成功\n", successCount, len(tasks))
    return nil
}

func executeTask(db *sqlx.DB, task *types.BackupTask) error {
    // 1. 创建备份执行器
    executor := NewBackupExecutor(db, task)
    
    // 2. 预检查
    if err := executor.PreCheck(); err != nil {
        return fmt.Errorf("预检查失败: %w", err)
    }
    
    // 3. 执行备份 (调用你的备份库)
    result, err := executor.Execute()
    if err != nil {
        // 记录失败
        _ = executor.RecordFailure(err)
        return err
    }
    
    // 4. 记录成功结果
    if err := executor.RecordSuccess(result); err != nil {
        return fmt.Errorf("记录备份结果失败: %w", err)
    }
    
    return nil
}
```

### 3. 备份执行器封装

#### 执行器结构
```go
type BackupExecutor struct {
    db       *sqlx.DB
    task     *types.BackupTask
    config   *ExecutionConfig
    versionID string
}

type ExecutionConfig struct {
    DryRun  bool
    Verbose bool
    Force   bool
}

type BackupResult struct {
    BackupFilename string
    BackupSize     int64
    StoragePath    string
    Checksum       string
    Duration       time.Duration
}

func NewBackupExecutor(db *sqlx.DB, task *types.BackupTask) *BackupExecutor {
    return &BackupExecutor{
        db:   db,
        task: task,
        config: &ExecutionConfig{
            DryRun:  dryRunFlag.Get(),
            Verbose: verboseFlag.Get(),
            Force:   forceFlag.Get(),
        },
        versionID: generateVersionID(),
    }
}
```

#### 核心执行方法
```go
func (e *BackupExecutor) Execute() (*BackupResult, error) {
    startTime := time.Now()
    
    if e.config.DryRun {
        return e.simulateBackup()
    }
    
    // 1. 准备存储目录
    if err := e.prepareStorageDir(); err != nil {
        return nil, err
    }
    
    // 2. 生成备份文件名
    backupFilename := e.generateBackupFilename()
    storagePath := filepath.Join(e.task.StorageDir, backupFilename)
    
    // 3. 调用你的备份库执行实际备份
    // 这里需要调用你现有的备份库
    backupSize, checksum, err := e.callYourBackupLibrary(storagePath)
    if err != nil {
        return nil, err
    }
    
    // 4. 清理旧备份
    if err := e.cleanupOldBackups(); err != nil {
        // 清理失败不影响主流程，只记录警告
        fmt.Printf("⚠️  清理旧备份失败: %v\n", err)
    }
    
    return &BackupResult{
        BackupFilename: backupFilename,
        BackupSize:     backupSize,
        StoragePath:    storagePath,
        Checksum:       checksum,
        Duration:       time.Since(startTime),
    }, nil
}

// 这个方法需要调用你的备份库
func (e *BackupExecutor) callYourBackupLibrary(outputPath string) (int64, string, error) {
    // TODO: 调用你现有的备份库
    // 参数应该包括:
    // - e.task.BackupDir (源目录)
    // - outputPath (输出文件路径)
    // - e.task.Compress (是否压缩)
    // - e.task.IncludeRules (包含规则)
    // - e.task.ExcludeRules (排除规则)
    // - e.task.MaxFileSize (最大文件大小)
    // - e.task.MinFileSize (最小文件大小)
    
    // 返回: 文件大小, 校验码, 错误
    return 0, "", fmt.Errorf("需要集成你的备份库")
}
```

### 4. 数据库记录管理

#### 记录备份结果
```go
func (e *BackupExecutor) RecordSuccess(result *BackupResult) error {
    record := &types.BackupRecord{
        TaskID:         e.task.ID,
        TaskName:       e.task.Name,
        VersionID:      e.versionID,
        BackupFilename: result.BackupFilename,
        BackupSize:     result.BackupSize,
        StoragePath:    result.StoragePath,
        Status:         true,
        Checksum:       result.Checksum,
    }
    
    return DB.InsertBackupRecord(e.db, record)
}

func (e *BackupExecutor) RecordFailure(err error) error {
    record := &types.BackupRecord{
        TaskID:         e.task.ID,
        TaskName:       e.task.Name,
        VersionID:      e.versionID,
        BackupFilename: "", // 失败时为空
        BackupSize:     0,
        StoragePath:    "",
        Status:         false,
        FailureMessage: err.Error(),
    }
    
    return DB.InsertBackupRecord(e.db, record)
}
```

### 5. 辅助功能

#### 版本ID生成
```go
func generateVersionID() string {
    return fmt.Sprintf("%d", time.Now().Unix())
}

func (e *BackupExecutor) generateBackupFilename() string {
    timestamp := time.Now().Format("20060102_150405")
    taskName := strings.ReplaceAll(e.task.Name, " ", "_")
    
    if e.task.Compress {
        return fmt.Sprintf("%s_%s.tar.gz", taskName, timestamp)
    }
    return fmt.Sprintf("%s_%s.tar", taskName, timestamp)
}
```

#### 清理旧备份
```go
func (e *BackupExecutor) cleanupOldBackups() error {
    // 根据 RetainCount 和 RetainDays 清理旧备份
    // 这个逻辑可能也需要调用你的备份库
    return nil
}
```

### 6. 集成到主程序

#### 修改 `cmd/bakctl/main.go`
```go
// 在导入中添加
"gitee.com/MM-Q/bakctl/cmd/subcmd/run"

// 在 main 函数中添加
runCmd := run.InitRunCmd()

// 在 AddSubCmd 中添加
if err := qflag.AddSubCmd(addCmd, editCmd, listCmd, logCmd, runCmd); err != nil {

// 在 switch 中添加
case runCmd.LongName(), runCmd.ShortName(): // run 命令
    if err := run.RunCmdMain(db); err != nil {
        fmt.Printf("err: %v\n", err)
        os.Exit(1)
    }
    return
```

## 实现优先级

### 第一阶段 (基础框架)
1. **命令行接口** - `flags.go` 和基础的 `run.go`
2. **任务选择逻辑** - 支持 ID/Name/All 方式
3. **主程序集成** - 修改 `main.go` 添加路由

### 第二阶段 (核心功能)
1. **执行器框架** - `BackupExecutor` 结构体和基础方法
2. **备份库集成** - 调用现有备份库的接口
3. **结果记录** - 成功和失败情况的数据库记录

### 第三阶段 (增强功能)
1. **并发执行** - 多任务并行处理
2. **进度显示** - 集成进度条库
3. **清理机制** - 旧备份自动清理

## 关键集成点

1. **备份库调用** - 需要了解你现有备份库的接口
2. **配置转换** - 将数据库中的任务配置转换为备份库需要的格式
3. **错误处理** - 统一的错误处理和记录机制
4. **进度反馈** - 如何从备份库获取进度信息

## 总结

基于现有架构，`run` 命令主要作为任务调度和结果记录的中间层，核心备份功能由现有库处理。实现重点在于：
- 清晰的命令行接口
- 可靠的任务选择和执行流程
- 完整的结果记录和错误处理
- 与现有备份库的无缝集成