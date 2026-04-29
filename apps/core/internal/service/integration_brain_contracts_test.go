//go:build integration

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
)

// skipIfBrainUnreachable skips the test when the local brain sidecar is not
// responding on the expected health endpoint.
func skipIfBrainUnreachable(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:7701/v1/health")
	if err != nil {
		t.Skipf("brain unreachable: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("brain health returned status %d", resp.StatusCode)
	}
}

// skipIfUnsupported skips the test when the brain reports the contract as
// unsupported rather than failing the test.
func skipIfUnsupported(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "unsupported") {
		t.Skipf("contract unsupported by brain: %v", err)
	}
}

func TestIntegration_CallPlanReview(t *testing.T) {
	t.Parallel()
	skipIfBrainUnreachable(t)

	svc := &sEasyMVPBrain{}
	ctx := context.Background()

	input := braincontracts.PlanReviewInput{
		PlanDraftID:            "draft_integ_001",
		PlanDraftVersion:       1,
		PlanDraftJSON:          json.RawMessage(`{"title":"integration test plan","steps":["step1"]}`),
		ProjectCategory:        "web_app",
		CategoryProfileVersion: 1,
		CategoryProfileJSON:    json.RawMessage(`{"category":"web_app","risk":"low"}`),
	}

	envelope, result, err := svc.CallPlanReview(ctx, input)
	skipIfUnsupported(t, err)
	if err != nil {
		t.Fatalf("CallPlanReview failed: %v", err)
	}

	if envelope.ResultKind != "plan_review_result" {
		t.Errorf("expected result_kind plan_review_result, got %s", envelope.ResultKind)
	}
	if strings.TrimSpace(result.Decision) == "" {
		t.Errorf("expected non-empty Decision")
	}
}

func TestIntegration_CallPlanCompile(t *testing.T) {
	t.Parallel()
	skipIfBrainUnreachable(t)

	svc := &sEasyMVPBrain{}
	ctx := context.Background()

	input := braincontracts.PlanCompileInput{
		PlanDraftJSON:        json.RawMessage(`{"title":"compile test plan","steps":["implement feature"]}`),
		PlanReviewResultJSON: json.RawMessage(`{"decision":"approved","review_result_id":"rev_001","review_version":1}`),
		CategoryProfileJSON:  json.RawMessage(`{"category":"web_app","risk":"low"}`),
		RoleContextJSON:      json.RawMessage(`{"roles":["developer"]}`),
	}

	envelope, result, err := svc.CallPlanCompile(ctx, input)
	skipIfUnsupported(t, err)
	if err != nil {
		t.Fatalf("CallPlanCompile failed: %v", err)
	}

	if envelope.ResultKind != "compiled_plan" {
		t.Errorf("expected result_kind compiled_plan, got %s", envelope.ResultKind)
	}
	if result.CompiledTasks == nil {
		t.Errorf("expected non-nil CompiledTasks")
	}
}

func TestIntegration_CallAcceptanceMapping(t *testing.T) {
	t.Parallel()
	skipIfBrainUnreachable(t)

	svc := &sEasyMVPBrain{}
	ctx := context.Background()

	input := braincontracts.AcceptanceMappingInput{
		ProjectCategory:     "web_app",
		CategoryProfileJSON: json.RawMessage(`{"category":"web_app","risk":"low"}`),
		ArtifactSummaryJSON: json.RawMessage(`{"artifacts":["api","ui"]}`),
		CoverageSummaryJSON: json.RawMessage(`{"coverage_pct":80}`),
	}

	envelope, result, err := svc.CallAcceptanceMapping(ctx, input)
	skipIfUnsupported(t, err)
	if err != nil {
		t.Fatalf("CallAcceptanceMapping failed: %v", err)
	}

	if envelope.ResultKind != "acceptance_mapping_result" {
		t.Errorf("expected result_kind acceptance_mapping_result, got %s", envelope.ResultKind)
	}
	if strings.TrimSpace(result.AcceptanceProfileID) == "" {
		t.Errorf("expected non-empty AcceptanceProfileID")
	}
	if len(result.RequiredSurfaces) == 0 {
		t.Errorf("expected non-empty RequiredSurfaces")
	}
	if len(result.RequiredJourneys) == 0 {
		t.Errorf("expected non-empty RequiredJourneys")
	}
	if len(result.RequiredEvidence) == 0 {
		t.Errorf("expected non-empty RequiredEvidence")
	}
}

func TestIntegration_CallCompletionAdjudication(t *testing.T) {
	t.Parallel()
	skipIfBrainUnreachable(t)

	svc := &sEasyMVPBrain{}
	ctx := context.Background()

	input := braincontracts.CompletionAdjudicationInput{
		ExecutionSummaryJSON:    json.RawMessage(`{"tasks_completed":3,"tasks_total":3}`),
		DeliverySummaryJSON:     json.RawMessage(`{"all_delivered":true}`),
		VerificationSummaryJSON: json.RawMessage(`{"tests_passed":true}`),
		AcceptanceSummaryJSON:   json.RawMessage(`{"surfaces_covered":true}`),
		ManualReleaseStateJSON:  json.RawMessage(`{"required":false}`),
	}

	envelope, result, err := svc.CallCompletionAdjudication(ctx, input)
	skipIfUnsupported(t, err)
	if err != nil {
		t.Fatalf("CallCompletionAdjudication failed: %v", err)
	}

	if envelope.ResultKind != "completion_decision" {
		t.Errorf("expected result_kind completion_decision, got %s", envelope.ResultKind)
	}
	if strings.TrimSpace(result.FinalStatus) == "" {
		t.Errorf("expected non-empty FinalStatus")
	}
}

func TestIntegration_CallRepairDesign(t *testing.T) {
	t.Parallel()
	skipIfBrainUnreachable(t)

	svc := &sEasyMVPBrain{}
	ctx := context.Background()

	input := braincontracts.RepairDesignInput{
		FailedTaskContextJSON: json.RawMessage(`{"task_id":"task_001","name":"build UI"}`),
		FailureReasonJSON:     json.RawMessage(`{"reason":"compilation error in component"}`),
		OriginalContractsJSON: json.RawMessage(`{"delivery":"build component","verification":"unit tests pass"}`),
		RuntimeSummaryJSON:    json.RawMessage(`{"attempt":1,"elapsed_seconds":120}`),
	}

	envelope, result, err := svc.CallRepairDesign(ctx, input)
	skipIfUnsupported(t, err)
	if err != nil {
		t.Fatalf("CallRepairDesign failed: %v", err)
	}

	if envelope.ResultKind != "repair_plan_draft" {
		t.Errorf("expected result_kind repair_plan_draft, got %s", envelope.ResultKind)
	}
	if strings.TrimSpace(result.RepairPlanDraftID) == "" {
		t.Errorf("expected non-empty RepairPlanDraftID")
	}
	if len(result.RepairPlanJSON) == 0 {
		t.Errorf("expected non-empty RepairPlanJSON")
	}
	if strings.TrimSpace(result.RepairReasoningSummary) == "" {
		t.Errorf("expected non-empty RepairReasoningSummary")
	}
}

func TestIntegration_CallWorkspaceExplanation(t *testing.T) {
	t.Parallel()
	skipIfBrainUnreachable(t)

	svc := &sEasyMVPBrain{}
	ctx := context.Background()

	input := braincontracts.WorkspaceExplanationInput{
		WorkspaceContextJSON:      json.RawMessage(`{"project_id":"proj_001","status":"in_progress"}`),
		RiskSummaryJSON:           json.RawMessage(`{"overall_risk":"low"}`),
		LatestDecisionSummaryJSON: json.RawMessage(`{"last_decision":"plan approved"}`),
	}

	envelope, result, err := svc.CallWorkspaceExplanation(ctx, input)
	skipIfUnsupported(t, err)
	if err != nil {
		t.Fatalf("CallWorkspaceExplanation failed: %v", err)
	}

	if envelope.ResultKind != "workspace_explanation" {
		t.Errorf("expected result_kind workspace_explanation, got %s", envelope.ResultKind)
	}
	if strings.TrimSpace(result.Headline) == "" {
		t.Errorf("expected non-empty Headline")
	}
	if strings.TrimSpace(result.Summary) == "" {
		t.Errorf("expected non-empty Summary")
	}
}
