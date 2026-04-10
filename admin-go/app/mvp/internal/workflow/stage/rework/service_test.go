package rework

import (
	"strings"
	"testing"
)

func TestClassifyOriginalTaskStatusForRework(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		status         string
		wantAllowReset bool
		wantConverged  bool
	}{
		{name: "failed can reset", status: "failed", wantAllowReset: true},
		{name: "escalated can reset", status: "escalated", wantAllowReset: true},
		{name: "pending already recovered", status: "pending", wantConverged: true},
		{name: "running already recovered", status: "running", wantConverged: true},
		{name: "completed already recovered", status: "completed", wantConverged: true},
		{name: "unknown blocks recovery", status: "bug_found"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotAllowReset, gotConverged := classifyOriginalTaskStatusForRework(tc.status)
			if gotAllowReset != tc.wantAllowReset || gotConverged != tc.wantConverged {
				t.Fatalf("classifyOriginalTaskStatusForRework(%q) = (%v, %v), want (%v, %v)",
					tc.status, gotAllowReset, gotConverged, tc.wantAllowReset, tc.wantConverged)
			}
		})
	}
}

func TestParseTaskPatch(t *testing.T) {
	t.Parallel()

	patch, err := parseTaskPatch(`{"description":"更新 README","affected_resources":["README.md"],"reason":"补齐说明"}`, "")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Description != "更新 README" || len(patch.AffectedResources) != 1 || patch.AffectedResources[0] != "README.md" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchSupportsFencedJSON(t *testing.T) {
	t.Parallel()

	patch, err := parseTaskPatch("修复建议如下：\n```json\n{\"description\":\"修复任务描述\",\"affected_resources\":[\"docs/plan.md\"],\"reason\":\"补齐交付说明\"}\n```", "")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Reason != "补齐交付说明" || patch.AffectedResources[0] != "docs/plan.md" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchSupportsEscapedFencedJSON(t *testing.T) {
	t.Parallel()

	content := "根据错误信息分析，这是一个修复方案。\\n\\n```json\\n" +
		"{\\\"task_repair\\\":{\\\"task_name\\\":\\\"backend-goframe-init\\\",\\\"description\\\":\\\"手工创建最小后端骨架\\\",\\\"affected_resources\\\":[\\\"backend/go.mod\\\",\\\"backend/main.go\\\"],\\\"reason\\\":\\\"消息内容被转义后仍应解析成功\\\"}}\\n```"
	patch, err := parseTaskPatch(content, "backend-goframe-init")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Description != "手工创建最小后端骨架" || patch.Reason != "消息内容被转义后仍应解析成功" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchSupportsNarrativeTaskRepairSample(t *testing.T) {
	t.Parallel()

	content := "我来分析这个任务失败的具体原因。\n\n## 失败原因分析\n\n这是说明文字。\n\n```json\n" +
		"{\"task_repair\":{\"task_name\":\"Implement GoFrame leaderboard domain\",\"description\":\"Build the backend game models, validation rules, file-backed leaderboard persistence, and deterministic sorting so submitted snake runs are safely stored and returned in rank order. 注意：本任务需在标准执行模式下运行，不依赖特殊e2e执行器配置。\",\"affected_resources\":[\"backend/internal/model/leaderboard.go\",\"backend/internal/service/leaderboard.go\",\"backend/internal/controller/leaderboard.go\",\"backend/internal/dao/leaderboard.go\",\"backend/manifest/config/config.yaml\"],\"reason\":\"失败原因：任务被标记为disabled_for_e2e执行模式，但执行器未注册该模式处理器，导致执行前校验失败。修复：移除对特殊e2e执行模式的依赖，确保使用标准执行模式；同时在description中显式声明执行模式要求，避免执行器模式匹配冲突。原任务代码实现目标不变，仅调整执行元数据兼容性。\"}}\n```"

	patch, err := parseTaskPatch(content, "Implement GoFrame leaderboard domain")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if !strings.Contains(patch.Description, "标准执行模式") {
		t.Fatalf("unexpected patch description: %+v", patch)
	}
	if len(patch.AffectedResources) != 5 {
		t.Fatalf("unexpected affected resources: %+v", patch.AffectedResources)
	}
}

func TestParseTaskPatchSupportsTaskRepairEnvelope(t *testing.T) {
	t.Parallel()

	patch, err := parseTaskPatch(`{"task_repair":{"task_name":"cli-root-init","description":"收缩根目录初始化范围","affected_resources":["package.json",".gitignore"],"reason":"pnpm-workspace.yaml 越界"}}`, "cli-root-init")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Description != "收缩根目录初始化范围" || len(patch.AffectedResources) != 2 {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchSupportsMatchingTaskPatchEnvelope(t *testing.T) {
	t.Parallel()

	patch, err := parseTaskPatch(`{"task_patches":[{"task_name":"backend-init","description":"忽略","reason":"无关"},{"task_name":"cli-root-init","description":"仅初始化根目录脚本","affected_resources":["package.json","scripts/dev.js"],"reason":"不能越界到 workspace 配置"}]}`, "cli-root-init")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Description != "仅初始化根目录脚本" || patch.Reason != "不能越界到 workspace 配置" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchRejectsMismatchedMultiTaskPatchEnvelope(t *testing.T) {
	t.Parallel()

	if _, err := parseTaskPatch(`{"task_patches":[{"task_name":"backend-init","description":"忽略","reason":"无关"},{"task_name":"frontend-init","description":"忽略","reason":"无关"}]}`, "cli-root-init"); err == nil {
		t.Fatal("expected mismatched task_patches to fail")
	}
}

func TestParseTaskPatchRejectsInvalidContent(t *testing.T) {
	t.Parallel()

	if _, err := parseTaskPatch("not json", ""); err == nil {
		t.Fatal("expected invalid content to fail")
	}
}
