package presetutil

import (
	"strings"
	"testing"
)

func TestPreferredRoleLevels(t *testing.T) {
	if got := strings.Join(PreferredRoleLevels("architect"), ","); got != "max,pro,lite" {
		t.Fatalf("unexpected architect order: %s", got)
	}
	if got := strings.Join(PreferredRoleLevels("coordinator"), ","); got != "lite,pro,max" {
		t.Fatalf("unexpected coordinator order: %s", got)
	}
	if got := strings.Join(PreferredRoleLevels("auditor"), ","); got != "pro,max,lite" {
		t.Fatalf("unexpected auditor order: %s", got)
	}
}

func TestBuildRoleSystemPromptFallsBackToCategory(t *testing.T) {
	prompt := BuildRoleSystemPrompt("game_dev", "implementer", "lite", "", "")
	if !strings.Contains(prompt, "游戏开发项目") {
		t.Fatalf("prompt should include category context, got: %s", prompt)
	}
	if !strings.Contains(prompt, "轻量级(lite)") {
		t.Fatalf("prompt should include level context, got: %s", prompt)
	}
}

func TestMergeSystemPromptKeepsBaseAndAppendsSupplement(t *testing.T) {
	merged := MergeSystemPrompt("分类提示词", "模型提示词")
	if !strings.Contains(merged, "分类提示词") {
		t.Fatalf("missing base prompt: %s", merged)
	}
	if !strings.Contains(merged, "模型能力补充") {
		t.Fatalf("missing supplement header: %s", merged)
	}
}

func TestBuildArchitectSystemPromptAppendsProjectContextAndFormat(t *testing.T) {
	base := BuildRoleSystemPrompt("product_design", "architect", "pro", "", "")
	prompt := BuildArchitectSystemPrompt("设计系统重构", "为企业后台重做信息架构和原型", "product_design", base)
	if !strings.Contains(prompt, "项目名称：设计系统重构") {
		t.Fatalf("missing project name: %s", prompt)
	}
	if !strings.Contains(prompt, "输出格式要求") {
		t.Fatalf("missing json suffix: %s", prompt)
	}
	if !strings.Contains(prompt, "禁止输出 <minimax:tool_call>") {
		t.Fatalf("missing tool-call guard: %s", prompt)
	}
}

func TestBuildArchitectSystemPromptAddsCodingBootstrapRules(t *testing.T) {
	base := BuildRoleSystemPrompt("software_dev", "architect", "max", "", "")
	prompt := BuildArchitectSystemPrompt("贪吃蛇", "贪吃蛇小游戏 react cli + goframe v2", "software_dev", base)
	if !strings.Contains(prompt, "脚手架 / 模板 / 官方工具初始化") {
		t.Fatalf("missing scaffold-first guidance: %s", prompt)
	}
	if !strings.Contains(prompt, "尽量并行") {
		t.Fatalf("missing parallel bootstrap guidance: %s", prompt)
	}
	if !strings.Contains(prompt, "affected_resources 只能写目录或文件相对路径") {
		t.Fatalf("missing affected_resources guard: %s", prompt)
	}
}

func TestBuildRoleSystemPromptAddsCodingImplementerRuntimeRules(t *testing.T) {
	prompt := BuildRoleSystemPrompt("software_dev", "implementer", "pro", "", "")
	if !strings.Contains(prompt, "优先使用这些方式快速初始化") {
		t.Fatalf("missing scaffold execution guidance: %s", prompt)
	}
	if !strings.Contains(prompt, "前端、后端、基础设施等独立根目录不要混成一次无边界改动") {
		t.Fatalf("missing scoped bootstrap guidance: %s", prompt)
	}
}
