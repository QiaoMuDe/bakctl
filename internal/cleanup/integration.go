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
// Package cleanup 实现了 bakctl 的备份文件清理集成功能。
//
// 该文件提供了清理功能与其他系统组件的集成接口，包括：
//   - 与数据库系统的集成
//   - 与文件系统操作的集成
//   - 与日志系统的集成
//   - 清理操作的统一入口点
//   - 错误处理和状态反馈
//
// 主要功能包括：
//   - 协调文件清理和数据库清理操作
//   - 提供统一的清理接口
//   - 处理清理过程中的异常情况
//   - 生成清理操作的详细报告
//   - 确保清理操作的原子性和一致性
//
// 该文件作为清理功能的集成层，连接底层清理逻辑和上层应用接口。
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
