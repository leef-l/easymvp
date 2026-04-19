package service

func deriveVerificationCurrentChannel(manualReviewRequired bool) string {
	if manualReviewRequired {
		return "manual_review"
	}
	return "github_actions"
}

func buildVerificationContractJSON(params verificationContractParams) string {
	requiredChecks := append([]string(nil), params.RequiredChecks...)
	requiredEvidence := append([]string(nil), params.RequiredEvidence...)
	if len(requiredChecks) == 0 {
		requiredChecks = []string{"acceptance_profile", "coverage_matrix", "evidence_artifacts"}
	}

	fallbackChannels := []string{"github_actions", "manual_review"}
	if params.ManualReviewRequired {
		fallbackChannels = []string{"manual_review", "github_actions"}
	}

	return mustMarshalJSONString(map[string]any{
		"contract_version":   "2026-04-20",
		"project_category":   params.ProjectCategory,
		"profile_version":    params.ProfileVersion,
		"verification_scope": "release_readiness",
		"verification_goal":  "Validate required checks, evidence, and release gates before allowing project completion.",
		"required_checks":    requiredChecks,
		"required_evidence":  requiredEvidence,
		"preferred_channel":  "high_spec_remote",
		"current_channel":    deriveVerificationCurrentChannel(params.ManualReviewRequired),
		"fallback_channels":  fallbackChannels,
		"manual_review_rules": map[string]any{
			"required": params.ManualReviewRequired,
			"summary":  params.ManualReviewSummary,
		},
		"result_schema_version": "verification-result.v1",
	}, "{}")
}

type verificationContractParams struct {
	ProjectCategory      string
	ProfileVersion       string
	RequiredChecks       []string
	RequiredEvidence     []string
	ManualReviewRequired bool
	ManualReviewSummary  string
}
