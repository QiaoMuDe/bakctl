package run

import (
	"fmt"
	"strconv"

	DB "gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/types"
	"github.com/jmoiron/sqlx"
)

// RunCmdMain run命令的主函数
func RunCmdMain(db *sqlx.DB) error {
	// 1. 参数校验
	if err := validateFlags(); err != nil {
		return fmt.Errorf("参数错误: %w", err)
	}

	// 2. 选择要执行的任务
	tasks, err := selectTasks(db)
	if err != nil {
		return fmt.Errorf("任务选择失败: %w", err)
	}

	// 3. 显示选中的任务信息
	fmt.Printf("找到 %d 个任务:\n", len(tasks))
	for i, task := range tasks {
		fmt.Printf("  %d. %s (ID: %d) - %s\n", i+1, task.Name, task.ID, task.BackupDir)
	}

	// TODO: 实现任务执行逻辑
	fmt.Println("\n任务执行功能正在开发中...")

	return nil
}

// selectTasks 根据标志选择要执行的任务
//
// 参数：
//   - db：数据库连接对象
//
// 返回值：
//   - []types.BackupTask：选中的任务列表
//   - error：如果查询过程中发生错误，则返回非 nil 错误信息
func selectTasks(db *sqlx.DB) ([]types.BackupTask, error) {
	// 根据单个任务ID查询
	if taskIDFlag.Get() > 0 {
		task, err := DB.GetTaskByID(db, taskIDFlag.Get())
		if err != nil {
			return nil, fmt.Errorf("获取任务ID %d 失败: %w", taskIDFlag.Get(), err)
		}
		return []types.BackupTask{*task}, nil
	}

	// 根据多个任务ID批量查询
	if len(taskIDsFlag.Get()) > 0 {
		tasks, err := DB.GetTasksByIDs(db, taskIDsFlag.Get())
		if err != nil {
			return nil, fmt.Errorf("批量获取任务失败: %w", err)
		}

		if len(tasks) == 0 {
			return nil, fmt.Errorf("系统中没有配置任何备份任务")
		}

		return tasks, nil
	}

	// 查询所有任务
	if allTasksFlag.Get() {
		tasks, err := DB.GetAllTasks(db)
		if err != nil {
			return nil, fmt.Errorf("获取所有任务失败: %w", err)
		}

		if len(tasks) == 0 {
			return nil, fmt.Errorf("系统中没有配置任何备份任务")
		}

		return tasks, nil
	}

	// 这种情况不应该发生，因为validateFlags已经检查过
	return nil, fmt.Errorf("内部错误: 未知的任务选择方式")
}

// validateFlags 检查三个标志的互斥性和参数有效性
//
// 返回值：
//   - error: 如果参数无效或冲突, 返回错误信息；否则返回nil
func validateFlags() error {
	// 统计有效参数的数量
	paramCount := 0

	// 检查单个任务ID
	if taskIDFlag.Get() > 0 {
		paramCount++
		// 校验任务ID必须为正数
		if taskIDFlag.Get() <= 0 {
			return fmt.Errorf("任务ID必须为正数, 当前值: %d", taskIDFlag.Get())
		}
	}

	// 检查多个任务ID
	if len(taskIDsFlag.Get()) > 0 {
		paramCount++
		// 校验每个ID都是有效的正整数
		for i, idStr := range taskIDsFlag.Get() {
			if idStr == "" {
				return fmt.Errorf("第%d个任务ID不能为空", i+1)
			}

			id, err := strconv.Atoi(idStr)
			if err != nil {
				return fmt.Errorf("第%d个任务ID格式无效: %s (必须为整数)", i+1, idStr)
			}

			if id <= 0 {
				return fmt.Errorf("第%d个任务ID必须为正数: %d", i+1, id)
			}
		}

		// 检查是否有重复的ID
		idSet := make(map[string]bool)
		for _, idStr := range taskIDsFlag.Get() {
			if idSet[idStr] {
				return fmt.Errorf("任务ID重复: %s", idStr)
			}
			idSet[idStr] = true
		}
	}

	// 检查运行所有任务标志
	if allTasksFlag.Get() {
		paramCount++
	}

	// 互斥性检查
	if paramCount == 0 {
		return fmt.Errorf("请指定要运行的任务: -id <任务ID> 或 -ids <任务ID列表> 或 -all")
	}

	if paramCount > 1 {
		return fmt.Errorf("不能同时指定多个任务选择参数, 请只使用其中一个: -id, -ids, -all")
	}

	return nil
}
