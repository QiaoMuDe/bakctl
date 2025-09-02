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

// RestoreCmdMain restore命令的主函数
//
// 参数：
//   - database: 数据库连接
//   - cl: colorlib.ColorLib 实例
//
// 返回:
//   - error: 错误信息，如果没有错误则返回nil
func RestoreCmdMain(database *sqlx.DB, cl *colorlib.ColorLib) error {
	startTime := time.Now()

	// 1. 检查 -id 和 -vid 是否为空，必须都不为空
	taskID := taskIDFlag.Get()
	versionID := versionIDFlag.Get()
	targetDir := targetDirFlag.Get()
	if taskID <= 0 {
		return fmt.Errorf("任务ID必须大于0, 请使用 -id 指定")
	}

	if versionID == "" {
		return fmt.Errorf("版本ID不能为空, 请使用 -vid 指定")
	}

	cl.Blue("开始恢复备份...")
	cl.Bluef("任务ID: %d\n", taskID)
	cl.Bluef("版本ID: %s\n", versionID)
	cl.Bluef("目标目录: %s\n", targetDir)

	// 2. 检查指定的任务ID是否存在
	cl.White("  → 检查任务是否存在...")
	if !DB.TaskExists(database, int64(taskID)) {
		return fmt.Errorf("任务ID %d 不存在", taskID)
	}

	// 3. 检查指定的vid是否存在并且是这个任务ID的
	cl.White("  → 检查备份记录...")
	record, err := DB.GetBackupRecordByTaskAndVersion(database, int64(taskID), versionID)
	if err != nil {
		return err
	}

	// 4. 检查备份文件是否存在
	cl.White("  → 验证备份文件...")
	if _, err := os.Stat(record.StoragePath); os.IsNotExist(err) {
		return fmt.Errorf("备份文件不存在: %s", record.StoragePath)
	}

	// 5. 验证备份文件校验值
	if record.Checksum != "" {
		cl.White("  → 验证文件完整性...")
		actualChecksum, err := hash.ChecksumProgress(record.StoragePath, baktypes.HashAlgorithm)
		if err != nil {
			return fmt.Errorf("计算备份文件校验值失败: %w", err)
		}
		if actualChecksum != record.Checksum {
			return fmt.Errorf("备份文件校验失败，文件可能已损坏或被篡改\n期望: %s\n实际: %s", record.Checksum, actualChecksum)
		}
		cl.Green("    文件完整性验证通过  ✔")
	} else {
		cl.Yellow("    该备份文件没有校验值记录，无法验证完整性  ✘")
	}

	// 6. 创建目标目录
	cl.White("  → 准备目标目录...")
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("无法获取目标目录的绝对路径: %w", err)
	}

	// 7. 根据任务记录里的备份存储地址，解压到-d指定的路径
	cl.White("  → 正在恢复备份文件...")
	if err := extractBackupFile(record.StoragePath, absTargetDir); err != nil {
		return fmt.Errorf("恢复失败: %w", err)
	}

	// 8. 显示结果
	duration := time.Since(startTime)
	cl.Green("恢复完成!")
	cl.Whitef("耗时: %v\n", duration)
	cl.Whitef("备份文件: %s\n", record.StoragePath)
	cl.Whitef("目标目录: %s\n", absTargetDir)

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
		ProgressStyle:         types.ProgressStyleASCII,      // 进度条样式
		DisablePathValidation: false,                         // 禁用路径验证
	}

	// 执行解压操作
	if err := comprx.UnpackOptions(backupPath, targetDir, opts); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	return nil
}
