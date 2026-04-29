package service

import (
	"context"
	"strings"
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
