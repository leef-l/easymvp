package chat

import (
	"context"
	"fmt"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/regression"
)

// RegressionScenarios 返回内部回归样例清单。
func (c *cWorkflow) RegressionScenarios(ctx context.Context, req *v1.WorkflowRegressionScenariosReq) (res *v1.WorkflowRegressionScenariosRes, err error) {
	manifestPath, err := regression.ResolveManifestPath()
	if err != nil {
		return &v1.WorkflowRegressionScenariosRes{Scenarios: []v1.RegressionScenarioItem{}}, nil
	}

	manifest, err := regression.LoadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("读取回归样例清单失败: %w", err)
	}

	readyCount, plannedCount := summarizeRegressionScenarioStatus(manifest.Scenarios)
	result := &v1.WorkflowRegressionScenariosRes{
		Version:      manifest.Version,
		UpdatedAt:    manifest.UpdatedAt,
		ManifestPath: manifestPath,
		ReadyCount:   readyCount,
		PlannedCount: plannedCount,
		Valid:        true,
		Message:      fmt.Sprintf("回归样例清单校验通过：ready=%d planned=%d", readyCount, plannedCount),
		Scenarios:    manifest.Scenarios,
	}

	report, validateErr := regression.ValidateManifest(manifestPath)
	if validateErr != nil {
		result.Valid = false
		result.Message = "回归样例清单校验失败: " + validateErr.Error()
		return result, nil
	}
	result.ReadyCount = report.ReadyCount
	result.PlannedCount = report.PlannedCount
	result.Message = fmt.Sprintf("回归样例清单校验通过：ready=%d planned=%d", report.ReadyCount, report.PlannedCount)

	return result, nil
}

func summarizeRegressionScenarioStatus(scenarios []v1.RegressionScenarioItem) (readyCount int, plannedCount int) {
	for _, scenario := range scenarios {
		switch scenario.Status {
		case "ready":
			readyCount++
		case "planned":
			plannedCount++
		}
	}
	return readyCount, plannedCount
}
