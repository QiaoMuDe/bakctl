// Package export 实现了 bakctl 的 export 子命令功能。
//
// 该包提供了导出备份任务配置的功能，支持：
//   - 将数据库中的任务配置导出为 TOML 格式文件
//   - 支持导出单个任务或批量导出多个任务
//   - 生成可用于重新导入的标准配置文件
//   - 提供配置文件的格式化和验证
//
// 主要功能包括：
//   - 从数据库读取任务配置
//   - 转换为标准的 TOML 配置格式
//   - 处理特殊字段的序列化（如 JSON 数组字段）
//   - 生成格式良好的配置文件
//   - 提供导出进度和结果反馈
//
// 导出的配置文件可以用于备份、迁移或批量创建任务。
package export

import (
	"fmt"
	"os"
	"path/filepath"
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

	// 检查是否没有指定导出模式
	if !hasCmd && !hasScript {
		return fmt.Errorf("请指定导出模式: --cmd/-c 或 --script/-s")
	}

	// 2. 验证任务选择参数
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

	// 3. 验证脚本模式的平台标志
	if scriptF.Get() {
		hasBat := batF.Get()
		hasShell := shF.Get()

		// 检查是否指定了平台标志
		if !hasBat && !hasShell {
			return fmt.Errorf("使用 --script/-s 时必须指定平台: -bat 或 -sh")
		}

		// 检查是否同时指定了多个平台标志
		if hasBat && hasShell {
			return fmt.Errorf("-bat 和 -sh 不能同时使用")
		}
	}

	// 4. 验证平台标志只能与script模式一起使用
	if (batF.Get() || shF.Get()) && !scriptF.Get() {
		return fmt.Errorf("-bat 和 -sh 只能与 --script/-s 一起使用")
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
		for _, i := range idsF.Get() {
			if i <= 0 {
				return nil, fmt.Errorf("任务ID必须大于0: %d", i)
			}
			if seen[i] {
				return nil, fmt.Errorf("重复的任务ID: %d", i)
			}

			seen[i] = true
			taskIDs = append(taskIDs, i)
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

	// 根据用户指定的平台标志选择脚本格式
	if batF.Get() {
		return printWindowsBatScript(tasks)
	}

	if shF.Get() {
		return printLinuxBashScript(tasks)
	}

	// 这里不应该到达，因为验证函数已经确保了平台标志的存在
	return fmt.Errorf("未指定脚本平台")
}

// printWindowsBatScript 打印 Windows BAT 脚本到终端
//
// 参数:
//   - tasks: 要导出的备份任务列表
//
// 返回:
//   - error: 导出过程中的错误信息
func printWindowsBatScript(tasks []types.BackupTask) error {
	if len(tasks) == 0 {
		return nil
	}

	// 获取程序名称
	programName := getProgramName()

	// 打印脚本头
	fmt.Println("@echo off")

	// 打印每个任务的执行命令
	for _, task := range tasks {
		fmt.Printf("start %s run -id %d\n", programName, task.ID)
	}

	return nil
}

// printLinuxBashScript 打印 Linux Bash 脚本到终端
//
// 参数:
//   - tasks: 要导出的备份任务列表
//
// 返回:
//   - error: 导出过程中的错误信息
func printLinuxBashScript(tasks []types.BackupTask) error {
	if len(tasks) == 0 {
		return nil
	}

	// 获取程序名称
	programName := getProgramName()

	// 打印脚本头
	fmt.Println("#!/usr/bin/env bash")

	// 打印每个任务的执行命令和进程ID保存
	for _, task := range tasks {
		fmt.Printf("( %s run -id %d ) &\n", programName, task.ID)
		fmt.Printf("pid%d=$!\n", task.ID)
	}

	// 构建并打印等待命令
	fmt.Print("wait")
	for _, task := range tasks {
		fmt.Printf(" $pid%d", task.ID)
	}
	fmt.Println()

	return nil
}
