package service

import (
	"context"
	"testing"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
)

type workspaceExplanationBrainStub struct {
	workspaceResult *braincontracts.WorkspaceExplanationResult
	workspaceErr    error
}

func (s *workspaceExplanationBrainStub) ResolveClientConfig(ctx context.Context) (*EasyMVPBrainClientConfig, error) {
	_ = ctx
	return nil, nil
}

func (s *workspaceExplanationBrainStub) ExecuteContract(ctx context.Context, cmd EasyMVPBrainExecuteCommand) (*EasyMVPBrainExecuteResult, error) {
	_ = ctx
	_ = cmd
	return nil, nil
}

func (s *workspaceExplanationBrainStub) CallPlanReview(ctx context.Context, input braincontracts.PlanReviewInput) (*braincontracts.BrainContractEnvelope, *braincontracts.PlanReviewResult, error) {
	_ = ctx
	_ = input
	return nil, nil, nil
}

func (s *workspaceExplanationBrainStub) CallPlanCompile(ctx context.Context, input braincontracts.PlanCompileInput) (*braincontracts.BrainContractEnvelope, *braincontracts.PlanCompileResult, error) {
	_ = ctx
	_ = input
	return nil, nil, nil
}

func (s *workspaceExplanationBrainStub) CallAcceptanceMapping(ctx context.Context, input braincontracts.AcceptanceMappingInput) (*braincontracts.BrainContractEnvelope, *braincontracts.AcceptanceMappingResult, error) {
	_ = ctx
	_ = input
	return nil, nil, nil
}

func (s *workspaceExplanationBrainStub) CallCompletionAdjudication(ctx context.Context, input braincontracts.CompletionAdjudicationInput) (*braincontracts.BrainContractEnvelope, *braincontracts.CompletionAdjudicationResult, error) {
	_ = ctx
	_ = input
	return nil, nil, nil
}

func (s *workspaceExplanationBrainStub) CallRepairDesign(ctx context.Context, input braincontracts.RepairDesignInput) (*braincontracts.BrainContractEnvelope, *braincontracts.RepairDesignResult, error) {
	_ = ctx
	_ = input
	return nil, nil, nil
}

func (s *workspaceExplanationBrainStub) CallWorkspaceExplanation(ctx context.Context, input braincontracts.WorkspaceExplanationInput) (*braincontracts.BrainContractEnvelope, *braincontracts.WorkspaceExplanationResult, error) {
	_ = ctx
	_ = input
	return nil, s.workspaceResult, s.workspaceErr
}

func (s *workspaceExplanationBrainStub) ValidateEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope) error {
	_ = ctx
	_ = envelope
	return nil
}

func (s *workspaceExplanationBrainStub) ValidatePlanReviewEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.PlanReviewResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func (s *workspaceExplanationBrainStub) ValidatePlanCompileEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.PlanCompileResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func (s *workspaceExplanationBrainStub) ValidateAcceptanceMappingEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.AcceptanceMappingResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func (s *workspaceExplanationBrainStub) ValidateCompletionAdjudicationEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.CompletionAdjudicationResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func (s *workspaceExplanationBrainStub) ValidateRepairDesignEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.RepairDesignResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func (s *workspaceExplanationBrainStub) ValidateWorkspaceExplanationEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.WorkspaceExplanationResult) error {
	_ = ctx
	_ = envelope
	_ = result
	return nil
}

func TestDeriveAcceptanceRunStatusFromAdjudication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  *braincontracts.CompletionAdjudicationResult
		expect string
	}{
		{
			name: "failed result stays failed",
			input: &braincontracts.CompletionAdjudicationResult{
				FinalStatus: "failed",
			},
			expect: "failed",
		},
		{
			name: "manual release pending waits for approval",
			input: &braincontracts.CompletionAdjudicationResult{
				FinalStatus:            "functional_passed",
				ManualReleaseRequired:  true,
				ManualReleaseCompleted: false,
			},
			expect: "awaiting_manual_release",
		},
		{
			name: "production passed completes",
			input: &braincontracts.CompletionAdjudicationResult{
				FinalStatus: "production_passed",
			},
			expect: "completed",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := deriveAcceptanceRunStatusFromAdjudication(tt.input); got != tt.expect {
				t.Fatalf("unexpected acceptance run status: got %s want %s", got, tt.expect)
			}
		})
	}
}

func TestShouldCreateRepairDraftAfterAdjudication(t *testing.T) {
	t.Parallel()

	if shouldCreateRepairDraftAfterAdjudication(nil) {
		t.Fatalf("nil adjudication should not create repair draft")
	}
	if !shouldCreateRepairDraftAfterAdjudication(&braincontracts.CompletionAdjudicationResult{FinalStatus: "failed"}) {
		t.Fatalf("failed adjudication should create repair draft")
	}
	if shouldCreateRepairDraftAfterAdjudication(&braincontracts.CompletionAdjudicationResult{FinalStatus: "production_passed"}) {
		t.Fatalf("production passed adjudication should not create repair draft")
	}
}

func TestAcceptanceProfilesNeedRefresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		profileVersion    string
		acceptanceProfile *acceptanceProfileRecord
		productionProfile *productionAcceptanceProfileRecord
		expect            bool
	}{
		{
			name:           "missing profiles require refresh",
			profileVersion: "v2",
			expect:         true,
		},
		{
			name:           "matching profiles stay fresh",
			profileVersion: "v2",
			acceptanceProfile: &acceptanceProfileRecord{
				ProfileVersion: "v2",
			},
			productionProfile: &productionAcceptanceProfileRecord{
				ProfileVersion: "v2",
			},
			expect: false,
		},
		{
			name:           "version mismatch refreshes",
			profileVersion: "v3",
			acceptanceProfile: &acceptanceProfileRecord{
				ProfileVersion: "v2",
			},
			productionProfile: &productionAcceptanceProfileRecord{
				ProfileVersion: "v2",
			},
			expect: true,
		},
		{
			name: "blank requested version only checks presence",
			acceptanceProfile: &acceptanceProfileRecord{
				ProfileVersion: "v1",
			},
			productionProfile: &productionAcceptanceProfileRecord{
				ProfileVersion: "v0",
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := acceptanceProfilesNeedRefresh(tt.profileVersion, tt.acceptanceProfile, tt.productionProfile); got != tt.expect {
				t.Fatalf("unexpected refresh decision: got %v want %v", got, tt.expect)
			}
		})
	}
}

func TestNormalizeAcceptanceRunMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "blank defaults to queued", input: "", expect: "queued"},
		{name: "known value normalizes case", input: " RUNNING ", expect: "running"},
		{name: "manual allowed", input: "manual", expect: "manual"},
		{name: "unknown value falls back", input: "unexpected_mode", expect: "queued"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeAcceptanceRunMode(tt.input); got != tt.expect {
				t.Fatalf("unexpected normalized mode: got %s want %s", got, tt.expect)
			}
		})
	}
}
