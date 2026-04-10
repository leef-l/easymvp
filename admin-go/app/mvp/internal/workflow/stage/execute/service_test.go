package execute

import "testing"

func TestBuildCompletionTransitionPlan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name               string
		stageStatus        string
		workflowStatus     string
		workflowStage      string
		workflowStageRunID int64
		stageRunID         int64
		wantComplete       bool
		wantTransition     bool
	}{
		{
			name:               "running execute stage should complete and transition",
			stageStatus:        "running",
			workflowStatus:     "executing",
			workflowStage:      "execute",
			workflowStageRunID: 101,
			stageRunID:         101,
			wantComplete:       true,
			wantTransition:     true,
		},
		{
			name:               "completed stage should still transition after crash recovery",
			stageStatus:        "completed",
			workflowStatus:     "executing",
			workflowStage:      "execute",
			workflowStageRunID: 202,
			stageRunID:         202,
			wantTransition:     true,
		},
		{
			name:               "already advanced workflow should not re-enter transition",
			stageStatus:        "completed",
			workflowStatus:     "accepting",
			workflowStage:      "accept",
			workflowStageRunID: 303,
			stageRunID:         303,
		},
		{
			name:               "stale callback from old stage should be ignored",
			stageStatus:        "completed",
			workflowStatus:     "executing",
			workflowStage:      "execute",
			workflowStageRunID: 404,
			stageRunID:         405,
		},
		{
			name:               "failed stage should not progress",
			stageStatus:        "failed",
			workflowStatus:     "executing",
			workflowStage:      "execute",
			workflowStageRunID: 506,
			stageRunID:         506,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := buildCompletionTransitionPlan(
				tc.stageStatus,
				tc.workflowStatus,
				tc.workflowStage,
				tc.workflowStageRunID,
				tc.stageRunID,
			)
			if got.completeStage != tc.wantComplete || got.transition != tc.wantTransition {
				t.Fatalf("plan = %+v, want complete=%t transition=%t", got, tc.wantComplete, tc.wantTransition)
			}
		})
	}
}
