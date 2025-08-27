package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"gitee.com/MM-Q/bakctl/cmd/subcmd/add"
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

	// 注册add命令
	if err := qflag.AddSubCmd(addCmd); err != nil {
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
