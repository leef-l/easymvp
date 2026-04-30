package service

import (
	"context"
	"fmt"
	"strings"

	acceptancev1 "github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
)

func deriveVerificationCurrentChannel(manualReviewRequired bool) string {
	if manualReviewRequired {
		return "manual_review"
	}
	return "github_actions"
}

func buildVerificationContractJSON(ctx context.Context, params verificationContractParams) string {
	requiredChecks := append([]string(nil), params.RequiredChecks...)
	requiredEvidence := append([]string(nil), params.RequiredEvidence...)
	if len(requiredChecks) == 0 {
		requiredChecks = []string{"acceptance_profile", "coverage_matrix", "evidence_artifacts"}
	}

	fallbackChannels := []string{"github_actions", "manual_review"}
	if params.ManualReviewRequired {
		fallbackChannels = []string{"manual_review", "github_actions"}
	}

	channelAvailable := true
	if params.ChannelUnavailable {
		channelAvailable = false
	}
	environmentAvailable := true
	if params.EnvironmentUnavailable {
		environmentAvailable = false
	}

	currentChannel := deriveVerificationCurrentChannelWithHighSpec(ctx, params.ManualReviewRequired)

	contract := map[string]any{
		"contract_version":   "2026-04-20",
		"project_category":   params.ProjectCategory,
		"profile_version":    params.ProfileVersion,
		"verification_scope": "release_readiness",
		"verification_goal":  "Validate required checks, evidence, and release gates before allowing project completion.",
		"required_checks":    requiredChecks,
		"required_evidence":  requiredEvidence,
		"preferred_channel":  "high_spec_remote",
		"current_channel":    currentChannel,
		"fallback_channels":  fallbackChannels,
		"channel_available":       channelAvailable,
		"environment_available":   environmentAvailable,
		"channel_unavailable_reason": params.ChannelUnavailableReason,
		"manual_review_rules": map[string]any{
			"required": params.ManualReviewRequired,
			"summary":  params.ManualReviewSummary,
		},
		"result_schema_version": "verification-result.v1",
	}
	if strings.TrimSpace(params.BrowserValidationURL) != "" {
		contract["browser_validation_url"] = strings.TrimSpace(params.BrowserValidationURL)
	}
	if len(params.VerifierChecks) > 0 {
		contract["verifier_checks"] = params.VerifierChecks
	}
	return mustMarshalJSONString(contract, "{}")
}

type verificationContractParams struct {
	ProjectCategory          string
	ProfileVersion           string
	RequiredChecks           []string
	RequiredEvidence         []string
	ManualReviewRequired     bool
	ManualReviewSummary      string
	ChannelUnavailable       bool
	EnvironmentUnavailable   bool
	ChannelUnavailableReason string
	BrowserValidationURL     string
	VerifierChecks           []VerifierCheck
}

// ---------------------------------------------------------------------------
// C-03: ContractGap — diff between contract requirements and actual results
// ---------------------------------------------------------------------------

// buildContractGapView compares verification_contract required_checks /
// required_evidence against the actual VerificationResultView to produce
// a structured gap summary for front-end rendering.
func buildContractGapView(v acceptancev1.VerificationResultView) acceptancev1.ContractGapView {
	failedSet := toStringSet(v.FailedChecks)
	missingSet := toStringSet(v.MissingEvidence)

	gap := acceptancev1.ContractGapView{
		BlockerChecks:   make([]acceptancev1.ContractGapItem, 0, len(failedSet)),
		WarningChecks:   make([]acceptancev1.ContractGapItem, 0),
		MissingEvidence: make([]acceptancev1.ContractGapItem, 0, len(missingSet)),
	}

	// Classify failed checks as blockers.
	for _, check := range v.FailedChecks {
		check = strings.TrimSpace(check)
		if check == "" {
			continue
		}
		gap.BlockerChecks = append(gap.BlockerChecks, acceptancev1.ContractGapItem{
			Key:      check,
			Label:    humanizeAcceptanceKey(check),
			Severity: "blocker",
			Status:   "failed",
			Detail:   "Required check failed during verification.",
		})
	}

	// Required checks that are neither failed nor evidently passed → warning.
	for _, check := range v.RequiredChecks {
		check = strings.TrimSpace(check)
		if check == "" || failedSet[strings.ToLower(check)] {
			continue
		}
		// If verification is not yet completed and the check is required, flag it.
		if !v.Completed && v.Status != "passed" {
			gap.WarningChecks = append(gap.WarningChecks, acceptancev1.ContractGapItem{
				Key:      check,
				Label:    humanizeAcceptanceKey(check),
				Severity: "warning",
				Status:   "not_run",
				Detail:   "Required check has not yet been executed.",
			})
		}
	}

	// Classify missing evidence items.
	for _, evidence := range v.MissingEvidence {
		evidence = strings.TrimSpace(evidence)
		if evidence == "" {
			continue
		}
		gap.MissingEvidence = append(gap.MissingEvidence, acceptancev1.ContractGapItem{
			Key:      evidence,
			Label:    humanizeAcceptanceKey(evidence),
			Severity: "missing",
			Status:   "missing",
			Detail:   "Required evidence has not been captured.",
		})
	}

	gap.HasGap = len(gap.BlockerChecks) > 0 || len(gap.MissingEvidence) > 0
	gap.Summary = deriveContractGapSummary(gap)
	return gap
}

func deriveContractGapSummary(gap acceptancev1.ContractGapView) string {
	parts := make([]string, 0, 3)
	if n := len(gap.BlockerChecks); n > 0 {
		parts = append(parts, fmt.Sprintf("%d blocker(s)", n))
	}
	if n := len(gap.WarningChecks); n > 0 {
		parts = append(parts, fmt.Sprintf("%d warning(s)", n))
	}
	if n := len(gap.MissingEvidence); n > 0 {
		parts = append(parts, fmt.Sprintf("%d missing evidence", n))
	}
	if len(parts) == 0 {
		return "No verification contract gap detected."
	}
	return "Contract gap: " + strings.Join(parts, ", ") + "."
}

func toStringSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item))
		if key != "" {
			m[key] = true
		}
	}
	return m
}
