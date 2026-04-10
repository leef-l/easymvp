package rework

import "testing"

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
