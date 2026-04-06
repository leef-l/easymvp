package engine

import (
	"errors"
	"strings"
)

// FormatErrorChain 返回错误链的串联字符串，保证调用方能看到完整的 unwrap 链。
func FormatErrorChain(err error) string {
	if err == nil {
		return ""
	}
	parts := []string{}
	for e := err; e != nil; e = errors.Unwrap(e) {
		parts = append(parts, e.Error())
	}
	return strings.Join(parts, " -> ")
}
