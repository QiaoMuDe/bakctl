// Package list 的命令行参数定义和解析功能。
//
// 该文件定义了 list 子命令支持的所有命令行参数，包括：
//   - 显示格式参数：表格格式、详细模式、简洁模式
//   - 过滤参数：按任务名称、状态等条件过滤
//   - 排序参数：按不同字段排序显示
//
// 提供灵活的列表显示选项，满足不同用户的查看需求。
package list

import (
	"flag"

	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	listCmd           *qflag.Cmd      // list命令
	listCmdTableStyle *qflag.EnumFlag // 日志表格样式
	listCmdSimple     *qflag.BoolFlag // 简化显示
)

func InitListCmd() *qflag.Cmd {
	listCmd = cmd.NewCmd("list", "ls", flag.ExitOnError)
	listCmd.SetDesc("列出所有备份任务")
	listCmd.SetChinese(true)

	// 添加标志
	listCmdTableStyle = listCmd.Enum("table-style", "ts", "ro", "列表表格样式, 支持的样式有:\n"+
		"\t\t\t\t\t[df  ] - 默认样式\n"+
		"\t\t\t\t\t[bd  ] - 加粗样式\n"+
		"\t\t\t\t\t[cb  ] - 亮色样式\n"+
		"\t\t\t\t\t[cd  ] - 暗色样式\n"+
		"\t\t\t\t\t[de  ] - 双边框样式\n"+
		"\t\t\t\t\t[lt  ] - 方边框样式\n"+
		"\t\t\t\t\t[ro  ] - 圆边框样式\n"+
		"\t\t\t\t\t[none] - 无边框样式", types.TableStyleList)
	listCmdSimple = listCmd.Bool("simple", "s", false, "简化显示，只显示核心信息")

	return listCmd
}
