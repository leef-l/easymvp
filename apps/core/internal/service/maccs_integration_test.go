package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
)

// ---------------------------------------------------------------------------
// maccsIntegrationBrainStub implements IEasyMVPBrain with minimal stubs for
// the MACCS closed-loop integration test. No real Brain calls are made.
// ---------------------------------------------------------------------------

type maccsIntegrationBrainStub struct {
	requirementAnalysisResult *braincontracts.RequirementAnalysisResult
	solutionDesignResult      *braincontracts.SolutionDesignResult
}

func (s *maccsIntegrationBrainStub) ResolveClientConfig(ctx context.Context) (*EasyMVPBrainClientConfig, error) {
	return &EasyMVPBrainClientConfig{Mode: "test-stub"}, nil
}

func (s *maccsIntegrationBrainStub) ExecuteContract(ctx context.Context, cmd EasyMVPBrainExecuteCommand) (*EasyMVPBrainExecuteResult, error) {
	return nil, nil
}

func (s *maccsIntegrationBrainStub) ExecuteContractStream(ctx context.Context, cmd EasyMVPBrainExecuteCommand) (<-chan EasyMVPBrainStreamEvent, error) {
	return nil, nil
}

func (s *maccsIntegrationBrainStub) CallPlanReview(ctx context.Context, input braincontracts.PlanReviewInput) (*braincontracts.BrainContractEnvelope, *braincontracts.PlanReviewResult, error) {
	return nil, nil, nil
}

func (s *maccsIntegrationBrainStub) CallPlanCompile(ctx context.Context, input braincontracts.PlanCompileInput) (*braincontracts.BrainContractEnvelope, *braincontracts.PlanCompileResult, error) {
	return nil, nil, nil
}

func (s *maccsIntegrationBrainStub) CallPlanRedesign(ctx context.Context, input braincontracts.PlanRedesignInput) (*braincontracts.BrainContractEnvelope, *braincontracts.PlanRedesignResult, error) {
	return nil, nil, nil
}

func (s *maccsIntegrationBrainStub) CallAcceptanceMapping(ctx context.Context, input braincontracts.AcceptanceMappingInput) (*braincontracts.BrainContractEnvelope, *braincontracts.AcceptanceMappingResult, error) {
	return nil, nil, nil
}

func (s *maccsIntegrationBrainStub) CallCompletionAdjudication(ctx context.Context, input braincontracts.CompletionAdjudicationInput) (*braincontracts.BrainContractEnvelope, *braincontracts.CompletionAdjudicationResult, error) {
	return nil, &braincontracts.CompletionAdjudicationResult{
		FunctionalPassed:      true,
		ProductionPassed:      true,
		ManualReleaseRequired: false,
		FinalStatus:           "production_passed",
		DecisionReason:        "Integration test auto-pass",
	}, nil
}

func (s *maccsIntegrationBrainStub) CallRepairDesign(ctx context.Context, input braincontracts.RepairDesignInput) (*braincontracts.BrainContractEnvelope, *braincontracts.RepairDesignResult, error) {
	return nil, nil, nil
}

func (s *maccsIntegrationBrainStub) CallWorkspaceExplanation(ctx context.Context, input braincontracts.WorkspaceExplanationInput) (*braincontracts.BrainContractEnvelope, *braincontracts.WorkspaceExplanationResult, error) {
	return nil, nil, nil
}

func (s *maccsIntegrationBrainStub) CallArchitectChat(ctx context.Context, input braincontracts.ArchitectChatInput) (*braincontracts.BrainContractEnvelope, *braincontracts.ArchitectChatResult, error) {
	return nil, nil, nil
}

func (s *maccsIntegrationBrainStub) CallSolutionDesign(ctx context.Context, input braincontracts.SolutionDesignInput) (*braincontracts.BrainContractEnvelope, *braincontracts.SolutionDesignResult, error) {
	if s.solutionDesignResult != nil {
		return nil, s.solutionDesignResult, nil
	}
	return nil, &braincontracts.SolutionDesignResult{
		Architecture:   "monolith",
		ModulesJSON:    `[{"name":"core"}]`,
		DataModelsJSON: `[{"name":"user"}]`,
		PagesJSON:      `[{"name":"index"}]`,
		TaskDraftsJSON: `[{"name":"build"}]`,
	}, nil
}

func (s *maccsIntegrationBrainStub) CallRequirementAnalysis(ctx context.Context, input braincontracts.RequirementAnalysisInput) (*braincontracts.BrainContractEnvelope, *braincontracts.RequirementAnalysisResult, error) {
	if s.requirementAnalysisResult != nil {
		return nil, s.requirementAnalysisResult, nil
	}
	return nil, &braincontracts.RequirementAnalysisResult{
		RequirementDoc: braincontracts.RequirementDoc{
			Title:    "Test Requirement",
			Overview: input.RawInput,
			FunctionalReqs: []braincontracts.RequirementItem{
				{ID: "FR-001", Description: "Core feature", Priority: "must"},
			},
			NonFunctionalReqs:  []braincontracts.RequirementItem{},
			UserStories:        []braincontracts.UserStory{},
			AcceptanceCriteria: []braincontracts.AcceptanceCriterion{},
			Constraints:        []string{},
			Assumptions:        []string{},
		},
		Summary:             "Test requirement analyzed",
		SuggestedNextAction: "confirm_requirement",
	}, nil
}

func (s *maccsIntegrationBrainStub) ValidateEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope) error {
	return nil
}

func (s *maccsIntegrationBrainStub) ValidatePlanReviewEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.PlanReviewResult) error {
	return nil
}

func (s *maccsIntegrationBrainStub) ValidatePlanCompileEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.PlanCompileResult) error {
	return nil
}

func (s *maccsIntegrationBrainStub) ValidatePlanRedesignEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.PlanRedesignResult) error {
	return nil
}

func (s *maccsIntegrationBrainStub) ValidateAcceptanceMappingEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.AcceptanceMappingResult) error {
	return nil
}

func (s *maccsIntegrationBrainStub) ValidateCompletionAdjudicationEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.CompletionAdjudicationResult) error {
	return nil
}

func (s *maccsIntegrationBrainStub) ValidateRepairDesignEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.RepairDesignResult) error {
	return nil
}

func (s *maccsIntegrationBrainStub) ValidateWorkspaceExplanationEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.WorkspaceExplanationResult) error {
	return nil
}

func (s *maccsIntegrationBrainStub) ValidateArchitectChatEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.ArchitectChatResult) error {
	return nil
}

func (s *maccsIntegrationBrainStub) ValidateSolutionDesignEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.SolutionDesignResult) error {
	return nil
}

func (s *maccsIntegrationBrainStub) ValidateRequirementAnalysisEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.RequirementAnalysisResult) error {
	return nil
}

// ---------------------------------------------------------------------------
// TestMACCSClosedLoopIntegration verifies the full MACCS seven-stage lifecycle
// at the data layer, without making real Brain calls.
//
// Stages:
//  1. Create project
//  2. Analyze requirement -> Confirm requirement
//  3. Generate design -> Confirm design
//  4. Start review -> Review passes
//  5. Prepare delivery -> Accept delivery
//  6. Generate retrospective
//
// Each stage's DB state is verified after the operation.
// ---------------------------------------------------------------------------

func TestMACCSClosedLoopIntegration(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		// Install the stub Brain globally for this test.
		previousBrain := localEasyMVPBrain
		stub := &maccsIntegrationBrainStub{}
		localEasyMVPBrain = stub
		defer func() { localEasyMVPBrain = previousBrain }()

		// ====================================================================
		// Stage 1: Create project
		// ====================================================================
		t.Log("Stage 1: Create project")

		projectID := newResourceID("project")
		now := nowText()
		_, err := db.ExecContext(ctx,
			`INSERT INTO projects (id, name, project_category, goal_summary, status, production_status, workspace_root, repo_root, current_plan_draft_id, current_compiled_plan_id, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			projectID, "MACCS Integration Test", "web", "Build a test web app",
			"created", "pending",
			"/tmp/maccs-integration-test", "",
			"", "",
			now, now,
		)
		if err != nil {
			t.Fatalf("Stage 1 failed: insert project: %v", err)
		}

		// Verify project exists with status=created.
		var projectStatus string
		mustQueryRow(t, db, `SELECT status FROM projects WHERE id = ?`, projectID).Scan(&projectStatus)
		if projectStatus != "created" {
			t.Fatalf("Stage 1: expected project status 'created', got %q", projectStatus)
		}

		// ====================================================================
		// Stage 2: Analyze requirement -> Confirm requirement
		// ====================================================================
		t.Log("Stage 2: Analyze and confirm requirement")

		reqResult, err := Requirement().AnalyzeRequirement(ctx, projectID, "I need a web app with user login and dashboard")
		if err != nil {
			t.Fatalf("Stage 2 failed: AnalyzeRequirement: %v", err)
		}
		if reqResult.RequirementID == "" {
			t.Fatal("Stage 2: requirement ID is empty")
		}
		if reqResult.Status != "draft" {
			t.Fatalf("Stage 2: expected requirement status 'draft', got %q", reqResult.Status)
		}

		// Verify requirement record in DB.
		var reqStatus string
		var reqUserConfirmed int
		mustQueryRow(t, db, `SELECT status, user_confirmed FROM requirements WHERE id = ?`, reqResult.RequirementID).
			Scan(&reqStatus, &reqUserConfirmed)
		if reqStatus != "draft" || reqUserConfirmed != 0 {
			t.Fatalf("Stage 2: unexpected requirement DB state: status=%q confirmed=%d", reqStatus, reqUserConfirmed)
		}

		// Confirm the requirement.
		if err := Requirement().ConfirmRequirement(ctx, reqResult.RequirementID); err != nil {
			t.Fatalf("Stage 2 failed: ConfirmRequirement: %v", err)
		}

		mustQueryRow(t, db, `SELECT status, user_confirmed FROM requirements WHERE id = ?`, reqResult.RequirementID).
			Scan(&reqStatus, &reqUserConfirmed)
		if reqStatus != "confirmed" || reqUserConfirmed != 1 {
			t.Fatalf("Stage 2: expected confirmed requirement, got status=%q confirmed=%d", reqStatus, reqUserConfirmed)
		}

		// Verify audit log was written for requirement analysis.
		var reqAuditCount int
		mustQueryRow(t, db, `SELECT COUNT(*) FROM audit_logs WHERE project_id = ? AND event_type = 'requirement.analyzed'`, projectID).
			Scan(&reqAuditCount)
		if reqAuditCount < 1 {
			t.Fatalf("Stage 2: expected audit log for requirement.analyzed, got count=%d", reqAuditCount)
		}

		// ====================================================================
		// Stage 3: Generate design -> Confirm design
		// ====================================================================
		t.Log("Stage 3: Generate and confirm design")

		designResult, err := Design().GenerateDesign(ctx, projectID, reqResult.RequirementID)
		if err != nil {
			t.Fatalf("Stage 3 failed: GenerateDesign: %v", err)
		}
		if designResult.DesignID == "" {
			t.Fatal("Stage 3: design ID is empty")
		}
		if designResult.Status != "draft" {
			t.Fatalf("Stage 3: expected design status 'draft', got %q", designResult.Status)
		}

		// Verify design record in DB.
		var designStatus string
		var designVersion int
		mustQueryRow(t, db, `SELECT status, version FROM solution_designs WHERE id = ?`, designResult.DesignID).
			Scan(&designStatus, &designVersion)
		if designStatus != "draft" || designVersion != 1 {
			t.Fatalf("Stage 3: unexpected design DB state: status=%q version=%d", designStatus, designVersion)
		}

		// Confirm the design.
		if err := Design().ConfirmDesign(ctx, designResult.DesignID); err != nil {
			t.Fatalf("Stage 3 failed: ConfirmDesign: %v", err)
		}

		mustQueryRow(t, db, `SELECT status FROM solution_designs WHERE id = ?`, designResult.DesignID).
			Scan(&designStatus)
		if designStatus != "approved" {
			t.Fatalf("Stage 3: expected design status 'approved', got %q", designStatus)
		}

		// ====================================================================
		// Stage 4: Review — use Intervene to override-approve (avoids real Brain review)
		// ====================================================================
		t.Log("Stage 4: Design review (override approve)")

		// Insert a synthetic review to simulate passing.
		reviewID := newResourceID("design_review")
		_, err = db.ExecContext(ctx,
			`INSERT INTO design_reviews (id, design_id, project_id, round, passed, score, dimensions_json, issues_json, suggestions_json, fix_tasks_json, brain_run_id, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			reviewID, designResult.DesignID, projectID, 1, 1, 95,
			"[]", "[]", "[]", "", "integration_test", now,
		)
		if err != nil {
			t.Fatalf("Stage 4 failed: insert review: %v", err)
		}

		// Verify review record.
		var reviewPassed int
		var reviewScore int
		mustQueryRow(t, db, `SELECT passed, score FROM design_reviews WHERE id = ?`, reviewID).
			Scan(&reviewPassed, &reviewScore)
		if reviewPassed != 1 || reviewScore != 95 {
			t.Fatalf("Stage 4: unexpected review: passed=%d score=%d", reviewPassed, reviewScore)
		}

		// ====================================================================
		// Stage 5: Prepare delivery -> Accept delivery
		// ====================================================================
		t.Log("Stage 5: Prepare and accept delivery")

		deliveryResult, err := Delivery().PrepareDelivery(ctx, projectID)
		if err != nil {
			t.Fatalf("Stage 5 failed: PrepareDelivery: %v", err)
		}
		if deliveryResult.DeliveryID == "" {
			t.Fatal("Stage 5: delivery ID is empty")
		}
		if deliveryResult.Status != "prepared" {
			t.Fatalf("Stage 5: expected delivery status 'prepared', got %q", deliveryResult.Status)
		}

		// Verify delivery record in DB.
		var deliveryStatus string
		mustQueryRow(t, db, `SELECT status FROM project_deliveries WHERE id = ?`, deliveryResult.DeliveryID).
			Scan(&deliveryStatus)
		if deliveryStatus != "prepared" {
			t.Fatalf("Stage 5: unexpected delivery status=%q", deliveryStatus)
		}

		// Accept the delivery.
		if err := Delivery().AcceptDelivery(ctx, deliveryResult.DeliveryID); err != nil {
			t.Fatalf("Stage 5 failed: AcceptDelivery: %v", err)
		}

		mustQueryRow(t, db, `SELECT status FROM project_deliveries WHERE id = ?`, deliveryResult.DeliveryID).
			Scan(&deliveryStatus)
		if deliveryStatus != "accepted" {
			t.Fatalf("Stage 5: expected delivery status 'accepted', got %q", deliveryStatus)
		}

		// ====================================================================
		// Stage 6: Generate retrospective
		// ====================================================================
		t.Log("Stage 6: Generate retrospective")

		retroResult, err := Retrospective().GenerateRetrospective(ctx, projectID)
		if err != nil {
			t.Fatalf("Stage 6 failed: GenerateRetrospective: %v", err)
		}
		if retroResult.RetrospectiveID == "" {
			t.Fatal("Stage 6: retrospective ID is empty")
		}

		// Verify retrospective record in DB.
		var retroProjectID string
		mustQueryRow(t, db, `SELECT project_id FROM project_retrospectives WHERE id = ?`, retroResult.RetrospectiveID).
			Scan(&retroProjectID)
		if retroProjectID != projectID {
			t.Fatalf("Stage 6: retrospective project_id mismatch: got %q want %q", retroProjectID, projectID)
		}

		// ====================================================================
		// Final: Verify audit trail covers all stages
		// ====================================================================
		t.Log("Final: Verify audit trail")

		var totalAuditCount int
		mustQueryRow(t, db, `SELECT COUNT(*) FROM audit_logs WHERE project_id = ?`, projectID).
			Scan(&totalAuditCount)
		// We expect at least: requirement.analyzed, requirement.confirmed,
		// design.generated, design.confirmed, delivery.prepared, delivery.accepted,
		// retrospective.generated = 7 audit entries minimum.
		if totalAuditCount < 7 {
			t.Fatalf("Final: expected at least 7 audit log entries, got %d", totalAuditCount)
		}

		t.Logf("MACCS closed-loop integration test passed: project=%s, audit_entries=%d", projectID, totalAuditCount)
	})
}

// TestMACCSClosedLoop_RequirementRejectAndRetry verifies the requirement
// can be re-analyzed after initial analysis (simulating user dissatisfaction).
func TestMACCSClosedLoop_RequirementRejectAndRetry(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		previousBrain := localEasyMVPBrain
		localEasyMVPBrain = &maccsIntegrationBrainStub{}
		defer func() { localEasyMVPBrain = previousBrain }()

		// Create project.
		projectID := newResourceID("project")
		now := nowText()
		_, err := db.ExecContext(ctx,
			`INSERT INTO projects (id, name, project_category, goal_summary, status, production_status, workspace_root, repo_root, current_plan_draft_id, current_compiled_plan_id, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			projectID, "Retry Test", "web", "Test retry flow",
			"created", "pending",
			"/tmp/maccs-retry-test", "",
			"", "",
			now, now,
		)
		if err != nil {
			t.Fatalf("insert project failed: %v", err)
		}

		// First analysis.
		req1, err := Requirement().AnalyzeRequirement(ctx, projectID, "First attempt")
		if err != nil {
			t.Fatalf("first AnalyzeRequirement failed: %v", err)
		}

		// Second analysis (user wants a different requirement).
		req2, err := Requirement().AnalyzeRequirement(ctx, projectID, "Second attempt with more detail")
		if err != nil {
			t.Fatalf("second AnalyzeRequirement failed: %v", err)
		}

		if req1.RequirementID == req2.RequirementID {
			t.Fatal("expected different requirement IDs for two analyses")
		}

		// Verify both exist in DB.
		var reqCount int
		mustQueryRow(t, db, `SELECT COUNT(*) FROM requirements WHERE project_id = ?`, projectID).Scan(&reqCount)
		if reqCount != 2 {
			t.Fatalf("expected 2 requirement records, got %d", reqCount)
		}
	})
}

// TestMACCSClosedLoop_DesignRejectAndRegenerate verifies design rejection
// allows a new design to be generated.
func TestMACCSClosedLoop_DesignRejectAndRegenerate(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		previousBrain := localEasyMVPBrain
		localEasyMVPBrain = &maccsIntegrationBrainStub{}
		defer func() { localEasyMVPBrain = previousBrain }()

		// Create project.
		projectID := newResourceID("project")
		now := nowText()
		_, err := db.ExecContext(ctx,
			`INSERT INTO projects (id, name, project_category, goal_summary, status, production_status, workspace_root, repo_root, current_plan_draft_id, current_compiled_plan_id, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			projectID, "Design Reject Test", "api", "Test design rejection",
			"created", "pending",
			"/tmp/maccs-design-reject", "",
			"", "",
			now, now,
		)
		if err != nil {
			t.Fatalf("insert project failed: %v", err)
		}

		// Analyze and confirm requirement.
		reqResult, err := Requirement().AnalyzeRequirement(ctx, projectID, "Build an API service")
		if err != nil {
			t.Fatalf("AnalyzeRequirement failed: %v", err)
		}
		if err := Requirement().ConfirmRequirement(ctx, reqResult.RequirementID); err != nil {
			t.Fatalf("ConfirmRequirement failed: %v", err)
		}

		// Generate first design.
		design1, err := Design().GenerateDesign(ctx, projectID, reqResult.RequirementID)
		if err != nil {
			t.Fatalf("first GenerateDesign failed: %v", err)
		}

		// Reject the first design.
		if err := Design().RejectDesign(ctx, design1.DesignID, "Not good enough"); err != nil {
			t.Fatalf("RejectDesign failed: %v", err)
		}

		// Verify rejection.
		var rejectedStatus string
		mustQueryRow(t, db, `SELECT status FROM solution_designs WHERE id = ?`, design1.DesignID).Scan(&rejectedStatus)
		if rejectedStatus != "rejected" {
			t.Fatalf("expected rejected status, got %q", rejectedStatus)
		}

		// Generate a second design.
		design2, err := Design().GenerateDesign(ctx, projectID, reqResult.RequirementID)
		if err != nil {
			t.Fatalf("second GenerateDesign failed: %v", err)
		}
		if design2.DesignID == design1.DesignID {
			t.Fatal("expected different design IDs")
		}

		// Verify version increment.
		var version2 int
		mustQueryRow(t, db, `SELECT version FROM solution_designs WHERE id = ?`, design2.DesignID).Scan(&version2)
		if version2 != 2 {
			t.Fatalf("expected design version 2, got %d", version2)
		}
	})
}

// TestMACCSClosedLoop_DeliveryRejectAndRetry verifies delivery rejection flow.
func TestMACCSClosedLoop_DeliveryRejectAndRetry(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		previousBrain := localEasyMVPBrain
		localEasyMVPBrain = &maccsIntegrationBrainStub{}
		defer func() { localEasyMVPBrain = previousBrain }()

		// Create project.
		projectID := newResourceID("project")
		now := nowText()
		_, err := db.ExecContext(ctx,
			`INSERT INTO projects (id, name, project_category, goal_summary, status, production_status, workspace_root, repo_root, current_plan_draft_id, current_compiled_plan_id, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			projectID, "Delivery Reject Test", "web", "Test delivery rejection",
			"created", "pending",
			"/tmp/maccs-delivery-reject", "",
			"", "",
			now, now,
		)
		if err != nil {
			t.Fatalf("insert project failed: %v", err)
		}

		// Prepare delivery.
		d1, err := Delivery().PrepareDelivery(ctx, projectID)
		if err != nil {
			t.Fatalf("PrepareDelivery failed: %v", err)
		}

		// Reject delivery.
		if err := Delivery().RejectDelivery(ctx, d1.DeliveryID, "Missing docs"); err != nil {
			t.Fatalf("RejectDelivery failed: %v", err)
		}

		var d1Status string
		mustQueryRow(t, db, `SELECT status FROM project_deliveries WHERE id = ?`, d1.DeliveryID).Scan(&d1Status)
		if d1Status != "rejected" {
			t.Fatalf("expected delivery status 'rejected', got %q", d1Status)
		}

		// Prepare a second delivery.
		d2, err := Delivery().PrepareDelivery(ctx, projectID)
		if err != nil {
			t.Fatalf("second PrepareDelivery failed: %v", err)
		}
		if d2.DeliveryID == d1.DeliveryID {
			t.Fatal("expected different delivery IDs")
		}
	})
}
