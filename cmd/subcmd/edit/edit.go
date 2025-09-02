package edit

import (
	"fmt"
	"strconv"

	DB "gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/bakctl/internal/utils"
	"gitee.com/MM-Q/colorlib"
	"github.com/jmoiron/sqlx"
)

// EditCmdMain 编辑命令主函数
//
// 参数:
//   - db: 数据库连接
//   - cl: 颜色库
//
// 返回值:
//   - error: 执行失败时返回错误信息，否则返回 nil
func EditCmdMain(db *sqlx.DB, cl *colorlib.ColorLib) error {
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
			cl.Redf("更新任务ID %d 失败: %v\n", taskID, err)
		} else {
			successCount++
			cl.Greenf("任务ID %d 更新成功\n", taskID)
		}
	}

	cl.Greenf("成功更新 %d/%d 个任务\n", successCount, len(taskIDs))
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
		return fmt.Errorf("任务ID %d 不存在或获取任务信息失败: %w", taskID, err)
	}

	// 准备更新的值，如果用户指定了新值且与当前值不同，则使用新值，否则使用当前值
	// 备份保留数量
	newRetainCount := updateInt(currentTask.RetainCount, retainCountF.Get(), -1)

	// 备份保留天数
	newRetainDays := updateInt(currentTask.RetainDays, retainDaysF.Get(), -1)

	// 最大文件大小
	newMaxFileSize := updateInt64(currentTask.MaxFileSize, maxSizeF.Get(), -1)

	// 最小文件大小
	newMinFileSize := updateInt64(currentTask.MinFileSize, minSizeF.Get(), -1)

	// 备份是否压缩
	newCompress, err := updateBooleanFromFlag(currentTask.Compress, compressF.Get, "压缩参数")
	if err != nil {
		return err // 如果解析失败，直接返回错误
	}

	// 包含规则
	newIncludeRules, includrErr := updateRuleString(currentTask.IncludeRules, includeF.Get(), "包含规则", clearIncludeF.Get())
	if includrErr != nil {
		return includrErr // 如果解析失败，直接返回错误
	}

	// 排除规则
	newExcludeRules, excludeErr := updateRuleString(currentTask.ExcludeRules, excludeF.Get(), "排除规则", clearExcludeF.Get())
	if excludeErr != nil {
		return excludeErr // 如果解析失败，直接返回错误
	}

	// 创建 UpdateTaskParams 结构体实例
	params := types.UpdateTaskParams{
		ID:           taskID,
		RetainCount:  newRetainCount,
		RetainDays:   newRetainDays,
		Compress:     newCompress,
		IncludeRules: newIncludeRules,
		ExcludeRules: newExcludeRules,
		MaxFileSize:  newMaxFileSize,
		MinFileSize:  newMinFileSize,
	}

	// 调用 db 包中的 UpdateTask 函数，传入结构体
	err = DB.UpdateTask(db, params)
	if err != nil {
		return fmt.Errorf("更新任务失败: %w", err)
	}

	return nil
}

// updateRuleString 辅助函数，用于更新规则字符串
//
// 参数:
//   - currentRuleStr: 当前规则字符串
//   - newRules: 新规则切片
//   - ruleType: 规则类型 (用于错误提示)
//   - clearFlag: 一个布尔值，如果为 true，表示要清空规则
//
// 返回值:
//   - string: 更新后的规则字符串
//   - error: 如果解析规则字符串失败，则返回错误信息，否则返回 nil
func updateRuleString(currentRuleStr string, newRules []string, ruleType string, clearFlag bool) (string, error) {
	// 如果 clearFlag 为 true，则直接返回空的 JSON 数组字符串，表示清空
	if clearFlag {
		return "[]", nil
	}

	// 如果没有新规则，则返回当前规则，不进行更新
	if len(newRules) == 0 {
		return currentRuleStr, nil
	}

	marshaledRules, err := utils.MarshalRules(newRules)
	if err != nil {
		return "", fmt.Errorf("无法解析%s: %v", ruleType, err)
	}

	if marshaledRules != currentRuleStr {
		return marshaledRules, nil // 新规则与当前规则不同，返回新规则
	}
	return currentRuleStr, nil // 新规则与当前规则相同，返回当前规则
}

// updateInt 辅助函数，用于更新 int 类型的值
//
// 参数:
//   - currentVal: 当前任务中的原始 int 值
//   - newVal: 从命令行参数或配置中获取的新 int 值
//   - unsetVal: 表示新值未设置或无效的特殊 int 值（例如 -1）
//
// 返回值:
//   - int: 更新后的 int 值
func updateInt(currentVal, newVal, unsetVal int) int {
	if newVal != unsetVal && newVal != currentVal {
		return newVal
	}
	return currentVal
}

// updateBooleanFromFlag 辅助函数，用于根据命令行或配置标志更新布尔值
//
// 参数:
//   - currentValue: 当前任务中的原始布尔值
//   - flagGetter: 一个函数，用于获取标志的字符串值（例如 compressF.Get）
//   - paramName: 参数的名称，用于错误信息（例如 "压缩参数"）
//
// 返回值:
//   - bool: 更新后的布尔值
//   - error: 解析失败时返回错误信息，否则返回 nil
func updateBooleanFromFlag(currentValue bool, flagGetter func() string, paramName string) (bool, error) {
	flagValue := flagGetter() // 获取标志的字符串值

	if flagValue == "" {
		return currentValue, nil // 如果标志值为空，表示用户未指定新值，则返回当前值
	}

	// 尝试将字符串解析为布尔值
	b, err := strconv.ParseBool(flagValue)
	if err != nil {
		// 如果解析失败，返回错误信息
		return currentValue, fmt.Errorf("无法解析%s: %s", paramName, flagValue)
	}

	return b, nil // 解析成功，返回新的布尔值
}

// updateInt64 辅助函数，用于更新 int64 类型的值
//
// 参数:
//   - currentVal: 当前任务中的原始 int64 值
//   - newVal: 从命令行参数或配置中获取的新 int64 值
//   - unsetVal: 表示新值未设置或无效的特殊 int64 值（例如 -1）
//
// 返回值:
//   - int64: 更新后的 int64 值
func updateInt64(currentVal, newVal, unsetVal int64) int64 {
	if newVal != unsetVal && newVal != currentVal {
		return newVal
	}
	return currentVal
}
