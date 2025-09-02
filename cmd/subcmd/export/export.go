package export

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	DB "gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/bakctl/internal/utils"
	"github.com/jmoiron/sqlx"
)

// ExportCmdMain 导出备份任务
//
// 参数:
//   - db: 数据库连接
//
// 返回:
//   - error: 导出过程中的错误信息
func ExportCmdMain(db *sqlx.DB) error {
	// 1. 验证所有参数
	if err := validateAllParameters(); err != nil {
		return err
	}

	// 4. 获取任务列表
	tasks, err := getTasksToExport(db)
	if err != nil {
		return err
	}

	// 5. 根据模式执行对应的逻辑
	// 导出添加任务命令模式
	if cmdF.Get() {
		return exportAddCommandsMode(tasks)
	}
	// 导出脚本模式
	if scriptF.Get() {
		return exportScriptMode(tasks)
	}

	// 默认打印帮助信息
	exportCmd.PrintHelp()
	return nil
}

// validateAllParameters 验证所有导出参数
//
// 返回:
//   - error: 验证失败时返回错误信息，否则返回 nil
func validateAllParameters() error {
	// 1. 验证导出模式，防止模式冲突
	hasCmd := cmdF.Get()
	hasScript := scriptF.Get()

	// 检查是否同时指定了多个模式
	if hasCmd && hasScript {
		return fmt.Errorf("--cmd/-c 和 --script/-s 不能同时使用")
	}

	// 2. 验证任务选择参数
	if hasCmd {
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
			return fmt.Errorf("请指定要导出的任务: -id, -ids 或 -all")
		}
		if count > 1 {
			return fmt.Errorf("-id, -ids 和 -all 只能选择一个")
		}
	}

	// 3. 验证执行指定模式下的参数
	if scriptF.Get() {
		path := pathF.Get() // 获取脚本路径

		// 检查脚本路径是否为空
		if path == "" {
			return fmt.Errorf("请指定脚本路径, 后缀为 .bat 或 .sh")
		}

		// 检查脚本路径是否已经存在
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("脚本路径已经存在: %s", path)
		}

		// 根据平台验证路径后缀
		if runtime.GOOS == "windows" {
			if !strings.HasSuffix(path, ".bat") {
				return fmt.Errorf("windows 平台脚本路径必须以 .bat 结尾")
			}
		} else {
			if !strings.HasSuffix(path, ".sh") {
				return fmt.Errorf("linux/unix 平台脚本路径必须以 .sh 结尾")
			}
		}
	}

	return nil
}

// getTasksToExport 获取要导出的备份任务列表
//
// 参数:
//   - db: 数据库连接
//
// 返回:
//   - []types.BackupTask: 要导出的备份任务列表
//   - error: 获取失败时返回错误信息，否则返回 nil
func getTasksToExport(db *sqlx.DB) ([]types.BackupTask, error) {
	// 获取所有任务
	if allF.Get() {
		return DB.GetAllTasks(db)
	}

	// 获取指定任务
	var taskIDs []int64
	if idF.Get() > 0 {
		taskIDs = []int64{int64(idF.Get())}
	}

	// 获取多个任务
	if len(idsF.Get()) > 0 {
		// 解析 idsF
		seen := make(map[int64]bool) // 检查重复ID
		for _, idStr := range idsF.Get() {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}

			// 解析 ID
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

	// 检查是否指定了任务ID
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("没有指定要导出的任务ID或指定的任务ID无效")
	}

	// 获取任务
	tasks, err := DB.GetTasksByIDs(db, taskIDs)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// exportAddCommandsMode 导出备份任务的添加命令模式
//
// 参数:
//   - tasks: 要导出的备份任务列表
//
// 返回:
//   - error: 导出过程中的错误信息
func exportAddCommandsMode(tasks []types.BackupTask) error {
	if len(tasks) == 0 {
		fmt.Println("没有找到要导出的任务")
		return nil
	}

	// 遍历任务列表，打印添加命令
	for _, task := range tasks {
		fmt.Printf("%s\n", buildAddCommand(task))
	}

	return nil
}

// getProgramName 获取当前程序的名称
//
// 返回:
//   - string: 程序名称
func getProgramName() string {
	if len(os.Args) == 0 {
		return "bakctl" // 默认名称
	}
	return filepath.Base(os.Args[0])
}

// buildAddCommand 构建添加命令
//
// 参数:
//   - task: 要添加的备份任务
//
// 返回:
//   - string: 添加命令
func buildAddCommand(task types.BackupTask) string {
	var parts []string
	// 动态获取程序名称
	programName := getProgramName()
	parts = append(parts, programName+" add")

	// 基本参数 (必需) - 根据最新的flags.go更新参数名
	parts = append(parts, fmt.Sprintf(`--name "%s"`, escapeQuotes(task.Name)))
	parts = append(parts, fmt.Sprintf(`--backup-dir "%s"`, escapeQuotes(task.BackupDir)))
	parts = append(parts, fmt.Sprintf(`--storage-dir "%s"`, escapeQuotes(filepath.Dir(task.StorageDir))))

	// 可选参数 (只有与默认值不同时才添加)
	if task.RetainCount != 3 { // 默认值
		parts = append(parts, fmt.Sprintf("--retain-count %d", task.RetainCount))
	}
	if task.RetainDays != 7 { // 默认值
		parts = append(parts, fmt.Sprintf("--retain-days %d", task.RetainDays))
	}
	if task.Compress {
		parts = append(parts, "--compress")
	}

	// 处理包含规则 - 每个规则作为单独的参数
	if task.IncludeRules != "[]" && task.IncludeRules != "" {
		rules, err := utils.UnmarshalRules(task.IncludeRules)
		if err == nil && len(rules) > 0 {
			// 使用逗号分隔的方式，因为add命令支持这种格式
			parts = append(parts, fmt.Sprintf(`--include "%s"`, escapeQuotes(strings.Join(rules, ","))))
		}
	}

	// 处理排除规则 - 每个规则作为单独的参数
	if task.ExcludeRules != "[]" && task.ExcludeRules != "" {
		rules, err := utils.UnmarshalRules(task.ExcludeRules)
		if err == nil && len(rules) > 0 {
			// 使用逗号分隔的方式，因为add命令支持这种格式
			parts = append(parts, fmt.Sprintf(`--exclude "%s"`, escapeQuotes(strings.Join(rules, ","))))
		}
	}

	// 文件大小限制 - 根据最新的flags.go更新参数名
	if task.MaxFileSize > 0 {
		parts = append(parts, fmt.Sprintf("--max-size %d", task.MaxFileSize))
	}
	if task.MinFileSize > 0 {
		parts = append(parts, fmt.Sprintf("--min-size %d", task.MinFileSize))
	}

	return strings.Join(parts, " ")
}

// escapeQuotes 转义双引号
func escapeQuotes(s string) string {
	// 转义双引号
	return strings.ReplaceAll(s, `"`, `\"`)
}

// exportScriptMode 导出一键备份脚本模式
//
// 参数:
//   - tasks: 要导出的备份任务列表
//
// 返回:
//   - error: 导出过程中的错误信息
func exportScriptMode(tasks []types.BackupTask) error {
	if len(tasks) == 0 {
		fmt.Println("没有找到要导出的任务")
		return nil
	}

	// 获取脚本路径
	scriptPath := pathF.Get()

	// 根据系统平台判断执行对应的脚本导出函数
	if runtime.GOOS == "windows" {
		return exportWindowsBatScript(tasks, scriptPath)
	} else {
		return exportLinuxBashScript(tasks, scriptPath)
	}

}

// exportWindowsBatScript 导出 Windows BAT 脚本
//
// 参数:
//   - tasks: 要导出的备份任务列表
//   - scriptPath: 脚本保存路径，为空时使用默认路径
//
// 返回:
//   - error: 导出过程中的错误信息
func exportWindowsBatScript(tasks []types.BackupTask, scriptPath string) error {
	if len(tasks) == 0 {
		fmt.Println("没有找到要导出的任务")
		return nil
	}

	// 确定输出路径
	outputPath := scriptPath
	if outputPath == "" {
		outputPath = "backup_script.bat"
	}

	// 检查输出路径是否为.bat文件
	if filepath.Ext(outputPath) != ".bat" {
		outputPath += ".bat"
	}

	// 打开文件
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 写入脚本内容
	if _, err := file.WriteString("@echo off\n"); err != nil {
		return fmt.Errorf("写入脚本内容失败: %w", err)
	}

	// 获取程序名称
	programName := getProgramName()

	// 循环写入任务ID对应的脚本内容
	for _, task := range tasks {
		if _, err := file.WriteString(fmt.Sprintf("start %s run -id %d\n", programName, task.ID)); err != nil {
			return fmt.Errorf("写入脚本内容失败: %w", err)
		}
	}

	// 打印提示信息
	fmt.Printf("已生成 Windows BAT 脚本到 %s 路径下\n", outputPath)
	return nil
}

// exportLinuxBashScript 导出 Linux Bash 脚本
//
// 参数:
//   - tasks: 要导出的备份任务列表
//   - scriptPath: 脚本保存路径，为空时使用默认路径
//
// 返回:
//   - error: 导出过程中的错误信息
func exportLinuxBashScript(tasks []types.BackupTask, scriptPath string) error {
	if len(tasks) == 0 {
		fmt.Println("没有找到要导出的任务")
		return nil
	}

	// 确定输出路径
	outputPath := scriptPath
	if outputPath == "" {
		outputPath = "backup_script.sh"
	}

	// 检查输出路径是否为.sh文件
	if filepath.Ext(outputPath) != ".sh" {
		outputPath += ".sh"
	}

	// 打开文件
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 写入脚本内容
	if _, err := file.WriteString("#!/usr/bin/env bash\n"); err != nil {
		return fmt.Errorf("写入脚本内容失败: %w", err)
	}

	// 获取程序名称
	programName := getProgramName()

	// 循环写入任务ID对应的脚本内容
	for _, task := range tasks {
		if _, err := file.WriteString(fmt.Sprintf("( %s run -id %d ) &\npid%d=$!\n", programName, task.ID, task.ID)); err != nil {
			return fmt.Errorf("写入脚本内容失败: %w", err)
		}
	}

	// 构建等待命令
	waitCmd := "wait"
	for _, task := range tasks {
		waitCmd += fmt.Sprintf(" $pid%d", task.ID)
	}

	// 写入等待命令
	if _, err := file.WriteString(waitCmd + "\n"); err != nil {
		return fmt.Errorf("写入脚本内容失败: %w", err)
	}

	// 打印提示信息
	fmt.Printf("已生成 Linux Bash 脚本到 %s 路径下\n", outputPath)
	return nil
}
