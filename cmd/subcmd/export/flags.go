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

	// 导出类型标志
	cmdF    *qflag.BoolFlag // 导出添加任务命令
	scriptF *qflag.BoolFlag // 导出一键备份脚本

	// 脚本平台标志
	batF *qflag.BoolFlag // 生成Windows BAT脚本
	shF  *qflag.BoolFlag // 生成Linux Bash脚本
)

func InitExportCmd() *cmd.Cmd {
	exportCmd = cmd.NewCmd("export", "exp", flag.ExitOnError)
	exportCmd.SetUseChinese(true)
	exportCmd.SetDescription("导出备份任务数据")

	// 任务选择标志 (三选一)
	idF = exportCmd.Int("", "id", 0, "指定单个任务ID进行导出")
	idsF = exportCmd.Slice("", "ids", []string{}, "指定多个任务ID进行导出, 用逗号分隔")
	allF = exportCmd.Bool("", "all", false, "导出所有任务")

	// 导出类型标志 (二选一)
	cmdF = exportCmd.Bool("cmd", "c", false, "导出添加任务命令")
	scriptF = exportCmd.Bool("script", "s", false, "导出一键备份脚本")

	// 脚本平台标志 (与--script配合使用，二选一)
	batF = exportCmd.Bool("", "bat", false, "生成Windows BAT脚本")
	shF = exportCmd.Bool("", "sh", false, "生成Linux Bash脚本")

	return exportCmd
}
