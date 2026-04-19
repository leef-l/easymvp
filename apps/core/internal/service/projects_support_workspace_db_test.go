package service

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	projectsv1 "github.com/leef-l/easymvp/apps/core/api/projects/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func TestBuildFallbackWorkspaceExplanationMirrorsInboxItems(t *testing.T) {
	t.Parallel()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{Id: "proj_123"},
	}
	actionInbox := []projectsv1.ActionInboxItem{
		{
			ItemID:            "issue_1",
			Title:             "First blocker",
			Severity:          "error",
			RecommendedAction: "open_repair_draft",
			TargetID:          "repair_1",
		},
		{
			ItemID:            "issue_2",
			Title:             "Second blocker",
			Severity:          "warning",
			RecommendedAction: "open_acceptance_center",
			TargetID:          "acc_1",
		},
		{
			ItemID:            "issue_3",
			Title:             "Third blocker",
			Severity:          "info",
			RecommendedAction: "open_project_plan",
			TargetID:          "proj_123",
		},
	}

	view := buildFallbackWorkspaceExplanation(data, actionInbox)

	if len(view.TopBlockers) != 3 {
		t.Fatalf("unexpected blocker count: got %d want %d", len(view.TopBlockers), 3)
	}
	if len(view.RecommendedActions) != 3 {
		t.Fatalf("unexpected recommended action count: got %d want %d", len(view.RecommendedActions), 3)
	}
	if view.RecommendedActions[2].ActionKey != "open_project_plan" {
		t.Fatalf("unexpected third action key: got %s want %s", view.RecommendedActions[2].ActionKey, "open_project_plan")
	}
	if view.RecommendedActions[2].Label != "Third blocker" {
		t.Fatalf("unexpected third action label: got %s want %s", view.RecommendedActions[2].Label, "Third blocker")
	}
}

func TestBuildFallbackWorkspaceExplanationUsesDefaultActionWhenInboxEmpty(t *testing.T) {
	t.Parallel()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{Id: "proj_456"},
	}

	view := buildFallbackWorkspaceExplanation(data, nil)

	if len(view.TopBlockers) != 1 || view.TopBlockers[0] != "No blocking issue is currently detected." {
		t.Fatalf("unexpected blockers: %#v", view.TopBlockers)
	}
	if len(view.RecommendedActions) != 1 {
		t.Fatalf("unexpected default action count: got %d want %d", len(view.RecommendedActions), 1)
	}
	if view.RecommendedActions[0].ActionKey != "open_project_plan" {
		t.Fatalf("unexpected default action key: got %s want %s", view.RecommendedActions[0].ActionKey, "open_project_plan")
	}
	if view.RecommendedActions[0].DeepLink != "proj_456" {
		t.Fatalf("unexpected default action deep link: got %s want %s", view.RecommendedActions[0].DeepLink, "proj_456")
	}
}

func TestBuildProjectStageProgressDefaultsUnknownStatusToDesign(t *testing.T) {
	t.Parallel()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{
			Id:        "proj_stage_unknown",
			Status:    "  ",
			CreatedAt: "2026-04-19T10:00:00Z",
			UpdatedAt: "2026-04-19T10:10:00Z",
		},
	}

	items := buildProjectStageProgress(data)

	if len(items) != 6 {
		t.Fatalf("unexpected stage count: got %d want %d", len(items), 6)
	}
	if items[0].StageKey != "design" || items[0].Status != "running" {
		t.Fatalf("unexpected first stage state: %#v", items[0])
	}
	if items[0].ActiveItemTitle != "Project is progressing" {
		t.Fatalf("unexpected active item title: got %s want %s", items[0].ActiveItemTitle, "Project is progressing")
	}
	for _, item := range items[1:] {
		if item.Status != "pending" {
			t.Fatalf("unexpected non-design stage state: %#v", item)
		}
	}
}

func TestBuildProjectStageProgressBlocksAcceptanceStage(t *testing.T) {
	t.Parallel()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{
			Id:        "proj_acceptance",
			Status:    "acceptance",
			CreatedAt: "2026-04-19T10:00:00Z",
			UpdatedAt: "2026-04-19T10:20:00Z",
		},
		AcceptanceIssues: []entity.AcceptanceIssues{
			{Id: "issue_1", Blocking: 1, Summary: "First blocker"},
			{Id: "issue_2", Blocking: 1, Summary: "Second blocker"},
		},
	}

	items := buildProjectStageProgress(data)

	if items[4].StageKey != "acceptance" || items[4].Status != "blocked" {
		t.Fatalf("unexpected acceptance stage state: %#v", items[4])
	}
	if items[4].BlockingIssueCnt != 2 {
		t.Fatalf("unexpected acceptance blocking count: got %d want %d", items[4].BlockingIssueCnt, 2)
	}
	for _, index := range []int{0, 1, 2, 3} {
		if items[index].Status != "completed" {
			t.Fatalf("unexpected completed stage state at index %d: %#v", index, items[index])
		}
	}
	if items[5].Status != "pending" {
		t.Fatalf("unexpected complete stage state: %#v", items[5])
	}
}

func TestBuildProjectAcceptanceCoverageFunctionalPassWithIssuesReducesEvidence(t *testing.T) {
	t.Parallel()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{
			Id:               "proj_cov",
			ProjectCategory:  "web",
			ProductionStatus: "not_ready",
		},
		LatestAcceptanceRun: &entity.AcceptanceRuns{
			Id:               "acc_run_1",
			FunctionalStatus: "functional_passed",
			ProductionStatus: "not_ready",
		},
		AcceptanceIssues: []entity.AcceptanceIssues{
			{Id: "issue_1", Blocking: 1},
			{Id: "issue_2", Blocking: 0},
		},
	}

	view := buildProjectAcceptanceCoverage(data)

	if view.RequiredSurfaces != 4 || view.RequiredJourneys != 6 || view.EvidenceRequired != 5 {
		t.Fatalf("unexpected requirements: %#v", view)
	}
	if view.CoveredSurfaces != 2 || view.CoveredJourneys != 3 {
		t.Fatalf("unexpected functional coverage: %#v", view)
	}
	if view.EvidenceReady != 0 {
		t.Fatalf("unexpected evidence count after issue reduction: got %d want %d", view.EvidenceReady, 0)
	}
	if view.ProductionPassed {
		t.Fatalf("expected production to remain not passed: %#v", view)
	}
}

func TestBuildProjectActionInboxPrioritizesRepairDraftAndRespectsLimit(t *testing.T) {
	t.Parallel()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{Id: "proj_inbox"},
		RepairDraft: &repairPlanDraftRecord{
			ID: "repair_1",
		},
		AcceptanceIssues: []entity.AcceptanceIssues{
			{Id: "issue_1", Blocking: 1, Summary: "Blocking issue", IssueKind: "runtime"},
			{Id: "issue_2", Blocking: 0, Summary: "Release note issue", IssueKind: "release"},
		},
		Tasks: []entity.DomainTasks{
			{Id: "task_1", Name: "Manual review A", ManualReviewRequired: 1, RiskLevel: "high", Status: "failed"},
			{Id: "task_2", Name: "Manual review B", ManualReviewRequired: 1, RiskLevel: "low", Status: "blocked"},
			{Id: "task_3", Name: "Manual review C", ManualReviewRequired: 1, RiskLevel: "low", Status: "running"},
			{Id: "task_4", Name: "Manual review D", ManualReviewRequired: 1, RiskLevel: "low", Status: "running"},
			{Id: "task_5", Name: "Manual review E", ManualReviewRequired: 1, RiskLevel: "low", Status: "running"},
			{Id: "task_6", Name: "Manual review F", ManualReviewRequired: 1, RiskLevel: "low", Status: "running"},
			{Id: "task_7", Name: "Manual review G", ManualReviewRequired: 1, RiskLevel: "low", Status: "running"},
			{Id: "task_8", Name: "Manual review H", ManualReviewRequired: 1, RiskLevel: "low", Status: "running"},
			{Id: "task_9", Name: "Manual review I", ManualReviewRequired: 1, RiskLevel: "low", Status: "running"},
			{Id: "task_10", Name: "Manual review J", ManualReviewRequired: 1, RiskLevel: "low", Status: "running"},
		},
	}

	items := buildProjectActionInbox(data)

	if len(items) != projectWorkspaceInboxLimit {
		t.Fatalf("unexpected inbox count: got %d want %d", len(items), projectWorkspaceInboxLimit)
	}
	if items[0].RecommendedAction != "open_repair_draft" || items[0].TargetID != "repair_1" {
		t.Fatalf("unexpected first inbox item: %#v", items[0])
	}
	if items[1].RecommendedAction != "open_acceptance_issue" {
		t.Fatalf("unexpected second inbox action: %#v", items[1])
	}
	if items[2].RecommendedAction != "open_acceptance_center" {
		t.Fatalf("unexpected third inbox action: %#v", items[2])
	}
}

func TestBuildProjectActionInboxIncludesManualReleaseWhenNeeded(t *testing.T) {
	t.Parallel()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{Id: "proj_release"},
		LatestAcceptanceRun: &entity.AcceptanceRuns{
			Id:                    "acc_run_release",
			ManualReleaseRequired: 1,
		},
		AcceptanceIssues: []entity.AcceptanceIssues{
			{Id: "issue_1", Blocking: 1, Summary: "Blocking issue", IssueKind: "runtime"},
		},
	}

	items := buildProjectActionInbox(data)

	foundManualRelease := false
	for _, item := range items {
		if item.ItemID == "manual_release_acc_run_release" {
			foundManualRelease = true
			if item.RecommendedAction != "open_acceptance_center" {
				t.Fatalf("unexpected manual release action: %#v", item)
			}
		}
	}
	if !foundManualRelease {
		t.Fatalf("expected manual release item to be included, got %#v", items)
	}
}

func TestBuildProjectActionInboxIncludesUnsupportedAndDeniedBindings(t *testing.T) {
	t.Parallel()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{Id: "proj_runtime_attention"},
		Tasks: []entity.DomainTasks{
			{Id: "task_unsupported", Name: "Collect replay evidence"},
			{Id: "task_denied", Name: "Open protected repository"},
		},
		RunBindings: []entity.BrainRunBindings{
			{Id: "bind_unsupported", TaskId: "task_unsupported", RunStatus: "run_unsupported"},
			{Id: "bind_denied", TaskId: "task_denied", RunStatus: "run_denied"},
		},
	}

	items := buildProjectActionInbox(data)
	if len(items) != 2 {
		t.Fatalf("unexpected inbox count: got %d want %d", len(items), 2)
	}
	if items[0].Severity != "warning" || items[0].RecommendedAction != "open_task_review" || items[0].TargetID != "task_unsupported" {
		t.Fatalf("unexpected unsupported binding inbox item: %#v", items[0])
	}
	if items[1].Severity != "error" || !items[1].IsBlocking || items[1].TargetID != "task_denied" {
		t.Fatalf("unexpected denied binding inbox item: %#v", items[1])
	}
}

func TestBuildProjectLiveActivityRequiresActionForUnsupportedAndDeniedBindings(t *testing.T) {
	t.Parallel()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{Id: "proj_runtime_activity"},
		Tasks: []entity.DomainTasks{
			{Id: "task_unsupported", Name: "Collect replay evidence"},
			{Id: "task_denied", Name: "Open protected repository"},
		},
		RunBindings: []entity.BrainRunBindings{
			{Id: "bind_unsupported", TaskId: "task_unsupported", RunStatus: "run_unsupported", LastSyncAt: "2026-04-19T10:00:00Z"},
			{Id: "bind_denied", TaskId: "task_denied", RunStatus: "run_denied", LastSyncAt: "2026-04-19T10:01:00Z"},
		},
	}

	items := buildProjectLiveActivity(data)
	if len(items) != 2 {
		t.Fatalf("unexpected live activity count: got %d want %d", len(items), 2)
	}
	if items[0].EventType != "brain_run_unsupported" || !items[0].RequiresAction {
		t.Fatalf("unexpected unsupported binding live activity: %#v", items[0])
	}
	if items[1].EventType != "brain_run_denied" || !items[1].RequiresAction {
		t.Fatalf("unexpected denied binding live activity: %#v", items[1])
	}
}

func TestBuildProjectLiveActivityPrefersAuditLogsAndFallsBackToBindings(t *testing.T) {
	t.Parallel()

	withLogs := &projectWorkspaceAggregate{
		Project: entity.Projects{Id: "proj_logs", CreatedAt: "2026-04-19T10:00:00Z", UpdatedAt: "2026-04-19T10:10:00Z"},
		AuditLogs: []entity.AuditLogs{
			{Id: "log_1", EventType: "", ActorKind: "", Summary: "", CreatedAt: "2026-04-19T10:05:00Z"},
			{Id: "log_2", EventType: "project.updated", ActorKind: "brain", Summary: "Updated", CreatedAt: "2026-04-19T10:06:00Z"},
		},
		RunBindings: []entity.BrainRunBindings{
			{Id: "bind_ignored", RunStatus: "failed"},
		},
	}

	logItems := buildProjectLiveActivity(withLogs)

	if len(logItems) != 2 {
		t.Fatalf("unexpected audit log activity count: got %d want %d", len(logItems), 2)
	}
	if logItems[0].EventType != "audit_event" || logItems[0].SourceBrain != "system" || logItems[0].Title != "Project activity updated" {
		t.Fatalf("unexpected first audit log fallback item: %#v", logItems[0])
	}
	if logItems[1].EventType != "project.updated" || logItems[1].SourceBrain != "brain" {
		t.Fatalf("unexpected second audit log item: %#v", logItems[1])
	}

	withBindings := &projectWorkspaceAggregate{
		Project: entity.Projects{Id: "proj_bindings"},
		Tasks: []entity.DomainTasks{
			{Id: "task_1", Name: "Execution task"},
		},
		RunBindings: []entity.BrainRunBindings{
			{Id: "bind_1", TaskId: "task_1", BrainKind: "coder", RunStatus: "failed", CreatedAt: "2026-04-19T10:05:00Z"},
			{Id: "bind_2", TaskId: "task_missing", BrainKind: "reviewer", RunStatus: "running", CreatedAt: "2026-04-19T10:06:00Z"},
		},
	}

	bindingItems := buildProjectLiveActivity(withBindings)

	if len(bindingItems) != 2 {
		t.Fatalf("unexpected binding activity count: got %d want %d", len(bindingItems), 2)
	}
	if bindingItems[0].Title != "Execution task" || !bindingItems[0].RequiresAction || bindingItems[0].EventType != "brain_run_failed" {
		t.Fatalf("unexpected first binding item: %#v", bindingItems[0])
	}
	if bindingItems[1].Title != "Brain run running" || bindingItems[1].RequiresAction {
		t.Fatalf("unexpected second binding item: %#v", bindingItems[1])
	}
}

func TestBuildProjectWorkspaceExplanationUsesDeniedFallbackWhenRuntimePolicyBlocks(t *testing.T) {
	previous := localEasyMVPBrain
	localEasyMVPBrain = &workspaceExplanationBrainStub{
		workspaceErr: gerror.New("BRN_004: brain execution was denied by runtime policy: policy rejected the request"),
	}
	defer func() {
		localEasyMVPBrain = previous
	}()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{Id: "proj_denied"},
		Tasks: []entity.DomainTasks{
			{Id: "task_denied", Name: "Open protected repository"},
		},
		RunBindings: []entity.BrainRunBindings{
			{Id: "bind_denied", TaskId: "task_denied", RunStatus: "run_denied"},
		},
	}

	view := buildProjectWorkspaceExplanation(context.Background(), data, projectsv1.ProjectSnapshot{}, nil, nil, nil, projectsv1.AcceptanceCoverage{})
	if !strings.Contains(strings.ToLower(view.Headline), "runtime policy") {
		t.Fatalf("expected denied fallback headline, got %#v", view)
	}
	if len(view.TopBlockers) == 0 || !strings.Contains(strings.ToLower(view.TopBlockers[0]), "denied") {
		t.Fatalf("expected denied blocker details, got %#v", view)
	}
	if len(view.RecommendedActions) == 0 || view.RecommendedActions[0].ActionKey != "open_task_review" {
		t.Fatalf("expected task review action, got %#v", view)
	}
}

func TestBuildProjectWorkspaceExplanationUsesUnsupportedFallbackWhenCapabilityMissing(t *testing.T) {
	previous := localEasyMVPBrain
	localEasyMVPBrain = &workspaceExplanationBrainStub{
		workspaceErr: gerror.New("BRN_004: brain execution reported unsupported capability: tool_unsupported"),
	}
	defer func() {
		localEasyMVPBrain = previous
	}()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{Id: "proj_unsupported"},
		Tasks: []entity.DomainTasks{
			{Id: "task_unsupported", Name: "Collect replay evidence"},
		},
		RunBindings: []entity.BrainRunBindings{
			{Id: "bind_unsupported", TaskId: "task_unsupported", RunStatus: "run_unsupported"},
		},
	}

	view := buildProjectWorkspaceExplanation(context.Background(), data, projectsv1.ProjectSnapshot{}, nil, nil, nil, projectsv1.AcceptanceCoverage{})
	if !strings.Contains(strings.ToLower(view.Headline), "runtime capability") {
		t.Fatalf("expected unsupported fallback headline, got %#v", view)
	}
	if len(view.TopBlockers) == 0 || !strings.Contains(strings.ToLower(view.TopBlockers[0]), "unsupported") {
		t.Fatalf("expected unsupported blocker details, got %#v", view)
	}
	if len(view.ExplainLinks) == 0 || view.ExplainLinks[0] != "runtime" {
		t.Fatalf("expected runtime explain link, got %#v", view)
	}
}

func TestMapProjectStatusToWorkspaceStageNormalizesCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  string
		expect string
	}{
		{input: " Acceptance ", expect: "acceptance"},
		{input: "COMPLETED", expect: "complete"},
		{input: "RuNnInG", expect: "execute"},
	}

	for _, tt := range tests {
		if got := mapProjectStatusToWorkspaceStage(tt.input); got != tt.expect {
			t.Fatalf("unexpected stage for %q: got %s want %s", tt.input, got, tt.expect)
		}
	}
}
