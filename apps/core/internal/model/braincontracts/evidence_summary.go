package braincontracts

// EvidenceSummary is the normalized runtime DTO for a collection of evidence
// items produced during verification or execution.
type EvidenceSummary struct {
	SummaryID        string         `json:"summary_id"`
	AcceptanceRunID  string         `json:"acceptance_run_id,omitempty"`
	VerificationRunID string        `json:"verification_run_id,omitempty"`
	Items            []EvidenceItem `json:"items"`
	CoverageRatio    float64        `json:"coverage_ratio"`
	MissingRequirements []string    `json:"missing_requirements"`
}

// EvidenceItem is a single piece of evidence.
type EvidenceItem struct {
	EvidenceID     string `json:"evidence_id"`
	EvidenceType   string `json:"evidence_type"`
	SourceBrain    string `json:"source_brain"`
	ArtifactURI    string `json:"artifact_uri"`
	Status         string `json:"status"`        // valid | invalid | pending | expired
	GeneratedAt    string `json:"generated_at"`
	RelatedCheckID string `json:"related_check_id,omitempty"`
}
