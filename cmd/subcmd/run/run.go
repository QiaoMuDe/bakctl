package run

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gitee.com/MM-Q/bakctl/internal/cleanup"
	DB "gitee.com/MM-Q/bakctl/internal/db"
	baktypes "gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/bakctl/internal/utils"
	"gitee.com/MM-Q/colorlib"
	"gitee.com/MM-Q/comprx"
	"gitee.com/MM-Q/comprx/types"
	"gitee.com/MM-Q/go-kit/hash"
	"gitee.com/MM-Q/go-kit/id"
	"github.com/jmoiron/sqlx"
)

// RunCmdMain run命令的主函数
//
// 参数:
//   - db: 数据库连接对象
//   - cl: 颜色库对象
//
// 返回值:
//   - error: 如果执行过程中发生错误，则返回非 nil 错误信息；成功则返回 nil
func RunCmdMain(db *sqlx.DB, cl *colorlib.ColorLib) error {
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
	cl.Bluef("找到 %d 个任务:\n", len(tasks))
	for i, task := range tasks {
		cl.Whitef("  %d. %s (ID: %d) - %s\n", i+1, task.Name, task.ID, task.BackupDir)
	}

	// 4. 执行选中的任务
	if err := executeTasks(tasks, db, cl); err != nil {
		return fmt.Errorf("任务执行失败: %w", err)
	}

	return nil
}

// executeTask 执行单个备份任务
//
// 参数：
//   - task：要执行的备份任务
//   - db：数据库连接对象
//   - cl: 颜色库对象
//
// 返回值：
//   - error：如果执行过程中发生错误，则返回非 nil 错误信息；成功则返回 nil
func executeTask(task baktypes.BackupTask, db *sqlx.DB, cl *colorlib.ColorLib) error {
	// 初始化结果结构体
	result := &baktypes.BackupResult{
		Success:    false,                    // 备份是否成功
		BackupPath: generateBackupPath(task), // 备份文件路径
	}

	// 使用defer确保无论成功失败都记录到数据库
	defer func() {
		if recordErr := recordBackupResult(db, task, result); recordErr != nil {
			// 记录失败的处理（可以记录日志等）
			cl.Redf("记录备份结果失败: %v\n", recordErr)
		}
	}()

	// 1. 验证源目录
	if err := validateSourceDir(task.BackupDir); err != nil {
		result.ErrorMsg = err.Error()
		return err
	}

	// 2. 解析过滤规则
	include, exclude, err := parseFilterRules(task.IncludeRules, task.ExcludeRules)
	if err != nil {
		result.ErrorMsg = err.Error()
		return err
	}

	// 3. 构建过滤器
	filters := types.FilterOptions{
		Include: include,          // 包含规则
		Exclude: exclude,          // 排除规则
		MinSize: task.MinFileSize, // 最小文件大小
		MaxSize: task.MaxFileSize, // 最大文件大小
	}

	// 4. 设置压缩等级
	level := types.CompressionLevelNone // 默认不压缩
	if task.Compress {
		level = types.CompressionLevelDefault // 使用默认压缩等级
	}

	// 5. 构建压缩配置
	opts := comprx.Options{
		CompressionLevel:      level,                    // 压缩等级
		OverwriteExisting:     false,                    // 覆盖已存在的文件
		ProgressEnabled:       true,                     // 显示进度条
		ProgressStyle:         types.ProgressStyleASCII, // 进度条样式
		DisablePathValidation: false,                    // 禁用路径验证
		Filter:                filters,                  // 过滤器
	}

	// 6. 执行备份操作
	if err := comprx.PackOptions(result.BackupPath, task.BackupDir, opts); err != nil {
		result.ErrorMsg = fmt.Sprintf("备份操作失败: %v", err)
		return err
	}

	// 7. 收集备份文件信息
	size, checksum, err := collectBackupInfo(result.BackupPath)
	if err != nil {
		result.ErrorMsg = err.Error()
		result.FileSize = size // 即使哈希失败也记录文件大小
		return err
	}

	// 8. 设置成功结果
	result.Success = true      // 备份成功
	result.FileSize = size     // 备份文件大小
	result.Checksum = checksum // 备份文件哈希值

	// 9. 清理历史备份
	cl.White("  → 清理历史备份...")
	taskAdapter := cleanup.NewBackupTaskAdapter(
		task.ID, task.Name, task.StorageDir,
		task.RetainCount, task.RetainDays,
	)
	if err := cleanup.CleanupBackupFilesWithLogging(taskAdapter, baktypes.BackupFileExt, cl); err != nil {
		cl.Yellowf("  → 清理警告: %v\n", err)
	}

	return nil
}

// executeTasks 批量执行备份任务
//
// 参数：
//   - tasks：要执行的备份任务切片
//   - db：数据库连接对象
//   - cl: 颜色库对象
//
// 返回值：
//   - error：如果执行过程中发生错误，则返回非 nil 错误信息；全部成功则返回 nil
func executeTasks(tasks []baktypes.BackupTask, db *sqlx.DB, cl *colorlib.ColorLib) error {
	successCount := 0 // 成功数量
	failureCount := 0 // 失败数量

	cl.Bluef("开始执行 %d 个备份任务...\n", len(tasks))

	for i, task := range tasks {
		cl.Whitef("[%d/%d] 正在执行任务: %s (ID: %d)\n", i+1, len(tasks), task.Name, task.ID)

		if err := executeTask(task, db, cl); err != nil {
			cl.Redf("任务执行失败: %v\n", err)
			failureCount++
		} else {
			cl.Green("任务执行成功")
			successCount++
		}
	}

	// 显示执行结果统计
	cl.Greenf("执行完成！成功: %d, 失败: %d\n", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("有 %d 个任务执行失败", failureCount)
	}

	return nil
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

// validateSourceDir 验证源目录是否存在
//
// 参数：
//   - dir：源目录路径
//
// 返回值：
//   - error：如果源目录不存在，则返回错误信息；否则返回 nil
func validateSourceDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("源目录不存在: %s", dir)
	}
	return nil
}

// parseFilterRules 解析包含和排除规则
//
// 参数:
//   - includeRules：包含规则字符串
//   - excludeRules：排除规则字符串
//
// 返回值:
//   - []string：包含规则列表
//   - []string：排除规则列表
//   - error：如果发生错误，则返回错误信息；否则返回 nil
func parseFilterRules(includeRules, excludeRules string) ([]string, []string, error) {
	include, err := utils.UnmarshalRules(includeRules)
	if err != nil {
		return nil, nil, fmt.Errorf("解析包含规则失败: %w", err)
	}

	exclude, err := utils.UnmarshalRules(excludeRules)
	if err != nil {
		return nil, nil, fmt.Errorf("解析排除规则失败: %w", err)
	}

	return include, exclude, nil
}

// generateBackupPath 生成备份文件路径
//
// 参数：
//   - task：要执行的备份任务
//
// 返回值：
//   - string：生成的备份文件路径
func generateBackupPath(task baktypes.BackupTask) string {
	// 使用时间字符串格式：YYYYMMDD_HHMMSS
	timeStr := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s%s", task.Name, timeStr, baktypes.BackupFileExt)
	return filepath.Join(task.StorageDir, filename)
}

// collectBackupInfo 收集备份文件信息（大小和哈希）
//
// 参数：
//   - filePath：备份文件路径
//
// 返回值：
//   - int64：文件大小
//   - string：文件哈希值
//   - error：如果发生错误，则返回错误信息；否则返回 nil
func collectBackupInfo(filePath string) (int64, string, error) {
	// 获取文件大小
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, "", fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 计算哈希值
	checksum, err := hash.ChecksumProgress(filePath, baktypes.HashAlgorithm)
	if err != nil {
		return info.Size(), "", fmt.Errorf("计算哈希失败: %w", err)
	}

	return info.Size(), checksum, nil
}

// recordBackupResult 统一记录备份结果（成功或失败）
//
// 参数：
//   - db：数据库连接对象
//   - task：要执行的备份任务
//   - result：备份执行结果
//
// 返回值：
//   - error：如果记录失败，则返回错误信息；成功则返回 nil
func recordBackupResult(db *sqlx.DB, task baktypes.BackupTask, result *baktypes.BackupResult) error {
	rec := baktypes.BackupRecord{
		TaskID:         task.ID,                          // 任务ID
		TaskName:       task.Name,                        // 任务名称
		VersionID:      id.GenMaskedID(),                 // 版本ID
		BackupFilename: filepath.Base(result.BackupPath), // 存储路径
		BackupSize:     result.FileSize,                  // 文件大小
		StoragePath:    result.BackupPath,                // 存储路径
		Status:         result.Success,                   // 状态
		FailureMessage: result.ErrorMsg,                  // 失败原因
		Checksum:       result.Checksum,                  // 校验码
	}

	return DB.InsertBackupRecord(db, &rec)
}

// selectTasks 根据标志选择要执行的任务
//
// 参数：
//   - db：数据库连接对象
//
// 返回值：
//   - []types.BackupTask：选中的任务列表
//   - error：如果查询过程中发生错误，则返回非 nil 错误信息
func selectTasks(db *sqlx.DB) ([]baktypes.BackupTask, error) {
	// 根据单个任务ID查询
	if taskIDFlag.Get() > 0 {
		task, err := DB.GetTaskByID(db, taskIDFlag.Get())
		if err != nil {
			return nil, fmt.Errorf("获取任务ID %d 失败: %w", taskIDFlag.Get(), err)
		}
		return []baktypes.BackupTask{*task}, nil
	}

	// 将字符串切片转换为整数切片
	taskIDs, err := utils.StringSliceToInt64(taskIDsFlag.Get())
	if err != nil {
		return nil, fmt.Errorf("转换任务ID列表失败: %w", err)
	}

	// 根据多个任务ID批量查询
	if len(taskIDsFlag.Get()) > 0 {
		tasks, err := DB.GetTasksByIDs(db, taskIDs)
		if err != nil {
			return nil, fmt.Errorf("批量获取任务失败: %w", err)
		}

		// 检查任务列表是否为空
		if len(tasks) == 0 {
			return nil, fmt.Errorf("没有找到指定的任务ID")
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
