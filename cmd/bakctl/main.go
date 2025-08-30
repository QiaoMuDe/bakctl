package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"gitee.com/MM-Q/bakctl/cmd/subcmd/add"
	"gitee.com/MM-Q/bakctl/cmd/subcmd/edit"
	"gitee.com/MM-Q/bakctl/cmd/subcmd/list"
	"gitee.com/MM-Q/bakctl/cmd/subcmd/log"
	"gitee.com/MM-Q/bakctl/cmd/subcmd/run"
	"gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/verman"
)

func main() {
	// 捕获panic
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic: %v\nstack: %s\n", err, debug.Stack())
			os.Exit(1)
		}
	}()

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

	// 注册子命令
	if err := qflag.AddSubCmd(addCmd, editCmd, listCmd, logCmd, runCmd); err != nil {
		fmt.Printf("err: %v\n", err)
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
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	// 获取命令名
	cmdName := qflag.Arg(0)
	if cmdName == "" {
		qflag.PrintHelp()
		os.Exit(0)
	}

	// 路由命令
	switch cmdName {
	case addCmd.LongName(), addCmd.ShortName(): // add 命令
		if err := add.AddCmdMain(db); err != nil {
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}
		return

	case editCmd.LongName(), editCmd.ShortName(): // edit 命令
		if err := edit.EditCmdMain(db); err != nil {
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}
		return

	case listCmd.LongName(), listCmd.ShortName(): // list 命令
		if err := list.ListCmdMain(db); err != nil {
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}
		return

	case logCmd.LongName(), logCmd.ShortName(): // log 命令
		if err := log.LogCmdMain(db); err != nil {
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}
		return

	case runCmd.LongName(), runCmd.ShortName(): // run 命令
		if err := run.RunCmdMain(db); err != nil {
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}
		return

	default:
		fmt.Printf("unknown command: %s\n", cmdName)
		os.Exit(1)
	}
}

// 初始化主命令
func initMainCmd() {
	// 全局主命令的参数设置
	qflag.SetUseChinese(true)
	qflag.SetEnableCompletion(true)

	// 获取版本信息
	v := verman.Get()
	qflag.SetVersionf("%s %s", v.AppName, v.GitVersion)

	// 设置描述
	qflag.SetDescription("bakctl is a backup system for Linux")
}
