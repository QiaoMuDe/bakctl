package add

import (
	"flag"

	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	addCmd  *cmd.Cmd          // 添加备份命令
	configF *qflag.StringFlag // 配置文件路径
	genF    *qflag.BoolFlag   // 生成配置文件
)

func InitAddCmd() *cmd.Cmd {
	addCmd = cmd.NewCmd("add", "a", flag.ExitOnError)
	addCmd.SetUseChinese(true)
	addCmd.SetDescription("添加备份任务")

	// 添加参数
	configF = addCmd.String("config", "c", "", "指定备份任务的文件路径")
	genF = addCmd.Bool("generate-template", "g", false, "生成备份任务模板")

	return addCmd
}
