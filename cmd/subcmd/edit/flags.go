// Package edit 的命令行参数定义和解析功能。
//
// 该文件定义了 edit 子命令支持的所有命令行参数，包括：
//   - 任务标识参数：任务ID用于指定要编辑的任务
//   - 可修改的配置参数：与 add 命令相同的所有配置选项
//   - 修改模式参数：增量修改或完全替换等选项
//
// 支持对现有任务的所有配置项进行修改，提供灵活的编辑能力。
package edit

import (
	"flag"

	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	editCmd *cmd.Cmd // 编辑备份任务命令

	// 任务ID选择
	idF  *qflag.IntFlag         // 单个任务ID
	idsF *qflag.StringSliceFlag // 多个任务ID (切片类型)

	// 可编辑的配置项
	retainCountF *qflag.IntFlag         // 保留备份数量
	retainDaysF  *qflag.IntFlag         // 保留天数
	compressF    *qflag.StringFlag      // 是否压缩 (使用字符串来区分未设置)
	includeF     *qflag.StringSliceFlag // 包含规则 (切片类型)
	excludeF     *qflag.StringSliceFlag // 排除规则 (切片类型)
	maxSizeF     *qflag.SizeFlag        // 最大文件大小
	minSizeF     *qflag.SizeFlag        // 最小文件大小

	// 特殊标志：用于清空规则
	clearIncludeF *qflag.BoolFlag // 清空包含规则
	clearExcludeF *qflag.BoolFlag // 清空排除规则
)

func InitEditCmd() *cmd.Cmd {
	editCmd = cmd.NewCmd("edit", "e", flag.ExitOnError)
	editCmd.SetUseChinese(true)
	editCmd.SetDescription("编辑备份任务配置")

	// 任务ID选择 (二选一)
	idF = editCmd.Int("", "id", 0, "指定单个任务ID进行编辑")
	idsF = editCmd.StringSlice("", "ids", []string{}, "指定多个任务ID进行批量编辑")

	// 可编辑的配置项
	retainCountF = editCmd.Int("retain-count", "r", -1, "保留备份数量 (-1表示不修改)")
	retainDaysF = editCmd.Int("retain-days", "t", -1, "保留天数 (-1表示不修改)")
	compressF = editCmd.String("compress", "c", "", "是否压缩备份 (true/false, 空字符串表示不修改)")
	includeF = editCmd.StringSlice("include", "i", []string{}, "包含规则, 多个规则用逗号分隔")
	excludeF = editCmd.StringSlice("exclude", "x", []string{}, "排除规则,	多个规则用逗号分隔")
	maxSizeF = editCmd.Size("max-size", "mx", -1, "最大文件大小 (字节, -1表示不修改)")
	minSizeF = editCmd.Size("min-size", "ms", -1, "最小文件大小 (字节, -1表示不修改)")

	// 特殊标志：用于清空规则
	clearIncludeF = editCmd.Bool("clear-include", "", false, "清空包含规则")
	clearExcludeF = editCmd.Bool("clear-exclude", "", false, "清空排除规则")

	return editCmd
}
