package engine

import (
	"fmt"
	"os"
	"strings"
)

// ValidateWorkDir 校验工作目录是否存在且可用。
func ValidateWorkDir(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("代码工作目录不能为空")
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("代码工作目录不可用: %s", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("代码工作目录不是目录: %s", path)
	}
	return nil
}
