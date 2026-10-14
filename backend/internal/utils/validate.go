package utils

import (
	"regexp"
	"strings"
)

// emailPattern 邮箱格式（与前端表单校验一致）
var emailPattern = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

// NormalizeEmail 去除邮箱首尾空白（含换行、全角空格）
func NormalizeEmail(email string) string {
	return strings.TrimSpace(email)
}

// IsEmail 校验是否为合法邮箱地址
func IsEmail(email string) bool {
	return emailPattern.MatchString(email)
}
