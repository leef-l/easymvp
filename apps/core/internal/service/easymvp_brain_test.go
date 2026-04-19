package service

import (
	"context"
	"strings"
	"testing"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
)

func TestValidateWorkspaceExplanationEnvelopeRejectsInvalidAction(t *testing.T) {
	t.Parallel()

	svc := &sEasyMVPBrain{}
	err := svc.ValidateWorkspaceExplanationEnvelope(context.Background(), &braincontracts.BrainContractEnvelope{
		SchemaVersion:   1,
		ResultKind:      "workspace_explanation",
		ResultVersion:   1,
		SourceRefs:      []braincontracts.BrainSourceRef{{Kind: "project", ID: "proj_1", Version: 1}},
		DecisionSummary: "workspace explanation ready",
		ResultJSON:      []byte(`{"headline":"h","summary":"s"}`),
	}, &braincontracts.WorkspaceExplanationResult{
		Headline: "Workspace headline",
		Summary:  "Workspace summary",
		RecommendedActions: []braincontracts.RecommendedActionItem{
			{
				ActionKey: "",
				Label:     "Open plan",
				Reason:    "Need a valid key",
			},
		},
	})
	if err == nil {
		t.Fatalf("expected invalid recommended action to fail validation")
	}
}

func TestValidateWorkspaceExplanationEnvelopeAcceptsValidResult(t *testing.T) {
	t.Parallel()

	svc := &sEasyMVPBrain{}
	err := svc.ValidateWorkspaceExplanationEnvelope(context.Background(), &braincontracts.BrainContractEnvelope{
		SchemaVersion:   1,
		ResultKind:      "workspace_explanation",
		ResultVersion:   1,
		SourceRefs:      []braincontracts.BrainSourceRef{{Kind: "project", ID: "proj_1", Version: 1}},
		DecisionSummary: "workspace explanation ready",
		ResultJSON:      []byte(`{"headline":"h","summary":"s"}`),
	}, &braincontracts.WorkspaceExplanationResult{
		Headline:    "Workspace headline",
		Summary:     "Workspace summary",
		TopBlockers: []string{"blocker"},
		RecommendedActions: []braincontracts.RecommendedActionItem{
			{
				ActionKey: "open_project_plan",
				Label:     "Open plan",
				Reason:    "Review latest plan state",
				DeepLink:  "proj_1",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected valid workspace explanation to pass validation, got %v", err)
	}
}

func TestExecuteTypedContractPreservesUnsupportedExecutionStatus(t *testing.T) {
	t.Parallel()

	_, _, err := executeTypedContract[braincontracts.WorkspaceExplanationResult](
		context.Background(),
		&workspaceExplanationBrainStub{
			executeResult: &EasyMVPBrainExecuteResult{
				Status: "unsupported",
				Error:  "workspace explanation tool is unavailable",
			},
		},
		"workspace_explanation",
		braincontracts.WorkspaceExplanationInput{},
		func(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.WorkspaceExplanationResult) error {
			return nil
		},
	)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "unsupported capability") {
		t.Fatalf("expected unsupported capability error, got %v", err)
	}
}

func TestExecuteTypedContractPreservesDeniedExecutionStatus(t *testing.T) {
	t.Parallel()

	_, _, err := executeTypedContract[braincontracts.WorkspaceExplanationResult](
		context.Background(),
		&workspaceExplanationBrainStub{
			executeResult: &EasyMVPBrainExecuteResult{
				Status:  "denied",
				Summary: "policy rejected the request",
			},
		},
		"workspace_explanation",
		braincontracts.WorkspaceExplanationInput{},
		func(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.WorkspaceExplanationResult) error {
			return nil
		},
	)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "denied by runtime policy") {
		t.Fatalf("expected denied-by-policy error, got %v", err)
	}
}
