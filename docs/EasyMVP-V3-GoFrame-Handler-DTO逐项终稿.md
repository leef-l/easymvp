# EasyMVP V3 GoFrame Handler DTO 逐项终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-核心API-DTO与TypeScript类型终稿](./EasyMVP-V3-核心API-DTO与TypeScript类型终稿.md)
> 关联文档：[EasyMVP-V3-API路由分组与命令查询边界设计](./EasyMVP-V3-API路由分组与命令查询边界设计.md)
> 目标：按 handler 粒度给出 GoFrame request/response struct 终稿，供 `apps/core/api/*/v1` 直接落文件。

## 1. 设计结论

本专题直接给出第一批 handler DTO 终稿：

1. `system/healthz`
2. `workspace/home-view`
3. `projects/create`
4. `projects/{id}/workspace-view`
5. `projects/{id}/plan-view`
6. `projects/{id}/acceptance-view`
7. `runtime/healthz`
8. `projects/{id}/runtime-runs`
9. `runtime-runs/{binding_id}`
10. `runtime-runs/{binding_id}/detail`
11. `runtime-runs/{binding_id}/events`
12. `runtime-runs/{binding_id}/sync`
13. `runtime-runs/{binding_id}/resume`
14. `projects/{id}/plan/compile`
15. `projects/{id}/acceptance-runs`
16. `manual-decisions`
17. `projects/{project_id}/runs/{run_id}/replay-summary`
18. `projects/{project_id}/runs/{run_id}/replay-timeline`
19. `projects/{project_id}/runs/{run_id}/replay-items/{replay_id}`
20. `projects/{project_id}/runs/{run_id}/replay-items/{replay_id}/raw`
21. `projects/{project_id}/runs/{run_id}/log-segments`
22. `projects/{project_id}/runs/{run_id}/log-segments/{segment_id}/raw`

## 2. System

### 2.1 `GET /api/v3/system/healthz`

```go
type HealthReq struct {
    g.Meta `path:"/api/v3/system/healthz" method:"get" tags:"System" summary:"System health check"`
}

type HealthRes struct {
    Status    string `json:"status"`
    Service   string `json:"service"`
    Version   string `json:"version"`
    Timestamp string `json:"timestamp"`
}
```

## 3. Workspace

### 3.1 `GET /api/v3/workspace/home-view`

```go
type HomeViewReq struct {
    g.Meta `path:"/api/v3/workspace/home-view" method:"get" tags:"Workspace" summary:"Workspace home view"`
}

type HomeViewRes struct {
    Summary          HomeSummary          `json:"summary"`
    ActiveProjects   []ProjectCard        `json:"active_projects"`
    NeedAttention    []NeedAttentionItem  `json:"need_attention"`
    RecentActivity   []LiveActivityItem   `json:"recent_activity"`
    ReleaseReadiness []ReleaseReadiness   `json:"release_readiness"`
}
```

## 4. Projects

### 4.1 `POST /api/v3/projects`

```go
type CreateProjectReq struct {
    g.Meta          `path:"/api/v3/projects" method:"post" tags:"Projects" summary:"Create project"`
    Name            string `json:"name" v:"required"`
    ProjectCategory string `json:"project_category" v:"required"`
    GoalSummary     string `json:"goal_summary" v:"required"`
    WorkspaceRoot   string `json:"workspace_root" v:"required"`
    RepoRoot        string `json:"repo_root"`
}

type CreateProjectRes struct {
    CommandId  string `json:"command_id"`
    Accepted   bool   `json:"accepted"`
    ResourceId string `json:"resource_id"`
    NextAction string `json:"next_action"`
}
```

### 4.2 `GET /api/v3/projects/{id}/workspace-view`

```go
type ProjectWorkspaceViewReq struct {
    g.Meta    `path:"/api/v3/projects/{id}/workspace-view" method:"get" tags:"Projects" summary:"Project workspace view"`
    ProjectId string `json:"id" in:"path" v:"required"`
}

type ProjectWorkspaceViewRes struct {
    ProjectSnapshot    ProjectSnapshot     `json:"project_snapshot"`
    StageProgress      []StageProgressItem `json:"stage_progress"`
    LiveActivity       []LiveActivityItem  `json:"live_activity"`
    ActionInbox        []ActionInboxItem   `json:"action_inbox"`
    AcceptanceCoverage AcceptanceCoverage  `json:"acceptance_coverage"`
}
```

## 5. Plan

## 5. Runtime

### 5.1 `GET /api/v3/runtime/healthz`

```go
type HealthzReq struct {
    g.Meta `path:"/api/v3/runtime/healthz" method:"get" tags:"Runtime" summary:"Runtime health check"`
}

type HealthzRes struct {
    Status  string `json:"status"`
    BaseUrl string `json:"base_url"`
}
```

### 5.2 `POST /api/v3/projects/{id}/runtime-runs`

```go
type StartRunReq struct {
    g.Meta    `path:"/api/v3/projects/{id}/runtime-runs" method:"post" tags:"Runtime" summary:"Start runtime run"`
    ProjectId string `json:"id" in:"path" v:"required"`
    TaskId    string `json:"task_id" v:"required"`
    BrainKind string `json:"brain_kind"`
    Prompt    string `json:"prompt" v:"required"`
    Workdir   string `json:"workdir"`
    MaxTurns  int    `json:"max_turns"`
    Provider  string `json:"provider"`
}

type StartRunRes struct {
    CommandId  string         `json:"command_id"`
    Accepted   bool           `json:"accepted"`
    ResourceId string         `json:"resource_id"`
    NextAction string         `json:"next_action"`
    RunBinding RunBindingView `json:"run_binding"`
}
```

### 5.3 `GET /api/v3/runtime-runs/{binding_id}`

```go
type GetRunBindingReq struct {
    g.Meta    `path:"/api/v3/runtime-runs/{binding_id}" method:"get" tags:"Runtime" summary:"Get runtime run binding"`
    BindingId string `json:"binding_id" in:"path" v:"required"`
}

type GetRunBindingRes struct {
    RunBinding RunBindingView `json:"run_binding"`
}
```

### 5.4 `GET /api/v3/runtime-runs/{binding_id}/detail`

```go
type GetRunBindingDetailReq struct {
    g.Meta     `path:"/api/v3/runtime-runs/{binding_id}/detail" method:"get" tags:"Runtime" summary:"Get runtime run binding detail"`
    BindingId  string `json:"binding_id" in:"path" v:"required"`
    EventLimit int    `json:"event_limit" in:"query"`
}

type GetRunBindingDetailRes struct {
    RunBinding   RunBindingView      `json:"run_binding"`
    RuntimeState *BrainRunStateView  `json:"runtime_state,omitempty"`
    RuntimeStale bool                `json:"runtime_stale"`
    RuntimeError string              `json:"runtime_error,omitempty"`
    RecentEvents []RunEventListItem  `json:"recent_events"`
    RefreshHint  string              `json:"refresh_hint"`
}
```

### 5.5 `GET /api/v3/runtime-runs/{binding_id}/events`

```go
type ListRunBindingEventsReq struct {
    g.Meta    `path:"/api/v3/runtime-runs/{binding_id}/events" method:"get" tags:"Runtime" summary:"List runtime run binding events"`
    BindingId string `json:"binding_id" in:"path" v:"required"`
    Limit     int    `json:"limit" in:"query"`
}

type ListRunBindingEventsRes struct {
    RunBindingId string             `json:"run_binding_id"`
    Events       []RunEventListItem `json:"events"`
}
```

### 5.6 `POST /api/v3/runtime-runs/{binding_id}/sync`

```go
type SyncRunBindingReq struct {
    g.Meta    `path:"/api/v3/runtime-runs/{binding_id}/sync" method:"post" tags:"Runtime" summary:"Sync runtime run binding"`
    BindingId string `json:"binding_id" in:"path" v:"required"`
}

type SyncRunBindingRes struct {
    RunBinding RunBindingView `json:"run_binding"`
}
```

### 5.7 `DELETE /api/v3/runtime-runs/{binding_id}`

```go
type CancelRunBindingReq struct {
    g.Meta    `path:"/api/v3/runtime-runs/{binding_id}" method:"delete" tags:"Runtime" summary:"Cancel runtime run binding"`
    BindingId string `json:"binding_id" in:"path" v:"required"`
}

type CancelRunBindingRes struct {
    CommandId  string         `json:"command_id"`
    Accepted   bool           `json:"accepted"`
    ResourceId string         `json:"resource_id"`
    NextAction string         `json:"next_action"`
    RunBinding RunBindingView `json:"run_binding"`
}
```

### 5.8 `POST /api/v3/runtime-runs/{binding_id}/resume`

```go
type ResumeRunBindingReq struct {
    g.Meta    `path:"/api/v3/runtime-runs/{binding_id}/resume" method:"post" tags:"Runtime" summary:"Resume runtime run binding"`
    BindingId string `json:"binding_id" in:"path" v:"required"`
}

type ResumeRunBindingRes struct {
    CommandId  string         `json:"command_id"`
    Accepted   bool           `json:"accepted"`
    ResourceId string         `json:"resource_id"`
    NextAction string         `json:"next_action"`
    RunBinding RunBindingView `json:"run_binding"`
}

type RunBindingView struct {
    BindingId  string `json:"binding_id"`
    ProjectId  string `json:"project_id"`
    TaskId     string `json:"task_id"`
    BrainKind  string `json:"brain_kind"`
    RunId      string `json:"run_id"`
    RunStatus  string `json:"run_status"`
    StartedAt  string `json:"started_at"`
    FinishedAt string `json:"finished_at,omitempty"`
    LastSyncAt string `json:"last_sync_at,omitempty"`
}

type RunEventListItem struct {
    EventId    string `json:"event_id"`
    SequenceNo int    `json:"sequence_no"`
    EventType  string `json:"event_type"`
    EventLevel string `json:"event_level,omitempty"`
    Summary    string `json:"summary"`
    Payload    string `json:"payload,omitempty"`
    CreatedAt  string `json:"created_at"`
}

type BrainRunStateView struct {
    RunId       string `json:"run_id"`
    ExecutionId string `json:"execution_id,omitempty"`
    Status      string `json:"status"`
    Brain       string `json:"brain,omitempty"`
    Prompt      string `json:"prompt,omitempty"`
    CreatedAt   string `json:"created_at,omitempty"`
}
```

## 5.6 Replay

### 5.6.1 `GET /api/v3/projects/{project_id}/runs/{run_id}/replay-summary`

```go
type GetReplaySummaryReq struct {
    g.Meta    `path:"/api/v3/projects/{project_id}/runs/{run_id}/replay-summary" method:"get" tags:"Replay" summary:"Get replay summary"`
    ProjectId string `json:"project_id" in:"path" v:"required"`
    RunId     string `json:"run_id" in:"path" v:"required"`
}

type GetReplaySummaryRes struct {
    RunId                 string                `json:"run_id"`
    ProjectId             string                `json:"project_id"`
    BrainKind             string                `json:"brain_kind,omitempty"`
    Status                string                `json:"status"`
    StartedAt             string                `json:"started_at,omitempty"`
    EndedAt               string                `json:"ended_at,omitempty"`
    EventCount            int                   `json:"event_count"`
    ReplayCount           int                   `json:"replay_count"`
    LogSegmentCount       int                   `json:"log_segment_count"`
    ArtifactStatusSummary ReplayArtifactSummary `json:"artifact_status_summary"`
    EntryPoints           []ReplayEntryPoint    `json:"entry_points"`
}

type ReplayArtifactSummary struct {
    Available int `json:"available"`
    Missing   int `json:"missing"`
    Pruned    int `json:"pruned"`
}

type ReplayEntryPoint struct {
    ReplayId   string `json:"replay_id"`
    ReplayType string `json:"replay_type"`
    Summary    string `json:"summary,omitempty"`
    FilePath   string `json:"file_path,omitempty"`
    CreatedAt  string `json:"created_at"`
}
```

### 5.6.2 `GET /api/v3/projects/{project_id}/runs/{run_id}/replay-timeline`

```go
type GetReplayTimelineReq struct {
    g.Meta     `path:"/api/v3/projects/{project_id}/runs/{run_id}/replay-timeline" method:"get" tags:"Replay" summary:"Get replay timeline"`
    ProjectId  string `json:"project_id" in:"path" v:"required"`
    RunId      string `json:"run_id" in:"path" v:"required"`
    Limit      int    `json:"limit" in:"query"`
    ReplayType string `json:"replay_type" in:"query"`
}

type GetReplayTimelineRes struct {
    RunId       string               `json:"run_id"`
    ProjectId   string               `json:"project_id"`
    Items       []ReplayTimelineItem `json:"items"`
    RefreshHint string               `json:"refresh_hint"`
}

type ReplayTimelineItem struct {
    ReplayId         string `json:"replay_id"`
    SeqNo            int    `json:"seq_no"`
    ReplayType       string `json:"replay_type"`
    Title            string `json:"title"`
    Summary          string `json:"summary,omitempty"`
    Status           string `json:"status"`
    PreviewAvailable bool   `json:"preview_available"`
    RawTarget        string `json:"raw_target,omitempty"`
    CreatedAt        string `json:"created_at"`
}
```

### 5.6.3 `GET /api/v3/projects/{project_id}/runs/{run_id}/replay-items/{replay_id}`

```go
type GetReplayDetailReq struct {
    g.Meta    `path:"/api/v3/projects/{project_id}/runs/{run_id}/replay-items/{replay_id}" method:"get" tags:"Replay" summary:"Get replay detail"`
    ProjectId string `json:"project_id" in:"path" v:"required"`
    RunId     string `json:"run_id" in:"path" v:"required"`
    ReplayId  string `json:"replay_id" in:"path" v:"required"`
}

type GetReplayDetailRes struct {
    ReplayId           string               `json:"replay_id"`
    ReplayKind         string               `json:"replay_kind"`
    Title              string               `json:"title"`
    Summary            string               `json:"summary,omitempty"`
    SourceObjectKind   string               `json:"source_object_kind,omitempty"`
    SourceObjectId     string               `json:"source_object_id,omitempty"`
    EventId            string               `json:"event_id,omitempty"`
    TraceId            string               `json:"trace_id,omitempty"`
    SpanId             string               `json:"span_id,omitempty"`
    Status             string               `json:"status"`
    RawPreview         ReplayRawPreview     `json:"raw_preview"`
    RelatedLogSegments []LogSegmentListItem `json:"related_log_segments"`
}

type ReplayRawPreview struct {
    MimeType         string `json:"mime_type,omitempty"`
    PreviewAvailable bool   `json:"preview_available"`
    RawTarget        string `json:"raw_target,omitempty"`
}
```

### 5.6.4 `GET /api/v3/projects/{project_id}/runs/{run_id}/replay-items/{replay_id}/raw`

```go
type GetReplayRawReq struct {
    g.Meta    `path:"/api/v3/projects/{project_id}/runs/{run_id}/replay-items/{replay_id}/raw" method:"get" tags:"Replay" summary:"Get replay raw content"`
    ProjectId string `json:"project_id" in:"path" v:"required"`
    RunId     string `json:"run_id" in:"path" v:"required"`
    ReplayId  string `json:"replay_id" in:"path" v:"required"`
    Limit     int    `json:"limit" in:"query"`
}

type GetReplayRawRes struct {
    ReplayId   string `json:"replay_id"`
    Status     string `json:"status"`
    MimeType   string `json:"mime_type,omitempty"`
    Encoding   string `json:"encoding"`
    Content    string `json:"content,omitempty"`
    Truncated  bool   `json:"truncated"`
}
```

### 5.6.5 `GET /api/v3/projects/{project_id}/runs/{run_id}/log-segments`

```go
type ListLogSegmentsReq struct {
    g.Meta    `path:"/api/v3/projects/{project_id}/runs/{run_id}/log-segments" method:"get" tags:"Replay" summary:"List log segments"`
    ProjectId string `json:"project_id" in:"path" v:"required"`
    RunId     string `json:"run_id" in:"path" v:"required"`
    Limit     int    `json:"limit" in:"query"`
}

type ListLogSegmentsRes struct {
    RunId       string               `json:"run_id"`
    ProjectId   string               `json:"project_id"`
    Segments    []LogSegmentListItem `json:"segments"`
    RefreshHint string               `json:"refresh_hint"`
}

type LogSegmentListItem struct {
    SegmentId  string `json:"segment_id"`
    StreamKind string `json:"stream_kind"`
    SeqNo      int    `json:"seq_no"`
    StartedAt  string `json:"started_at,omitempty"`
    EndedAt    string `json:"ended_at,omitempty"`
    Status     string `json:"status"`
    Size       int64  `json:"size"`
    RawTarget  string `json:"raw_target,omitempty"`
}
```

### 5.6.6 `GET /api/v3/projects/{project_id}/runs/{run_id}/log-segments/{segment_id}/raw`

```go
type GetLogSegmentRawReq struct {
    g.Meta    `path:"/api/v3/projects/{project_id}/runs/{run_id}/log-segments/{segment_id}/raw" method:"get" tags:"Replay" summary:"Get log segment raw content"`
    ProjectId string `json:"project_id" in:"path" v:"required"`
    RunId     string `json:"run_id" in:"path" v:"required"`
    SegmentId string `json:"segment_id" in:"path" v:"required"`
    Limit     int    `json:"limit" in:"query"`
}

type GetLogSegmentRawRes struct {
    SegmentId  string `json:"segment_id"`
    StreamKind string `json:"stream_kind"`
    Status     string `json:"status"`
    Content    string `json:"content,omitempty"`
    Truncated  bool   `json:"truncated"`
}
```

## 6. Plan

### 6.1 `GET /api/v3/projects/{id}/plan-view`

```go
type PlanViewReq struct {
    g.Meta    `path:"/api/v3/projects/{id}/plan-view" method:"get" tags:"Plan" summary:"Plan view"`
    ProjectId string `json:"id" in:"path" v:"required"`
}

type PlanViewRes struct {
    Draft          PlanDraftView      `json:"draft"`
    Review         PlanReviewView     `json:"review"`
    Compiled       CompiledPlanView   `json:"compiled"`
    TaskProjection []CompiledTaskView `json:"task_projection"`
    DiffSummary    DiffSummary        `json:"diff_summary"`
}
```

### 6.2 `POST /api/v3/projects/{id}/plan/compile`

```go
type CompilePlanReq struct {
    g.Meta         `path:"/api/v3/projects/{id}/plan/compile" method:"post" tags:"Plan" summary:"Compile plan"`
    ProjectId      string `json:"id" in:"path" v:"required"`
    PlanDraftId    string `json:"plan_draft_id" v:"required"`
    ForceRecompile bool   `json:"force_recompile"`
}

type CompilePlanRes struct {
    CommandId  string `json:"command_id"`
    Accepted   bool   `json:"accepted"`
    ResourceId string `json:"resource_id"`
    NextAction string `json:"next_action"`
}
```

## 7. Acceptance

### 7.1 `GET /api/v3/projects/{id}/acceptance-view`

```go
type AcceptanceViewReq struct {
    g.Meta    `path:"/api/v3/projects/{id}/acceptance-view" method:"get" tags:"Acceptance" summary:"Acceptance view"`
    ProjectId string `json:"id" in:"path" v:"required"`
}

type AcceptanceViewRes struct {
    AcceptanceRun AcceptanceRunView   `json:"acceptance_run"`
    CoverageMatrix []CoverageItem     `json:"coverage_matrix"`
    Issues         []AcceptanceIssue  `json:"issues"`
    EvidenceCards  []EvidenceCard     `json:"evidence_cards"`
    ReleaseGate    ReleaseGateView    `json:"release_gate"`
}
```

### 7.2 `POST /api/v3/projects/{id}/acceptance-runs`

```go
type StartAcceptanceReq struct {
    g.Meta         `path:"/api/v3/projects/{id}/acceptance-runs" method:"post" tags:"Acceptance" summary:"Start acceptance"`
    ProjectId      string `json:"id" in:"path" v:"required"`
    ProfileVersion string `json:"profile_version" v:"required"`
    Mode           string `json:"mode" v:"required"`
}

type StartAcceptanceRes struct {
    CommandId  string `json:"command_id"`
    Accepted   bool   `json:"accepted"`
    ResourceId string `json:"resource_id"`
    NextAction string `json:"next_action"`
}
```

## 8. Manual Decisions

### 8.1 `POST /api/v3/manual-decisions`

```go
type ApplyManualDecisionReq struct {
    g.Meta       `path:"/api/v3/manual-decisions" method:"post" tags:"ManualDecision" summary:"Apply manual decision"`
    ProjectId    string `json:"project_id" v:"required"`
    DecisionKind string `json:"decision_kind" v:"required"`
    TargetId     string `json:"target_id" v:"required"`
    Comment      string `json:"comment"`
}

type ApplyManualDecisionRes struct {
    CommandId  string `json:"command_id"`
    Accepted   bool   `json:"accepted"`
    ResourceId string `json:"resource_id"`
    NextAction string `json:"next_action"`
}
```

## 9. 共享嵌套类型建议

这些类型建议落到：

1. `apps/core/api/workspace/v1/types.go`
2. `apps/core/api/plan/v1/types.go`
3. `apps/core/api/acceptance/v1/types.go`
4. `apps/core/api/runtime/v1/runtime.go`

共享类型包括：

1. `ProjectCard`
2. `LiveActivityItem`
3. `ProjectSnapshot`
4. `StageProgressItem`
5. `ActionInboxItem`
6. `AcceptanceCoverage`
7. `CoverageItem`
8. `EvidenceCard`
9. `RunBindingView`

## 10. 实施顺序

1. 先建 `system`
2. 再建 `workspace`
3. 再建 `projects`
4. 再建 `runtime`
5. 再建 `plan`
6. 最后建 `acceptance` 与 `manual-decisions`
