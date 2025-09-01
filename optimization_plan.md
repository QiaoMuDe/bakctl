# executeTask 函数优化方案

## 设计要求分析

### 核心设计原则
**无论备份成功还是失败，都必须插入记录到数据库**
- 成功时：记录成功状态、文件信息、哈希值等
- 失败时：记录失败状态、错误信息、部分可获取的信息

### 当前问题分析

#### 1. 代码问题
- **函数过长**：103行代码，职责过多
- **变量名冗长**：`compressedFileName`、`compressedFilePath`、`compressionLevel`
- **错误变量命名不统一**：`hashErr`、`statErr` 应该复用 `err`
- **缺少统一的记录机制**：只在成功时记录，失败时没有记录
- **文件扩展名不一致**：代码中用 `.zip` 但注释说 `.tar.gz`
- **缺少常量定义**：硬编码的 `"sha1"` 和文件扩展名

#### 2. 优化目标
- 实现统一的记录机制（成功/失败都记录）
- 提取辅助函数，分离职责
- 简化变量名，提高可读性
- 使用 defer 确保记录一定执行
- 添加常量定义
- 减少主函数行数

## 优化方案

### 1. 核心设计：统一记录机制

使用 `defer` 和结果结构体确保无论成功失败都记录：

```go
// BackupResult 备份执行结果
type BackupResult struct {
    Success    bool
    ErrorMsg   string
    BackupPath string
    FileSize   int64
    Checksum   string
}

// recordBackupResult 统一记录备份结果（成功或失败）
func recordBackupResult(db *sqlx.DB, task baktypes.BackupTask, result *BackupResult) error {
    rec := baktypes.BackupRecord{
        TaskID:         task.ID,
        TaskName:       task.Name,
        VersionID:      id.GenMaskedID(),
        BackupFilename: filepath.Base(result.BackupPath),
        BackupSize:     result.FileSize,
        StoragePath:    result.BackupPath,
        Status:         result.Success,
        FailureMessage: result.ErrorMsg,
        Checksum:       result.Checksum,
    }
    
    return DB.InsertBackupRecord(db, &rec)
}
```

### 2. 提取辅助函数

```go
// validateSourceDir 验证源目录是否存在
func validateSourceDir(dir string) error {
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        return fmt.Errorf("源目录不存在: %s", dir)
    }
    return nil
}

// parseFilterRules 解析包含和排除规则
func parseFilterRules(includeRules, excludeRules string) ([]string, []string, error) {
    include, err := ut.UnmarshalRules(includeRules)
    if err != nil {
        return nil, nil, fmt.Errorf("解析包含规则失败: %w", err)
    }
    
    exclude, err := ut.UnmarshalRules(excludeRules)
    if err != nil {
        return nil, nil, fmt.Errorf("解析排除规则失败: %w", err)
    }
    
    return include, exclude, nil
}

// generateBackupPath 生成备份文件路径
func generateBackupPath(task baktypes.BackupTask) string {
    filename := fmt.Sprintf("%s_%d%s", task.Name, time.Now().Unix(), BackupFileExt)
    return filepath.Join(task.StorageDir, filename)
}

// collectBackupInfo 收集备份文件信息（大小和哈希）
func collectBackupInfo(filePath string) (int64, string, error) {
    // 获取文件大小
    info, err := os.Stat(filePath)
    if err != nil {
        return 0, "", fmt.Errorf("获取文件信息失败: %w", err)
    }
    
    // 计算哈希值
    checksum, err := hash.ChecksumProgress(filePath, HashAlgorithm)
    if err != nil {
        return info.Size(), "", fmt.Errorf("计算哈希失败: %w", err)
    }
    
    return info.Size(), checksum, nil
}
```

### 2. 添加常量定义

```go
const (
    HashAlgorithm = "sha1"
    BackupFileExt = ".zip"  // 需要与实际使用的扩展名保持一致
)
```

### 3. 变量名优化对照表

| 原变量名 | 优化后 | 说明 |
|---------|--------|------|
| `compressedFileName` | `filename` | 更简洁 |
| `compressedFilePath` | `backupPath` | 语义更明确 |
| `compressionLevel` | `level` | 上下文已明确 |
| `hashVal` | `checksum` | 语义更准确 |
| `fileInfo` | `info` | 上下文已明确 |
| `hashErr`, `statErr` | `err` | 复用变量 |

### 4. 重构后的主函数（关键改进：使用defer确保记录）

```go
func executeTask(task baktypes.BackupTask, db *sqlx.DB) error {
    // 初始化结果结构体
    result := &BackupResult{
        Success:    false,
        BackupPath: generateBackupPath(task),
    }
    
    // 使用defer确保无论成功失败都记录到数据库
    defer func() {
        if recordErr := recordBackupResult(db, task, result); recordErr != nil {
            // 记录失败的处理（可以记录日志等）
            fmt.Printf("记录备份结果失败: %v
", recordErr)
        }
    }()
    
    // 1. 验证源目录
    if err := validateSourceDir(task.BackupDir); err != nil {
        result.ErrorMsg = err.Error()
        return err
    }
    
    // 2. 解析过滤规则
    include, exclude, err := parseFilterRules(task.IncludeRules, task.ExcludeRules)
    if err != nil {
        result.ErrorMsg = err.Error()
        return err
    }
    
    // 3. 构建过滤器
    filters := types.FilterOptions{
        Include: include,
        Exclude: exclude,
        MinSize: task.MinFileSize,
        MaxSize: task.MaxFileSize,
    }
    
    // 4. 设置压缩等级
    level := types.CompressionLevelNone
    if task.Compress {
        level = types.CompressionLevelDefault
    }
    
    // 5. 构建压缩配置
    opts := comprx.Options{
        CompressionLevel:      level,
        OverwriteExisting:     true,
        ProgressEnabled:       true,
        ProgressStyle:         types.ProgressStyleASCII,
        DisablePathValidation: false,
        Filter:                filters,
    }
    
    // 6. 执行备份操作
    if err := comprx.PackOptions(result.BackupPath, task.BackupDir, opts); err != nil {
        result.ErrorMsg = fmt.Sprintf("备份操作失败: %v", err)
        return err
    }
    
    // 7. 收集备份文件信息
    size, checksum, err := collectBackupInfo(result.BackupPath)
    if err != nil {
        result.ErrorMsg = err.Error()
        result.FileSize = size  // 即使哈希失败也记录文件大小
        return err
    }
    
    // 8. 设置成功结果
    result.Success = true
    result.FileSize = size
    result.Checksum = checksum
    
    return nil
}
```

## 优化效果

### 核心改进
1. **统一记录机制**：使用 `defer` 确保无论成功失败都记录
2. **结果追踪**：`BackupResult` 结构体记录执行状态和错误信息

### 代码行数对比
- **优化前**：103行
- **优化后**：约60行（主函数）+ 5个辅助函数

### 改进点
1. **可靠性保证**：defer 机制确保记录一定执行
2. **职责分离**：每个函数职责单一，易于测试和维护
3. **可读性提升**：变量名更简洁，逻辑更清晰
4. **错误追踪**：详细记录失败原因和执行进度
5. **复用性增强**：辅助函数可在其他地方复用
6. **常量化配置**：避免硬编码，便于维护

### 设计优势
1. **数据完整性**：无论何种情况都有记录
2. **调试友好**：记录详细的错误信息
3. **监控支持**：便于统计成功率和失败原因
4. **维护性好**：清晰的函数分工

### 注意事项
1. 需要确认文件扩展名是使用 `.zip` 还是 `.tar.gz`
2. `BackupResult` 结构体可能需要根据实际的 `BackupRecord` 字段调整

4. 可以考虑为辅助函数添加单元测试