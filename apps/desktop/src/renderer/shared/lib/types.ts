export type VerificationResultSummary = {
  status?: string;
  decision?: string;
  completed?: boolean;
  summary?: string;
  required_checks?: string[];
  required_evidence?: string[];
  preferred_verification_channel?: string;
  missing_evidence?: string[];
  failed_checks?: string[];
  verification_contract_json?: string;
  source_run_id?: string;
  updated_at?: string;
};

export type CompletionVerdictSummary = {
  decision?: string;
  completed?: boolean;
  summary?: string;
  final_status?: string;
  reason?: string;
  release_ready?: boolean;
  manual_review_required?: boolean;
  manual_release_required?: boolean;
  manual_release_completed?: boolean;
  blocker_count?: number;
  next_action?: string;
  source_run_id?: string;
  updated_at?: string;
};

export type RuntimeEscalationSummary = {
  status?: string;
  reason_class?: string;
  source_brain?: string;
  source_task_id?: string;
  run_binding_id?: string;
  run_status?: string;
  severity?: string;
  summary?: string;
  action?: string;
  task_id?: string;
  run_id?: string;
  updated_at?: string;
  policy_denied?: boolean;
};

export type FaultSummary = {
  status?: string;
  blocking_issue_count?: number;
  advisory_issue_count?: number;
  top_issue?: string;
  fault_loop_detected?: boolean;
  fault_kind?: string;
  severity?: string;
  summary?: string;
  failed_checks?: string[];
  affected_tasks?: string[];
  updated_at?: string;
};

export type RepairPlanDraftSummary = {
  id?: string;
  status?: string;
  reason_class?: string;
  repair_strategy?: string;
  reasoning_summary?: string;
  summary?: string;
  updated_tasks?: string[];
  manual_review_required?: boolean;
  updated_at?: string;
};

export type AcceptanceOverview = {
  project_id: string;
  current_stage: string;
  overall_status: string;
  functional_status: string;
  production_status: string;
  release_gate_status: string;
  next_action: string;
  blocking_issue_count: number;
  covered_item_count: number;
  required_item_count: number;
  evidence_card_count: number;
  manual_release_required: boolean;
};

export type WorkspaceOverview = {
  project_id: string;
  current_stage: string;
  stage_status: string;
  risk_level: string;
  production_status: string;
  next_action: string;
  action_required_count: number;
  blocking_issue_count: number;
  manual_release_required: boolean;
  acceptance_run_status?: string;
  manual_review_required: boolean;
  verification_conflict: boolean;
  fault_loop_detected: boolean;
  policy_denied: boolean;
};

export type WorkspaceView = {
  overview?: WorkspaceOverview;
  project_snapshot: {
    project_id: string;
    name: string;
    project_category: string;
    current_stage: string;
    progress_percent: number;
    risk_level: string;
    production_status: string;
    manual_release_required: boolean;
    manual_review_required?: boolean;
    verification_conflict?: boolean;
    fault_loop_detected?: boolean;
    policy_denied?: boolean;
  };
  stage_progress: Array<{
    stage_key: string;
    stage_name: string;
    status: string;
    duration_seconds: number;
    active_item_title: string;
    blocking_issue_count: number;
  }>;
  live_activity: Array<{
    event_id: string;
    event_type: string;
    title: string;
    source_brain: string;
    source_task_id: string;
    occurred_at: string;
    requires_action: boolean;
  }>;
  action_inbox: Array<{
    item_id: string;
    title: string;
    severity: string;
    is_blocking: boolean;
    recommended_action: string;
    target_id: string;
  }>;
  acceptance_coverage: {
    category: string;
    covered_surfaces: number;
    required_surfaces: number;
    covered_journeys: number;
    required_journeys: number;
    evidence_ready: number;
    evidence_required: number;
    production_passed: boolean;
  };
  workspace_explanation: {
    headline: string;
    summary: string;
    top_blockers: string[];
    recommended_actions: Array<{
      action_key: string;
      label: string;
      reason: string;
      deep_link?: string;
    }>;
    explain_links?: string[];
  };
  verification_result?: VerificationResultSummary;
  completion_verdict?: CompletionVerdictSummary;
  runtime_escalation?: RuntimeEscalationSummary;
  fault_summary?: FaultSummary;
  repair_plan_draft?: RepairPlanDraftSummary;
};

export type PlanView = {
  draft: {
    id: string;
    version: number;
    status: string;
    goal_summary: string;
  };
  review: {
    id: string;
    review_version: number;
    decision: string;
    blocking_issue_count: number;
    advisory_issue_count: number;
  };
  compiled: {
    id: string;
    compiled_version: number;
    status: string;
    risk_summary?: string;
  };
  repair_draft: {
    id: string;
    status: string;
    reasoning_summary: string;
    replaced_constraints?: string[];
    updated_at?: string;
  };
  task_projection: Array<{
    task_id: string;
    task_key: string;
    task_name: string;
    phase: string;
    task_kind: string;
    role_type: string;
    brain_kind: string;
    risk_level: string;
    status: string;
    delivery_summary: string;
    verification_summary: string;
    affected_resources: string[];
    manual_review_required: boolean;
    mapped_domain_task_id?: string;
    mapped_domain_task_status?: string;
  }>;
  diff_summary: {
    total_changes: number;
    split_count: number;
    override_count: number;
    drop_count: number;
    unchanged_count: number;
    review_issue_count: number;
    summary: string;
    items: Array<{
      diff_kind: string;
      before_label: string;
      after_label: string;
      reason: string;
      source_review_issue_id?: string;
    }>;
  };
};

export type CommandResponse = {
  command_id: string;
  accepted: boolean;
  resource_id: string;
  next_action: string;
};

export type CreateProjectResponse = CommandResponse;

export type StartRunResponse = CommandResponse & {
  run_binding: {
    binding_id: string;
    project_id: string;
    task_id: string;
    brain_kind: string;
    run_id: string;
    run_status: string;
    started_at: string;
    finished_at?: string;
    last_sync_at?: string;
  };
};

export type SyncRunResponse = {
  run_binding: StartRunResponse["run_binding"];
};

export type AcceptanceView = {
  overview?: AcceptanceOverview;
  acceptance_run: {
    id: string;
    task_id?: string;
    profile_version?: string;
    status: string;
    functional_status: string;
    production_status: string;
    manual_release_required: boolean;
    finished_at?: string;
    latest_judgement_kind?: string;
    latest_judgement_result?: string;
    latest_judgement_summary?: string;
    latest_judgement_at?: string;
    release_gate_status?: string;
    next_action?: string;
    blocking_issue_count?: number;
  };
  coverage_matrix: Array<{
    key: string;
    kind: string;
    name: string;
    coverage_status: string;
    evidence_count: number;
  }>;
  issues: Array<{
    id: string;
    acceptance_run_id?: string;
    severity: string;
    issue_kind?: string;
    blocking: boolean;
    summary: string;
    detail_json?: string;
    created_at?: string;
  }>;
  evidence_cards: Array<{
    id: string;
    surface: string;
    journey?: string;
    evidence_type: string;
    file_path: string;
    captured_at: string;
  }>;
  release_gate: {
    status: string;
    next_action: string;
    summary: string;
  };
  verification_result?: VerificationResultSummary;
  completion_verdict?: CompletionVerdictSummary;
  runtime_escalation?: RuntimeEscalationSummary;
  fault_summary?: FaultSummary;
  repair_plan_draft?: RepairPlanDraftSummary;
};

export type RepairDraftView = {
  repair_draft: {
    id: string;
    status: string;
    reasoning_summary: string;
    replaced_constraints?: string[];
    failed_task_context_json?: string;
    failure_reason_json?: string;
    original_contracts_json?: string;
    runtime_summary_json?: string;
    repair_plan_json?: string;
    created_by?: string;
    created_at?: string;
    updated_at?: string;
  };
};

export type WorkspaceHomeView = {
  overview?: {
    total_projects: number;
    active_projects: number;
    blocked_projects: number;
    pending_actions: number;
    attention_item_count: number;
    release_watch_count: number;
    production_ready_count: number;
  };
  summary: {
    total_projects: number;
    active_projects: number;
    blocked_projects: number;
    pending_actions: number;
  };
  active_projects: Array<{
    project_id: string;
    name: string;
    project_category: string;
    current_stage: string;
    stage_status: string;
    progress_percent: number;
    production_status: string;
  }>;
  need_attention: Array<{
    item_id: string;
    project_id: string;
    title: string;
    severity: string;
    is_blocking: boolean;
    recommended_action: string;
  }>;
  recent_activity: Array<{
    event_id: string;
    project_id: string;
    event_type: string;
    title: string;
    source_brain: string;
    occurred_at: string;
    needs_attention: boolean;
  }>;
  release_readiness: Array<{
    project_id: string;
    name: string;
    production_status: string;
    missing_items: number;
  }>;
};

export type AuditLogsView = {
  items: Array<{
    id: string;
    project_id: string;
    event_type: string;
    actor_kind: string;
    summary: string;
    payload_json?: string;
    created_at: string;
  }>;
  refresh_hint: string;
};

export type CreateProjectPayload = {
  name: string;
  project_category: string;
  goal_summary: string;
  workspace_root: string;
  repo_root?: string;
};

export type RuntimeHealthView = {
  status: string;
  base_url: string;
};

export type ProjectDiagnosticsView = {
  items: Array<{
    id: string;
    scope: string;
    severity: string;
    error_code: string;
    summary: string;
    detail_json?: string;
    project_id?: string;
    task_id?: string;
    run_id?: string;
    binding_id?: string;
    created_at: string;
  }>;
  refresh_hint: string;
};

export type ExecutionView = {
  runtime_health: RuntimeHealthView;
  runtime_error?: string;
  recent_bindings: Array<{
    binding_id: string;
    project_id: string;
    task_id: string;
    brain_kind: string;
    run_id: string;
    run_status: string;
    started_at: string;
    finished_at?: string;
    last_sync_at?: string;
  }>;
  latest_binding?: {
    run_binding: {
      binding_id: string;
      project_id: string;
      task_id: string;
      brain_kind: string;
      run_id: string;
      run_status: string;
      started_at: string;
      finished_at?: string;
      last_sync_at?: string;
    };
    runtime_state?: {
      run_id: string;
      execution_id?: string;
      status: string;
      brain?: string;
      prompt?: string;
      created_at?: string;
    };
    runtime_stale: boolean;
    runtime_error?: string;
    recent_events: Array<{
      event_id: string;
      sequence_no: number;
      event_type: string;
      event_level?: string;
      summary: string;
      payload?: string;
      created_at: string;
    }>;
    refresh_hint: string;
  };
  replay_error?: string;
  replay_summary?: {
    run_id: string;
    project_id: string;
    brain_kind?: string;
    status: string;
    started_at?: string;
    ended_at?: string;
    event_count: number;
    replay_count: number;
    log_segment_count: number;
    artifact_status_summary: {
      available: number;
      missing: number;
      pruned: number;
    };
    entry_points: Array<{
      replay_id: string;
      domain_task_id?: string;
      compiled_task_id?: string;
      replay_type: string;
      summary?: string;
      file_path?: string;
      created_at: string;
    }>;
  };
  replay_timeline: Array<{
    replay_id: string;
    domain_task_id?: string;
    compiled_task_id?: string;
    seq_no: number;
    replay_type: string;
    title: string;
    summary?: string;
    status: string;
    preview_available: boolean;
    raw_target?: string;
    created_at: string;
  }>;
  log_segments: Array<{
    segment_id: string;
    stream_kind: string;
    seq_no: number;
    started_at?: string;
    ended_at?: string;
    status: string;
    size: number;
    raw_target?: string;
  }>;
};

export type ReplayDetailView = {
  replay_id: string;
  replay_kind: string;
  title: string;
  summary?: string;
  domain_task_id?: string;
  compiled_task_id?: string;
  source_object_kind?: string;
  source_object_id?: string;
  event_id?: string;
  trace_id?: string;
  span_id?: string;
  status: string;
  raw_preview: {
    mime_type?: string;
    preview_available: boolean;
    raw_target?: string;
  };
  related_log_segments: Array<{
    segment_id: string;
    stream_kind: string;
    seq_no: number;
    started_at?: string;
    ended_at?: string;
    status: string;
    size: number;
    raw_target?: string;
  }>;
};

export type ReplayRawView = {
  replay_id: string;
  status: string;
  mime_type?: string;
  encoding: string;
  content?: string;
  truncated: boolean;
};

export type LogSegmentRawView = {
  segment_id: string;
  stream_kind: string;
  status: string;
  content?: string;
  truncated: boolean;
};

export type SystemHealthView = {
  status: string;
  service: string;
  version: string;
  timestamp: string;
};
