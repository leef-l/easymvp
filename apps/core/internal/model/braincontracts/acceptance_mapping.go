package braincontracts

import "encoding/json"

type AcceptanceMappingInput struct {
	ProjectCategory     string          `json:"project_category"`
	CategoryProfileJSON json.RawMessage `json:"category_profile_json"`
	ArtifactSummaryJSON json.RawMessage `json:"artifact_summary_json"`
	CoverageSummaryJSON json.RawMessage `json:"coverage_summary_json"`
}

type AcceptanceMappingResult struct {
	AcceptanceProfileID           string   `json:"acceptance_profile_id"`
	ProductionAcceptanceProfileID string   `json:"production_acceptance_profile_id"`
	RequiredSurfaces              []string `json:"required_surfaces"`
	RequiredJourneys              []string `json:"required_journeys"`
	RequiredEvidence              []string `json:"required_evidence"`
	ReleaseRequirements           []string `json:"release_requirements,omitempty"`
}
