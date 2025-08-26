package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/bakctl/internal/utils"
	"gitee.com/MM-Q/colorlib"
)

func main() {
	// 捕获panic
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic: %v\nstack: %s\n", err, debug.Stack())
			os.Exit(1)
		}
	}()

	// 获取打印库
	CL := colorlib.New()

	// 初始化数据库配置
	db, err := utils.InitSQLite(types.DBFilename, types.DataDirPath)
	if err != nil {
		CL.PrintErrorf("初始化数据库失败: %v", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

}
