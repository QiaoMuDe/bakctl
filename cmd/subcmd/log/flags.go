package log

import (
	"flag"

	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	logCmd           *cmd.Cmd          // log命令
	logCmdTableStyle *qflag.EnumFlag   // 日志表格样式
	logCmdTaskID     *qflag.IntFlag    // 任务ID标志
	logCmdTaskName   *qflag.StringFlag // 任务名称标志
	logCmdLimit      *qflag.IntFlag    // 限制条数标志
)

// InitLogCmd 初始化日志命令
func InitLogCmd() *cmd.Cmd {
	logCmd = cmd.NewCmd("log", "lg", flag.ExitOnError)
	logCmd.SetDescription("显示备份记录日志")
	logCmd.SetUseChinese(true)

	// 添加标志
	logCmdTableStyle = logCmd.Enum("table-style", "ts", "df", "日志表格样式", types.TableStyleList)
	logCmdTaskID = logCmd.Int("id", "", 0, "指定任务ID来过滤备份记录")
	logCmdTaskName = logCmd.String("name", "n", "", "指定任务名称来过滤备份记录")
	logCmdLimit = logCmd.Int("limit", "l", 10, "限制显示的备份记录条数")

	return logCmd
}
