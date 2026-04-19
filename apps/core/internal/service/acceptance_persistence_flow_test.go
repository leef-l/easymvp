package service

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/do"
)

var transactionalServiceTestMu sync.Mutex

type completionAdjudicationBrainStub struct {
	result *braincontracts.CompletionAdjudicationResult
	err    error
}

func (s *completionAdjudicationBrainStub) ResolveClientConfig(ctx context.Context) (*EasyMVPBrainClientConfig, error) {
	_ = ctx
	return nil, nil
}

func (s *completionAdjudicationBrainStub) ExecuteContract(ctx context.Context, cmd EasyMVPBrainExecuteCommand) (*EasyMVPBrainExecuteResult, error) {
	_ = ctx
	_ = cmd
	return nil, nil
}

func (s *completionAdjudicationBrainStub) CallPlanReview(ctx context.Context, input braincontracts.PlanReviewInput) (*braincontracts.BrainContractEnvelope, *braincontracts.PlanReviewResult, error) {
	_ = ctx
	_ = input
	return nil, nil, nil
}

func (s *completionAdjudicationBrainStub) CallPlanCompile(ctx context.Context, input braincontracts.PlanCompileInput) (*braincontracts.BrainContractEnvelope, *braincontracts.PlanCompileResult, error) {
	_ = ctx
	_ = input
	return nil, nil, nil
}

func (s *completionAdjudicationBrainStub) CallAcceptanceMapping(ctx context.Context, input braincontracts.AcceptanceMappingInput) (*braincontracts.BrainContractEnvelope, *braincontracts.AcceptanceMappingResult, error) {
	_ = ctx
	_ = input
	return nil, nil, nil
}

func (s *completionAdjudicationBrainStub) CallCompletionAdjudication(ctx context.Context, input braincontracts.CompletionAdjudicationInput) (*braincontracts.BrainContractEnvelope, *braincontracts.CompletionAdjudicationResult, error) {
	_ = ctx
	_ = input
	return nil, s.result, s.err
}

func (s *completionAdjudicationBrainStub) CallRepairDesign(ctx context.Context, input braincontracts.RepairDesignInput) (*braincontracts.BrainContractEnvelope, *braincontracts.RepairDesignResult, error) {
	_ = ctx
	_ = input
	return nil, nil, nil
}

func (s *completionAdjudicationBrainStub) CallWorkspaceExplanation(ctx context.Context, input braincontracts.WorkspaceExplanationInput) (*braincontracts.BrainContractEnvelope, *braincontracts.WorkspaceExplanationResult, error) {
	_ = ctx
	_ = input
	return nil, nil, nil
}

func (s *completionAdjudicationBrainStub) ValidateEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope) error {
	_ = ctx
	_ = envelope
	return nil
}

func (s *completionAdjudicationBrainStub) ValidatePlanReviewEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.PlanReviewResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func (s *completionAdjudicationBrainStub) ValidatePlanCompileEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.PlanCompileResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func (s *completionAdjudicationBrainStub) ValidateAcceptanceMappingEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.AcceptanceMappingResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func (s *completionAdjudicationBrainStub) ValidateCompletionAdjudicationEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.CompletionAdjudicationResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func (s *completionAdjudicationBrainStub) ValidateRepairDesignEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.RepairDesignResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func (s *completionAdjudicationBrainStub) ValidateWorkspaceExplanationEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.WorkspaceExplanationResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func TestApplyManualReleasePersistsAllAffectedRows(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		seedAcceptanceFlowProject(t, ctx, db, acceptanceFlowSeed{
			projectID:             "proj_manual_release_success",
			taskID:                "task_manual_release_success",
			runID:                 "run_manual_release_success",
			projectStatus:         "acceptance",
			projectProduction:     "functional_passed",
			runStatus:             "awaiting_manual_release",
			functionalStatus:      "functional_passed",
			productionStatus:      "functional_passed",
			manualReleaseRequired: 1,
		})

		runID, err := applyManualRelease(ctx, ApplyManualReleaseCommand{
			ProjectID: "proj_manual_release_success",
			Comment:   "Ship it",
		})
		if err != nil {
			t.Fatalf("applyManualRelease failed: %v", err)
		}
		if runID != "run_manual_release_success" {
			t.Fatalf("unexpected acceptance run id: got %q want %q", runID, "run_manual_release_success")
		}

		var (
			runStatus         string
			runProduction     string
			runFinishedAt     sql.NullString
			projectStatus     string
			projectProduction string
			judgementCount    int
			auditCount        int
			gateKind          string
			gateStatus        string
			gateComment       sql.NullString
		)

		mustQueryRow(t, db, `SELECT status, production_status, finished_at FROM acceptance_runs WHERE id = ?`, "run_manual_release_success").
			Scan(&runStatus, &runProduction, &runFinishedAt)
		mustQueryRow(t, db, `SELECT status, production_status FROM projects WHERE id = ?`, "proj_manual_release_success").
			Scan(&projectStatus, &projectProduction)
		mustQueryRow(t, db, `SELECT COUNT(*) FROM acceptance_judgements WHERE acceptance_run_id = ? AND judgement_kind = ? AND judgement_result = ?`, "run_manual_release_success", "release_gate", "approved").
			Scan(&judgementCount)
		mustQueryRow(t, db, `SELECT COUNT(*) FROM audit_logs WHERE project_id = ? AND event_type = ?`, "proj_manual_release_success", "manual_release.approved").
			Scan(&auditCount)
		mustQueryRow(t, db, `SELECT gate_kind, gate_status, comment FROM task_manual_gates WHERE project_id = ? AND task_id = ?`, "proj_manual_release_success", "task_manual_release_success").
			Scan(&gateKind, &gateStatus, &gateComment)

		if runStatus != "completed" || runProduction != "production_passed" || !runFinishedAt.Valid {
			t.Fatalf("unexpected acceptance run state: status=%q production=%q finished_at=%#v", runStatus, runProduction, runFinishedAt)
		}
		if projectStatus != "completed" || projectProduction != "production_passed" {
			t.Fatalf("unexpected project state: status=%q production=%q", projectStatus, projectProduction)
		}
		if judgementCount != 1 {
			t.Fatalf("unexpected release judgement count: got %d want %d", judgementCount, 1)
		}
		if auditCount != 1 {
			t.Fatalf("unexpected audit count: got %d want %d", auditCount, 1)
		}
		if gateKind != "manual_release" || gateStatus != "approved" || gateComment.String != "Ship it" {
			t.Fatalf("unexpected manual gate row: kind=%q status=%q comment=%#v", gateKind, gateStatus, gateComment)
		}
	})
}

func TestApplyManualReleaseReturnsExistingApprovalWithoutDuplicatingRows(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		seedAcceptanceFlowProject(t, ctx, db, acceptanceFlowSeed{
			projectID:                "proj_manual_release_idempotent",
			taskID:                   "task_manual_release_idempotent",
			runID:                    "run_manual_release_idempotent",
			projectStatus:            "acceptance",
			projectProduction:        "functional_passed",
			runStatus:                "awaiting_manual_release",
			functionalStatus:         "functional_passed",
			productionStatus:         "functional_passed",
			manualReleaseRequired:    1,
			existingReleaseJudgement: true,
		})

		runID, err := applyManualRelease(ctx, ApplyManualReleaseCommand{
			ProjectID: "proj_manual_release_idempotent",
			Comment:   "Ignored duplicate approval",
		})
		if err != nil {
			t.Fatalf("applyManualRelease failed: %v", err)
		}
		if runID != "run_manual_release_idempotent" {
			t.Fatalf("unexpected acceptance run id: got %q want %q", runID, "run_manual_release_idempotent")
		}

		var judgementCount, auditCount, gateCount int
		mustQueryRow(t, db, `SELECT COUNT(*) FROM acceptance_judgements WHERE acceptance_run_id = ?`, "run_manual_release_idempotent").Scan(&judgementCount)
		mustQueryRow(t, db, `SELECT COUNT(*) FROM audit_logs WHERE project_id = ?`, "proj_manual_release_idempotent").Scan(&auditCount)
		mustQueryRow(t, db, `SELECT COUNT(*) FROM task_manual_gates WHERE project_id = ? AND task_id = ?`, "proj_manual_release_idempotent", "task_manual_release_idempotent").Scan(&gateCount)

		if judgementCount != 1 || auditCount != 0 || gateCount != 0 {
			t.Fatalf("duplicate manual release wrote unexpected rows: judgements=%d audit=%d gates=%d", judgementCount, auditCount, gateCount)
		}
	})
}

func TestAdjudicateAcceptanceAggregatePersistsAwaitingManualReleaseState(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		seedAcceptanceFlowProject(t, ctx, db, acceptanceFlowSeed{
			projectID:             "proj_adjudicate_manual",
			taskID:                "task_adjudicate_manual",
			runID:                 "run_adjudicate_manual",
			projectStatus:         "acceptance",
			projectProduction:     "functional_passed",
			runStatus:             "running",
			functionalStatus:      "functional_passed",
			productionStatus:      "functional_passed",
			manualReleaseRequired: 0,
		})

		aggregate, err := loadAcceptanceAggregate(ctx, db, "proj_adjudicate_manual")
		if err != nil {
			t.Fatalf("loadAcceptanceAggregate failed: %v", err)
		}

		previous := localEasyMVPBrain
		localEasyMVPBrain = &completionAdjudicationBrainStub{
			result: &braincontracts.CompletionAdjudicationResult{
				FunctionalPassed:      true,
				ProductionPassed:      true,
				ManualReleaseRequired: true,
				FinalStatus:           "production_passed",
				DecisionReason:        "Human release approval is still required.",
			},
		}
		defer func() {
			localEasyMVPBrain = previous
		}()

		result, err := adjudicateAcceptanceAggregate(ctx, db, aggregate)
		if err != nil {
			t.Fatalf("adjudicateAcceptanceAggregate failed: %v", err)
		}
		if !result.ManualReleaseRequired || result.FinalStatus != "production_passed" {
			t.Fatalf("unexpected adjudication result: %#v", result)
		}

		var (
			runStatus                 string
			runFunctionalStatus       string
			runProductionStatus       string
			runManualReleaseRequired  int
			runFinishedAt             sql.NullString
			projectStatus             string
			projectProductionStatus   string
			totalJudgements           int
			releaseGateJudgementCount int
		)
		mustQueryRow(t, db, `SELECT status, functional_status, production_status, manual_release_required, finished_at FROM acceptance_runs WHERE id = ?`, "run_adjudicate_manual").
			Scan(&runStatus, &runFunctionalStatus, &runProductionStatus, &runManualReleaseRequired, &runFinishedAt)
		mustQueryRow(t, db, `SELECT status, production_status FROM projects WHERE id = ?`, "proj_adjudicate_manual").
			Scan(&projectStatus, &projectProductionStatus)
		mustQueryRow(t, db, `SELECT COUNT(*) FROM acceptance_judgements WHERE acceptance_run_id = ?`, "run_adjudicate_manual").
			Scan(&totalJudgements)
		mustQueryRow(t, db, `SELECT COUNT(*) FROM acceptance_judgements WHERE acceptance_run_id = ? AND judgement_kind = ? AND judgement_result = ?`, "run_adjudicate_manual", "release_gate", "awaiting_manual_release").
			Scan(&releaseGateJudgementCount)

		if runStatus != "awaiting_manual_release" || runFunctionalStatus != "functional_passed" || runProductionStatus != "production_passed" || runManualReleaseRequired != 1 || runFinishedAt.Valid {
			t.Fatalf("unexpected acceptance run state: status=%q functional=%q production=%q manual_release=%d finished_at=%#v", runStatus, runFunctionalStatus, runProductionStatus, runManualReleaseRequired, runFinishedAt)
		}
		if projectStatus != "acceptance" || projectProductionStatus != "production_passed" {
			t.Fatalf("unexpected project state: status=%q production=%q", projectStatus, projectProductionStatus)
		}
		if totalJudgements != 4 || releaseGateJudgementCount != 1 {
			t.Fatalf("unexpected judgement counts: total=%d release_gate=%d", totalJudgements, releaseGateJudgementCount)
		}
	})
}

func TestAdjudicateAcceptanceAggregatePersistsCompletedState(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		seedAcceptanceFlowProject(t, ctx, db, acceptanceFlowSeed{
			projectID:             "proj_adjudicate_complete",
			taskID:                "task_adjudicate_complete",
			runID:                 "run_adjudicate_complete",
			projectStatus:         "acceptance",
			projectProduction:     "functional_passed",
			runStatus:             "running",
			functionalStatus:      "functional_passed",
			productionStatus:      "functional_passed",
			manualReleaseRequired: 0,
		})

		aggregate, err := loadAcceptanceAggregate(ctx, db, "proj_adjudicate_complete")
		if err != nil {
			t.Fatalf("loadAcceptanceAggregate failed: %v", err)
		}

		previous := localEasyMVPBrain
		localEasyMVPBrain = &completionAdjudicationBrainStub{
			result: &braincontracts.CompletionAdjudicationResult{
				FunctionalPassed:      true,
				ProductionPassed:      true,
				ManualReleaseRequired: false,
				FinalStatus:           "production_passed",
				DecisionReason:        "Verification passed and completion is clear.",
			},
		}
		defer func() {
			localEasyMVPBrain = previous
		}()

		result, err := adjudicateAcceptanceAggregate(ctx, db, aggregate)
		if err != nil {
			t.Fatalf("adjudicateAcceptanceAggregate failed: %v", err)
		}
		if result.FinalStatus != "production_passed" {
			t.Fatalf("unexpected adjudication result: %#v", result)
		}

		var (
			runStatus           string
			runProductionStatus string
			runFinishedAt       sql.NullString
			projectStatus       string
			projectProduction   string
			totalJudgements     int
		)
		mustQueryRow(t, db, `SELECT status, production_status, finished_at FROM acceptance_runs WHERE id = ?`, "run_adjudicate_complete").
			Scan(&runStatus, &runProductionStatus, &runFinishedAt)
		mustQueryRow(t, db, `SELECT status, production_status FROM projects WHERE id = ?`, "proj_adjudicate_complete").
			Scan(&projectStatus, &projectProduction)
		mustQueryRow(t, db, `SELECT COUNT(*) FROM acceptance_judgements WHERE acceptance_run_id = ?`, "run_adjudicate_complete").
			Scan(&totalJudgements)

		if runStatus != "completed" || runProductionStatus != "production_passed" || !runFinishedAt.Valid {
			t.Fatalf("unexpected acceptance run state: status=%q production=%q finished_at=%#v", runStatus, runProductionStatus, runFinishedAt)
		}
		if projectStatus != "completed" || projectProduction != "production_passed" {
			t.Fatalf("unexpected project state: status=%q production=%q", projectStatus, projectProduction)
		}
		if totalJudgements != 3 {
			t.Fatalf("unexpected judgement count: got %d want %d", totalJudgements, 3)
		}
	})
}

type acceptanceFlowSeed struct {
	projectID                string
	taskID                   string
	runID                    string
	projectStatus            string
	projectProduction        string
	runStatus                string
	functionalStatus         string
	productionStatus         string
	manualReleaseRequired    int
	existingReleaseJudgement bool
}

func withAcceptanceFlowDB(t *testing.T, fn func(ctx context.Context, db *sql.DB)) {
	t.Helper()

	transactionalServiceTestMu.Lock()
	defer transactionalServiceTestMu.Unlock()

	ctx := context.Background()
	root := t.TempDir()
	dataRoot := filepath.Join(root, "var")
	dbPath := filepath.Join(dataRoot, "data", "easymvp.db")
	migrationPath, err := filepath.Abs(filepath.Join("..", "..", "manifest", "migrations"))
	if err != nil {
		t.Fatalf("resolve migration path failed: %v", err)
	}

	adapter, err := gcfg.NewAdapterContent(fmt.Sprintf(`{
  "easymvp": {
    "dataRoot": %q,
    "dbPath": %q,
    "migrationPath": %q,
    "brainServeBaseURL": "http://127.0.0.1:7701",
    "safeMode": true
  },
  "server": {
    "address": ":8000"
  }
}`, dataRoot, dbPath, migrationPath))
	if err != nil {
		t.Fatalf("create config adapter failed: %v", err)
	}

	previousAdapter := g.Cfg().GetAdapter()
	startupConfigStore.mu.Lock()
	previousStartup := startupConfigStore.value
	startupConfigStore.mu.Unlock()

	g.Cfg().SetAdapter(adapter)
	SetStartupConfig(StartupConfig{
		DataRoot:          dataRoot,
		DBPath:            dbPath,
		MigrationPath:     migrationPath,
		BrainServeBaseURL: "http://127.0.0.1:7701",
		ServerAddress:     ":8000",
		SafeMode:          true,
	})
	t.Cleanup(func() {
		g.Cfg().SetAdapter(previousAdapter)
		startupConfigStore.mu.Lock()
		startupConfigStore.value = previousStartup
		startupConfigStore.mu.Unlock()
	})

	if err = Bootstrap(ctx); err != nil {
		t.Fatalf("bootstrap test database failed: %v", err)
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		t.Fatalf("openProjectDatabase failed: %v", err)
	}
	defer closeFn()

	fn(ctx, db)
}

func seedAcceptanceFlowProject(t *testing.T, ctx context.Context, db *sql.DB, seed acceptanceFlowSeed) {
	t.Helper()

	now := "2026-04-20T10:00:00Z"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin seed transaction failed: %v", err)
	}

	if err = insertProjectRow(ctx, tx, &do.Projects{
		Id:               seed.projectID,
		Name:             "Seed Project",
		ProjectCategory:  "web",
		GoalSummary:      "Seed acceptance flow state",
		Status:           seed.projectStatus,
		ProductionStatus: seed.projectProduction,
		WorkspaceRoot:    filepath.Join("/tmp", seed.projectID),
		CreatedAt:        now,
		UpdatedAt:        now,
	}); err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert project failed: %v", err)
	}

	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO domain_tasks (
id, project_id, source_compiled_plan_id, source_compiled_task_id, source_task_key, compiled_version, name, phase, task_kind, role_type, brain_kind, risk_level, status, manual_review_required, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		seed.taskID,
		seed.projectID,
		"",
		"",
		"",
		1,
		"Acceptance Task",
		"acceptance",
		"verification",
		"executor",
		"coder",
		"medium",
		"done",
		0,
		now,
		now,
	); err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert domain task failed: %v", err)
	}

	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO acceptance_runs (
id, project_id, task_id, profile_version, status, functional_status, production_status, manual_release_required, created_at, finished_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)`,
		seed.runID,
		seed.projectID,
		seed.taskID,
		"web/v1",
		seed.runStatus,
		seed.functionalStatus,
		seed.productionStatus,
		seed.manualReleaseRequired,
		now,
	); err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert acceptance run failed: %v", err)
	}

	if seed.existingReleaseJudgement {
		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO acceptance_judgements (
id, project_id, acceptance_run_id, judgement_kind, judgement_result, summary, detail_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			"judge_existing_release",
			seed.projectID,
			seed.runID,
			"release_gate",
			"approved",
			"Manual release already approved",
			`{"manual_release_completed":true}`,
			now,
		); err != nil {
			_ = tx.Rollback()
			t.Fatalf("insert existing release judgement failed: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		t.Fatalf("commit seed transaction failed: %v", err)
	}
}

func mustQueryRow(t *testing.T, db *sql.DB, query string, args ...any) *sql.Row {
	t.Helper()
	return db.QueryRowContext(context.Background(), query, args...)
}
