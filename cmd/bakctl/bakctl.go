// bakctl 是一个跨平台的备份管理工具的主程序入口。
//
// 该程序提供了完整的备份生命周期管理功能，包括：
//   - 备份任务的创建、编辑、删除和列表显示
//   - 备份的执行、监控和日志查看
//   - 备份文件的恢复和清理
//   - 配置的导入和导出
//
// 程序采用子命令架构，支持以下命令：
//   - add: 创建新的备份任务
//   - edit: 编辑现有备份任务
//   - list: 列出所有备份任务
//   - run: 执行备份任务
//   - log: 查看备份日志
//   - restore: 恢复备份文件
//   - delete: 删除备份任务
//   - export: 导出任务配置
//
// 使用示例：
//
//	bakctl add --name "文档备份" --backup-dir "/home/docs"
//	bakctl run --task-id 1
//	bakctl list
package bakctl

import (
	"fmt"
	"os"
	"runtime/debug"

	"gitee.com/MM-Q/bakctl/cmd/subcmd/add"
	"gitee.com/MM-Q/bakctl/cmd/subcmd/delete"
	"gitee.com/MM-Q/bakctl/cmd/subcmd/edit"
	"gitee.com/MM-Q/bakctl/cmd/subcmd/export"
	"gitee.com/MM-Q/bakctl/cmd/subcmd/list"
	"gitee.com/MM-Q/bakctl/cmd/subcmd/log"
	"gitee.com/MM-Q/bakctl/cmd/subcmd/restore"
	"gitee.com/MM-Q/bakctl/cmd/subcmd/run"
	"gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/colorlib"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/verman"
)

func BakctlMain() {
	// 捕获panic
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic: %v\nstack: %s\n", err, debug.Stack())
			os.Exit(1)
		}
	}()

	// 获取输出渲染器
	CL := colorlib.New()

	// 初始化全局主命令
	initMainCmd()

	// 获取add命令
	addCmd := add.InitAddCmd()

	// 获取edit命令
	editCmd := edit.InitEditCmd()

	// 获取list命令
	listCmd := list.InitListCmd()

	// 获取log命令
	logCmd := log.InitLogCmd()

	// 获取run命令
	runCmd := run.InitRunCmd()

	// 获取delete命令
	deleteCmd := delete.InitDeleteCmd()

	// 获取export命令
	exportCmd := export.InitExportCmd()

	// 获取restore命令
	restoreCmd := restore.InitRestoreCmd()

	// 注册子命令
	if err := qflag.AddSubCmd(addCmd, editCmd, listCmd, logCmd, runCmd, deleteCmd, exportCmd, restoreCmd); err != nil {
		CL.PrintError(err)
		os.Exit(1)
	}

	// 解析参数
	if parseErr := qflag.Parse(); parseErr != nil {
		fmt.Printf("err: %v\n", parseErr)
		os.Exit(1)
	}

	// 初始化数据库配置
	db, err := db.InitSQLite(types.DBFilename, types.DataDirPath)
	if err != nil {
		CL.PrintError(err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	// 设置颜色（默认启用，除非用户指定禁用）
	CL.SetColor(!noColorF.Get())

	// 获取命令名, 如果没有命令名, 则打印帮助信息
	cmdName := qflag.Arg(0)
	if cmdName == "" {
		qflag.PrintHelp()
		os.Exit(0)
	}

	// 路由命令
	switch cmdName {
	case addCmd.LongName(), addCmd.ShortName(): // add 命令
		if err := add.AddCmdMain(db, CL); err != nil {
			CL.PrintError(err)
			os.Exit(1)
		}
		return

	case editCmd.LongName(), editCmd.ShortName(): // edit 命令
		if err := edit.EditCmdMain(db, CL); err != nil {
			CL.PrintError(err)
			os.Exit(1)
		}
		return

	case listCmd.LongName(), listCmd.ShortName(): // list 命令
		if err := list.ListCmdMain(db, CL); err != nil {
			CL.PrintError(err)
			os.Exit(1)
		}
		return

	case logCmd.LongName(), logCmd.ShortName(): // log 命令
		if err := log.LogCmdMain(db, CL); err != nil {
			CL.PrintError(err)
			os.Exit(1)
		}
		return

	case runCmd.LongName(), runCmd.ShortName(): // run 命令
		if err := run.RunCmdMain(db, CL); err != nil {
			CL.PrintError(err)
			os.Exit(1)
		}
		return

	case deleteCmd.LongName(), deleteCmd.ShortName(): // delete 命令
		if err := delete.DeleteCmdMain(db, CL); err != nil {
			CL.PrintError(err)
			os.Exit(1)
		}
		return

	case exportCmd.LongName(), exportCmd.ShortName(): // export 命令
		if err := export.ExportCmdMain(db); err != nil {
			CL.PrintError(err)
			os.Exit(1)
		}
		return

	case restoreCmd.LongName(), restoreCmd.ShortName(): // restore 命令
		if err := restore.RestoreCmdMain(db, CL); err != nil {
			CL.PrintError(err)
			os.Exit(1)
		}
		return

	default:
		CL.PrintErrorf("unknown command: %s\n", cmdName)
		os.Exit(1)
	}
}

var (
	noColorF *qflag.BoolFlag // 禁用颜色
)

// 初始化主命令
func initMainCmd() {
	// 全局主命令的参数设置
	qflag.SetChinese(true)    // 使用中文版帮助信息
	qflag.SetCompletion(true) // 开启自动补全

	// 获取版本信息
	qflag.SetVersion(verman.V.Version())

	// 设置描述
	qflag.SetDesc("bakctl 是一个跨平台的备份管理工具，支持数据库存储和全面的备份操作")

	// 添加禁用颜色选项
	noColorF = qflag.Bool("no-color", "nc", false, "禁用颜色输出")
}
