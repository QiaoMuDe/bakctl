// Package cleanup 提供备份文件清理的集成功能。
//
// 该文件提供了清理功能的高级封装，包括带日志输出的清理函数、
// 清理预览功能等。主要用于与其他模块的集成，提供用户友好的
// 清理操作接口。
//
// 主要功能：
//   - CleanupBackupFilesWithLogging: 执行清理并输出彩色日志
//   - GetCleanupPreview: 预览将要删除的文件（不实际删除）
//   - 清理结果格式化和错误处理
//   - 参数验证和安全检查
//
// 该文件作为cleanup包与外部调用者之间的桥梁，
// 提供了完整的清理工作流程和用户交互体验。
package cleanup

import (
	"fmt"
	"sort"

	"gitee.com/MM-Q/colorlib"
)

// BackupTask 备份任务接口，避免循环依赖
type BackupTask interface {
	GetID() int64
	GetName() string
	GetStorageDir() string
	GetRetainCount() int
	GetRetainDays() int
}

// CleanupBackupFilesWithLogging 静默清理历史备份文件
//
// 参数:
//   - task: 备份任务对象
//   - backupFileExt: 备份文件扩展名
//   - cl: 颜色库对象
//
// 返回值:
//   - error: 清理过程中的错误
func CleanupBackupFilesWithLogging(task BackupTask, backupFileExt string, cl *colorlib.ColorLib) error {
	// 验证参数
	if err := ValidateCleanupParams(task.GetStorageDir(), task.GetName(), task.GetRetainCount(), task.GetRetainDays()); err != nil {
		return fmt.Errorf("清理参数验证失败: %w", err)
	}

	// 如果两个保留策略都为0，静默跳过清理
	if task.GetRetainCount() <= 0 && task.GetRetainDays() <= 0 {
		return nil
	}

	// 执行清理
	result, err := CleanupBackupFiles(
		task.GetStorageDir(),
		task.GetName(),
		task.GetRetainCount(),
		task.GetRetainDays(),
		backupFileExt,
	)

	if err != nil {
		return fmt.Errorf("清理执行失败: %w", err)
	}

	// 如果有删除失败的文件，返回错误信息
	if len(result.ErrorFiles) > 0 {
		return fmt.Errorf("清理完成，但有 %d 个文件删除失败: %v", len(result.ErrorFiles), result.ErrorFiles)
	}

	return nil
}

// GetCleanupPreview 获取清理预览信息（不实际删除文件）
//
// 参数:
//   - task: 备份任务对象
//   - backupFileExt: 备份文件扩展名
//
// 返回值:
//   - []BackupFileInfo: 将要删除的文件信息列表
//   - error: 预览过程中的错误
func GetCleanupPreview(task BackupTask, backupFileExt string) ([]BackupFileInfo, error) {
	// 验证参数
	if err := ValidateCleanupParams(task.GetStorageDir(), task.GetName(), task.GetRetainCount(), task.GetRetainDays()); err != nil {
		return nil, fmt.Errorf("清理参数验证失败: %w", err)
	}

	// 如果两个保留策略都为0，返回空列表
	if task.GetRetainCount() <= 0 && task.GetRetainDays() <= 0 {
		return []BackupFileInfo{}, nil
	}

	// 收集备份文件信息
	backupFiles, err := collectBackupFiles(task.GetStorageDir(), task.GetName(), backupFileExt)
	if err != nil {
		return nil, fmt.Errorf("收集备份文件失败: %w", err)
	}

	if len(backupFiles) == 0 {
		return []BackupFileInfo{}, nil
	}

	// 按时间戳降序排序（最新的在前面）
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].CreatedTime.After(backupFiles[j].CreatedTime)
	})

	// 确定需要删除的文件
	filesToDelete := determineFilesToDelete(backupFiles, task.GetRetainCount(), task.GetRetainDays())

	return filesToDelete, nil
}
