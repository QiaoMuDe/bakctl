package utils

import (
	"encoding/json"
)

// MarshalRules 将 Go 字符串切片编码为 JSON 数组字符串。
// 如果编码失败，返回空字符串和错误。
//
// 参数:
//   - rules: 要编码的字符串切片
//
// 返回值:
//   - string: 编码后的 JSON 数组字符串
//   - error: 编码过程中的错误（如果有）
func MarshalRules(rules []string) (string, error) {
	// 检查切片是否为空
	if len(rules) == 0 {
		return "[]", nil
	}

	jsonBytes, err := json.Marshal(rules)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// UnmarshalRules 将 JSON 数组字符串解码为 Go 字符串切片。
// 如果解码失败，返回 nil 切片和错误。
//
// 参数:
//   - jsonString: 要解码的 JSON 数组字符串
//
// 返回值:
//   - []string: 解码后的字符串切片
//   - error: 解码过程中的错误（如果有）
func UnmarshalRules(jsonString string) ([]string, error) {
	// 检查 JSON 字符串是否为空或只包含方括号
	if jsonString == "" || jsonString == "[]" {
		return []string{}, nil
	}

	var rules []string
	err := json.Unmarshal([]byte(jsonString), &rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}
