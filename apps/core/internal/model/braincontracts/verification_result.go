package braincontracts

// VerificationResult is the normalized runtime DTO for a single verification
// execution. It aligns with the verification_contract_json that the task was
// compiled with.
type VerificationResult struct {
	VerificationRunID    string          `json:"verification_run_id"`
	ContractID           string          `json:"contract_id"`
	Channel              string          `json:"channel"`
	ExecutorBrain        string          `json:"executor_brain"`
	StartedAt            string          `json:"started_at"`
	EndedAt              string          `json:"ended_at"`
	Checks               []VerificationCheck `json:"checks"`
	EvidenceRefs         []string        `json:"evidence_refs"`
	BlockingIssueCount   int             `json:"blocking_issue_count"`
	WarningCount         int             `json:"warning_count"`
	Verdict              string          `json:"verdict"` // passed | passed_with_warning | failed | manual_review_required | channel_unavailable
	ReasonSummary        string          `json:"reason_summary"`
}

// VerificationCheck is a single check item inside a VerificationResult.
type VerificationCheck struct {
	CheckID   string `json:"check_id"`
	Status    string `json:"status"`    // passed | failed | skipped | not_run
	Blocking  bool   `json:"blocking"`
	Summary   string `json:"summary"`
	RawRef    string `json:"raw_ref,omitempty"`
}
