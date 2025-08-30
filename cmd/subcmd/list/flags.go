package list

import (
	"flag"

	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	listCmd           *cmd.Cmd        // list命令
	listCmdTableStyle *qflag.EnumFlag // 日志表格样式
)

func InitListCmd() *cmd.Cmd {
	listCmd = cmd.NewCmd("list", "ls", flag.ExitOnError)
	listCmd.SetDescription("列出所有备份任务")
	listCmd.SetUseChinese(true)

	// 添加标志
	listCmdTableStyle = listCmd.Enum("table-style", "ts", "df", "列表表格样式", types.TableStyleList)

	return listCmd
}
