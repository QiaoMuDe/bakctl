package run

import (
	"fmt"

	baktypes "gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/comprx"
	"gitee.com/MM-Q/comprx/types"
	"github.com/jmoiron/sqlx"
)

// 常量定义
const (
	HashAlgorithm = "sha1"
	BackupFileExt = ".zip"
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
	if err := executeTasks(tasks, db); err != nil {
		return fmt.Errorf("任务执行失败: %w", err)
	}

	return nil
}

// executeTask 执行单个备份任务
//
// 参数：
//   - task：要执行的备份任务
//   - db：数据库连接对象
//
// 返回值：
//   - error：如果执行过程中发生错误，则返回非 nil 错误信息；成功则返回 nil
func executeTask(task baktypes.BackupTask, db *sqlx.DB) error {
	// 初始化结果结构体
	result := &baktypes.BackupResult{
		Success:    false,                    // 备份是否成功
		BackupPath: generateBackupPath(task), // 备份文件路径
	}

	// 使用defer确保无论成功失败都记录到数据库
	defer func() {
		if recordErr := recordBackupResult(db, task, result); recordErr != nil {
			// 记录失败的处理（可以记录日志等）
			fmt.Printf("记录备份结果失败: %v\n", recordErr)
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
		OverwriteExisting:     true,                     // 覆盖已存在的文件
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

	// 9. 预留的清理历史备份环境
	// ...

	return nil
}

// executeTasks 批量执行备份任务
//
// 参数：
//   - tasks：要执行的备份任务切片
//   - db：数据库连接对象
//
// 返回值：
//   - error：如果执行过程中发生错误，则返回非 nil 错误信息；全部成功则返回 nil
func executeTasks(tasks []baktypes.BackupTask, db *sqlx.DB) error {
	successCount := 0 // 成功数量
	failureCount := 0 // 失败数量

	fmt.Printf("开始执行 %d 个备份任务...\n", len(tasks))

	for i, task := range tasks {
		fmt.Printf("[%d/%d] 正在执行任务: %s (ID: %d)\n", i+1, len(tasks), task.Name, task.ID)

		if err := executeTask(task, db); err != nil {
			fmt.Printf("❌ 任务执行失败: %v\n", err)
			failureCount++
		} else {
			fmt.Print("✅ 任务执行成功\n")
			successCount++
		}
	}

	// 显示执行结果统计
	fmt.Printf("执行完成！成功: %d, 失败: %d\n", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("有 %d 个任务执行失败", failureCount)
	}

	return nil
}
