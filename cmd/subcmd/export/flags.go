// Package export 的命令行参数定义和解析功能。
//
// 该文件定义了 export 子命令支持的所有命令行参数，包括：
//   - 任务选择参数：任务ID、任务ID列表、全部任务导出
//   - 输出控制参数：输出文件路径、输出格式选项
//   - 导出范围参数：是否包含敏感信息、导出模板等
//
// 提供灵活的导出选项，支持不同场景下的配置导出需求。
package export

import (
	"flag"

	"gitee.com/MM-Q/qflag"
)

var (
	exportCmd *qflag.Cmd // 导出备份任务命令

	// 任务选择标志
	idF  *qflag.IntFlag        // 单个任务ID
	idsF *qflag.Int64SliceFlag // 多个任务ID
	allF *qflag.BoolFlag       // 导出所有任务

	// 导出类型标志
	cmdF    *qflag.BoolFlag // 导出添加任务命令
	scriptF *qflag.BoolFlag // 导出一键备份脚本

	// 脚本平台标志
	batF *qflag.BoolFlag // 生成Windows BAT脚本
	shF  *qflag.BoolFlag // 生成Linux Bash脚本
)

func InitExportCmd() *qflag.Cmd {
	exportCmd = qflag.NewCmd("export", "exp", flag.ExitOnError)
	exportCmd.SetChinese(true)
	exportCmd.SetDesc("导出备份任务数据")

	// 任务选择标志 (三选一)
	idF = exportCmd.Int("", "id", "指定单个任务ID进行导出", 0)
	idsF = exportCmd.Int64Slice("", "ids", "指定多个任务ID进行导出, 用逗号分隔", []int64{})
	allF = exportCmd.Bool("", "all", "导出所有任务", false)

	// 导出类型标志 (二选一)
	cmdF = exportCmd.Bool("cmd", "c", "导出添加任务命令", false)
	scriptF = exportCmd.Bool("script", "s", "导出一键备份脚本", false)

	// 脚本平台标志 (与--script配合使用，二选一)
	batF = exportCmd.Bool("", "bat", "生成Windows BAT脚本", false)
	shF = exportCmd.Bool("", "sh", "生成Linux Bash脚本", false)

	return exportCmd
}
