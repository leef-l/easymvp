package service

import (
	"testing"

	projectsv1 "github.com/leef-l/easymvp/apps/core/api/projects/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func TestBuildWorkspaceVerificationResultPrefersRepairWhenBlockingIssuesExist(t *testing.T) {
	t.Parallel()

	view := buildWorkspaceVerificationResult(&projectWorkspaceAggregate{
		Project: entity.Projects{
			Id:              "proj_workspace_failed",
			ProjectCategory: "web",
			UpdatedAt:       "2026-04-20T00:00:00Z",
		},
		AcceptanceIssues: []entity.AcceptanceIssues{
			{Id: "issue_1", IssueKind: "api_regression", Blocking: 1, Summary: "API regression blocks release"},
		},
	}, projectsv1.AcceptanceCoverage{
		ProductionPassed: true,
		EvidenceReady:    2,
	})

	if view.Status != "failed" {
		t.Fatalf("workspace verification status = %q, want failed", view.Status)
	}
	if view.Decision != "repair_required" {
		t.Fatalf("workspace verification decision = %q, want repair_required", view.Decision)
	}
	if len(view.FailedChecks) != 1 || view.FailedChecks[0] != "api_regression" {
		t.Fatalf("workspace failed checks = %#v", view.FailedChecks)
	}
}

func TestBuildWorkspaceCompletionVerdictRequiresManualReviewForManualRelease(t *testing.T) {
	t.Parallel()

	verdict := buildWorkspaceCompletionVerdict(&projectWorkspaceAggregate{
		Project: entity.Projects{
			Id:              "proj_workspace_manual_release",
			ProjectCategory: "web",
			UpdatedAt:       "2026-04-20T00:00:00Z",
		},
		LatestAcceptanceRun: &entity.AcceptanceRuns{
			Id:                    "run_manual_release",
			Status:                "completed",
			FunctionalStatus:      "functional_passed",
			ProductionStatus:      "production_passed",
			ManualReleaseRequired: 1,
			FinishedAt:            "2026-04-20T00:05:00Z",
		},
	}, projectsv1.AcceptanceCoverage{
		ProductionPassed: true,
		EvidenceReady:    5,
	})

	if verdict.Decision != "manual_review" {
		t.Fatalf("workspace completion decision = %q, want manual_review", verdict.Decision)
	}
	if verdict.NextAction != "apply_manual_release" {
		t.Fatalf("workspace completion next action = %q, want apply_manual_release", verdict.NextAction)
	}
	if verdict.Completed {
		t.Fatal("workspace completion should not be completed while manual release is pending")
	}
}

func TestBuildWorkspaceCompletionVerdictCompletesCleanPass(t *testing.T) {
	t.Parallel()

	verdict := buildWorkspaceCompletionVerdict(&projectWorkspaceAggregate{
		Project: entity.Projects{
			Id:              "proj_workspace_complete",
			ProjectCategory: "web",
			UpdatedAt:       "2026-04-20T00:00:00Z",
		},
		LatestAcceptanceRun: &entity.AcceptanceRuns{
			Id:               "run_workspace_complete",
			Status:           "completed",
			FunctionalStatus: "functional_passed",
			ProductionStatus: "production_passed",
			FinishedAt:       "2026-04-20T00:06:00Z",
		},
	}, projectsv1.AcceptanceCoverage{
		ProductionPassed: true,
		EvidenceReady:    5,
	})

	if verdict.Decision != "complete" {
		t.Fatalf("workspace completion decision = %q, want complete", verdict.Decision)
	}
	if verdict.FinalStatus != "completed" || !verdict.Completed {
		t.Fatalf("workspace completion verdict = %#v, want completed", verdict)
	}
}
