package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// fileReadMaxSize 单个文件最大读取 512KB
const fileReadMaxSize = 512 * 1024

// fileReadMaxTotal 单条消息最大读取总量 2MB
const fileReadMaxTotal = 2 * 1024 * 1024

// dirMaxFiles 目录最大展开文件数
const dirMaxFiles = 50

// readCmdPattern 匹配 "读取：路径" 或 "读取: 路径" 格式
var readCmdPattern = regexp.MustCompile(`(?m)^读取[：:]\s*(.+)$`)

// ExpandFileReads 解析消息中的 "读取：路径" 指令，将文件/目录内容展开拼接到消息中
// 原始指令保留，内容追加到消息末尾
func ExpandFileReads(content string) string {
	matches := readCmdPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return content
	}

	var blocks []string
	totalSize := 0

	for _, m := range matches {
		path := strings.TrimSpace(m[1])
		if path == "" {
			continue
		}

		block, size := readPath(path, fileReadMaxTotal-totalSize)
		if block != "" {
			blocks = append(blocks, block)
			totalSize += size
		}

		if totalSize >= fileReadMaxTotal {
			blocks = append(blocks, "\n> ⚠️ 已达到读取总量上限(2MB)，后续文件跳过")
			break
		}
	}

	if len(blocks) == 0 {
		return content
	}

	return content + "\n\n---\n以下是读取的文件内容：\n" + strings.Join(blocks, "\n")
}

// readPath 读取文件或目录，返回格式化的内容和字节数
func readPath(path string, remaining int) (string, int) {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Sprintf("\n> ⚠️ 无法读取 `%s`: %v", path, err), 0
	}

	if info.IsDir() {
		return readDir(path, remaining)
	}
	return readFile(path, remaining)
}

// readFile 读取单个文件
func readFile(path string, remaining int) (string, int) {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Sprintf("\n> ⚠️ 无法读取 `%s`: %v", path, err), 0
	}

	size := info.Size()
	if size == 0 {
		return fmt.Sprintf("\n### 📄 `%s`\n（空文件）", path), 0
	}

	// 限制单文件和剩余总量
	readSize := int(size)
	if readSize > fileReadMaxSize {
		readSize = fileReadMaxSize
	}
	if readSize > remaining {
		readSize = remaining
	}

	data := make([]byte, readSize)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Sprintf("\n> ⚠️ 无法打开 `%s`: %v", path, err), 0
	}
	defer f.Close()

	n, err := f.Read(data)
	if err != nil {
		return fmt.Sprintf("\n> ⚠️ 读取 `%s` 失败: %v", path, err), 0
	}
	data = data[:n]

	ext := filepath.Ext(path)
	lang := extToLang(ext)
	truncated := ""
	if int64(readSize) < size {
		truncated = fmt.Sprintf("\n> （文件共 %d 字节，仅展示前 %d 字节）", size, n)
	}

	return fmt.Sprintf("\n### 📄 `%s`%s\n```%s\n%s\n```", path, truncated, lang, string(data)), n
}

// readDir 递归读取目录下的文件
func readDir(dir string, remaining int) (string, int) {
	var blocks []string
	totalRead := 0
	fileCount := 0

	// 先输出目录树结构
	tree := buildDirTree(dir, "", 0)
	blocks = append(blocks, fmt.Sprintf("\n### 📁 `%s` 目录结构\n```\n%s\n```", dir, tree))

	// 递归读取所有文本文件
	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无法访问的文件
		}
		if info.IsDir() {
			// 跳过隐藏目录和常见无用目录
			name := info.Name()
			if name != "." && (strings.HasPrefix(name, ".") || skipDirs[name]) {
				return filepath.SkipDir
			}
			return nil
		}
		if fileCount >= dirMaxFiles {
			return filepath.SkipAll
		}
		if !isTextFile(info.Name()) {
			return nil
		}

		block, size := readFile(path, remaining-totalRead)
		blocks = append(blocks, block)
		totalRead += size
		fileCount++

		if totalRead >= remaining {
			return filepath.SkipAll
		}
		return nil
	})
	_ = walkErr

	if fileCount >= dirMaxFiles {
		blocks = append(blocks, fmt.Sprintf("\n> 文件数超过 %d 个上限，后续文件跳过", dirMaxFiles))
	}

	return strings.Join(blocks, "\n"), totalRead
}

// skipDirs 递归时跳过的目录
var skipDirs = map[string]bool{
	"node_modules": true, "vendor": true, ".git": true,
	"__pycache__": true, "dist": true, "build": true,
	".idea": true, ".vscode": true, ".next": true,
	"target": true, "bin": true, "obj": true,
}

// buildDirTree 构建目录树字符串（最多3层）
func buildDirTree(dir string, prefix string, depth int) string {
	if depth > 3 {
		return prefix + "..."
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return prefix + "(无法读取)"
	}

	var lines []string
	for i, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") || skipDirs[name] {
			continue
		}

		isLast := i == len(entries)-1
		connector := "├── "
		childPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		if e.IsDir() {
			lines = append(lines, prefix+connector+name+"/")
			sub := buildDirTree(filepath.Join(dir, name), childPrefix, depth+1)
			if sub != "" {
				lines = append(lines, sub)
			}
		} else {
			lines = append(lines, prefix+connector+name)
		}
	}
	return strings.Join(lines, "\n")
}

// extToLang 文件扩展名映射代码块语言
func extToLang(ext string) string {
	m := map[string]string{
		".go":    "go",
		".js":    "javascript",
		".ts":    "typescript",
		".vue":   "vue",
		".py":    "python",
		".java":  "java",
		".yaml":  "yaml",
		".yml":   "yaml",
		".json":  "json",
		".xml":   "xml",
		".sql":   "sql",
		".sh":    "bash",
		".md":    "markdown",
		".html":  "html",
		".css":   "css",
		".scss":  "scss",
		".less":  "less",
		".toml":  "toml",
		".ini":   "ini",
		".conf":  "conf",
		".rs":    "rust",
		".rb":    "ruby",
		".php":   "php",
		".swift": "swift",
		".kt":    "kotlin",
		".c":     "c",
		".cpp":   "cpp",
		".h":     "c",
		".hpp":   "cpp",
		".proto": "protobuf",
		".tf":    "hcl",
		".lua":   "lua",
	}
	if lang, ok := m[strings.ToLower(ext)]; ok {
		return lang
	}
	return ""
}

// isTextFile 简单判断是否为文本文件（按扩展名）
func isTextFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	textExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".vue": true, ".py": true,
		".java": true, ".yaml": true, ".yml": true, ".json": true, ".xml": true,
		".sql": true, ".sh": true, ".md": true, ".html": true, ".css": true,
		".scss": true, ".less": true, ".toml": true, ".ini": true, ".conf": true,
		".txt": true, ".log": true, ".env": true, ".mod": true, ".sum": true,
		".rs": true, ".rb": true, ".php": true, ".swift": true, ".kt": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true, ".proto": true,
		".tf": true, ".lua": true, ".tsx": true, ".jsx": true, ".svelte": true,
		".tpl": true, ".tmpl": true, ".cfg": true, ".properties": true,
	}
	return textExts[ext]
}
