package executor

import (
	"strings"
	"testing"
)

func TestBuildStrictAiderTaskPromptIncludesAllowedFilesAndGuards(t *testing.T) {
	prompt := buildStrictAiderTaskPrompt("实现 /health 接口", []string{"demo/main.go", "README.md"})
	for _, want := range []string{
		"实现 /health 接口",
		"demo/main.go",
		"README.md",
		"运行方式：",
		"验证：",
		"只允许创建或修改以下路径",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %s", want, prompt)
		}
	}
}

func TestBuildStrictAiderTaskPromptWithoutFilesForbidsExpansion(t *testing.T) {
	prompt := buildStrictAiderTaskPrompt("修复任务", nil)
	if !strings.Contains(prompt, "未声明允许写入的文件路径") {
		t.Fatalf("expected no-file guard in prompt: %s", prompt)
	}
}
