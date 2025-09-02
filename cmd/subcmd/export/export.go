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
	if hasID {
		count++
	}
	if hasIDs {
		count++
	}
	if hasAll {
		count++
	}

	if count == 0 {
		return fmt.Errorf("请指定要导出的任务: --id, --ids 或 --all")
	}
	if count > 1 {
		return fmt.Errorf("--id, --ids 和 --all 只能选择一个")
	}

	return nil
}

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
