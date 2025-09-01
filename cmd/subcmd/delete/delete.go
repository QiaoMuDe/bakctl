package delete

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gitee.com/MM-Q/bakctl/internal/types"
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
func DeleteCmdMain(db *sqlx.DB) error {
	// 验证参数
	if err := validateFlags(); err != nil {
		return fmt.Errorf("参数验证失败: %w", err)
	}

	// 获取要删除的任务ID列表
	taskIDs, err := getTaskIDs()
	if err != nil {
		return fmt.Errorf("获取任务ID失败: %w", err)
	}

	// 查询要删除的任务
	tasks, err := selectTasksToDelete(db, taskIDs)
	if err != nil {
		return fmt.Errorf("查询任务失败: %w", err)
	}

	// 用户确认
	confirmed, err := confirmDeletion(db, tasks, keepFilesF.Get())
	if err != nil {
		return fmt.Errorf("确认删除失败: %w", err)
	}

	if !confirmed {
		fmt.Println("删除操作已取消")
		return nil
	}

	// 执行删除
	summary, err := deleteTasks(db, tasks)
	if err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}

	// 打印结果汇总
	printSummary(summary)

	return nil
}

// getTaskIDs 获取要删除的任务ID列表
//
// 返回:
//   - []int: 要删除的任务ID列表
//   - error: 获取失败时返回错误信息
func getTaskIDs() ([]int, error) {
	id := idF.Get()      // 从命令行参数获取任务ID
	idsStr := idsF.Get() // 从命令行参数获取任务ID列表

	// 如果指定了任务ID，则只处理一个任务
	if id > 0 {
		return []int{id}, nil
	}

	// 如果指定了任务ID列表，则处理多个任务
	var taskIDs []int
	for _, idStr := range idsStr {
		idStr = strings.TrimSpace(idStr) // 移除首尾空格
		if idStr == "" {
			continue
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("无效的任务ID: %s", idStr)
		}
		taskIDs = append(taskIDs, id)
	}

	return taskIDs, nil
}

// selectTasksToDelete 选择要删除的任务
//
// 参数:
//   - db: 数据库连接
//   - taskIDs: 要删除的任务ID列表
//
// 返回:
//   - []types.BackupTask: 要删除的任务列表
//   - error: 选择失败时返回错误信息
func selectTasksToDelete(db *sqlx.DB, taskIDs []int) ([]types.BackupTask, error) {
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("任务ID列表为空")
	}

	// 使用 sqlx.In 构建IN查询
	query := `
		SELECT id, name, source_path, target_path, schedule_type, schedule_value, 
		       enabled, created_at, updated_at
		FROM backup_tasks 
		WHERE id IN (?)
		ORDER BY id
	`

	query, args, err := sqlx.In(query, taskIDs)
	if err != nil {
		return nil, fmt.Errorf("构建IN查询失败: %w", err)
	}
	query = db.Rebind(query)

	var tasks []types.BackupTask
	err = db.Select(&tasks, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询备份任务失败: %w", err)
	}

	// 检查是否所有任务都存在
	if len(tasks) != len(taskIDs) {
		foundIDs := make(map[int]bool)
		for _, task := range tasks {
			foundIDs[int(task.ID)] = true
		}

		var missingIDs []int
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

// getBackupRecords 获取任务的备份记录
func getBackupRecords(db *sqlx.DB, taskID int) ([]types.BackupRecord, error) {
	query := `
		SELECT ID, task_id, task_name, version_id, backup_filename, backup_size, 
		       storage_path, status, failure_message, checksum, created_at
		FROM backup_records 
		WHERE task_id = ?
		ORDER BY created_at DESC
	`

	var records []types.BackupRecord
	err := db.Select(&records, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取备份记录失败: %w", err)
	}

	return records, nil
}

// confirmDeletion 显示删除信息并确认
func confirmDeletion(db *sqlx.DB, tasks []types.BackupTask, keepFiles bool) (bool, error) {
	if forceF.Get() {
		return true, nil
	}

	fmt.Println("即将删除以下备份任务：")
	fmt.Println()

	totalRecords := 0
	totalFiles := 0

	for _, task := range tasks {
		records, err := getBackupRecords(db, int(task.ID))
		if err != nil {
			return false, fmt.Errorf("获取任务 %d 的备份记录失败: %w", task.ID, err)
		}

		fileCount := len(records)
		if keepFiles {
			fileCount = 0
		}

		fmt.Printf("任务ID: %d, 名称: \"%s\", 备份记录: %d个",
			task.ID, task.Name, len(records))

		if !keepFiles {
			fmt.Printf(", 预计删除文件: %d个", fileCount)
		}
		fmt.Println()

		totalRecords += len(records)
		totalFiles += fileCount
	}

	fmt.Println()
	fmt.Printf("总计: %d个任务, %d个备份记录", len(tasks), totalRecords)
	if !keepFiles {
		fmt.Printf(", %d个备份文件", totalFiles)
	}
	fmt.Println()
	fmt.Println()

	if !keepFiles {
		fmt.Println("警告: 此操作不可逆！备份文件将被永久删除。")
	} else {
		fmt.Println("注意: 只删除数据库记录，备份文件将被保留。")
	}
	fmt.Println()

	fmt.Print("确认删除? (y/N): ")

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, nil // 默认为不确认
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// deleteTasks 批量删除任务
func deleteTasks(db *sqlx.DB, tasks []types.BackupTask) (DeleteSummary, error) {
	summary := DeleteSummary{
		TotalTasks: len(tasks),
		Results:    make([]DeleteResult, 0, len(tasks)),
	}

	for i, task := range tasks {
		fmt.Printf("[%d/%d] 正在删除任务: %s (ID: %d)\n", i+1, len(tasks), task.Name, task.ID)

		result := deleteTask(db, task)
		summary.Results = append(summary.Results, result)

		if result.Success {
			summary.SuccessTasks++
			summary.TotalFiles += result.FilesDeleted
			summary.TotalRecords += result.RecordsDeleted
		} else {
			summary.FailedTasks++
		}
	}

	return summary, nil
}

// deleteTask 删除单个任务
func deleteTask(db *sqlx.DB, task types.BackupTask) DeleteResult {
	result := DeleteResult{
		TaskID:   int(task.ID),
		TaskName: task.Name,
		Success:  false,
	}

	// 获取备份记录
	records, err := getBackupRecords(db, int(task.ID))
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
			fmt.Printf("  ⚠ 部分文件删除失败: %v\n", err)
		}

		if err == nil {
			fmt.Printf("  ✓ 删除备份文件: %d个\n", deleted)
		} else {
			fmt.Printf("  ✗ 删除备份文件失败\n")
		}
	} else {
		fmt.Printf("  ✓ 跳过文件删除: 0个\n")
	}

	// 删除备份记录
	recordsDeleted, err := deleteBackupRecords(db, int(task.ID))
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("删除备份记录失败: %v", err)
		return result
	}
	result.RecordsDeleted = recordsDeleted
	fmt.Printf("  ✓ 删除备份记录: %d个\n", recordsDeleted)

	// 删除备份任务
	err = deleteBackupTask(db, int(task.ID))
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("删除备份任务失败: %v", err)
		return result
	}
	fmt.Printf("  ✓ 删除任务配置: 1个\n")

	result.Success = true
	return result
}

// deleteBackupFiles 删除备份文件
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

		// 删除文件或目录
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

// deleteBackupRecords 删除备份记录
func deleteBackupRecords(db *sqlx.DB, taskID int) (int, error) {
	query := `DELETE FROM backup_records WHERE task_id = ?`

	result, err := db.Exec(query, taskID)
	if err != nil {
		return 0, fmt.Errorf("删除备份记录失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取删除记录数失败: %w", err)
	}

	return int(affected), nil
}

// deleteBackupTask 删除备份任务
func deleteBackupTask(db *sqlx.DB, taskID int) error {
	query := `DELETE FROM backup_tasks WHERE id = ?`

	result, err := db.Exec(query, taskID)
	if err != nil {
		return fmt.Errorf("删除备份任务失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除任务数失败: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("任务ID %d 不存在", taskID)
	}

	return nil
}

// printSummary 打印删除汇总
func printSummary(summary DeleteSummary) {
	fmt.Println()
	fmt.Printf("删除完成！成功: %d个任务, 失败: %d个任务\n", summary.SuccessTasks, summary.FailedTasks)

	if summary.FailedTasks > 0 {
		fmt.Println("\n失败详情:")
		for _, result := range summary.Results {
			if !result.Success {
				fmt.Printf("  任务ID %d (%s): %s\n", result.TaskID, result.TaskName, result.ErrorMsg)
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

	// 检查是否指定了ID或IDs
	if id == 0 && len(idsStr) == 0 {
		return fmt.Errorf("必须指定 -id 或 -ids 参数")
	}

	// 检查ID和IDs是否互斥
	if id != 0 && len(idsStr) > 0 {
		return fmt.Errorf("-id 和 -ids 参数不能同时使用")
	}

	// 检查ID值是否有效
	if id < 0 {
		return fmt.Errorf("任务ID必须大于0")
	}

	// 检查IDs值是否有效
	if len(idsStr) > 0 {
		seen := make(map[int]bool)
		for _, idStr := range idsStr {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}

			id, err := strconv.Atoi(idStr)
			if err != nil {
				return fmt.Errorf("无效的任务ID: %s", idStr)
			}
			if id <= 0 {
				return fmt.Errorf("任务ID必须大于0: %d", id)
			}

			// 检查重复
			if seen[id] {
				return fmt.Errorf("任务ID列表中存在重复的ID: %d", id)
			}
			seen[id] = true
		}
	}

	return nil
}
