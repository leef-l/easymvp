package braincontracts

import "encoding/json"

// VerificationContract defines the structured verification requirements for a
// compiled task. It replaces the raw json.RawMessage with a typed contract so
// that accepting / reworking / completed stages can consume it directly.
type VerificationContract struct {
	ContractID        string                    `json:"contract_id"`
	RequiredChecks    []VerificationCheckItem   `json:"required_checks"`
	RequiredEvidence  []VerificationEvidenceItem `json:"required_evidence"`
	PreferredChannel  string                    `json:"preferred_channel"`
	FallbackChannels  []string                  `json:"fallback_channels"`
	BlockingRules     []string                  `json:"blocking_rules"`
	WarningTolerance  int                       `json:"warning_tolerance"`
	ManualReviewRules []string                  `json:"manual_review_rules"`
}

// VerificationCheckItem is a single required check inside a verification contract.
type VerificationCheckItem struct {
	CheckID       string `json:"check_id"`
	Kind          string `json:"kind"`          // unit_test | integration_test | e2e_test | static_analysis | security_scan | performance_benchmark
	Blocking      bool   `json:"blocking"`
	ExecutorBrain string `json:"executor_brain"` // verifier | browser | code | central
	Summary       string `json:"summary"`
}

// VerificationEvidenceItem is a single required evidence item.
type VerificationEvidenceItem struct {
	EvidenceID    string `json:"evidence_id"`
	Kind          string `json:"kind"`          // screenshot | log | artifact | replay | audit_trail
	Required      bool   `json:"required"`
	ExecutorBrain string `json:"executor_brain"`
	Summary       string `json:"summary"`
}

// ParseVerificationContract attempts to parse a JSON blob into a typed
// VerificationContract. It is a convenience helper for the transition period
// where CompiledTaskItem.VerificationContract is still json.RawMessage.
func ParseVerificationContract(raw json.RawMessage) (*VerificationContract, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var vc VerificationContract
	if err := json.Unmarshal(raw, &vc); err != nil {
		return nil, err
	}
	return &vc, nil
}
