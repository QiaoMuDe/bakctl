// Package log 实现了 bakctl 的 log 子命令的命令行参数解析功能。
//
// 该文件定义了 log 命令支持的所有命令行标志和参数，包括：
//   - 任务ID过滤选项
//   - 时间范围过滤选项
//   - 状态过滤选项
//   - 输出格式选项
//   - 日志级别控制选项
//
// 通过这些参数，用户可以精确控制要查看的日志内容和显示格式。
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
	logCmdSimple     *qflag.BoolFlag   // 简化显示
	logCmdFailed     *qflag.BoolFlag   // 只显示失败的备份记录
)

// InitLogCmd 初始化日志命令
func InitLogCmd() *cmd.Cmd {
	logCmd = cmd.NewCmd("log", "lg", flag.ExitOnError)
	logCmd.SetDescription("查看备份记录日志")
	logCmd.SetUseChinese(true)

	// 添加标志
	logCmdTableStyle = logCmd.Enum("table-style", "ts", "df", "日志表格样式", types.TableStyleList)
	logCmdTaskID = logCmd.Int("id", "", 0, "指定任务ID来过滤备份记录")
	logCmdTaskName = logCmd.String("name", "n", "", "指定任务名称来过滤备份记录")
	logCmdLimit = logCmd.Int("limit", "l", 10, "限制显示的备份记录条数")
	logCmdSimple = logCmd.Bool("simple", "s", false, "简化显示，只显示核心信息")
	logCmdFailed = logCmd.Bool("failed", "fd", false, "只显示失败的备份记录")

	return logCmd
}
