package rework

import "testing"

func TestParseTaskPatch(t *testing.T) {
	t.Parallel()

	patch, err := parseTaskPatch(`{"description":"更新 README","affected_resources":["README.md"],"reason":"补齐说明"}`)
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Description != "更新 README" || len(patch.AffectedResources) != 1 || patch.AffectedResources[0] != "README.md" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchSupportsFencedJSON(t *testing.T) {
	t.Parallel()

	patch, err := parseTaskPatch("修复建议如下：\n```json\n{\"description\":\"修复任务描述\",\"affected_resources\":[\"docs/plan.md\"],\"reason\":\"补齐交付说明\"}\n```")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Reason != "补齐交付说明" || patch.AffectedResources[0] != "docs/plan.md" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchRejectsInvalidContent(t *testing.T) {
	t.Parallel()

	if _, err := parseTaskPatch("not json"); err == nil {
		t.Fatal("expected invalid content to fail")
	}
}
