package edit

import (
	"flag"

	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	editCmd *cmd.Cmd // 编辑备份任务命令

	// 任务ID选择
	idF  *qflag.IntFlag   // 单个任务ID
	idsF *qflag.SliceFlag // 多个任务ID (切片类型)

	// 可编辑的配置项
	retainCountF *qflag.IntFlag    // 保留备份数量
	retainDaysF  *qflag.IntFlag    // 保留天数
	compressF    *qflag.StringFlag // 是否压缩 (使用字符串来区分未设置)
	includeF     *qflag.SliceFlag  // 包含规则 (切片类型)
	excludeF     *qflag.SliceFlag  // 排除规则 (切片类型)
	maxSizeF     *qflag.Int64Flag  // 最大文件大小
	minSizeF     *qflag.Int64Flag  // 最小文件大小

	// 特殊标志：用于清空规则
	clearIncludeF *qflag.BoolFlag // 清空包含规则
	clearExcludeF *qflag.BoolFlag // 清空排除规则
)

func InitEditCmd() *cmd.Cmd {
	editCmd = cmd.NewCmd("edit", "e", flag.ExitOnError)
	editCmd.SetUseChinese(true)
	editCmd.SetDescription("编辑备份任务配置")

	// 任务ID选择 (二选一)
	idF = editCmd.Int("id", "I", 0, "指定单个任务ID进行编辑")
	idsF = editCmd.Slice("ids", "S", []string{}, "指定多个任务ID进行批量编辑")

	// 可编辑的配置项
	retainCountF = editCmd.Int("retain-count", "r", -1, "保留备份数量 (-1表示不修改)")
	retainDaysF = editCmd.Int("retain-days", "t", -1, "保留天数 (-1表示不修改)")
	compressF = editCmd.String("compress", "C", "", "是否压缩备份 (true/false, 空字符串表示不修改)")
	includeF = editCmd.Slice("include", "i", []string{}, "包含规则")
	excludeF = editCmd.Slice("exclude", "x", []string{}, "排除规则")
	maxSizeF = editCmd.Int64("max-size", "M", -1, "最大文件大小 (字节, -1表示不修改)")
	minSizeF = editCmd.Int64("min-size", "m", -1, "最小文件大小 (字节, -1表示不修改)")

	// 特殊标志：用于清空规则
	clearIncludeF = editCmd.Bool("clear-include", "", false, "清空包含规则")
	clearExcludeF = editCmd.Bool("clear-exclude", "", false, "清空排除规则")

	return editCmd
}
