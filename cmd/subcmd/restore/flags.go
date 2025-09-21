// Package restore 实现了 bakctl 的 restore 子命令的命令行参数解析功能。
//
// 该文件定义了 restore 命令支持的所有命令行标志和参数，包括：
//   - 备份文件路径选项
//   - 恢复目标目录选项
//   - 恢复模式选项（完整/增量）
//   - 文件过滤和排除选项
//   - 恢复验证选项
//   - 覆盖策略选项
//
// 通过这些参数，用户可以精确控制恢复操作的行为和目标。
package restore

import (
	"flag"

	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	restoreCmd *qflag.Cmd // restore命令

	// 必需参数
	taskIDFlag    *qflag.IntFlag    // 任务ID
	versionIDFlag *qflag.StringFlag // 版本ID
	latestFlag    *qflag.BoolFlag   // 恢复最新备份标志

	// 可选参数
	targetDirFlag *qflag.StringFlag // 目标目录
)

// InitRestoreCmd 初始化restore子命令
func InitRestoreCmd() *qflag.Cmd {
	restoreCmd = cmd.NewCmd("restore", "rs", flag.ExitOnError)
	restoreCmd.SetDescription("恢复备份文件")
	restoreCmd.SetUseChinese(true)

	// 必需参数
	taskIDFlag = restoreCmd.Int("", "id", 0, "指定要恢复的备份任务ID")
	versionIDFlag = restoreCmd.String("", "vid", "", "指定要恢复的备份版本ID (与--latest/-l互斥)")
	latestFlag = restoreCmd.Bool("latest", "l", false, "恢复最新的备份 (与-vid互斥)")

	// 可选参数
	targetDirFlag = restoreCmd.String("", "d", ".", "指定恢复到的目标目录 (默认为当前目录)")

	return restoreCmd
}
