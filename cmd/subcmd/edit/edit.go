package edit

import (
	"fmt"
	"strconv"

	DB "gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/utils"
	"github.com/jmoiron/sqlx"
)

func EditCmdMain(db *sqlx.DB) error {
	// 获取要编辑的任务ID列表
	taskIDs, err := getTaskIDs()
	if err != nil {
		return err
	}

	// 检查是否有指定要更新的任务
	if len(taskIDs) == 0 {
		return fmt.Errorf("请指定要编辑的任务ID: 使用 --id 指定单个任务或 --ids 指定多个任务")
	}

	// 检查是否有指定要更新的配置项
	if !hasAnyUpdateFlags() {
		return fmt.Errorf("没有指定要更新的配置项")
	}

	// 执行批量更新
	successCount := 0
	for _, taskID := range taskIDs {
		if err := updateTask(db, taskID); err != nil {
			fmt.Printf("更新任务ID %d 失败: %v\n", taskID, err)
		} else {
			successCount++
			fmt.Printf("任务ID %d 更新成功\n", taskID)
		}
	}

	fmt.Printf("成功更新 %d/%d 个任务\n", successCount, len(taskIDs))
	return nil
}

// getTaskIDs 获取要编辑的任务ID列表
//
// 返回值:
//   - []int64: 任务ID列表
//   - error: 获取失败时返回错误信息，否则返回 nil
func getTaskIDs() ([]int64, error) {
	var taskIDs []int64

	// 检查是否同时指定了单个ID和多个ID
	if idF.Get() > 0 && len(idsF.Get()) > 0 {
		return nil, fmt.Errorf("不能同时使用 -id 和 -ids 参数")
	}

	// 优先使用单个ID
	if singleID := idF.Get(); singleID > 0 {
		taskIDs = append(taskIDs, int64(singleID))
		return taskIDs, nil
	}

	// 处理多个ID (使用切片类型)
	idsSlice := idsF.Get()
	if len(idsSlice) <= 0 {
		return nil, fmt.Errorf("没有指定要编辑的任务ID: 使用 -id 指定单个任务或 -ids 指定多个任务")
	}

	// 遍历每个ID并解析
	for _, idStr := range idsSlice {
		id, err := strconv.ParseInt(idStr, 10, 64) // 获取任务ID
		if err != nil {
			return nil, fmt.Errorf("无效的任务ID: %s", idStr)
		}

		if id <= 0 {
			return nil, fmt.Errorf("任务ID必须大于0: %d", id)
		}

		// 添加到任务ID列表
		taskIDs = append(taskIDs, id)
	}

	return taskIDs, nil
}

// hasAnyUpdateFlags 检查是否有任何更新标志被设置
//
// 返回值:
//   - bool: 如果有任何一个更新标志被设置，则返回 true，否则返回 false
func hasAnyUpdateFlags() bool {
	return retainCountF.Get() != -1 ||
		retainDaysF.Get() != -1 ||
		compressF.Get() != "" ||
		len(includeF.Get()) > 0 ||
		len(excludeF.Get()) > 0 ||
		clearIncludeF.Get() ||
		clearExcludeF.Get() ||
		maxSizeF.Get() != -1 ||
		minSizeF.Get() != -1
}

// updateTask 更新单个任务
//
// 参数:
//   - db: 数据库连接
//   - taskID: 要更新的任务ID
//
// 返回值:
//   - error: 更新失败时返回错误信息，否则返回 nil
func updateTask(db *sqlx.DB, taskID int64) error {
	// 获取当前任务信息
	currentTask, err := DB.GetTaskByID(db, taskID)
	if err != nil {
		return fmt.Errorf("获取任务信息失败: %w", err)
	}

	// 准备更新的值，如果用户指定了新值且与当前值不同，则使用新值，否则使用当前值
	// 保留数量
	newRetainCount := currentTask.RetainCount
	if retainCount := retainCountF.Get(); retainCount != -1 && retainCount != currentTask.RetainCount {
		newRetainCount = retainCount
	}

	// 保留天数
	newRetainDays := currentTask.RetainDays
	if retainDays := retainDaysF.Get(); retainDays != -1 && retainDays != currentTask.RetainDays {
		newRetainDays = retainDays
	}

	// 是否压缩
	newCompress := currentTask.Compress
	if compressStr := compressF.Get(); compressStr != "" {
		if compressStr == "true" {
			newCompress = true
		} else if compressStr == "false" {
			newCompress = false
		}
	}

	// 包含规则
	newIncludeRules := currentTask.IncludeRules
	if includeRules := includeF.Get(); len(includeRules) > 0 {
		if includeJSON, err := utils.MarshalRules(includeRules); err != nil {
			fmt.Printf("警告: 包含规则编码失败: %v\n", err)
		} else if includeJSON != currentTask.IncludeRules {
			newIncludeRules = includeJSON
		}
	}

	// 排除规则
	newExcludeRules := currentTask.ExcludeRules
	if excludeRules := excludeF.Get(); len(excludeRules) > 0 {
		if excludeJSON, err := utils.MarshalRules(excludeRules); err != nil {
			fmt.Printf("警告: 排除规则编码失败: %v\n", err)
		} else if excludeJSON != currentTask.ExcludeRules {
			newExcludeRules = excludeJSON
		}
	}

	// 最大文件大小
	newMaxFileSize := currentTask.MaxFileSize
	if maxSize := maxSizeF.Get(); maxSize != -1 && maxSize != currentTask.MaxFileSize {
		newMaxFileSize = maxSize
	}

	// 最小文件大小
	newMinFileSize := currentTask.MinFileSize
	if minSize := minSizeF.Get(); minSize != -1 && minSize != currentTask.MinFileSize {
		newMinFileSize = minSize
	}

	// 固定的SQL更新语句
	sql := `UPDATE backup_tasks SET 
		retain_count = ?, 
		retain_days = ?, 
		compress = ?, 
		include_rules = ?, 
		exclude_rules = ?, 
		max_file_size = ?, 
		min_file_size = ?, 
		updated_at = CURRENT_TIMESTAMP 
		WHERE ID = ?`

	// 执行更新
	result, err := db.Exec(sql,
		newRetainCount,
		newRetainDays,
		newCompress,
		newIncludeRules,
		newExcludeRules,
		newMaxFileSize,
		newMinFileSize,
		taskID)

	if err != nil {
		return fmt.Errorf("执行更新失败: %w", err)
	}

	// 检查是否有行被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("任务 %s (ID: %d) 未被更新", currentTask.Name, taskID)
	}

	return nil
}
