package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// systemPathBlacklist 系统路径黑名单，禁止作为工作目录
var systemPathBlacklist = []string{
	"/",
	"/bin",
	"/sbin",
	"/usr",
	"/usr/bin",
	"/usr/sbin",
	"/usr/local/bin",
	"/etc",
	"/var",
	"/tmp",
	"/root",
	"/home",
	"/boot",
	"/proc",
	"/sys",
	"/dev",
}

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

	// 系统路径黑名单检查（先检查原始路径）
	if err := checkPathBlacklist(path); err != nil {
		return "", false, err
	}

	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return "", false, fmt.Errorf("代码工作目录不是目录: %s", path)
		}
		// 解析符号链接后再次检查黑名单，防止 symlink 绕过
		if realPath, evalErr := filepath.EvalSymlinks(path); evalErr == nil {
			realPath = filepath.Clean(realPath)
			if blErr := checkPathBlacklist(realPath); blErr != nil {
				return "", false, fmt.Errorf("代码工作目录通过符号链接指向系统目录: %s → %s", path, realPath)
			}
			path = realPath
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

// checkPathBlacklist 检查路径是否在系统黑名单中（含子目录）
func checkPathBlacklist(path string) error {
	cleaned := filepath.Clean(path)
	for _, blocked := range systemPathBlacklist {
		if cleaned == blocked || strings.HasPrefix(cleaned, blocked+string(filepath.Separator)) {
			return fmt.Errorf("禁止使用系统目录作为工作目录: %s", path)
		}
	}
	return nil
}

// GenerateWorkDir 为非编码类项目自动生成工作目录
// 编码类项目由用户手动指定，非编码类项目自动生成到 /www/wwwroot/project/easymvp/workspace/
func GenerateWorkDir(projectCategory string, projectID int64) string {
	family := GetCategoryFamily(projectCategory)
	if family == CategoryFamilyCoding {
		return "" // 编码类不自动生成，由用户指定
	}
	return fmt.Sprintf("/www/wwwroot/project/easymvp/workspace/%s/%d", string(family), projectID)
}
