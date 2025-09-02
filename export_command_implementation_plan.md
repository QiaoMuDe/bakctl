# Export 子命令实现方案

## 🎯 概述

本文档详细描述了 CBK 备份管理工具的 export 子命令实现方案。该命令功能：
- 导出备份任务的添加命令到终端

## 📁 文件结构设计

```
cmd/subcmd/export/
├── flags.go          // 标志定义
└── export.go         // 主要实现逻辑
```

## 🚩 标志定义方案 (flags.go)

```go
package export

import (
    "flag"
    "gitee.com/MM-Q/qflag"
    "gitee.com/MM-Q/qflag/cmd"
)

var (
    exportCmd *cmd.Cmd // 导出备份任务命令

    // 任务选择标志
    idF     *qflag.IntFlag   // 单个任务ID
    idsF    *qflag.SliceFlag // 多个任务ID
    allF    *qflag.BoolFlag  // 导出所有任务
)

func InitExportCmd() *cmd.Cmd {
    exportCmd = cmd.NewCmd("export", "exp", flag.ExitOnError)
    exportCmd.SetUseChinese(true)
    exportCmd.SetDescription("导出备份任务的添加命令")

    // 任务选择标志 (三选一)
    idF = exportCmd.Int("id", "I", 0, "指定单个任务ID进行导出")
    idsF = exportCmd.Slice("ids", "S", []string{}, "指定多个任务ID进行导出，用逗号分隔")
    allF = exportCmd.Bool("all", "A", false, "导出所有任务")

    return exportCmd
}
```

## 🔧 核心功能设计

### 导出添加命令功能

```go
// 导出格式示例：
// bakctl add -n "备份任务1" -s "/path/to/source" -d "/path/to/dest" -r 5 -t 30 --compress
// bakctl add -n "备份任务2" -s "/path/to/source2" -d "/path/to/dest2" -r 3 -t 7

func buildAddCommand(task types.BackupTask) string {
    // 构建完整的 cbk add 命令
    // 处理特殊字符转义
    // 处理布尔值、数组等复杂参数
}
```

## 📋 详细实现方案

### 1. 主函数逻辑 (export.go)

```go
package export

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
    
    "gitee.com/MM-Q/bakctl/internal/types"
    "github.com/jmoiron/sqlx"
)

func ExportCmdMain(db *sqlx.DB) error {
    // 1. 参数验证
    if err := validateExportFlags(); err != nil {
        return err
    }

    // 2. 获取任务列表
    tasks, err := getTasksToExport(db)
    if err != nil {
        return err
    }

    // 3. 导出添加命令
    return exportAddCommands(tasks)
}

func validateExportFlags() error {
    // 验证任务选择
    hasID := idF.Get() > 0
    hasIDs := len(idsF.Get()) > 0
    hasAll := allF.Get()

    count := 0
    if hasID { count++ }
    if hasIDs { count++ }
    if hasAll { count++ }

    if count == 0 {
        return fmt.Errorf("请指定要导出的任务: --id, --ids 或 --all")
    }
    if count > 1 {
        return fmt.Errorf("--id, --ids 和 --all 只能选择一个")
    }

    return nil
}
```

### 2. 任务获取逻辑

```go
func getTasksToExport(db *sqlx.DB) ([]types.BackupTask, error) {
    if allF.Get() {
        return getAllTasks(db)
    }

    var taskIDs []int64
    if idF.Get() > 0 {
        taskIDs = []int64{int64(idF.Get())}
    } else {
        // 解析 idsF
        seen := make(map[int64]bool) // 检查重复ID
        for _, idStr := range idsF.Get() {
            idStr = strings.TrimSpace(idStr)
            if idStr == "" {
                continue
            }
            
            id, err := strconv.ParseInt(idStr, 10, 64)
            if err != nil {
                return nil, fmt.Errorf("无效的任务ID: %s", idStr)
            }
            if id <= 0 {
                return nil, fmt.Errorf("任务ID必须大于0: %d", id)
            }
            if seen[id] {
                return nil, fmt.Errorf("重复的任务ID: %d", id)
            }
            
            seen[id] = true
            taskIDs = append(taskIDs, id)
        }
    }

    return getTasksByIDs(db, taskIDs)
}

func getAllTasks(db *sqlx.DB) ([]types.BackupTask, error) {
    query := `
        SELECT ID, name, retain_count, retain_days, backup_dir, storage_dir, 
               compress, include_rules, exclude_rules, max_file_size, min_file_size
        FROM backup_tasks 
        ORDER BY ID
    `
    
    var tasks []types.BackupTask
    err := db.Select(&tasks, query)
    if err != nil {
        return nil, fmt.Errorf("获取所有任务失败: %w", err)
    }
    
    return tasks, nil
}

func getTasksByIDs(db *sqlx.DB, taskIDs []int64) ([]types.BackupTask, error) {
    if len(taskIDs) == 0 {
        return []types.BackupTask{}, nil
    }

    query := `
        SELECT ID, name, retain_count, retain_days, backup_dir, storage_dir, 
               compress, include_rules, exclude_rules, max_file_size, min_file_size
        FROM backup_tasks 
        WHERE ID IN (?)
        ORDER BY ID
    `
    
    query, args, err := sqlx.In(query, taskIDs)
    if err != nil {
        return nil, fmt.Errorf("构建查询失败: %w", err)
    }
    query = db.Rebind(query)

    var tasks []types.BackupTask
    err = db.Select(&tasks, query, args...)
    if err != nil {
        return nil, fmt.Errorf("获取任务失败: %w", err)
    }

    // 检查是否所有任务都存在
    if len(tasks) != len(taskIDs) {
        foundIDs := make(map[int64]bool)
        for _, task := range tasks {
            foundIDs[task.ID] = true
        }

        var missingIDs []int64
        for _, id := range taskIDs {
            if !foundIDs[id] {
                missingIDs = append(missingIDs, id)
            }
        }

        if len(missingIDs) > 0 {
            return tasks, fmt.Errorf("以下任务ID不存在: %v", missingIDs)
        }
    }

    return tasks, nil
}
```

### 3. 添加命令导出实现

```go
func exportAddCommands(tasks []types.BackupTask) error {
    if len(tasks) == 0 {
        fmt.Println("没有找到要导出的任务")
        return nil
    }

    fmt.Printf("# CBK 备份任务添加命令\n")
    fmt.Printf("# 生成时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
    
    for i, task := range tasks {
        fmt.Printf("# 任务 %d: %s\n", i+1, task.Name)
        fmt.Printf("%s\n\n", buildAddCommand(task))
    }

    return nil
}

// getProgramName 获取当前程序的名称
func getProgramName() string {
    if len(os.Args) == 0 {
        return "bakctl" // 默认名称
    }
    return filepath.Base(os.Args[0])
}

func buildAddCommand(task types.BackupTask) string {
    var parts []string
    // 动态获取程序名称
    programName := getProgramName()
    parts = append(parts, programName+" add")
    
    // 基本参数 (必需)
    parts = append(parts, fmt.Sprintf(`-n "%s"`, escapeQuotes(task.Name)))
    parts = append(parts, fmt.Sprintf(`-s "%s"`, escapeQuotes(task.BackupDir)))
    parts = append(parts, fmt.Sprintf(`-d "%s"`, escapeQuotes(task.StorageDir)))
    
    // 可选参数 (只有与默认值不同时才添加)
    if task.RetainCount != 3 { // 默认值
        parts = append(parts, fmt.Sprintf("-r %d", task.RetainCount))
    }
    if task.RetainDays != 7 { // 默认值
        parts = append(parts, fmt.Sprintf("-t %d", task.RetainDays))
    }
    if task.Compress {
        parts = append(parts, "--compress")
    }
    
    // 处理包含规则 - 使用逗号分隔的单个参数
    if task.IncludeRules != "[]" && task.IncludeRules != "" {
        rules := parseRulesFromJSON(task.IncludeRules)
        if len(rules) > 0 {
            parts = append(parts, fmt.Sprintf(`-i "%s"`, escapeQuotes(strings.Join(rules, ","))))
        }
    }
    
    // 处理排除规则 - 使用逗号分隔的单个参数
    if task.ExcludeRules != "[]" && task.ExcludeRules != "" {
        rules := parseRulesFromJSON(task.ExcludeRules)
        if len(rules) > 0 {
            parts = append(parts, fmt.Sprintf(`-e "%s"`, escapeQuotes(strings.Join(rules, ","))))
        }
    }
    
    // 文件大小限制
    if task.MaxFileSize > 0 {
        parts = append(parts, fmt.Sprintf("-M %d", task.MaxFileSize))
    }
    if task.MinFileSize > 0 {
        parts = append(parts, fmt.Sprintf("-m %d", task.MinFileSize))
    }
    
    return strings.Join(parts, " ")
}

func escapeQuotes(s string) string {
    // 转义双引号
    return strings.ReplaceAll(s, `"`, `\"`)
}

func parseRulesFromJSON(jsonStr string) []string {
    // 解析JSON格式的规则数组
    if jsonStr == "" || jsonStr == "[]" {
        return []string{}
    }
    
    var rules []string
    err := json.Unmarshal([]byte(jsonStr), &rules)
    if err != nil {
        // 解析失败时返回空数组
        return []string{}
    }
    
    return rules
}
```

## 📝 使用示例

### 基本用法

```bash
# 导出单个任务的添加命令
bakctl export --id 1

# 导出多个任务的添加命令
bakctl export --ids 1,2,3

# 导出所有任务的添加命令
bakctl export --all
```

## 📤 输出示例

### 添加命令导出示例

```bash
# CBK 备份任务添加命令
# 生成时间: 2025-01-02 15:30:45

# 任务 1: 文档备份
bakctl add -n "文档备份" -s "/home/user/documents" -d "/backup/docs" -r 5 -t 30 --compress -i "*.doc,*.pdf" -e "*.tmp"

# 任务 2: 代码备份  
bakctl add -n "代码备份" -s "/home/user/projects" -d "/backup/code" -r 10 -t 90 --compress -e "node_modules,.git"
```

## 🔍 错误处理

### 常见错误及处理

1. **任务不存在**
   ```
   错误: 以下任务ID不存在: [5, 8]
   ```

2. **参数冲突**
   ```
   错误: --id, --ids 和 --all 只能选择一个
   ```

## 🚀 扩展性考虑

### 未来可能的扩展

1. **输出选项**
   - 支持输出到文件
   - 支持不同格式（JSON、YAML）

2. **脚本导出**
   - 导出执行脚本
   - 支持多平台脚本

3. **高级功能**
   - 任务依赖关系导出
   - 定时任务配置导出
   - 配置验证和测试

## 📋 实施检查清单

- [ ] 创建 export 子命令目录结构
- [ ] 实现 flags.go 标志定义
- [ ] 实现 export.go 主逻辑
- [ ] 添加单元测试
- [ ] 添加集成测试
- [ ] 更新文档和帮助信息
- [ ] 错误处理完善

## 🎯 总结

这个 export 子命令实现方案提供了：

1. **灵活的任务选择** - 支持单个、多个或全部任务
2. **简洁的输出** - 直接输出到终端，便于查看和复制
3. **完整的命令重现** - 导出的命令可以直接执行来重建任务
4. **友好的用户体验** - 清晰的错误信息和格式化输出
5. **良好的扩展性** - 易于添加新功能

该方案遵循了项目的整体架构风格，提供了简洁高效的功能实现。