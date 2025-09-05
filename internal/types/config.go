// Package types 定义了 bakctl 工具的核心数据类型和配置结构。
// 包含备份任务配置、数据库记录结构以及相关的验证和转换功能。
package types

import (
	"fmt"
	"strings"
)

// RootConfig 根配置结构体, 用于解析TOML配置文件
type RootConfig struct {
	AddTaskConfig AddTaskConfig `toml:"AddTaskConfig"`
}

// AddTaskConfig 表示添加备份任务的配置结构
// 对应TOML配置文件中的[AddTaskConfig]部分
type AddTaskConfig struct {
	Name         string   `toml:"name" comment:"任务名称(必填, 唯一, 不可重复)"`                                                      // 任务名称
	BackupDir    string   `toml:"backup_dir" comment:"备份源目录(必填, 单个路径, 支持Windows和Linux路径)"`                                // 备份源目录
	StorageDir   string   `toml:"storage_dir" comment:"备份存储目录(必填, 单个路径, 备份文件最终存放位置, 如果为'', 则默认使用 ~/.bakctl)"`             // 备份存储目录
	RetainCount  int      `toml:"retain_count" comment:"保留备份文件的数量(可选, 默认0个; 设置为0表示不按数量限制)"`                               // 保留备份文件的数量
	RetainDays   int      `toml:"retain_days" comment:"保留备份文件的天数(可选, 默认0天; 设置为0表示不按天数限制)"`                                // 保留备份文件的天数
	Compress     bool     `toml:"compress" comment:"是否压缩(可选, 默认false)"`                                                   // 是否压缩
	IncludeRules []string `toml:"include_rules" comment:"包含规则(可选, 仅备份符合规则的文件; 空数组表示备份所有文件)"`                              // 包含规则
	ExcludeRules []string `toml:"exclude_rules" comment:"排除规则(可选, 不备份符合规则的文件; 优先级高于包含规则, 即\"先包含后排除\")"`                   // 排除规则
	MaxFileSize  int64    `toml:"max_file_size" comment:"最大文件大小(可选, 超过此尺寸的文件不备份, 单位字节, 默认为0表示不限制; 示例: 1073741824 = 1GB)"` // 最大文件大小
	MinFileSize  int64    `toml:"min_file_size" comment:"最小文件大小(可选, 小于此尺寸的文件不备份, 单位字节, 默认为0表示不限制; 示例: 1024 = 1KB)"`       // 最小文件大小
}

type TaskConfig struct {
	Name         string   // 任务名称
	BackupDir    string   // 备份源目录
	StorageDir   string   // 备份存储目录
	RetainCount  int      // 保留备份数量
	RetainDays   int      // 保留备份天数
	Compress     bool     // 是否压缩
	IncludeRules []string // 包含规则
	ExcludeRules []string // 排除规则
	MaxFileSize  int64    // 最大文件大小
	MinFileSize  int64    // 最小文件大小
}

// invalidChars 全局map，用于定义不允许的特殊字符。
// 路径分隔符 '/' 和 '\' 在路径验证时被明确允许。
var invalidChars = map[rune]struct{}{
	'!': {}, '@': {}, '#': {}, '$': {}, '%': {}, '^': {}, '&': {}, '*': {}, '(': {}, ')': {},
	'+': {}, '=': {}, '{': {}, '}': {}, '[': {}, ']': {}, '|': {}, ';': {},
	'\'': {}, '"': {}, '<': {}, '>': {}, ',': {}, '?': {}, '~': {}, '`': {},
}

// isValidString 检查字符串是否为空，并且不包含任何不允许的特殊字符。
//
// 参数:
//   - s: 要检查的字符串
//   - allowPathSeparators: 是否允许路径分隔符 '/' 和 '\'
//
// 返回值:
//   - error: 如果字符串为空或包含不允许的特殊字符，则返回错误信息
func isValidString(s string, allowPathSeparators bool) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("不能为空")
	}
	for _, r := range s {
		if allowPathSeparators && (r == '/' || r == '\\') {
			continue
		}
		if _, ok := invalidChars[r]; ok {
			return fmt.Errorf("不能包含特殊字符 '%c'", r)
		}
	}
	return nil
}

// Validate 验证 AddTaskConfig 结构体的字段。
func (cfg *AddTaskConfig) Validate() error {
	// 验证任务名称
	if err := isValidString(cfg.Name, false); err != nil {
		return fmt.Errorf("任务名称 %w", err)
	}

	// 验证备份源目录
	if err := isValidString(cfg.BackupDir, true); err != nil {
		return fmt.Errorf("备份源目录 %w", err)
	}

	return nil
}
