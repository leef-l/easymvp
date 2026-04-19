# EasyMVP V3 核心 API DTO 与 TypeScript 类型终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-本地API-DTO与错误返回设计](./EasyMVP-V3-本地API-DTO与错误返回设计.md)
> 关联文档：[EasyMVP-V3-错误码与诊断分级设计](./EasyMVP-V3-错误码与诊断分级设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：把 V3 核心本地 API 继续细化到 GoFrame 请求结构、JSON 字段和前端 TypeScript 类型定义的终稿层。

## 1. 设计结论

这份文档是实现用 DTO 终稿，不再只写样例 JSON。

它要同时服务：

1. GoFrame request / response struct
2. 前端 `TypeScript` client types
3. 事件流 payload

## 2. 通用类型

### 2.1 Error Envelope

```ts
export type ErrorEnvelope = {
  error: {
    code: string
    level: "info" | "warning" | "error" | "critical"
    scope: string
    message: string
    debug_message?: string
    recovery_hint?: string
  }
  meta: {
    request_id: string
  }
}
```

### 2.2 Query Envelope

```ts
export type QueryEnvelope<T> = {
  data: T
  meta: {
    request_id: string
    generated_at: string
  }
}
```

### 2.3 Command Result Envelope

```ts
export type CommandEnvelope<T> = {
  result: T
  meta: {
    request_id: string
  }
}
```

## 3. Workspace Home DTO 终稿

```ts
export type WorkspaceHomeView = {
  summary: {
    active_project_count: number
    blocked_project_count: number
    waiting_decision_count: number
  }
  active_projects: ProjectCard[]
  need_attention: NeedAttentionItem[]
  recent_activity: LiveActivityItem[]
  release_readiness: ReleaseReadinessItem[]
}

export type ProjectCard = {
  project_id: string
  name: string
  project_category: string
  stage: string
  progress_percent: number
  production_status: string
  blocker_count: number
  waiting_action_count: number
  updated_at: string
}

export type NeedAttentionItem = {
  id: string
  project_id: string
  severity: "warning" | "error" | "critical"
  title: string
  summary: string
  recommended_action: string
}
```

## 4. Project Workspace DTO 终稿

```ts
export type ProjectWorkspaceView = {
  project_snapshot: ProjectSnapshot
  stage_progress: StageProgress[]
  live_activity: LiveActivityItem[]
  action_inbox: ActionInboxItem[]
  acceptance_coverage: AcceptanceCoverage
}

export type ProjectSnapshot = {
  project_id: string
  name: string
  project_category: string
  current_stage: string
  production_status: string
  progress_percent: number
}

export type StageProgress = {
  stage: string
  status: string
  progress_percent: number
  blocker_count: number
  started_at?: string
  finished_at?: string
}

export type LiveActivityItem = {
  id: string
  timestamp: string
  level: "info" | "warning" | "error"
  source_kind: string
  summary: string
}

export type ActionInboxItem = {
  id: string
  severity: "warning" | "error" | "critical"
  title: string
  summary: string
  action_label: string
  action_kind: string
}
```

### 4.1 当前总纲口径补充

上面这组首页/单项目 DTO 可以继续保留为旧 V3 首版聚合形态，但按当前钱学森总纲应补一条解释边界：

1. `production_status` 仍可保留为当前现实字段
2. 但不应再被实现者误读为最终完成状态
3. 首页与单项目工作台的更准确方向，应逐步补齐 `decision / completed / manual_checkpoint_required / has_runtime_escalation`

## 5. Plan DTO 终稿

```ts
export type PlanView = {
  draft: PlanDraftView
  review: PlanReviewView
  compiled: CompiledPlanView
  task_projection: CompiledTaskView[]
  diff_summary: DiffSummary
}

export type PlanDraftView = {
  id: string
  version: number
  status: string
  goal_summary: string
}

export type PlanReviewView = {
  id: string
  review_version: number
  decision: string
  blocking_issue_count: number
  advisory_issue_count: number
}

export type CompiledPlanView = {
  id: string
  compiled_version: number
  status: string
  risk_summary?: string
}
```

## 6. Runtime DTO 终稿

```ts
export type RuntimeHealthView = {
  status: string
  base_url: string
}

export type RunBindingView = {
  binding_id: string
  project_id: string
  task_id: string
  brain_kind: string
  run_id: string
  run_status: string
  started_at: string
  finished_at?: string
  last_sync_at?: string
}

export type BrainRunStateView = {
  run_id: string
  execution_id?: string
  status: string
  brain?: string
  prompt?: string
  created_at?: string
}

export type StartRuntimeRunResult = {
  command_id: string
  accepted: boolean
  resource_id: string
  next_action: string
  run_binding: RunBindingView
}

export type RunEventListItem = {
  event_id: string
  sequence_no: number
  event_type: string
  event_level?: string
  summary: string
  payload?: string
  created_at: string
}

export type ListRuntimeRunEventsView = {
  run_binding_id: string
  events: RunEventListItem[]
}

export type RuntimeRunDetailView = {
  run_binding: RunBindingView
  runtime_state?: BrainRunStateView
  runtime_stale: boolean
  runtime_error?: string
  recent_events: RunEventListItem[]
  refresh_hint: string
}

export type ReplayArtifactSummary = {
  available: number
  missing: number
  pruned: number
}

export type ReplayEntryPoint = {
  replay_id: string
  replay_type: string
  summary?: string
  file_path?: string
  created_at: string
}

export type ReplaySummaryView = {
  run_id: string
  project_id: string
  brain_kind?: string
  status: string
  started_at?: string
  ended_at?: string
  event_count: number
  replay_count: number
  log_segment_count: number
  artifact_status_summary: ReplayArtifactSummary
  entry_points: ReplayEntryPoint[]
}

export type ReplayTimelineItem = {
  replay_id: string
  seq_no: number
  replay_type: string
  title: string
  summary?: string
  status: string
  preview_available: boolean
  raw_target?: string
  created_at: string
}

export type ReplayTimelineView = {
  run_id: string
  project_id: string
  items: ReplayTimelineItem[]
  refresh_hint: string
}

export type ReplayRawPreview = {
  mime_type?: string
  preview_available: boolean
  raw_target?: string
}

export type LogSegmentListItem = {
  segment_id: string
  stream_kind: string
  seq_no: number
  started_at?: string
  ended_at?: string
  status: string
  size: number
  raw_target?: string
}

export type ReplayDetailView = {
  replay_id: string
  replay_kind: string
  title: string
  summary?: string
  source_object_kind?: string
  source_object_id?: string
  event_id?: string
  trace_id?: string
  span_id?: string
  status: string
  raw_preview: ReplayRawPreview
  related_log_segments: LogSegmentListItem[]
}

export type ReplayRawView = {
  replay_id: string
  status: string
  mime_type?: string
  encoding: string
  content?: string
  truncated: boolean
}

export type LogSegmentListView = {
  run_id: string
  project_id: string
  segments: LogSegmentListItem[]
  refresh_hint: string
}

export type LogSegmentRawView = {
  segment_id: string
  stream_kind: string
  status: string
  content?: string
  truncated: boolean
}

export type ResumeRuntimeRunResult = {
  command_id: string
  accepted: boolean
  resource_id: string
  next_action: string
  run_binding: RunBindingView
}

export type SyncRuntimeRunResult = {
  run_binding: RunBindingView
}
```

## 7. Acceptance DTO 终稿

```ts
export type AcceptanceView = {
  acceptance_run: AcceptanceRunView
  coverage_matrix: CoverageMatrixItem[]
  issues: AcceptanceIssueView[]
  evidence_cards: EvidenceCardView[]
  release_gate: ReleaseGateView
}

export type AcceptanceRunView = {
  id: string
  status: string
  functional_status: string
  production_status: string
  manual_release_required: boolean
}

export type CoverageMatrixItem = {
  key: string
  kind: "surface" | "journey"
  name: string
  coverage_status: "pass" | "partial" | "missing"
  evidence_count: number
}

export type AcceptanceIssueView = {
  id: string
  severity: string
  blocking: boolean
  summary: string
}

export type ReleaseGateView = {
  status: string
  next_action: string
  summary: string
}
```

### 7.1 当前总纲口径补充

上面这组 DTO 可以继续保留为旧 V3 的首版接口形态，但当前钱学森总纲下不应再把它理解为最终权威对象集合。

建议这样理解：

1. `AcceptanceRunView` 仍可保留为历史兼容/查询聚合中的一部分
2. `AcceptanceView` 的真正主锚点应逐步补到 `VerificationResult + CompletionVerdict`
3. `release_gate` 不能单独定义“是否 completed”

如果后续继续扩 DTO，建议优先补齐：

1. `verification_contract`
2. `verification_result`
3. `completion_verdict`
4. `runtime_escalations`
5. `missing_evidence`
6. `failed_checks`

核心边界：

- `production_status` 不应再被实现者误读为“最终完成状态”
- 最终是否完成，仍以 `CompletionVerdict.completed` 为准

## 8. GoFrame Request Struct 建议

### 8.1 Create Project

```go
type CreateProjectReq struct {
    g.Meta          `path:"/api/v3/projects" method:"post" tags:"Projects" summary:"Create project"`
    Name            string `json:"name" v:"required"`
    ProjectCategory string `json:"project_category" v:"required"`
    GoalSummary     string `json:"goal_summary" v:"required"`
    WorkspaceRoot   string `json:"workspace_root" v:"required"`
    RepoRoot        string `json:"repo_root"`
}
```

### 8.2 Workspace Home

```go
type WorkspaceHomeViewReq struct {
    g.Meta `path:"/api/v3/workspace/home-view" method:"get" tags:"Workspace" summary:"Workspace home view"`
}
```

## 9. 事件流 Payload 终稿

```ts
export type StreamEvent =
  | WorkspaceUpdatedEvent
  | ProjectUpdatedEvent
  | RunStatusChangedEvent
  | AcceptanceUpdatedEvent

export type WorkspaceUpdatedEvent = {
  event_type: "workspace_updated"
  project_id?: string
  timestamp: string
}

export type RunStatusChangedEvent = {
  event_type: "run_status_changed"
  project_id: string
  run_binding_id: string
  status: string
  timestamp: string
}
```

## 10. 建议文件落点

### Go

1. `apps/core/api/system/v1`
2. `apps/core/api/projects/v1`
3. `apps/core/api/workspace/v1`
4. `apps/core/api/plan/v1`
5. `apps/core/api/acceptance/v1`
6. `apps/core/api/runtime/v1`

### TypeScript

1. `apps/desktop/src/shared/contracts/api.ts`
2. `apps/desktop/src/shared/contracts/workspace.ts`
3. `apps/desktop/src/shared/contracts/plan.ts`
4. `apps/desktop/src/shared/contracts/acceptance.ts`

## 10. 后续细分专题

1. 全量 handler DTO 逐项展开
2. 自动生成 TS 类型策略
3. event stream 序列化规范
