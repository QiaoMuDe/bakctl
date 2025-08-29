package utils

import (
	"os"
	"path/filepath"
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
