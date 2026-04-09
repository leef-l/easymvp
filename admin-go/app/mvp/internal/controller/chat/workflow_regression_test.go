package chat

import (
	"testing"

	v1 "easymvp/app/mvp/api/mvp/v1"
)

func TestSummarizeRegressionScenarioStatus(t *testing.T) {
	t.Parallel()

	ready, planned := summarizeRegressionScenarioStatus([]v1.RegressionScenarioItem{
		{ScenarioCode: "a", Status: "ready"},
		{ScenarioCode: "b", Status: "planned"},
		{ScenarioCode: "c", Status: "ready"},
		{ScenarioCode: "d", Status: "unknown"},
	})
	if ready != 2 || planned != 1 {
		t.Fatalf("unexpected summary: ready=%d planned=%d", ready, planned)
	}
}
