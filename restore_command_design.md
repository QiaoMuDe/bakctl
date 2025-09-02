# Restore 子命令设计方案

## 🎯 概述

本文档详细描述了 CBK 备份管理工具的 restore 子命令设计方案。该命令用于恢复备份数据到指定位置，支持多种恢复模式和选项。

## 📋 功能需求分析

### 核心功能
1. **按任务恢复** - 恢复指定任务的最新备份
2. **按备份记录恢复** - 恢复指定的备份记录
3. **按时间恢复** - 恢复指定时间点的备份
4. **选择性恢复** - 恢复备份中的特定文件/目录
5. **预览模式** - 查看将要恢复的内容而不实际恢复

### 恢复选项
1. **目标位置** - 指定恢复到的目标目录
2. **覆盖策略** - 处理目标位置已存在文件的策略
3. **权限保持** - 是否保持原始文件权限
4. **时间戳保持** - 是否保持原始文件时间戳
5. **验证模式** - 恢复后验证文件完整性

## 🚩 标志定义方案

### 文件结构
```
cmd/subcmd/restore/
├── flags.go          // 标志定义
└── restore.go         // 主要实现逻辑
```

### flags.go 实现

```go
package restore

import (
	"flag"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	restoreCmd *cmd.Cmd // 恢复备份命令

	// 备份源选择 (四选一)
	taskIDf    *qflag.IntFlag    // 任务ID (恢复最新备份)
	recordIDf  *qflag.IntFlag    // 备份记录ID
	taskNameF  *qflag.StringFlag // 任务名称 (恢复最新备份)
	timeF      *qflag.StringFlag // 时间点 (格式: 2006-01-02 15:04:05)

	// 恢复目标
	targetF    *qflag.StringFlag // 目标目录 (必需)
	
	// 文件选择
	includeF   *qflag.SliceFlag  // 包含规则 (只恢复匹配的文件)
	excludeF   *qflag.SliceFlag  // 排除规则 (排除匹配的文件)
	pathF      *qflag.SliceFlag  // 指定路径 (只恢复指定的文件/目录)

	// 恢复选项
	overwriteF *qflag.StringFlag // 覆盖策略: skip|overwrite|prompt|backup
	preserveF  *qflag.BoolFlag   // 保持权限和时间戳
	verifyF    *qflag.BoolFlag   // 恢复后验证
	dryRunF    *qflag.BoolFlag   // 预览模式 (不实际恢复)
	
	// 输出选项
	verboseF   *qflag.BoolFlag   // 详细输出
	quietF     *qflag.BoolFlag   // 静默模式
)

func InitRestoreCmd() *cmd.Cmd {
	restoreCmd = cmd.NewCmd("restore", "r", flag.ExitOnError)
	restoreCmd.SetUseChinese(true)
	restoreCmd.SetDescription("恢复备份数据")

	// 备份源选择 (四选一)
	taskIDf = restoreCmd.Int("task-id", "t", 0, "指定任务ID，恢复该任务的最新备份")
	recordIDf = restoreCmd.Int("record-id", "r", 0, "指定备份记录ID进行恢复")
	taskNameF = restoreCmd.String("task-name", "n", "", "指定任务名称，恢复该任务的最新备份")
	timeF = restoreCmd.String("time", "T", "", "指定时间点恢复 (格式: 2006-01-02 15:04:05)")

	// 恢复目标 (必需)
	targetF = restoreCmd.String("target", "d", "", "恢复目标目录 (必需)")

	// 文件选择
	includeF = restoreCmd.Slice("include", "i", []string{}, "包含规则，只恢复匹配的文件")
	excludeF = restoreCmd.Slice("exclude", "e", []string{}, "排除规则，排除匹配的文件")
	pathF = restoreCmd.Slice("path", "p", []string{}, "指定要恢复的文件或目录路径")

	// 恢复选项
	overwriteF = restoreCmd.String("overwrite", "o", "prompt", "覆盖策略: skip(跳过)|overwrite(覆盖)|prompt(询问)|backup(备份)")
	preserveF = restoreCmd.Bool("preserve", "P", true, "保持文件权限和时间戳")
	verifyF = restoreCmd.Bool("verify", "V", false, "恢复后验证文件完整性")
	dryRunF = restoreCmd.Bool("dry-run", "D", false, "预览模式，显示将要恢复的内容但不实际恢复")

	// 输出选项
	verboseF = restoreCmd.Bool("verbose", "v", false, "详细输出恢复过程")
	quietF = restoreCmd.Bool("quiet", "q", false, "静默模式，只输出错误信息")

	return restoreCmd
}
```

## 🔧 核心功能设计

### 主要数据结构

```go
// RestoreConfig 恢复配置
type RestoreConfig struct {
	// 源信息
	TaskID     int64  `json:"task_id"`
	RecordID   int64  `json:"record_id"`
	TaskName   string `json:"task_name"`
	TimePoint  string `json:"time_point"`

	// 目标信息
	TargetDir  string `json:"target_dir"`

	// 文件过滤
	IncludeRules []string `json:"include_rules"`
	ExcludeRules []string `json:"exclude_rules"`
	SpecificPaths []string `json:"specific_paths"`

	// 恢复选项
	OverwriteMode string `json:"overwrite_mode"` // skip, overwrite, prompt, backup
	PreserveAttrs bool   `json:"preserve_attrs"`
	VerifyAfter   bool   `json:"verify_after"`
	DryRun        bool   `json:"dry_run"`

	// 输出选项
	Verbose bool `json:"verbose"`
	Quiet   bool `json:"quiet"`
}

// RestoreItem 恢复项目
type RestoreItem struct {
	SourcePath   string `json:"source_path"`   // 备份中的路径
	TargetPath   string `json:"target_path"`   // 恢复目标路径
	IsDirectory  bool   `json:"is_directory"`
	Size         int64  `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	Permissions  os.FileMode `json:"permissions"`
}

// RestoreResult 恢复结果
type RestoreResult struct {
	TotalFiles    int           `json:"total_files"`
	RestoredFiles int           `json:"restored_files"`
	SkippedFiles  int           `json:"skipped_files"`
	FailedFiles   int           `json:"failed_files"`
	TotalSize     int64         `json:"total_size"`
	Duration      time.Duration `json:"duration"`
	Errors        []string      `json:"errors"`
}
```

## 📝 详细实现方案

### 1. 主函数逻辑 (restore.go)

```go
package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/MM-Q/bakctl/internal/types"
	"github.com/jmoiron/sqlx"
)

func RestoreCmdMain(db *sqlx.DB) error {
	// 1. 参数验证
	config, err := validateAndBuildConfig()
	if err != nil {
		return err
	}

	// 2. 获取备份记录
	record, err := getBackupRecord(db, config)
	if err != nil {
		return err
	}

	// 3. 分析备份内容
	items, err := analyzeBackupContent(record, config)
	if err != nil {
		return err
	}

	// 4. 预览模式或实际恢复
	if config.DryRun {
		return previewRestore(items, config)
	}

	// 5. 执行恢复
	result, err := executeRestore(items, config)
	if err != nil {
		return err
	}

	// 6. 输出结果
	return printRestoreResult(result, config)
}

func validateAndBuildConfig() (*RestoreConfig, error) {
	config := &RestoreConfig{
		TaskID:        int64(taskIDf.Get()),
		RecordID:      int64(recordIDf.Get()),
		TaskName:      taskNameF.Get(),
		TimePoint:     timeF.Get(),
		TargetDir:     targetF.Get(),
		IncludeRules:  includeF.Get(),
		ExcludeRules:  excludeF.Get(),
		SpecificPaths: pathF.Get(),
		OverwriteMode: overwriteF.Get(),
		PreserveAttrs: preserveF.Get(),
		VerifyAfter:   verifyF.Get(),
		DryRun:        dryRunF.Get(),
		Verbose:       verboseF.Get(),
		Quiet:         quietF.Get(),
	}

	// 验证备份源选择 (四选一)
	sourceCount := 0
	if config.TaskID > 0 { sourceCount++ }
	if config.RecordID > 0 { sourceCount++ }
	if config.TaskName != "" { sourceCount++ }
	if config.TimePoint != "" { sourceCount++ }

	if sourceCount == 0 {
		return nil, fmt.Errorf("请指定备份源: --task-id, --record-id, --task-name 或 --time")
	}
	if sourceCount > 1 {
		return nil, fmt.Errorf("--task-id, --record-id, --task-name 和 --time 只能选择一个")
	}

	// 验证目标目录
	if config.TargetDir == "" {
		return nil, fmt.Errorf("请指定恢复目标目录: --target")
	}

	// 验证覆盖策略
	validModes := []string{"skip", "overwrite", "prompt", "backup"}
	valid := false
	for _, mode := range validModes {
		if config.OverwriteMode == mode {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("无效的覆盖策略: %s，支持的策略: %s", 
			config.OverwriteMode, strings.Join(validModes, ", "))
	}

	// 验证时间格式
	if config.TimePoint != "" {
		_, err := time.Parse("2006-01-02 15:04:05", config.TimePoint)
		if err != nil {
			return nil, fmt.Errorf("无效的时间格式: %s，请使用格式: 2006-01-02 15:04:05", config.TimePoint)
		}
	}

	// 验证输出选项冲突
	if config.Verbose && config.Quiet {
		return nil, fmt.Errorf("--verbose 和 --quiet 不能同时使用")
	}

	return config, nil
}
```

### 2. 备份记录获取逻辑

```go
func getBackupRecord(db *sqlx.DB, config *RestoreConfig) (*types.BackupRecord, error) {
	var record *types.BackupRecord
	var err error

	switch {
	case config.RecordID > 0:
		// 直接通过记录ID获取
		record, err = getRecordByID(db, config.RecordID)
	
	case config.TaskID > 0:
		// 通过任务ID获取最新备份
		record, err = getLatestRecordByTaskID(db, config.TaskID)
	
	case config.TaskName != "":
		// 通过任务名称获取最新备份
		record, err = getLatestRecordByTaskName(db, config.TaskName)
	
	case config.TimePoint != "":
		// 通过时间点获取最接近的备份
		record, err = getRecordByTimePoint(db, config.TimePoint)
	}

	if err != nil {
		return nil, err
	}

	if record == nil {
		return nil, fmt.Errorf("未找到匹配的备份记录")
	}

	return record, nil
}

func getRecordByID(db *sqlx.DB, recordID int64) (*types.BackupRecord, error) {
	query := `
		SELECT r.*, t.name as task_name 
		FROM backup_records r 
		JOIN backup_tasks t ON r.task_id = t.ID 
		WHERE r.ID = ?
	`
	
	var record types.BackupRecord
	err := db.Get(&record, query, recordID)
	if err != nil {
		return nil, fmt.Errorf("获取备份记录失败: %w", err)
	}
	
	return &record, nil
}

func getLatestRecordByTaskID(db *sqlx.DB, taskID int64) (*types.BackupRecord, error) {
	query := `
		SELECT r.*, t.name as task_name 
		FROM backup_records r 
		JOIN backup_tasks t ON r.task_id = t.ID 
		WHERE r.task_id = ? AND r.status = 'success'
		ORDER BY r.created_at DESC 
		LIMIT 1
	`
	
	var record types.BackupRecord
	err := db.Get(&record, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务最新备份失败: %w", err)
	}
	
	return &record, nil
}

func getLatestRecordByTaskName(db *sqlx.DB, taskName string) (*types.BackupRecord, error) {
	query := `
		SELECT r.*, t.name as task_name 
		FROM backup_records r 
		JOIN backup_tasks t ON r.task_id = t.ID 
		WHERE t.name = ? AND r.status = 'success'
		ORDER BY r.created_at DESC 
		LIMIT 1
	`
	
	var record types.BackupRecord
	err := db.Get(&record, query, taskName)
	if err != nil {
		return nil, fmt.Errorf("获取任务 '%s' 最新备份失败: %w", taskName, err)
	}
	
	return &record, nil
}

func getRecordByTimePoint(db *sqlx.DB, timePoint string) (*types.BackupRecord, error) {
	targetTime, _ := time.Parse("2006-01-02 15:04:05", timePoint)
	
	query := `
		SELECT r.*, t.name as task_name 
		FROM backup_records r 
		JOIN backup_tasks t ON r.task_id = t.ID 
		WHERE r.status = 'success' AND r.created_at <= ?
		ORDER BY ABS(strftime('%s', r.created_at) - strftime('%s', ?)) ASC
		LIMIT 1
	`
	
	var record types.BackupRecord
	err := db.Get(&record, query, targetTime, targetTime)
	if err != nil {
		return nil, fmt.Errorf("获取时间点 '%s' 附近的备份失败: %w", timePoint, err)
	}
	
	return &record, nil
}
```

### 3. 备份内容分析

```go
func analyzeBackupContent(record *types.BackupRecord, config *RestoreConfig) ([]RestoreItem, error) {
	// 1. 读取备份文件列表
	backupPath := record.BackupPath
	if !filepath.IsAbs(backupPath) {
		return nil, fmt.Errorf("备份路径不是绝对路径: %s", backupPath)
	}

	// 2. 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("备份文件不存在: %s", backupPath)
	}

	// 3. 分析备份内容 (根据备份格式处理)
	var items []RestoreItem
	var err error

	if record.Compress {
		items, err = analyzeCompressedBackup(backupPath, config)
	} else {
		items, err = analyzeDirectoryBackup(backupPath, config)
	}

	if err != nil {
		return nil, fmt.Errorf("分析备份内容失败: %w", err)
	}

	// 4. 应用文件过滤规则
	filteredItems := applyFileFilters(items, config)

	return filteredItems, nil
}

func analyzeCompressedBackup(backupPath string, config *RestoreConfig) ([]RestoreItem, error) {
	// 处理压缩备份文件 (tar.gz, zip 等)
	// 这里需要根据实际的压缩格式实现
	// 示例实现框架:
	
	var items []RestoreItem
	
	// TODO: 实现压缩文件分析逻辑
	// 1. 打开压缩文件
	// 2. 遍历文件列表
	// 3. 构建 RestoreItem 列表
	
	return items, nil
}

func analyzeDirectoryBackup(backupPath string, config *RestoreConfig) ([]RestoreItem, error) {
	var items []RestoreItem
	
	err := filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 计算相对路径
		relPath, err := filepath.Rel(backupPath, path)
		if err != nil {
			return err
		}
		
		// 跳过根目录
		if relPath == "." {
			return nil
		}
		
		// 构建目标路径
		targetPath := filepath.Join(config.TargetDir, relPath)
		
		item := RestoreItem{
			SourcePath:  path,
			TargetPath:  targetPath,
			IsDirectory: info.IsDir(),
			Size:        info.Size(),
			ModTime:     info.ModTime(),
			Permissions: info.Mode(),
		}
		
		items = append(items, item)
		return nil
	})
	
	return items, err
}

func applyFileFilters(items []RestoreItem, config *RestoreConfig) []RestoreItem {
	var filtered []RestoreItem
	
	for _, item := range items {
		// 检查特定路径过滤
		if len(config.SpecificPaths) > 0 {
			matched := false
			for _, path := range config.SpecificPaths {
				if strings.Contains(item.SourcePath, path) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		
		// 检查包含规则
		if len(config.IncludeRules) > 0 {
			matched := false
			for _, rule := range config.IncludeRules {
				if matched, _ := filepath.Match(rule, filepath.Base(item.SourcePath)); matched {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		
		// 检查排除规则
		excluded := false
		for _, rule := range config.ExcludeRules {
			if matched, _ := filepath.Match(rule, filepath.Base(item.SourcePath)); matched {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		
		filtered = append(filtered, item)
	}
	
	return filtered
}
```

### 4. 恢复执行逻辑

```go
func executeRestore(items []RestoreItem, config *RestoreConfig) (*RestoreResult, error) {
	result := &RestoreResult{
		TotalFiles: len(items),
	}
	
	startTime := time.Now()
	
	for _, item := range items {
		if !config.Quiet {
			fmt.Printf("恢复: %s -> %s\n", item.SourcePath, item.TargetPath)
		}
		
		err := restoreItem(item, config)
		if err != nil {
			result.FailedFiles++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", item.SourcePath, err))
			
			if config.Verbose {
				fmt.Printf("错误: %v\n", err)
			}
			continue
		}
		
		result.RestoredFiles++
		result.TotalSize += item.Size
	}
	
	result.Duration = time.Since(startTime)
	
	// 验证恢复结果
	if config.VerifyAfter {
		if err := verifyRestoreResult(items, config); err != nil {
			return result, fmt.Errorf("恢复验证失败: %w", err)
		}
	}
	
	return result, nil
}

func restoreItem(item RestoreItem, config *RestoreConfig) error {
	// 1. 检查目标路径是否存在
	if _, err := os.Stat(item.TargetPath); err == nil {
		// 文件已存在，根据覆盖策略处理
		action, err := handleExistingFile(item.TargetPath, config.OverwriteMode)
		if err != nil {
			return err
		}
		
		switch action {
		case "skip":
			return nil
		case "backup":
			if err := backupExistingFile(item.TargetPath); err != nil {
				return fmt.Errorf("备份现有文件失败: %w", err)
			}
		}
	}
	
	// 2. 创建目标目录
	targetDir := filepath.Dir(item.TargetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}
	
	// 3. 恢复文件或目录
	if item.IsDirectory {
		return restoreDirectory(item, config)
	} else {
		return restoreFile(item, config)
	}
}

func restoreFile(item RestoreItem, config *RestoreConfig) error {
	// 复制文件
	src, err := os.Open(item.SourcePath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer src.Close()
	
	dst, err := os.Create(item.TargetPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()
	
	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}
	
	// 保持文件属性
	if config.PreserveAttrs {
		if err := os.Chmod(item.TargetPath, item.Permissions); err != nil {
			return fmt.Errorf("设置文件权限失败: %w", err)
		}
		
		if err := os.Chtimes(item.TargetPath, item.ModTime, item.ModTime); err != nil {
			return fmt.Errorf("设置文件时间失败: %w", err)
		}
	}
	
	return nil
}

func restoreDirectory(item RestoreItem, config *RestoreConfig) error {
	// 创建目录
	err := os.MkdirAll(item.TargetPath, item.Permissions)
	if err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	
	// 保持目录属性
	if config.PreserveAttrs {
		if err := os.Chmod(item.TargetPath, item.Permissions); err != nil {
			return fmt.Errorf("设置目录权限失败: %w", err)
		}
		
		if err := os.Chtimes(item.TargetPath, item.ModTime, item.ModTime); err != nil {
			return fmt.Errorf("设置目录时间失败: %w", err)
		}
	}
	
	return nil
}
```

## 📋 使用示例

### 基本用法

```bash
# 恢复任务ID为1的最新备份到指定目录
bakctl restore --task-id 1 --target /restore/path

# 恢复指定备份记录到目录
bakctl restore --record-id 123 --target /restore/path

# 恢复指定任务名称的最新备份
bakctl restore --task-name "文档备份" --target /restore/path

# 恢复指定时间点的备份
bakctl restore --time "2024-01-15 14:30:00" --target /restore/path
```

### 高级用法

```bash
# 预览恢复内容 (不实际恢复)
bakctl restore --task-id 1 --target /restore/path --dry-run

# 只恢复特定文件类型
bakctl restore --task-id 1 --target /restore/path --include "*.txt,*.doc"

# 排除特定文件
bakctl restore --task-id 1 --target /restore/path --exclude "*.tmp,*.log"

# 只恢复指定路径
bakctl restore --task-id 1 --target /restore/path --path "documents/important"

# 覆盖现有文件并验证
bakctl restore --task-id 1 --target /restore/path --overwrite overwrite --verify

# 备份现有文件后恢复
bakctl restore --task-id 1 --target /restore/path --overwrite backup --verbose
```

## 🔍 错误处理

### 常见错误及处理

1. **备份不存在**
   ```
   错误: 未找到匹配的备份记录
   ```

2. **目标目录权限不足**
   ```
   错误: 创建目标目录失败: permission denied
   ```

3. **磁盘空间不足**
   ```
   错误: 复制文件失败: no space left on device
   ```

4. **备份文件损坏**
   ```
   错误: 恢复验证失败: 文件校验和不匹配
   ```

## 🚀 扩展性考虑

### 未来可能的扩展

1. **增量恢复** - 支持增量备份的恢复
2. **网络恢复** - 支持从远程位置恢复
3. **并行恢复** - 支持多线程并行恢复
4. **恢复日志** - 详细的恢复操作日志
5. **恢复计划** - 支持定时恢复任务

## 📋 实施检查清单

- [ ] 实现 flags.go 标志定义
- [ ] 实现 restore.go 主逻辑
- [ ] 实现备份记录查询功能
- [ ] 实现文件过滤逻辑
- [ ] 实现恢复执行逻辑
- [ ] 实现验证功能
- [ ] 添加单元测试
- [ ] 添加集成测试
- [ ] 更新文档和帮助信息
- [ ] 性能优化和错误处理完善

## 🎯 总结

这个 restore 子命令设计方案提供了：

1. **灵活的备份源选择** - 支持多种方式指定要恢复的备份
2. **精确的文件控制** - 支持包含/排除规则和路径过滤
3. **智能的冲突处理** - 多种覆盖策略处理现有文件
4. **完整的属性保持** - 保持原始文件权限和时间戳
5. **可靠的验证机制** - 恢复后验证确保数据完整性
6. **友好的用户体验** - 预览模式和详细的进度输出
7. **良好的扩展性** - 易于添加新功能和优化

该方案遵循了项目的整体架构风格，提供了完整、可靠、易用的备份恢复功能。