package restore

// import (
// 	"flag"

// 	"gitee.com/MM-Q/qflag"
// 	"gitee.com/MM-Q/qflag/cmd"
// )

// var (
// 	restoreCmd *cmd.Cmd // restore命令

// 	// 必需参数
// 	taskIDFlag    *qflag.IntFlag    // 任务ID
// 	versionIDFlag *qflag.StringFlag // 版本ID

// 	// 可选参数
// 	targetDirFlag *qflag.StringFlag // 目标目录
// )

// // InitRestoreCmd 初始化restore子命令
// func InitRestoreCmd() *cmd.Cmd {
// 	restoreCmd = cmd.NewCmd("restore", "rs", flag.ExitOnError)
// 	restoreCmd.SetDescription("恢复备份文件")
// 	restoreCmd.SetUseChinese(true)

// 	// 必需参数
// 	taskIDFlag = restoreCmd.Int("", "id", 0, "指定要恢复的备份任务ID (必需)")
// 	versionIDFlag = restoreCmd.String("", "vid", "", "指定要恢复的备份版本ID (必需)")

// 	// 可选参数
// 	targetDirFlag = restoreCmd.String("", "d", ".", "指定恢复到的目标目录 (默认为当前目录)")

// 	return restoreCmd
// }
