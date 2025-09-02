package export

import (
	"flag"

	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	exportCmd *cmd.Cmd // 导出备份任务命令

	// 任务选择标志
	idF  *qflag.IntFlag   // 单个任务ID
	idsF *qflag.SliceFlag // 多个任务ID
	allF *qflag.BoolFlag  // 导出所有任务
)

func InitExportCmd() *cmd.Cmd {
	exportCmd = cmd.NewCmd("export", "exp", flag.ExitOnError)
	exportCmd.SetUseChinese(true)
	exportCmd.SetDescription("导出备份任务的添加命令")

	// 任务选择标志 (三选一)
	idF = exportCmd.Int("id", "I", 0, "指定单个任务ID进行导出")
	idsF = exportCmd.Slice("ids", "S", []string{}, "指定多个任务ID进行导出，用逗号分隔")
	allF = exportCmd.Bool("all", "A", false, "导出所有任务")

	return exportCmd
}
