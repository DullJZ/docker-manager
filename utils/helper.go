package utils

import (
	"regexp"
)

// CleanANSIEscapes 清理 ANSI 转义序列
func CleanANSIEscapes(text string) string {
	// 匹配所有 ANSI 转义序列的正则表达式
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)
	// 移除所有 ANSI 转义序列
	cleaned := ansiRegex.ReplaceAllString(text, "")
	return cleaned
}

// SuccessResponse 创建成功响应
func SuccessResponse(data interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"status": "success",
	}
	if data != nil {
		result["data"] = data
	}
	return result
}

// ErrorResponse 创建错误响应
func ErrorResponse(message string) map[string]interface{} {
	return map[string]interface{}{
		"status": "fail",
		"error":  message,
	}
}

// ErrorResponseWithRaw 创建包含原始错误信息的错误响应
func ErrorResponseWithRaw(message, rawMessage string) map[string]interface{} {
	return map[string]interface{}{
		"status":      "fail",
		"message":     message,
		"raw_message": rawMessage,
	}
}
