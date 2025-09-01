package delete

import (
	"flag"

	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	deleteCmd *cmd.Cmd // 删除备份任务命令

	// 任务ID选择
	idF  *qflag.IntFlag   // 单个任务ID
	idsF *qflag.SliceFlag // 多个任务ID列表

	// 删除选项
	forceF     *qflag.BoolFlag // 强制删除，跳过确认提示
	keepFilesF *qflag.BoolFlag // 只删除数据库记录，保留备份文件
)

// InitDeleteCmd 初始化删除备份任务命令
func InitDeleteCmd() *cmd.Cmd {
	deleteCmd = cmd.NewCmd("delete", "del", flag.ExitOnError)
	deleteCmd.SetUseChinese(true)
	deleteCmd.SetDescription("删除备份任务")

	// 任务ID选择 (二选一)
	idF = deleteCmd.Int("id", "I", 0, "删除指定ID的单个备份任务")
	idsF = deleteCmd.Slice("ids", "S", []string{}, "批量删除多个备份任务（逗号分隔）")

	// 删除选项
	forceF = deleteCmd.Bool("force", "f", false, "强制删除，跳过确认提示")
	keepFilesF = deleteCmd.Bool("keep-files", "k", false, "只删除数据库记录，保留备份文件")

	return deleteCmd
}
