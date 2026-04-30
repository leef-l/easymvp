package service

import (
	"testing"
)

// ---------------------------------------------------------------------------
// DefaultAcceptanceCriteria tests
// ---------------------------------------------------------------------------

func TestDefaultAcceptanceCriteria_WebApp(t *testing.T) {
	criteria := DefaultAcceptanceCriteria("web_app")
	if len(criteria.Layers) == 0 {
		t.Fatal("expected non-empty layers for web_app")
	}

	// web_app should have unit_test, integration_test, e2e_test, security_test, performance_test
	layerNames := extractLayerNames(criteria)
	expectedLayers := []string{"unit_test", "integration_test", "e2e_test", "security_test", "performance_test"}
	for _, expected := range expectedLayers {
		if !containsString(layerNames, expected) {
			t.Fatalf("web_app criteria missing expected layer %q", expected)
		}
	}

	// unit_test and integration_test should be required
	for _, layer := range criteria.Layers {
		if layer.Name == "unit_test" && !layer.Required {
			t.Fatal("unit_test layer should be required for web_app")
		}
		if layer.Name == "integration_test" && !layer.Required {
			t.Fatal("integration_test layer should be required for web_app")
		}
	}
}

func TestDefaultAcceptanceCriteria_Game(t *testing.T) {
	criteria := DefaultAcceptanceCriteria("game")
	if len(criteria.Layers) == 0 {
		t.Fatal("expected non-empty layers for game")
	}

	// Game should have performance_test as required
	for _, layer := range criteria.Layers {
		if layer.Name == "performance_test" && !layer.Required {
			t.Fatal("performance_test should be required for game")
		}
	}
}

func TestDefaultAcceptanceCriteria_API(t *testing.T) {
	criteria := DefaultAcceptanceCriteria("api")
	if len(criteria.Layers) == 0 {
		t.Fatal("expected non-empty layers for api")
	}

	// API should have security_test as required
	for _, layer := range criteria.Layers {
		if layer.Name == "security_test" && !layer.Required {
			t.Fatal("security_test should be required for api")
		}
	}
}

func TestDefaultAcceptanceCriteria_DataPipeline(t *testing.T) {
	criteria := DefaultAcceptanceCriteria("data_pipeline")
	if len(criteria.Layers) == 0 {
		t.Fatal("expected non-empty layers for data_pipeline")
	}

	// e2e_test should be required for data_pipeline
	for _, layer := range criteria.Layers {
		if layer.Name == "e2e_test" && !layer.Required {
			t.Fatal("e2e_test should be required for data_pipeline")
		}
	}
}

func TestDefaultAcceptanceCriteria_UnknownType(t *testing.T) {
	criteria := DefaultAcceptanceCriteria("unknown_type")
	if len(criteria.Layers) == 0 {
		t.Fatal("expected non-empty layers for unknown type (should fall back to general)")
	}

	// general type should have unit_test and integration_test as required
	for _, layer := range criteria.Layers {
		if layer.Name == "unit_test" && !layer.Required {
			t.Fatal("unit_test should be required for general")
		}
	}
}

// ---------------------------------------------------------------------------
// EvaluateAcceptanceResults tests
// ---------------------------------------------------------------------------

func TestEvaluateAcceptanceResults_AllPassed(t *testing.T) {
	criteria := AcceptanceCriteria{
		Layers: []AcceptanceLayer{
			{
				Name:     "unit_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "core_logic", Weight: 20},
					{Name: "edge_cases", Weight: 10},
				},
			},
			{
				Name:     "integration_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "api_test", Weight: 20},
				},
			},
		},
	}

	results := []LayerResult{
		{LayerName: "unit_test", Passed: true, Score: 30},
		{LayerName: "integration_test", Passed: true, Score: 20},
	}

	verdict := EvaluateAcceptanceResults(criteria, results)
	if !verdict.Passed {
		t.Fatalf("expected passed verdict, got failed: %s", verdict.Summary)
	}
	if verdict.TotalScore != 50 {
		t.Fatalf("expected total score 50, got %d", verdict.TotalScore)
	}
	if verdict.MaxScore != 50 {
		t.Fatalf("expected max score 50, got %d", verdict.MaxScore)
	}
}

func TestEvaluateAcceptanceResults_RequiredLayerFailed(t *testing.T) {
	criteria := AcceptanceCriteria{
		Layers: []AcceptanceLayer{
			{
				Name:     "unit_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "core_logic", Weight: 20},
				},
			},
		},
	}

	results := []LayerResult{
		{LayerName: "unit_test", Passed: false, Score: 5},
	}

	verdict := EvaluateAcceptanceResults(criteria, results)
	if verdict.Passed {
		t.Fatal("expected failed verdict when required layer fails")
	}
	if verdict.Summary != "Required layer unit_test failed" {
		t.Fatalf("unexpected summary: %s", verdict.Summary)
	}
}

func TestEvaluateAcceptanceResults_RequiredLayerMissing(t *testing.T) {
	criteria := AcceptanceCriteria{
		Layers: []AcceptanceLayer{
			{
				Name:     "unit_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "core_logic", Weight: 20},
				},
			},
		},
	}

	// No results at all
	verdict := EvaluateAcceptanceResults(criteria, nil)
	if verdict.Passed {
		t.Fatal("expected failed verdict when required layer has no results")
	}
	if verdict.Summary != "Required layer unit_test has no results" {
		t.Fatalf("unexpected summary: %s", verdict.Summary)
	}
}

func TestEvaluateAcceptanceResults_OptionalLayerMissing(t *testing.T) {
	criteria := AcceptanceCriteria{
		Layers: []AcceptanceLayer{
			{
				Name:     "unit_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "core_logic", Weight: 20},
				},
			},
			{
				Name:     "performance_test",
				Required: false,
				Criteria: []Criterion{
					{Name: "speed", Weight: 10},
				},
			},
		},
	}

	results := []LayerResult{
		{LayerName: "unit_test", Passed: true, Score: 20},
		// performance_test missing - should be OK since optional
	}

	verdict := EvaluateAcceptanceResults(criteria, results)
	// 20 out of 30 = 66.7% which is below 70% threshold
	// so this should actually fail due to score threshold
	if verdict.Passed {
		t.Fatal("expected failed verdict due to score below 70%")
	}
}

func TestEvaluateAcceptanceResults_ScoreThreshold(t *testing.T) {
	criteria := AcceptanceCriteria{
		Layers: []AcceptanceLayer{
			{
				Name:     "unit_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "core_logic", Weight: 100},
				},
			},
		},
	}

	// Score 69 out of 100 = 69%, below 70%
	results := []LayerResult{
		{LayerName: "unit_test", Passed: true, Score: 69},
	}
	verdict := EvaluateAcceptanceResults(criteria, results)
	if verdict.Passed {
		t.Fatal("expected failed verdict for score 69%")
	}

	// Score 70 out of 100 = 70%, at threshold
	results = []LayerResult{
		{LayerName: "unit_test", Passed: true, Score: 70},
	}
	verdict = EvaluateAcceptanceResults(criteria, results)
	if !verdict.Passed {
		t.Fatalf("expected passed verdict for score 70%%: %s", verdict.Summary)
	}
}

func TestEvaluateAcceptanceResults_EmptyCriteria(t *testing.T) {
	criteria := AcceptanceCriteria{Layers: nil}
	verdict := EvaluateAcceptanceResults(criteria, nil)
	// No layers, maxScore=0, should pass.
	if !verdict.Passed {
		t.Fatalf("expected passed verdict for empty criteria: %s", verdict.Summary)
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func extractLayerNames(criteria AcceptanceCriteria) []string {
	names := make([]string, len(criteria.Layers))
	for i, layer := range criteria.Layers {
		names[i] = layer.Name
	}
	return names
}

func containsString(ss []string, target string) bool {
	for _, s := range ss {
		if s == target {
			return true
		}
	}
	return false
}
