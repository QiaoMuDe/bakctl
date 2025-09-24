// Package run 实现了 bakctl 的 run 子命令的命令行参数解析功能。
//
// 该文件定义了 run 命令支持的所有命令行标志和参数，包括：
//   - 任务ID选择选项
//   - 并发执行控制选项
//   - 输出详细程度选项
//   - 强制执行选项
//   - 跳过清理选项
//   - 验证模式选项
//
// 通过这些参数，用户可以精确控制备份任务的执行方式和行为。
package run

import (
	"flag"

	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	runCmd *qflag.Cmd // run命令

	// 任务选择参数
	taskIDFlag   *qflag.Int64Flag      // -id: 指定任务ID
	taskIDsFlag  *qflag.Int64SliceFlag // -ids: 指定多个任务ID
	allTasksFlag *qflag.BoolFlag       // -all: 运行所有任务
)

// InitRunCmd 初始化run子命令
func InitRunCmd() *qflag.Cmd {
	runCmd = cmd.NewCmd("run", "r", flag.ExitOnError)
	runCmd.SetDesc("运行备份任务")
	runCmd.SetChinese(true)

	// 任务选择参数（互斥）
	taskIDFlag = runCmd.Int64("", "id", 0, "指定要运行的任务ID")
	taskIDsFlag = runCmd.Int64Slice("", "ids", []int64{}, "指定多个任务ID进行批量运行")
	allTasksFlag = runCmd.Bool("", "all", false, "运行所有任务")

	return runCmd
}
