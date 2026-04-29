package braincontracts

import "encoding/json"

// BrainContractEnvelope is the shared top-level response envelope for all
// easymvp-brain contracts.
type BrainContractEnvelope struct {
	SchemaVersion   int              `json:"schema_version"`
	ResultKind      string           `json:"result_kind"`
	ResultVersion   int              `json:"result_version"`
	SourceRefs      []BrainSourceRef `json:"source_refs"`
	DecisionSummary string           `json:"decision_summary"`
	TraceID         string           `json:"trace_id,omitempty"`
	DeploymentMode  string           `json:"deployment_mode,omitempty"`
	BrainEndpoint    string           `json:"brain_endpoint,omitempty"`
	NormalizedStatus string           `json:"normalized_status,omitempty"`
	ResultJSON       json.RawMessage  `json:"result_json"`
}

type BrainSourceRef struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type ArtifactRef struct {
	Kind string `json:"kind,omitempty"`
	ID   string `json:"id,omitempty"`
	Path string `json:"path,omitempty"`
}

type BrainContractError struct {
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	TraceID      string `json:"trace_id,omitempty"`
	Retryable    bool   `json:"retryable"`
}
