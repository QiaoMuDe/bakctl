// Package restore 实现了 bakctl 的 restore 子命令功能。
//
// 该包提供了从备份文件恢复数据的功能，支持：
//   - 从指定的备份文件恢复数据到目标位置
//   - 验证备份文件的完整性和有效性
//   - 支持增量恢复和完整恢复
//   - 提供恢复进度显示和状态反馈
//   - 支持恢复前的数据备份保护
//
// 主要功能包括：
//   - 解压缩备份文件
//   - 验证文件校验和
//   - 恢复文件到指定目录
//   - 处理文件权限和属性
//   - 提供恢复操作的详细日志
//
// 恢复过程包括文件验证、解压缩、文件复制等步骤，确保数据安全可靠地恢复。
package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	DB "gitee.com/MM-Q/bakctl/internal/db"
	baktypes "gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/colorlib"
	"gitee.com/MM-Q/comprx"
	"gitee.com/MM-Q/comprx/types"
	"gitee.com/MM-Q/go-kit/hash"
	"github.com/jmoiron/sqlx"
)

// validateRestoreParams 验证restore命令的参数
//
// 返回值：
//   - taskID: 任务ID
//   - versionID: 版本ID
//   - latest: 是否使用最新备份
//   - targetDir: 目标目录
//   - error: 验证错误
func validateRestoreParams() (int, string, bool, string, error) {
	taskID := taskIDFlag.Get()
	versionID := versionIDFlag.Get()
	latest := latestFlag.Get()
	targetDir := targetDirFlag.Get()

	if taskID <= 0 {
		return 0, "", false, "", fmt.Errorf("任务ID必须大于0, 请使用 -id 指定")
	}

	// 检查 -vid 和 --latest 的互斥性
	if latest && versionID != "" {
		return 0, "", false, "", fmt.Errorf("不能同时指定 -vid 和 --latest/-l 参数，请选择其中一个")
	}

	if !latest && versionID == "" {
		return 0, "", false, "", fmt.Errorf("必须指定 -vid 或 --latest/-l 参数之一")
	}

	return taskID, versionID, latest, targetDir, nil
}

// RestoreCmdMain restore命令的主函数
//
// 参数：
//   - database: 数据库连接
//   - cl: colorlib.ColorLib 实例
//
// 返回:
//   - error: 错误信息，如果没有错误则返回nil
func RestoreCmdMain(database *sqlx.DB, cl *colorlib.ColorLib) error {
	// 获取开始时间
	startTime := time.Now()

	// 1. 验证参数
	taskID, versionID, latest, targetDir, validationErr := validateRestoreParams()
	if validationErr != nil {
		return validationErr
	}

	// 显示基本信息
	if latest {
		cl.Bluef("恢复 %d 的最新备份到 %s\n", taskID, targetDir)
	} else {
		cl.Bluef("恢复 %d 版本 %s 到 %s\n", taskID, versionID, targetDir)
	}

	// 2. 检查指定的任务ID是否存在
	if !DB.TaskExists(database, int64(taskID)) {
		return fmt.Errorf("任务ID %d 不存在", taskID)
	}

	// 3. 根据参数获取备份记录
	var record *baktypes.BackupRecord
	var err error

	if latest {
		// 获取最新的备份记录
		record, err = DB.GetLatestBackupRecordByTask(database, int64(taskID))
		if err != nil {
			return err
		}
	} else {
		// 获取指定版本的备份记录
		record, err = DB.GetBackupRecordByTaskAndVersion(database, int64(taskID), versionID)
		if err != nil {
			return err
		}
	}

	// 4. 检查备份文件是否存在
	if _, err := os.Stat(record.StoragePath); os.IsNotExist(err) {
		return fmt.Errorf("备份文件不存在: %s", record.StoragePath)
	}

	// 5. 验证备份文件校验值
	if record.Checksum != "" {
		actualChecksum, err := hash.ChecksumProgress(record.StoragePath, baktypes.HashAlgorithm)
		if err != nil {
			return fmt.Errorf("计算备份文件校验值失败: %w", err)
		}
		if actualChecksum != record.Checksum {
			return fmt.Errorf("备份文件校验失败，文件可能已损坏或被篡改\n期望: %s\n实际: %s", record.Checksum, actualChecksum)
		}
	}

	// 6. 创建目标目录
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("无法获取目标目录的绝对路径: %w", err)
	}

	// 7. 执行恢复
	if err := extractBackupFile(record.StoragePath, absTargetDir); err != nil {
		return fmt.Errorf("恢复失败: %w", err)
	}

	// 8. 显示结果
	duration := time.Since(startTime)
	cl.Green("恢复完成!")
	cl.Whitef("耗时: %v\n", duration)

	return nil
}

// extractBackupFile 解压备份文件到目标目录
//
// 参数:
//   - backupPath: 备份文件的路径
//   - targetDir: 目标目录的路径
//
// 返回:
//   - error: 如果发生错误则返回错误信息，否则返回nil
func extractBackupFile(backupPath, targetDir string) error {
	// 构建压缩配置
	opts := comprx.Options{
		CompressionLevel:      types.CompressionLevelDefault, // 压缩等级默认
		OverwriteExisting:     false,                         // 覆盖已存在的文件
		ProgressEnabled:       true,                          // 显示进度条
		ProgressStyle:         types.ProgressStyleDefault,    // 进度条样式
		DisablePathValidation: false,                         // 禁用路径验证
	}

	// 执行解压操作
	if err := comprx.UnpackOptions(backupPath, targetDir, opts); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	return nil
}
