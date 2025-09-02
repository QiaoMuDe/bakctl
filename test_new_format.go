package main

import (
	"fmt"
	"regexp"
	"time"
)

func main() {
	// 测试新的时间字符串格式
	now := time.Now()

	// 生成文件名（模拟新的格式）
	timeStr := now.Format("20060102_150405")
	filename := fmt.Sprintf("task1_%s.zip", timeStr)

	fmt.Printf("生成的文件名: %s\n", filename)

	// 测试解析
	parsedTime, err := time.Parse("20060102_150405", timeStr)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}

	fmt.Printf("原始时间: %s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Printf("解析时间: %s\n", parsedTime.Format("2006-01-02 15:04:05"))

	// 测试正则表达式匹配
	pattern := `^task1_(\d{8}_\d{6})\.zip$`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(filename)

	if len(matches) == 2 {
		fmt.Printf("✅ 正则匹配成功，提取的时间字符串: %s\n", matches[1])

		// 验证提取的时间字符串可以正确解析
		extractedTime, err := time.Parse("20060102_150405", matches[1])
		if err != nil {
			fmt.Printf("❌ 提取的时间字符串解析失败: %v\n", err)
		} else {
			fmt.Printf("✅ 提取的时间解析成功: %s\n", extractedTime.Format("2006-01-02 15:04:05"))
		}
	} else {
		fmt.Printf("❌ 正则匹配失败\n")
	}

	fmt.Println("\n=== 格式对比 ===")
	fmt.Printf("旧格式 (Unix时间戳): task1_%d.zip\n", now.Unix())
	fmt.Printf("新格式 (时间字符串): task1_%s.zip\n", timeStr)
	fmt.Printf("新格式更易读: %s 对应 %s\n", timeStr, now.Format("2006年01月02日 15:04:05"))
}
