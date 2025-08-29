package add

import (
	"flag"

	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/qflag/cmd"
)

var (
	addCmd  *cmd.Cmd          // 添加备份命令
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
	includeF *qflag.SliceFlag // 包含规则
	excludeF *qflag.SliceFlag // 排除规则

	// 文件大小限制
	maxSizeF *qflag.Int64Flag // 最大文件大小
	minSizeF *qflag.Int64Flag // 最小文件大小
)

// InitAddCmd 初始化添加备份命令
func InitAddCmd() *cmd.Cmd {
	addCmd = cmd.NewCmd("add", "a", flag.ExitOnError)
	addCmd.SetUseChinese(true)
	addCmd.SetDescription("添加备份任务")

	// 配置文件相关
	configF = addCmd.String("config", "c", "", "指定备份任务的文件路径")
	genF = addCmd.Bool("generate-template", "g", false, "生成备份任务模板")

	// 基本任务信息
	nameF = addCmd.String("name", "n", "", "任务名称 (必需)")
	backupDirF = addCmd.String("backup-dir", "s", "", "备份源目录 (必需)")
	storageDirF = addCmd.String("storage-dir", "d", "", "存储目录 (必需)")

	// 保留策略
	retainCountF = addCmd.Int("retain-count", "r", 3, "保留备份数量")
	retainDaysF = addCmd.Int("retain-days", "t", 7, "保留天数")

	// 压缩选项
	compressF = addCmd.Bool("compress", "C", false, "是否压缩备份")

	// 文件过滤规则
	includeF = addCmd.Slice("include", "i", []string{}, "包含规则, 多个规则用逗号分隔")
	excludeF = addCmd.Slice("exclude", "e", []string{}, "排除规则, 多个规则用逗号分隔")

	// 文件大小限制
	maxSizeF = addCmd.Int64("max-size", "M", 0, "最大文件大小 (字节, 0表示无限制)")
	minSizeF = addCmd.Int64("min-size", "m", 0, "最小文件大小 (字节, 0表示无限制)")

	return addCmd
}
