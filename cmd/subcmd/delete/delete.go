// Package delete 实现了 bakctl 的 delete 子命令功能。
//
// 该包提供了删除备份任务和相关数据的完整功能，支持多种删除模式：
//   - 仅删除任务配置：保留备份文件，只删除数据库中的任务记录
//   - 删除任务和备份：同时删除任务配置和所有相关的备份文件
//   - 批量删除：支持同时删除多个任务
//   - 选择性删除：可以选择删除特定版本的备份记录
//
// 主要功能包括：
//   - 任务存在性验证和权限检查
//   - 备份文件的安全删除和错误处理
//   - 数据库记录的一致性维护
//   - 详细的删除进度和结果统计
//   - 用户确认和安全提示机制
//
// 删除操作具有不可逆性，因此包含多重安全检查和用户确认流程。
package delete

import (
	"fmt"
	"os"
	"strings"

	DB "gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/colorlib"
	"github.com/jmoiron/sqlx"
)

// DeleteResult 删除结果
type DeleteResult struct {
	TaskID         int    `json:"task_id"`             // 任务ID
	TaskName       string `json:"task_name"`           // 任务名称
	FilesDeleted   int    `json:"files_deleted"`       // 删除的文件数量
	FilesSkipped   int    `json:"files_skipped"`       // 跳过的文件数量
	RecordsDeleted int    `json:"records_deleted"`     // 删除的记录数量
	Success        bool   `json:"success"`             // 是否成功
	ErrorMsg       string `json:"error_msg,omitempty"` // 错误信息
}

// DeleteSummary 删除汇总
type DeleteSummary struct {
	TotalTasks   int            `json:"total_tasks"`   // 总任务数
	SuccessTasks int            `json:"success_tasks"` // 成功任务数
	FailedTasks  int            `json:"failed_tasks"`  // 失败任务数
	TotalFiles   int            `json:"total_files"`   // 总文件数
	TotalRecords int            `json:"total_records"` // 总记录数
	Results      []DeleteResult `json:"results"`       // 删除结果列表
}

// DeleteCmdMain delete命令的主函数
//
// 参数:
//   - db: 数据库连接
//   - cl: 颜色库
//
// 返回:
//   - error: 执行过程中发生的错误
func DeleteCmdMain(db *sqlx.DB, cl *colorlib.ColorLib) error {
	// 验证参数
	if err := validateFlags(); err != nil {
		return fmt.Errorf("参数错误: %w", err)
	}

	// 检查是否是删除失败记录模式
	if failedF.Get() {
		return deleteFailedRecords(db, cl)
	}

	// 原有的删除任务逻辑
	return deleteTasksMode(db, cl)
}

// deleteFailedRecords 删除失败的备份记录（保留任务）
func deleteFailedRecords(db *sqlx.DB, cl *colorlib.ColorLib) error {
	// 获取所有失败的备份记录
	failedRecords, err := DB.GetFailedBackupRecords(db)
	if err != nil {
		return fmt.Errorf("获取失败备份记录失败: %w", err)
	}

	if len(failedRecords) == 0 {
		cl.Yellow("没有找到失败的备份记录")
		return nil
	}

	// 显示要删除的失败记录
	cl.Redf("找到 %d 条失败的备份记录:\n", len(failedRecords))
	for _, record := range failedRecords {
		cl.Whitef("  - 任务ID: %d, 版本ID: %s, 失败原因: %s\n",
			record.TaskID, record.VersionID, record.FailureMessage)
	}

	// 用户确认
	if !forceF.Get() {
		cl.Yellow("\n确认删除这些失败的备份记录吗? (y/N): ")
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			// 如果读取输入失败，默认取消操作
			cl.White("操作已取消")
			return nil
		}
		if input != "y" && input != "Y" {
			cl.White("操作已取消")
			return nil
		}
	}

	// 构建要删除的记录ID列表
	var recordIDs []int64
	for _, record := range failedRecords {
		recordIDs = append(recordIDs, record.ID)
	}

	// 批量删除失败记录
	deletedCount, err := DB.DeleteBackupRecordsByIDs(db, recordIDs)
	if err != nil {
		return fmt.Errorf("删除失败记录时出错: %w", err)
	}

	cl.Greenf("成功删除 %d 条失败的备份记录\n", deletedCount)
	return nil
}

// deleteTasksMode 删除任务模式（原有逻辑）
func deleteTasksMode(db *sqlx.DB, cl *colorlib.ColorLib) error {
	// 获取要删除的任务ID列表
	taskIDs, err := getTaskIDsForTasks()
	if err != nil {
		return fmt.Errorf("任务ID解析失败: %w", err)
	}

	// 获取任务列表
	tasks, err := DB.GetTasksByIDs(db, taskIDs)
	if err != nil {
		return fmt.Errorf("查找任务失败: %w", err)
	}

	// 验证任务列表
	if len(tasks) == 0 {
		return fmt.Errorf("未找到要删除的任务")
	}

	// 用户确认
	confirmed, err := confirmDeletion(db, tasks, keepFilesF.Get(), cl)
	if err != nil {
		return fmt.Errorf("确认操作失败: %w", err)
	}

	if !confirmed {
		return nil // 用户取消，不需要额外提示
	}

	// 执行删除
	summary, err := deleteTasks(db, tasks, cl)
	if err != nil {
		return fmt.Errorf("执行删除操作失败: %w", err)
	}

	// 打印结果汇总
	printSummary(summary, cl)

	return nil
}

// getTaskIDsForTasks 获取要删除的任务ID列表（用于删除任务模式）
//
// 返回:
//   - []int64: 要删除的任务ID列表
//   - error: 获取失败时返回错误信息
func getTaskIDsForTasks() ([]int64, error) {
	id := idF.Get()   // 从命令行参数获取任务ID
	ids := idsF.Get() // 从命令行参数获取任务ID列表

	// 如果指定了任务ID，则只处理一个任务
	if id > 0 {
		return []int64{id}, nil
	}

	// 如果指定了任务ID列表，则处理多个任务
	var taskIDs []int64
	seen := make(map[int64]bool) // 用于检查重复ID
	for _, i := range ids {
		if i <= 0 {
			return nil, fmt.Errorf("任务ID必须大于0: %d", i)
		}

		// 检查重复
		if seen[i] {
			return nil, fmt.Errorf("任务ID列表中存在重复的ID: %d", i)
		}
		seen[i] = true
		taskIDs = append(taskIDs, i)
	}

	return taskIDs, nil
}

// confirmDeletion 显示删除信息并确认
//
// 参数:
//   - db: 数据库连接
//   - tasks: 要删除的任务列表
//   - keepFiles: 是否保留备份文件
//   - cl: 颜色库
//
// 返回:
//   - bool: 用户是否确认删除
//   - error: 确认失败时返回错误信息
func confirmDeletion(db *sqlx.DB, tasks []types.BackupTask, keepFiles bool, cl *colorlib.ColorLib) (bool, error) {
	if forceF.Get() {
		return true, nil
	}

	cl.Blue("即将删除以下备份任务: ")
	fmt.Println()

	// 批量获取所有任务的备份记录
	recordsByTask, err := DB.GetBatchBackupRecords(db, tasks)
	if err != nil {
		return false, fmt.Errorf("获取备份记录失败: %w", err)
	}

	totalRecords := 0 // 记录总数
	totalFiles := 0   // 文件总数

	// 遍历所有任务
	for _, task := range tasks {
		records := recordsByTask[task.ID]
		fileCount := len(records)
		if keepFiles {
			fileCount = 0
		}

		cl.Whitef("任务ID: %d, 名称: '%s', 备份记录: %d个", task.ID, task.Name, len(records))

		if !keepFiles {
			cl.Whitef(", 预计删除文件: %d个", fileCount)
		}
		fmt.Println()

		totalRecords += len(records)
		totalFiles += fileCount
	}

	fmt.Println()
	cl.Bluef("总计: %d个任务, %d个备份记录", len(tasks), totalRecords)
	if !keepFiles {
		cl.Bluef(", %d个备份文件", totalFiles)
	}
	fmt.Println()
	fmt.Println()

	if !keepFiles {
		cl.Red("警告: 此操作不可逆！备份文件将被永久删除。")
	} else {
		cl.Yellow("注意: 只删除数据库记录，备份文件将被保留。")
	}
	fmt.Println()

	cl.Blue("确认删除? (输入 y 或 yes 确认，其他任意键取消): ")

	// 读取用户输入
	var response string
	_, err = fmt.Scanln(&response)
	if err != nil {
		cl.White("操作已取消")
		return false, nil // 默认为不确认
	}

	response = strings.ToLower(strings.TrimSpace(response))
	confirmed := response == "y" || response == "yes"

	if !confirmed {
		cl.White("操作已取消")
	}

	return confirmed, nil
}

// deleteTasks 批量删除任务
//
// 参数:
//   - db: 数据库连接
//   - tasks: 要删除的任务列表
//   - cl: 颜色库
//
// 返回:
//   - DeleteSummary: 删除结果摘要
//   - error: 删除失败时返回错误信息
func deleteTasks(db *sqlx.DB, tasks []types.BackupTask, cl *colorlib.ColorLib) (DeleteSummary, error) {
	// 初始化结果
	summary := DeleteSummary{
		TotalTasks: len(tasks),
		Results:    make([]DeleteResult, 0, len(tasks)),
	}

	// 遍历任务列表
	for i, task := range tasks {
		cl.Bluef("[%d/%d] 正在删除任务: %s (ID: %d)\n", i+1, len(tasks), task.Name, task.ID)

		// 删除单个任务
		result := deleteTask(db, task, cl)
		summary.Results = append(summary.Results, result)

		if result.Success {
			summary.SuccessTasks++                        // 成功的任务数加一
			summary.TotalFiles += result.FilesDeleted     // 总的文件数加一
			summary.TotalRecords += result.RecordsDeleted // 总的记录数加一
		} else {
			summary.FailedTasks++ // 失败的任务数加一
		}
	}

	return summary, nil
}

// deleteTask 删除单个任务
//
// 参数:
//   - db: 数据库连接
//   - task: 要删除的任务
//   - cl: 颜色库
//
// 返回:
//   - DeleteResult: 删除结果
func deleteTask(db *sqlx.DB, task types.BackupTask, cl *colorlib.ColorLib) DeleteResult {
	result := DeleteResult{
		TaskID:   int(task.ID),
		TaskName: task.Name,
		Success:  false,
	}

	// 获取备份记录
	records, err := DB.GetBackupRecordsByTaskID(db, task.ID)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("获取备份记录失败: %v", err)
		return result
	}

	// 删除备份文件（如果不保留文件）
	if !keepFilesF.Get() {
		deleted, skipped, err := deleteBackupFiles(records)
		result.FilesDeleted = deleted
		result.FilesSkipped = skipped

		if err != nil {
			// 文件删除失败不影响数据库操作，只记录错误
			cl.Redf("部分文件删除失败: %v\n", err)
		}

		if err == nil {
			cl.Whitef("删除备份文件: %d个\n", deleted)
		} else {
			cl.Red("删除备份文件失败")
		}
	} else {
		cl.White("跳过文件删除: 0个")
	}

	// 删除备份记录
	recordsDeleted, err := DB.DeleteBackupRecords(db, task.ID)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("删除备份记录失败: %v", err)
		return result
	}
	result.RecordsDeleted = recordsDeleted
	cl.Whitef("删除备份记录: %d个\n", recordsDeleted)

	// 删除备份任务
	err = DB.DeleteBackupTask(db, task.ID)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("删除备份任务失败: %v", err)
		return result
	}
	cl.White("删除任务配置: 1个")

	result.Success = true
	return result
}

// deleteBackupFiles 删除备份文件
//
// 参数:
//   - records: 备份记录列表
//
// 返回:
//   - int: 成功删除的文件数
//   - int: 跳过的文件数
//   - error: 删除失败时返回错误信息
func deleteBackupFiles(records []types.BackupRecord) (int, int, error) {
	var deleted, skipped int
	var errors []string

	for _, record := range records {
		if record.StoragePath == "" {
			skipped++
			continue
		}

		// 检查文件是否存在
		if _, err := os.Stat(record.StoragePath); os.IsNotExist(err) {
			skipped++
			continue
		}

		// 删除项
		err := os.RemoveAll(record.StoragePath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("删除文件 %s 失败: %v", record.StoragePath, err))
			skipped++
			continue
		}

		deleted++
	}

	var err error
	if len(errors) > 0 {
		err = fmt.Errorf("部分文件删除失败: %s", strings.Join(errors, "; "))
	}

	return deleted, skipped, err
}

// printSummary 打印删除汇总
//
// 参数:
//   - summary: 删除结果摘要
//   - cl: 颜色库
func printSummary(summary DeleteSummary, cl *colorlib.ColorLib) {
	fmt.Println()
	cl.Greenf("删除完成！成功: %d个任务, 失败: %d个任务\n", summary.SuccessTasks, summary.FailedTasks)

	if summary.FailedTasks > 0 {
		cl.Red("\n失败详情:")
		for _, result := range summary.Results {
			if !result.Success {
				cl.Redf("  任务ID %d (%s): %s\n", result.TaskID, result.TaskName, result.ErrorMsg)
			}
		}
	}
}

// validateFlags 验证命令行参数
//
// 返回:
//   - error: 验证失败时返回错误信息
func validateFlags() error {
	id := idF.Get()
	idsStr := idsF.Get()
	failed := failedF.Get()

	// 计算指定的参数数量
	paramCount := 0
	if id != 0 {
		paramCount++
	}
	if len(idsStr) > 0 {
		paramCount++
	}
	if failed {
		paramCount++
	}

	// 检查是否指定了至少一个参数
	if paramCount == 0 {
		return fmt.Errorf("必须指定 -id、-ids 或 --failed 参数之一")
	}

	// 检查参数互斥性
	if paramCount > 1 {
		return fmt.Errorf("-id、-ids 和 --failed 参数不能同时使用")
	}

	// 检查单个ID值是否有效
	if id < 0 {
		return fmt.Errorf("任务ID必须大于0")
	}

	return nil
}
