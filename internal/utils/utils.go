package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// StringSliceToInt64 将字符串切片转换为 int64 切片
// 会跳过无法解析的字符串并返回错误信息
//
// 参数:
//   - strs: 字符串切片
//
// 返回值:
//   - []int64: 转换后的 int64 切片
//   - error: 转换过程中的错误
func StringSliceToInt64(strs []string) ([]int64, error) {
	if len(strs) == 0 {
		return nil, nil
	}

	ids := make([]int64, 0, len(strs))
	var errors []string

	for _, str := range strs {
		if str == "" {
			continue // 跳过空字符串
		}

		// 尝试将字符串转换为64位整数
		id, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			errors = append(errors, fmt.Sprintf("'%s'", str))
			continue
		}

		ids = append(ids, id)
	}

	if len(errors) > 0 {
		return ids, fmt.Errorf("无法解析以下字符串为整数: %s", strings.Join(errors, ", "))
	}

	return ids, nil
}
