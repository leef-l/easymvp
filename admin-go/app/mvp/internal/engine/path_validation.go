package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateWorkDir 校验工作目录是否存在且可用。
func ValidateWorkDir(path string) error {
	_, _, err := ensureWorkDir(path, false)
	return err
}

// EnsureWorkDir 校验工作目录，必要时自动创建。
func EnsureWorkDir(path string) (string, bool, error) {
	return ensureWorkDir(path, true)
}

func ensureWorkDir(path string, autoCreate bool) (string, bool, error) {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" || path == "." {
		return "", false, fmt.Errorf("代码工作目录不能为空")
	}

	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return "", false, fmt.Errorf("代码工作目录不是目录: %s", path)
		}
		return path, false, nil
	}
	if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("代码工作目录不可用: %s", path)
	}
	if !autoCreate {
		return "", false, fmt.Errorf("代码工作目录不存在: %s", path)
	}
	if err = os.MkdirAll(path, 0o755); err != nil {
		return "", false, fmt.Errorf("代码工作目录创建失败: %s", path)
	}
	return path, true, nil
}
