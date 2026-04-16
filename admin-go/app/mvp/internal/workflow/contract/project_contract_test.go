package contract

import (
	"strings"
	"testing"
)

func TestInferFromTextsExtractsRequiredAndForbiddenTechnologies(t *testing.T) {
	t.Parallel()

	contract := InferFromTexts("不要 sqlite，必须使用 redis")
	if !strings.Contains(strings.Join(contract.ForbiddenTechnologies, ","), "sqlite") {
		t.Fatalf("expected sqlite to be forbidden: %+v", contract)
	}
	if !strings.Contains(strings.Join(contract.RequiredTechnologies, ","), "redis") {
		t.Fatalf("expected redis to be required: %+v", contract)
	}
}

func TestMergeObjectiveFieldsPreservesTechnicalContract(t *testing.T) {
	t.Parallel()

	raw := `{"technicalContract":{"forbiddenTechnologies":["sqlite"]},"qualityFloor":0.8}`
	merged, err := MergeObjectiveFields(raw, map[string]interface{}{"qualityFloor": 0.9, "tokenBudget": 1024})
	if err != nil {
		t.Fatalf("MergeObjectiveFields error: %v", err)
	}
	if !strings.Contains(merged, `"technicalContract"`) || !strings.Contains(merged, `"sqlite"`) {
		t.Fatalf("technical contract should be preserved: %s", merged)
	}
	if !strings.Contains(merged, `"tokenBudget":1024`) {
		t.Fatalf("new objective field should be merged: %s", merged)
	}
}

func TestDetectConflictsFindsForbiddenTechnology(t *testing.T) {
	t.Parallel()

	contract := &ProjectContract{ForbiddenTechnologies: []string{"sqlite"}}
	conflicts := DetectConflicts(contract, "配置 SQLite in-memory 测试数据库")
	if len(conflicts) != 1 || !strings.Contains(conflicts[0], "sqlite") {
		t.Fatalf("expected sqlite conflict, got %+v", conflicts)
	}
}

func TestDetectMissingRequiredFindsAbsentTechnology(t *testing.T) {
	t.Parallel()

	contract := &ProjectContract{RequiredTechnologies: []string{"redis"}}
	missing := DetectMissingRequired(contract, "前端页面初始化", "后端 API 框架")
	if len(missing) != 1 || missing[0] != "redis" {
		t.Fatalf("expected redis to be missing, got %+v", missing)
	}
}

func TestAppendConstraintBlock(t *testing.T) {
	t.Parallel()

	contract := &ProjectContract{
		ForbiddenTechnologies: []string{"sqlite"},
		RequiredTechnologies:  []string{"redis"},
	}
	got := AppendConstraintBlock("实现缓存层", contract)
	for _, fragment := range []string{"## 项目级硬约束", "禁止使用: sqlite", "必须优先使用: redis"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("constraint block missing %q: %s", fragment, got)
		}
	}
}
