// Package delete 的命令行参数定义和解析功能。
//
// 该文件定义了 delete 子命令支持的所有命令行参数，包括：
//   - 任务选择参数：任务ID、任务ID列表、全部任务选择
//   - 删除范围参数：是否删除备份文件、是否删除特定版本
//   - 安全控制参数：强制删除、确认提示等
//
// 提供灵活的删除选项和安全机制，确保用户能够精确控制删除范围。
package delete

import (
	"flag"

	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	deleteCmd *qflag.Cmd // 删除备份任务命令

	// 任务ID选择
	idF  *qflag.Int64Flag      // 单个任务ID
	idsF *qflag.Int64SliceFlag // 多个任务ID列表

	// 删除选项
	forceF     *qflag.BoolFlag // 强制删除，跳过确认提示
	keepFilesF *qflag.BoolFlag // 只删除数据库记录，保留备份文件
	failedF    *qflag.BoolFlag // 删除所有失败的备份记录
)

// InitDeleteCmd 初始化删除备份任务命令
func InitDeleteCmd() *qflag.Cmd {
	deleteCmd = cmd.NewCmd("delete", "del", flag.ExitOnError)
	deleteCmd.SetUseChinese(true)
	deleteCmd.SetDescription("删除备份任务")

	// 任务ID选择 (二选一)
	idF = deleteCmd.Int64("", "id", 0, "删除指定ID的单个备份任务")
	idsF = deleteCmd.Int64Slice("", "ids", []int64{}, "批量删除多个备份任务 (逗号分隔)")

	// 删除选项
	forceF = deleteCmd.Bool("force", "f", false, "强制删除，跳过确认提示")
	keepFilesF = deleteCmd.Bool("keep-files", "k", false, "只删除数据库记录，保留备份文件")
	failedF = deleteCmd.Bool("failed", "fd", false, "删除所有失败的备份记录")

	return deleteCmd
}
