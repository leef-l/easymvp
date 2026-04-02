package util

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

// ReplacePlaceholders 将路径中的 {app} 和 {module} 占位符替换为实际名称
func ReplacePlaceholders(path, app, module string) string {
	result := strings.ReplaceAll(path, "{app}", app)
	result = strings.ReplaceAll(result, "{module}", module)
	return result
}

// TemplateMapping 模板文件名 → 输出相对路径
type TemplateMapping struct {
	TplFile    string
	OutputPath string
}

// TemplateCache 模板缓存，避免重复解析
type TemplateCache struct {
	mu    sync.RWMutex
	cache map[string]*template.Template
}

// NewTemplateCache 创建模板缓存
func NewTemplateCache() *TemplateCache {
	return &TemplateCache{cache: make(map[string]*template.Template)}
}

// Get 获取或解析模板
func (tc *TemplateCache) Get(tplPath string) (*template.Template, error) {
	tc.mu.RLock()
	if t, ok := tc.cache[tplPath]; ok {
		tc.mu.RUnlock()
		return t, nil
	}
	tc.mu.RUnlock()

	tc.mu.Lock()
	defer tc.mu.Unlock()
	// 双重检查
	if t, ok := tc.cache[tplPath]; ok {
		return t, nil
	}
	t, err := template.ParseFiles(tplPath)
	if err != nil {
		return nil, err
	}
	tc.cache[tplPath] = t
	return t, nil
}

// GenerateFiles 通用文件生成逻辑
func GenerateFiles(mappings []TemplateMapping, tplDir, outDir, appName, moduleName string, force bool, data interface{}, cache ...*TemplateCache) ([]string, error) {
	var generated []string

	for _, m := range mappings {
		tplPath := filepath.Join(tplDir, m.TplFile)

		var tpl *template.Template
		var err error
		if len(cache) > 0 && cache[0] != nil {
			tpl, err = cache[0].Get(tplPath)
		} else {
			tpl, err = template.ParseFiles(tplPath)
		}
		if err != nil {
			return generated, fmt.Errorf("解析模板 %s 失败: %w", m.TplFile, err)
		}

		relPath := ReplacePlaceholders(m.OutputPath, appName, moduleName)
		outPath := filepath.Join(outDir, relPath)

		// Enhance 文件保护：force 模式下跳过 *_enhance.* 文件
		if force && isEnhanceFile(outPath) {
			fmt.Printf("  [保护] %s（enhance 文件不覆盖）\n", outPath)
			continue
		}

		if !force {
			if _, err := os.Stat(outPath); err == nil {
				fmt.Printf("  [跳过] %s（已存在，使用 --force 覆盖）\n", outPath)
				continue
			}
		}

		dir := filepath.Dir(outPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return generated, fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}

		var buf bytes.Buffer
		if err := tpl.Execute(&buf, data); err != nil {
			return generated, fmt.Errorf("渲染模板 %s 失败: %w", m.TplFile, err)
		}

		if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
			return generated, fmt.Errorf("写入文件 %s 失败: %w", outPath, err)
		}

		generated = append(generated, outPath)
		fmt.Printf("  [生成] %s\n", outPath)
	}

	return generated, nil
}

// GenerateToMemory 生成到内存（用于 dry-run diff 预览）
func GenerateToMemory(mappings []TemplateMapping, tplDir, outDir, appName, moduleName string, data interface{}, cache ...*TemplateCache) (map[string][]byte, error) {
	result := make(map[string][]byte)
	for _, m := range mappings {
		tplPath := filepath.Join(tplDir, m.TplFile)
		var tpl *template.Template
		var err error
		if len(cache) > 0 && cache[0] != nil {
			tpl, err = cache[0].Get(tplPath)
		} else {
			tpl, err = template.ParseFiles(tplPath)
		}
		if err != nil {
			return nil, fmt.Errorf("解析模板 %s 失败: %w", m.TplFile, err)
		}
		relPath := ReplacePlaceholders(m.OutputPath, appName, moduleName)
		outPath := filepath.Join(outDir, relPath)
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("渲染模板 %s 失败: %w", m.TplFile, err)
		}
		result[outPath] = buf.Bytes()
	}
	return result, nil
}

// isEnhanceFile 判断是否是 enhance 文件
func isEnhanceFile(path string) bool {
	base := filepath.Base(path)
	return strings.Contains(base, "_enhance.") || strings.Contains(base, "enhance.")
}
