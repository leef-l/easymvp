package rolecatalog

import "testing"

func TestMergeDefinitionsAllowsOverridesAndExtensions(t *testing.T) {
	t.Parallel()

	merged := mergeDefinitions([]Definition{
		{RoleType: "architect", DisplayName: "架构师", Color: "purple", PreferredLevels: []string{"max", "pro", "lite"}, Sort: 10},
	}, []Definition{
		{RoleType: "architect", DisplayName: "首席架构师", Color: "gold"},
		{RoleType: "tester", DisplayName: "测试师", Color: "blue", Sort: 20},
	})

	if len(merged) != 2 {
		t.Fatalf("expected 2 definitions, got %d", len(merged))
	}
	if merged[0].DisplayName != "首席架构师" || merged[0].Color != "gold" {
		t.Fatalf("architect override not applied: %+v", merged[0])
	}
	if merged[1].RoleType != "tester" || merged[1].DisplayName != "测试师" {
		t.Fatalf("tester extension missing: %+v", merged[1])
	}
}

func TestMergeDefinitionsAllowsDisablingAcceptanceJudge(t *testing.T) {
	t.Parallel()

	merged := mergeDefinitions([]Definition{
		{RoleType: "experience_reviewer", DisplayName: "体验评审师", AcceptanceJudge: true, Sort: 60},
	}, []Definition{
		{RoleType: "experience_reviewer", DisplayName: "体验评审师", AcceptanceJudge: false, Sort: 60},
	})

	if len(merged) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(merged))
	}
	if merged[0].AcceptanceJudge {
		t.Fatalf("expected acceptance judge to be disabled, got %+v", merged[0])
	}
}

func TestNormalizeForSaveValidatesAndKeepsBuiltinDefaults(t *testing.T) {
	t.Parallel()

	normalized, err := normalizeForSave([]Definition{
		{RoleType: "experience_reviewer", DisplayName: "体验评审师", AcceptanceJudge: false},
	})
	if err != nil {
		t.Fatalf("normalizeForSave returned error: %v", err)
	}
	if len(normalized) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(normalized))
	}
	if normalized[0].Color != "magenta" {
		t.Fatalf("expected builtin color fallback, got %+v", normalized[0])
	}
	if normalized[0].Sort != 60 {
		t.Fatalf("expected builtin sort fallback, got %+v", normalized[0])
	}
	if len(normalized[0].PreferredLevels) != 3 {
		t.Fatalf("expected builtin preferred levels, got %+v", normalized[0])
	}
	if normalized[0].AcceptanceJudge {
		t.Fatalf("expected explicit false to be preserved, got %+v", normalized[0])
	}
}

func TestNormalizeForSaveRejectsInvalidRoleTypeAndLevel(t *testing.T) {
	t.Parallel()

	if _, err := normalizeForSave([]Definition{
		{RoleType: "Bad-Role", DisplayName: "坏角色"},
	}); err == nil {
		t.Fatalf("expected invalid roleType error")
	}

	if _, err := normalizeForSave([]Definition{
		{RoleType: "qa_guardian", DisplayName: "质量守门人", PreferredLevels: []string{"expert"}},
	}); err == nil {
		t.Fatalf("expected invalid preferred level error")
	}
}
