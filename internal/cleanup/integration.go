package cleanup

import (
	"fmt"

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

// CleanupBackupFilesWithLogging 带日志输出的清理函数
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

	// 如果两个保留策略都为0，跳过清理
	if task.GetRetainCount() <= 0 && task.GetRetainDays() <= 0 {
		cl.White("  → 跳过清理 (未设置保留策略)")
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

	// 输出清理结果
	if result.DeletedFiles > 0 {
		cl.Whitef("  → %s\n", FormatCleanupResult(result))
	} else {
		cl.White("  → 无需清理历史备份")
	}

	// 如果有删除失败的文件，输出警告
	if len(result.ErrorFiles) > 0 {
		cl.Yellowf("  → 警告: %d 个文件删除失败\n", len(result.ErrorFiles))
		for _, errorFile := range result.ErrorFiles {
			cl.Yellowf("    - %s\n", errorFile)
		}
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

	// 按时间戳降序排序
	sortBackupFilesByTimestamp(backupFiles)

	// 确定需要删除的文件
	filesToDelete := determineFilesToDelete(backupFiles, task.GetRetainCount(), task.GetRetainDays())

	return filesToDelete, nil
}

// sortBackupFilesByTimestamp 按时间戳降序排序备份文件
func sortBackupFilesByTimestamp(backupFiles []BackupFileInfo) {
	for i := 0; i < len(backupFiles)-1; i++ {
		for j := i + 1; j < len(backupFiles); j++ {
			if backupFiles[i].Timestamp < backupFiles[j].Timestamp {
				backupFiles[i], backupFiles[j] = backupFiles[j], backupFiles[i]
			}
		}
	}
}
