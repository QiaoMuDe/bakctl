package log

import (
	"flag"

	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	logCmd           *cmd.Cmd        // log命令
	logCmdTableStyle *qflag.EnumFlag // 日志表格样式
)

func InitLogCmd() *cmd.Cmd {
	logCmd = cmd.NewCmd("log", "lg", flag.ExitOnError)
	logCmd.SetDescription("显示备份记录日志")
	logCmd.SetUseChinese(true)

	// 添加标志
	logCmdTableStyle = logCmd.Enum("table-style", "ts", "ro", "日志表格样式", types.TableStyleList)

	return logCmd
}
