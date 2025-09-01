package run

import (
	"flag"

	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	runCmd *cmd.Cmd // run命令

	// 任务选择参数
	taskIDFlag   *qflag.Int64Flag // -id: 指定任务ID
	taskIDsFlag  *qflag.SliceFlag // -ids: 指定多个任务ID
	allTasksFlag *qflag.BoolFlag  // -all: 运行所有任务
)

// InitRunCmd 初始化run子命令
func InitRunCmd() *cmd.Cmd {
	runCmd = cmd.NewCmd("run", "r", flag.ExitOnError)
	runCmd.SetDescription("运行备份任务")
	runCmd.SetUseChinese(true)

	// 任务选择参数（互斥）
	taskIDFlag = runCmd.Int64("", "id", 0, "指定要运行的任务ID")
	taskIDsFlag = runCmd.Slice("", "ids", []string{}, "指定多个任务ID进行批量运行")
	allTasksFlag = runCmd.Bool("", "all", false, "运行所有任务")

	return runCmd
}
