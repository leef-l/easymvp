package v1

import "github.com/gogf/gf/v2/frame/g"

type ProjectDiagnosticItem struct {
	ID                string `json:"id"`
	Scope             string `json:"scope"`
	Severity          string `json:"severity"`
	ErrorCode         string `json:"error_code"`
	Summary           string `json:"summary"`
	Category          string `json:"category,omitempty"`
	Component         string `json:"component,omitempty"`
	Field             string `json:"field,omitempty"`
	RecommendedAction string `json:"recommended_action,omitempty"`
	RelatedPage       string `json:"related_page,omitempty"`
	DetailJSON        string `json:"detail_json,omitempty"`
	ProjectID         string `json:"project_id,omitempty"`
	TaskID            string `json:"task_id,omitempty"`
	RunID             string `json:"run_id,omitempty"`
	BindingID         string `json:"binding_id,omitempty"`
	CreatedAt         string `json:"created_at"`
}

type ListProjectDiagnosticsReq struct {
	g.Meta    `path:"/api/v3/projects/{project_id}/diagnostic-records" tags:"System" method:"get" summary:"List project diagnostic records"`
	ProjectID string `json:"project_id" in:"path" v:"required"`
	Limit     int    `json:"limit" in:"query"`
}

type ListProjectDiagnosticsRes struct {
	Items            []ProjectDiagnosticItem            `json:"items"`
	CategoryCounts   map[string]int                     `json:"category_counts,omitempty"`
	LatestAuditLogs  []ProjectAuditFact                 `json:"latest_audit_logs,omitempty"`
	LinkedRuns       []ProjectLinkedRun                 `json:"linked_runs,omitempty"`
	ArtifactContexts []ProjectArtifactDiagnosticContext `json:"artifact_contexts,omitempty"`
	EvidenceOverview ProjectEvidenceOverview            `json:"evidence_overview"`
	VerificationRead ProjectVerificationRead            `json:"verification_read"`
	RefreshHint      string                             `json:"refresh_hint"`
}

type ProjectAuditFact struct {
	ID        string `json:"id"`
	EventType string `json:"event_type"`
	ActorKind string `json:"actor_kind"`
	Summary   string `json:"summary"`
	CreatedAt string `json:"created_at"`
}

type ProjectLinkedRun struct {
	RunID                string `json:"run_id"`
	BindingID            string `json:"binding_id,omitempty"`
	TaskID               string `json:"task_id,omitempty"`
	RunStatus            string `json:"run_status,omitempty"`
	ReplayCount          int    `json:"replay_count"`
	LogSegmentCount      int    `json:"log_segment_count"`
	ArtifactReady        int    `json:"artifact_ready"`
	ArtifactMissing      int    `json:"artifact_missing"`
	ArtifactPruned       int    `json:"artifact_pruned"`
	LatestReplayID       string `json:"latest_replay_id,omitempty"`
	LatestReplayType     string `json:"latest_replay_type,omitempty"`
	LatestReplayTitle    string `json:"latest_replay_title,omitempty"`
	LatestEventID        string `json:"latest_event_id,omitempty"`
	LatestEventType      string `json:"latest_event_type,omitempty"`
	LatestEventAt        string `json:"latest_event_at,omitempty"`
	LatestCheckpointID   string `json:"latest_checkpoint_id,omitempty"`
	LatestCheckpointType string `json:"latest_checkpoint_type,omitempty"`
	LatestCheckpointAt   string `json:"latest_checkpoint_at,omitempty"`
}

type ProjectEvidenceOverview struct {
	TotalCount      int                  `json:"total_count"`
	LatestEvidence  []ProjectEvidenceRef `json:"latest_evidence"`
	MissingFiles    []ProjectEvidenceRef `json:"missing_files,omitempty"`
	MissingRequired []string             `json:"missing_required,omitempty"`
	FailedChecks    []string             `json:"failed_checks,omitempty"`
}

type ProjectEvidenceRef struct {
	ID           string `json:"id"`
	Surface      string `json:"surface"`
	Journey      string `json:"journey,omitempty"`
	EvidenceType string `json:"evidence_type"`
	FilePath     string `json:"file_path"`
	CapturedAt   string `json:"captured_at"`
}

type ProjectArtifactDiagnosticContext struct {
	Kind              string `json:"kind"`
	Source            string `json:"source"`
	ID                string `json:"id"`
	RunID             string `json:"run_id,omitempty"`
	Status            string `json:"status"`
	FilePath          string `json:"file_path,omitempty"`
	Summary           string `json:"summary,omitempty"`
	RecommendedAction string `json:"recommended_action"`
}

type ProjectVerificationRead struct {
	Decision                 string   `json:"decision,omitempty"`
	Status                   string   `json:"status,omitempty"`
	CompletionDecision       string   `json:"completion_decision,omitempty"`
	CompletionStatus         string   `json:"completion_status,omitempty"`
	RepairDraftStatus        string   `json:"repair_draft_status,omitempty"`
	RepairStrategy           string   `json:"repair_strategy,omitempty"`
	FaultKind                string   `json:"fault_kind,omitempty"`
	FaultSummary             string   `json:"fault_summary,omitempty"`
	FaultLoopDetected        bool     `json:"fault_loop_detected"`
	VerificationContractJSON string   `json:"verification_contract_json,omitempty"`
	MissingEvidence          []string `json:"missing_evidence,omitempty"`
	FailedChecks             []string `json:"failed_checks,omitempty"`
	RequiredChecks           []string `json:"required_checks,omitempty"`
	RequiredEvidence         []string `json:"required_evidence,omitempty"`
}
