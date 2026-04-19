package v1

import "github.com/gogf/gf/v2/frame/g"

type HealthzReq struct {
	g.Meta `path:"/api/v3/runtime/healthz" method:"get" tags:"Runtime" summary:"Runtime health check"`
}

type HealthzRes struct {
	Status  string `json:"status"`
	BaseURL string `json:"base_url"`
}

type ExecutionViewReq struct {
	g.Meta      `path:"/api/v3/projects/{id}/execution-view" method:"get" tags:"Runtime" summary:"Execution view"`
	ProjectID   string `json:"id" in:"path" v:"required"`
	BindingID   string `json:"binding_id" in:"query"`
	EventLimit  int    `json:"event_limit" in:"query"`
	ReplayLimit int    `json:"replay_limit" in:"query"`
	LogLimit    int    `json:"log_limit" in:"query"`
}

type ExecutionViewRes struct {
	RuntimeHealth  HealthzRes              `json:"runtime_health"`
	RuntimeError   string                  `json:"runtime_error,omitempty"`
	RecentBindings []RunBindingView        `json:"recent_bindings"`
	LatestBinding  *GetRunBindingDetailRes `json:"latest_binding,omitempty"`
	ReplayError    string                  `json:"replay_error,omitempty"`
	ReplaySummary  *ExecutionReplaySummary `json:"replay_summary,omitempty"`
	ReplayTimeline []ExecutionReplayItem   `json:"replay_timeline"`
	LogSegments    []ExecutionLogSegment   `json:"log_segments"`
}

type StartRunReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/runtime-runs" method:"post" tags:"Runtime" summary:"Start runtime run"`
	ProjectID string `json:"id" in:"path" v:"required"`
	TaskID    string `json:"task_id" v:"required"`
	BrainKind string `json:"brain_kind"`
	Prompt    string `json:"prompt" v:"required"`
	Workdir   string `json:"workdir"`
	MaxTurns  int    `json:"max_turns"`
	Provider  string `json:"provider"`
}

type StartRunRes struct {
	CommandID  string         `json:"command_id"`
	Accepted   bool           `json:"accepted"`
	ResourceID string         `json:"resource_id"`
	NextAction string         `json:"next_action"`
	RunBinding RunBindingView `json:"run_binding"`
}

type GetRunBindingReq struct {
	g.Meta    `path:"/api/v3/runtime-runs/{binding_id}" method:"get" tags:"Runtime" summary:"Get runtime run binding"`
	BindingID string `json:"binding_id" in:"path" v:"required"`
}

type GetRunBindingRes struct {
	RunBinding RunBindingView `json:"run_binding"`
}

type GetRunBindingDetailReq struct {
	g.Meta     `path:"/api/v3/runtime-runs/{binding_id}/detail" method:"get" tags:"Runtime" summary:"Get runtime run binding detail"`
	BindingID  string `json:"binding_id" in:"path" v:"required"`
	EventLimit int    `json:"event_limit" in:"query"`
}

type GetRunBindingDetailRes struct {
	RunBinding   RunBindingView     `json:"run_binding"`
	RuntimeState *BrainRunStateView `json:"runtime_state,omitempty"`
	RuntimeStale bool               `json:"runtime_stale"`
	RuntimeError string             `json:"runtime_error,omitempty"`
	RecentEvents []RunEventListItem `json:"recent_events"`
	RefreshHint  string             `json:"refresh_hint"`
}

type ListRunBindingEventsReq struct {
	g.Meta    `path:"/api/v3/runtime-runs/{binding_id}/events" method:"get" tags:"Runtime" summary:"List runtime run binding events"`
	BindingID string `json:"binding_id" in:"path" v:"required"`
	Limit     int    `json:"limit" in:"query"`
}

type ListRunBindingEventsRes struct {
	RunBindingID string             `json:"run_binding_id"`
	Events       []RunEventListItem `json:"events"`
}

type SyncRunBindingReq struct {
	g.Meta    `path:"/api/v3/runtime-runs/{binding_id}/sync" method:"post" tags:"Runtime" summary:"Sync runtime run binding"`
	BindingID string `json:"binding_id" in:"path" v:"required"`
}

type SyncRunBindingRes struct {
	RunBinding RunBindingView `json:"run_binding"`
}

type ResumeRunBindingReq struct {
	g.Meta    `path:"/api/v3/runtime-runs/{binding_id}/resume" method:"post" tags:"Runtime" summary:"Resume runtime run binding"`
	BindingID string `json:"binding_id" in:"path" v:"required"`
}

type ResumeRunBindingRes struct {
	CommandID  string         `json:"command_id"`
	Accepted   bool           `json:"accepted"`
	ResourceID string         `json:"resource_id"`
	NextAction string         `json:"next_action"`
	RunBinding RunBindingView `json:"run_binding"`
}

type CancelRunBindingReq struct {
	g.Meta    `path:"/api/v3/runtime-runs/{binding_id}" method:"delete" tags:"Runtime" summary:"Cancel runtime run binding"`
	BindingID string `json:"binding_id" in:"path" v:"required"`
}

type CancelRunBindingRes struct {
	CommandID  string         `json:"command_id"`
	Accepted   bool           `json:"accepted"`
	ResourceID string         `json:"resource_id"`
	NextAction string         `json:"next_action"`
	RunBinding RunBindingView `json:"run_binding"`
}

type RunBindingView struct {
	BindingID  string `json:"binding_id"`
	ProjectID  string `json:"project_id"`
	TaskID     string `json:"task_id"`
	BrainKind  string `json:"brain_kind"`
	RunID      string `json:"run_id"`
	RunStatus  string `json:"run_status"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at,omitempty"`
	LastSyncAt string `json:"last_sync_at,omitempty"`
}

type RunEventListItem struct {
	EventID    string `json:"event_id"`
	SequenceNo int    `json:"sequence_no"`
	EventType  string `json:"event_type"`
	EventLevel string `json:"event_level,omitempty"`
	Summary    string `json:"summary"`
	Payload    string `json:"payload,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type BrainRunStateView struct {
	RunID       string `json:"run_id"`
	ExecutionID string `json:"execution_id,omitempty"`
	Status      string `json:"status"`
	Brain       string `json:"brain,omitempty"`
	Prompt      string `json:"prompt,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type ExecutionReplaySummary struct {
	RunID           string `json:"run_id"`
	ProjectID       string `json:"project_id"`
	BrainKind       string `json:"brain_kind,omitempty"`
	Status          string `json:"status"`
	StartedAt       string `json:"started_at,omitempty"`
	EndedAt         string `json:"ended_at,omitempty"`
	EventCount      int    `json:"event_count"`
	ReplayCount     int    `json:"replay_count"`
	LogSegmentCount int    `json:"log_segment_count"`
	ArtifactReady   int    `json:"artifact_ready"`
	ArtifactMissing int    `json:"artifact_missing"`
	ArtifactPruned  int    `json:"artifact_pruned"`
}

type ExecutionReplayItem struct {
	ReplayID         string `json:"replay_id"`
	ReplayType       string `json:"replay_type"`
	Title            string `json:"title"`
	Summary          string `json:"summary,omitempty"`
	Status           string `json:"status"`
	PreviewAvailable bool   `json:"preview_available"`
	RawTarget        string `json:"raw_target,omitempty"`
	CreatedAt        string `json:"created_at"`
}

type ExecutionLogSegment struct {
	SegmentID  string `json:"segment_id"`
	StreamKind string `json:"stream_kind"`
	SeqNo      int    `json:"seq_no"`
	Status     string `json:"status"`
	Size       int64  `json:"size"`
	RawTarget  string `json:"raw_target,omitempty"`
}
