// Package utils 提供了 bakctl 工具的通用工具函数。
// 包含字节格式化、哈希计算、JSON 处理和系统路径操作等实用功能。
package utils

import (
	"os"
	"path/filepath"
	"time"
)

// GetUserHomeDir 获取用户家目录，如果失败则返回当前工作目录作为保底
//
// 返回值:
//   - string: 用户家目录路径
//
// 说明:
//   - 尝试获取用户家目录，如果失败则返回当前工作目录作为保底
//   - 如果所有路径获取都失败，则返回当前目录"."作为最后的保底
func GetUserHomeDir() string {
	// 尝试获取用户家目录
	homeDir, err := os.UserHomeDir()

	// 先判断是否成功获取家目录
	if err == nil {
		// 成功获取时，确保返回绝对路径
		absHome, absErr := filepath.Abs(homeDir)
		if absErr == nil {
			return absHome
		}
		// 如果转换绝对路径失败，直接返回原始家目录
		return homeDir
	}

	// 家目录获取失败，尝试获取当前工作目录
	wd, wdErr := os.Getwd()
	if wdErr == nil {
		return wd
	}

	// 所有路径获取都失败时，返回当前目录"."作为最后的保底
	return "."
}

// ConvertUTCToLocal 将 UTC 时间字符串转换为本地时间字符串
//
// 参数:
//   - utcTimeStr: UTC 时间字符串，格式固定为 "2006-01-02 15:04:05"
//
// 返回值:
//   - string: 本地时间字符串，格式为 "2006-01-02 15:04:05"
//
// 说明:
//   - 解析 SQLite 返回的 UTC 时间字符串并转换为本地时间
//   - 如果解析失败，返回原字符串
func ConvertUTCToLocal(utcTimeStr string) string {
	if utcTimeStr == "" {
		return ""
	}

	// 解析 UTC 时间字符串（SQLite CURRENT_TIMESTAMP 格式）
	utcTime, err := time.Parse("2006-01-02 15:04:05", utcTimeStr)
	if err != nil {
		// 解析失败时返回原字符串
		return utcTimeStr
	}

	// 转换为本地时间
	localTime := utcTime.Local()

	// 格式化为相同的格式
	return localTime.Format("2006-01-02 15:04:05")
}
