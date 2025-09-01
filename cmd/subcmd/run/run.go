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

	// 4. 执行选中的任务
	if err := executeTasks(tasks); err != nil {
		return fmt.Errorf("任务执行失败: %w", err)
	}

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

// executeTask 执行单个备份任务
//
// 参数：
//   - task：要执行的备份任务
//
// 返回值：
//   - error：如果执行过程中发生错误，则返回非 nil 错误信息；成功则返回 nil
func executeTask(task types.BackupTask) error {
	// TODO: 实现单个任务的执行逻辑
	// 这里将包含：
	// 1. 检查源目录是否存在
	// 2. 创建备份目录（如果不存在）
	// 3. 执行备份操作（复制文件）
	// 4. 记录执行日志
	// 5. 更新任务状态
	return nil
}

// executeTasks 批量执行备份任务
//
// 参数：
//   - tasks：要执行的备份任务切片
//
// 返回值：
//   - error：如果执行过程中发生错误，则返回非 nil 错误信息；全部成功则返回 nil
func executeTasks(tasks []types.BackupTask) error {
	successCount := 0
	failureCount := 0
	
	fmt.Printf("\n开始执行 %d 个备份任务...\n", len(tasks))
	
	for i, task := range tasks {
		fmt.Printf("\n[%d/%d] 正在执行任务: %s (ID: %d)\n", i+1, len(tasks), task.Name, task.ID)
		
		if err := executeTask(task); err != nil {
			fmt.Printf("❌ 任务执行失败: %v\n", err)
			failureCount++
		} else {
			fmt.Printf("✅ 任务执行成功\n")
			successCount++
		}
	}
	
	// 显示执行结果统计
	fmt.Printf("\n执行完成！成功: %d, 失败: %d\n", successCount, failureCount)
	
	if failureCount > 0 {
		return fmt.Errorf("有 %d 个任务执行失败", failureCount)
	}
	
	return nil
}