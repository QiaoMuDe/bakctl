// Package add 的命令行参数定义和解析功能。
//
// 该文件定义了 add 子命令支持的所有命令行参数，包括：
//   - 基本配置参数：任务名称、备份目录、存储目录
//   - 压缩和保留策略参数：压缩开关、保留数量、保留天数
//   - 文件过滤参数：包含规则、排除规则、文件大小限制
//   - 配置文件参数：从 TOML 文件读取配置
//
// 所有参数都提供了详细的帮助信息和默认值，支持短参数和长参数两种形式。
// 参数验证确保用户输入的有效性和一致性。
package add

import (
	"flag"

	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	addCmd  *qflag.Cmd        // 添加备份命令
	configF *qflag.StringFlag // 配置文件路径
	genF    *qflag.BoolFlag   // 生成配置文件

	// 基本任务信息
	nameF       *qflag.StringFlag // 任务名称
	backupDirF  *qflag.StringFlag // 备份源目录
	storageDirF *qflag.StringFlag // 存储目录

	// 保留策略
	retainCountF *qflag.IntFlag // 保留备份数量
	retainDaysF  *qflag.IntFlag // 保留天数

	// 压缩选项
	compressF *qflag.BoolFlag // 是否压缩

	// 文件过滤规则
	includeF *qflag.StringSliceFlag // 包含规则
	excludeF *qflag.StringSliceFlag // 排除规则

	// 文件大小限制
	maxSizeF *qflag.SizeFlag // 最大文件大小
	minSizeF *qflag.SizeFlag // 最小文件大小
)

// InitAddCmd 初始化添加备份命令
func InitAddCmd() *qflag.Cmd {
	addCmd = cmd.NewCmd("add", "a", flag.ExitOnError)
	addCmd.SetUseChinese(true)
	addCmd.SetDescription("添加备份任务")

	// 配置文件相关
	configF = addCmd.String("config", "C", "", "指定备份任务的文件路径")
	genF = addCmd.Bool("generate-template", "g", false, "生成备份任务模板")

	// 基本任务信息
	nameF = addCmd.String("name", "n", "", "任务名称 (必需)")
	backupDirF = addCmd.String("backup-dir", "b", "", "备份源目录 (必需)")
	storageDirF = addCmd.String("storage-dir", "s", "", "存储目录 (必需)")

	// 保留策略
	retainCountF = addCmd.Int("retain-count", "r", 3, "保留备份数量")
	retainDaysF = addCmd.Int("retain-days", "t", 7, "保留天数")

	// 压缩选项
	compressF = addCmd.Bool("compress", "c", false, "是否压缩备份")

	// 文件过滤规则
	includeF = addCmd.StringSlice("include", "i", []string{}, "包含规则, 多个规则用逗号分隔")
	excludeF = addCmd.StringSlice("exclude", "e", []string{}, "排除规则, 多个规则用逗号分隔")

	// 文件大小限制
	maxSizeF = addCmd.Size("max-size", "mx", 0, "最大文件大小 (0表示无限制)")
	minSizeF = addCmd.Size("min-size", "ms", 0, "最小文件大小 (0表示无限制)")

	return addCmd
}
