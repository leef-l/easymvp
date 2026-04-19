package v1

import "github.com/gogf/gf/v2/frame/g"

type GetReplaySummaryReq struct {
	g.Meta    `path:"/api/v3/projects/{project_id}/runs/{run_id}/replay-summary" method:"get" tags:"Replay" summary:"Get replay summary"`
	ProjectID string `json:"project_id" in:"path" v:"required"`
	RunID     string `json:"run_id" in:"path" v:"required"`
}

type GetReplaySummaryRes struct {
	RunID                 string                 `json:"run_id"`
	ProjectID             string                 `json:"project_id"`
	BrainKind             string                 `json:"brain_kind,omitempty"`
	Status                string                 `json:"status"`
	StartedAt             string                 `json:"started_at,omitempty"`
	EndedAt               string                 `json:"ended_at,omitempty"`
	EventCount            int                    `json:"event_count"`
	ReplayCount           int                    `json:"replay_count"`
	LogSegmentCount       int                    `json:"log_segment_count"`
	ArtifactStatusSummary ReplayArtifactSummary  `json:"artifact_status_summary"`
	EntryPoints           []ReplayEntryPointItem `json:"entry_points"`
}

type GetReplayTimelineReq struct {
	g.Meta     `path:"/api/v3/projects/{project_id}/runs/{run_id}/replay-timeline" method:"get" tags:"Replay" summary:"Get replay timeline"`
	ProjectID  string `json:"project_id" in:"path" v:"required"`
	RunID      string `json:"run_id" in:"path" v:"required"`
	Limit      int    `json:"limit" in:"query"`
	ReplayType string `json:"replay_type" in:"query"`
}

type GetReplayTimelineRes struct {
	RunID       string               `json:"run_id"`
	ProjectID   string               `json:"project_id"`
	Items       []ReplayTimelineItem `json:"items"`
	RefreshHint string               `json:"refresh_hint"`
}

type GetReplayDetailReq struct {
	g.Meta    `path:"/api/v3/projects/{project_id}/runs/{run_id}/replay-items/{replay_id}" method:"get" tags:"Replay" summary:"Get replay detail"`
	ProjectID string `json:"project_id" in:"path" v:"required"`
	RunID     string `json:"run_id" in:"path" v:"required"`
	ReplayID  string `json:"replay_id" in:"path" v:"required"`
}

type GetReplayDetailRes struct {
	ReplayID           string               `json:"replay_id"`
	ReplayKind         string               `json:"replay_kind"`
	Title              string               `json:"title"`
	Summary            string               `json:"summary,omitempty"`
	DomainTaskID       string               `json:"domain_task_id,omitempty"`
	CompiledTaskID     string               `json:"compiled_task_id,omitempty"`
	SourceObjectKind   string               `json:"source_object_kind,omitempty"`
	SourceObjectID     string               `json:"source_object_id,omitempty"`
	EventID            string               `json:"event_id,omitempty"`
	TraceID            string               `json:"trace_id,omitempty"`
	SpanID             string               `json:"span_id,omitempty"`
	Status             string               `json:"status"`
	RawPreview         ReplayRawPreview     `json:"raw_preview"`
	RelatedLogSegments []LogSegmentListItem `json:"related_log_segments"`
}

type GetReplayRawReq struct {
	g.Meta    `path:"/api/v3/projects/{project_id}/runs/{run_id}/replay-items/{replay_id}/raw" method:"get" tags:"Replay" summary:"Get replay raw content"`
	ProjectID string `json:"project_id" in:"path" v:"required"`
	RunID     string `json:"run_id" in:"path" v:"required"`
	ReplayID  string `json:"replay_id" in:"path" v:"required"`
	Limit     int    `json:"limit" in:"query"`
}

type GetReplayRawRes struct {
	ReplayID  string `json:"replay_id"`
	Status    string `json:"status"`
	MimeType  string `json:"mime_type,omitempty"`
	Encoding  string `json:"encoding"`
	Content   string `json:"content,omitempty"`
	Truncated bool   `json:"truncated"`
}

type ListLogSegmentsReq struct {
	g.Meta    `path:"/api/v3/projects/{project_id}/runs/{run_id}/log-segments" method:"get" tags:"Replay" summary:"List log segments"`
	ProjectID string `json:"project_id" in:"path" v:"required"`
	RunID     string `json:"run_id" in:"path" v:"required"`
	Limit     int    `json:"limit" in:"query"`
}

type ListLogSegmentsRes struct {
	RunID       string               `json:"run_id"`
	ProjectID   string               `json:"project_id"`
	Segments    []LogSegmentListItem `json:"segments"`
	RefreshHint string               `json:"refresh_hint"`
}

type GetLogSegmentRawReq struct {
	g.Meta    `path:"/api/v3/projects/{project_id}/runs/{run_id}/log-segments/{segment_id}/raw" method:"get" tags:"Replay" summary:"Get log segment raw content"`
	ProjectID string `json:"project_id" in:"path" v:"required"`
	RunID     string `json:"run_id" in:"path" v:"required"`
	SegmentID string `json:"segment_id" in:"path" v:"required"`
	Limit     int    `json:"limit" in:"query"`
}

type GetLogSegmentRawRes struct {
	SegmentID  string `json:"segment_id"`
	StreamKind string `json:"stream_kind"`
	Status     string `json:"status"`
	Content    string `json:"content,omitempty"`
	Truncated  bool   `json:"truncated"`
}

type ReplayArtifactSummary struct {
	Available int `json:"available"`
	Missing   int `json:"missing"`
	Pruned    int `json:"pruned"`
}

type ReplayEntryPointItem struct {
	DomainTaskID   string `json:"domain_task_id,omitempty"`
	CompiledTaskID string `json:"compiled_task_id,omitempty"`
	ReplayID   string `json:"replay_id"`
	ReplayType string `json:"replay_type"`
	Summary    string `json:"summary,omitempty"`
	FilePath   string `json:"file_path,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type ReplayTimelineItem struct {
	ReplayID         string `json:"replay_id"`
	DomainTaskID     string `json:"domain_task_id,omitempty"`
	CompiledTaskID   string `json:"compiled_task_id,omitempty"`
	SeqNo            int    `json:"seq_no"`
	ReplayType       string `json:"replay_type"`
	Title            string `json:"title"`
	Summary          string `json:"summary,omitempty"`
	Status           string `json:"status"`
	PreviewAvailable bool   `json:"preview_available"`
	RawTarget        string `json:"raw_target,omitempty"`
	CreatedAt        string `json:"created_at"`
}

type ReplayRawPreview struct {
	MimeType         string `json:"mime_type,omitempty"`
	PreviewAvailable bool   `json:"preview_available"`
	RawTarget        string `json:"raw_target,omitempty"`
}

type LogSegmentListItem struct {
	SegmentID  string `json:"segment_id"`
	StreamKind string `json:"stream_kind"`
	SeqNo      int    `json:"seq_no"`
	StartedAt  string `json:"started_at,omitempty"`
	EndedAt    string `json:"ended_at,omitempty"`
	Status     string `json:"status"`
	Size       int64  `json:"size"`
	RawTarget  string `json:"raw_target,omitempty"`
}
