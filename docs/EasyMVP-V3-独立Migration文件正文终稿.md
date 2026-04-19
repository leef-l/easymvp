# EasyMVP V3 独立 Migration 文件正文终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-完整SQLite建表与索引SQL终稿](./EasyMVP-V3-完整SQLite建表与索引SQL终稿.md)
> 关联文档：[EasyMVP-V3-首批Migration清单与建表SQL设计](./EasyMVP-V3-首批Migration清单与建表SQL设计.md)
> 目标：把首批 migration 从“总 SQL 终稿”继续拆成可直接保存为独立文件的正文版本。

## 1. 设计结论

本专题直接给出每个 migration 文件应包含的 SQL 正文块。

推荐目录：

```text
apps/core/manifest/migrations/
  0001_init_system_tables.sql
  0002_init_project_tables.sql
  0003_init_plan_tables.sql
  0004_init_task_tables.sql
  0005_init_runtime_tables.sql
  0006_init_acceptance_tables.sql
  0007_init_evidence_replay_audit_tables.sql
  0008_add_core_indexes.sql
```

## 2. `0001_init_system_tables.sql`

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  checksum TEXT NOT NULL,
  applied_at TEXT NOT NULL,
  duration_ms INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  error_message TEXT
);

CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value_json TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS diagnostic_records (
  id TEXT PRIMARY KEY,
  scope TEXT NOT NULL,
  severity TEXT NOT NULL,
  error_code TEXT,
  summary TEXT NOT NULL,
  detail_json TEXT,
  created_at TEXT NOT NULL
);
```

## 3. `0002_init_project_tables.sql`

```sql
CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  project_category TEXT NOT NULL,
  goal_summary TEXT NOT NULL,
  status TEXT NOT NULL,
  production_status TEXT NOT NULL,
  workspace_root TEXT NOT NULL,
  repo_root TEXT,
  current_plan_draft_id TEXT,
  current_compiled_plan_id TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS project_profiles (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  category_profile_version TEXT NOT NULL,
  acceptance_profile_version TEXT NOT NULL,
  role_profile_version TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS project_workspaces (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  workspace_root TEXT NOT NULL,
  evidence_root TEXT NOT NULL,
  runs_root TEXT NOT NULL,
  replay_root TEXT NOT NULL,
  diagnostics_root TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

## 4. `0003_init_plan_tables.sql`

```sql
CREATE TABLE IF NOT EXISTS workflow_plan_drafts (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  version INTEGER NOT NULL,
  source_kind TEXT NOT NULL,
  source_run_id TEXT,
  project_category TEXT NOT NULL,
  goal_summary TEXT NOT NULL,
  input_requirements_json TEXT NOT NULL,
  draft_tasks_json TEXT NOT NULL,
  status TEXT NOT NULL,
  created_by TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  UNIQUE(project_id, version)
);

CREATE TABLE IF NOT EXISTS workflow_plan_review_results (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  plan_draft_id TEXT NOT NULL,
  review_version INTEGER NOT NULL,
  review_run_id TEXT,
  decision TEXT NOT NULL,
  blocking_issue_count INTEGER NOT NULL DEFAULT 0,
  advisory_issue_count INTEGER NOT NULL DEFAULT 0,
  issues_json TEXT NOT NULL,
  split_suggestions_json TEXT,
  override_suggestions_json TEXT,
  status TEXT NOT NULL,
  reviewed_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(plan_draft_id) REFERENCES workflow_plan_drafts(id) ON DELETE CASCADE,
  UNIQUE(plan_draft_id, review_version)
);

CREATE TABLE IF NOT EXISTS workflow_compiled_plans (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  plan_draft_id TEXT NOT NULL,
  plan_review_result_id TEXT NOT NULL,
  compiled_version INTEGER NOT NULL,
  compile_run_id TEXT,
  project_category TEXT NOT NULL,
  status TEXT NOT NULL,
  risk_summary_json TEXT,
  compile_diff_json TEXT,
  generated_at TEXT NOT NULL,
  activated_at TEXT,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(plan_draft_id) REFERENCES workflow_plan_drafts(id) ON DELETE CASCADE,
  FOREIGN KEY(plan_review_result_id) REFERENCES workflow_plan_review_results(id) ON DELETE CASCADE,
  UNIQUE(project_id, compiled_version)
);

CREATE TABLE IF NOT EXISTS workflow_compiled_tasks (
  id TEXT PRIMARY KEY,
  compiled_plan_id TEXT NOT NULL,
  task_key TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  phase TEXT NOT NULL,
  task_kind TEXT NOT NULL,
  role_type TEXT NOT NULL,
  brain_kind TEXT NOT NULL,
  risk_level TEXT NOT NULL,
  affected_resources_json TEXT NOT NULL,
  delivery_contract_json TEXT NOT NULL,
  verification_contract_json TEXT NOT NULL,
  manual_review_required INTEGER NOT NULL DEFAULT 0,
  depends_on_task_keys_json TEXT,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(compiled_plan_id) REFERENCES workflow_compiled_plans(id) ON DELETE CASCADE,
  UNIQUE(compiled_plan_id, task_key)
);
```

## 5. `0004_init_task_tables.sql`

```sql
CREATE TABLE IF NOT EXISTS domain_tasks (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  source_compiled_plan_id TEXT NOT NULL,
  source_compiled_task_id TEXT NOT NULL,
  source_task_key TEXT NOT NULL,
  compiled_version INTEGER NOT NULL,
  name TEXT NOT NULL,
  phase TEXT NOT NULL,
  task_kind TEXT NOT NULL,
  role_type TEXT NOT NULL,
  brain_kind TEXT NOT NULL,
  risk_level TEXT NOT NULL,
  status TEXT NOT NULL,
  manual_review_required INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS task_dependencies (
  task_id TEXT NOT NULL,
  depends_on_task_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY(task_id, depends_on_task_id),
  FOREIGN KEY(task_id) REFERENCES domain_tasks(id) ON DELETE CASCADE,
  FOREIGN KEY(depends_on_task_id) REFERENCES domain_tasks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS task_manual_gates (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  task_id TEXT NOT NULL,
  gate_kind TEXT NOT NULL,
  gate_status TEXT NOT NULL,
  comment TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(task_id) REFERENCES domain_tasks(id) ON DELETE CASCADE
);
```

## 6. `0005_init_runtime_tables.sql`

```sql
CREATE TABLE IF NOT EXISTS brain_run_bindings (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  task_id TEXT NOT NULL,
  brain_kind TEXT NOT NULL,
  brain_run_id TEXT NOT NULL,
  run_status TEXT NOT NULL,
  started_at TEXT NOT NULL,
  finished_at TEXT,
  last_sync_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(task_id) REFERENCES domain_tasks(id) ON DELETE CASCADE,
  UNIQUE(brain_run_id)
);

CREATE TABLE IF NOT EXISTS run_checkpoints (
  id TEXT PRIMARY KEY,
  run_binding_id TEXT NOT NULL,
  checkpoint_type TEXT NOT NULL,
  payload_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(run_binding_id) REFERENCES brain_run_bindings(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS run_event_index (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  run_binding_id TEXT NOT NULL,
  sequence_no INTEGER NOT NULL,
  event_type TEXT NOT NULL,
  event_level TEXT,
  summary TEXT NOT NULL,
  payload_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(run_binding_id) REFERENCES brain_run_bindings(id) ON DELETE CASCADE,
  UNIQUE(run_binding_id, sequence_no)
);
```

## 7. `0006_init_acceptance_tables.sql`

```sql
CREATE TABLE IF NOT EXISTS acceptance_runs (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  task_id TEXT,
  profile_version TEXT NOT NULL,
  status TEXT NOT NULL,
  functional_status TEXT NOT NULL,
  production_status TEXT NOT NULL,
  manual_release_required INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  finished_at TEXT,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(task_id) REFERENCES domain_tasks(id)
);

CREATE TABLE IF NOT EXISTS acceptance_issues (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  acceptance_run_id TEXT NOT NULL,
  severity TEXT NOT NULL,
  issue_kind TEXT NOT NULL,
  blocking INTEGER NOT NULL DEFAULT 0,
  summary TEXT NOT NULL,
  detail_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(acceptance_run_id) REFERENCES acceptance_runs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS acceptance_judgements (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  acceptance_run_id TEXT NOT NULL,
  judgement_kind TEXT NOT NULL,
  judgement_result TEXT NOT NULL,
  summary TEXT NOT NULL,
  detail_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(acceptance_run_id) REFERENCES acceptance_runs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS acceptance_surface_coverage (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  acceptance_run_id TEXT NOT NULL,
  surface TEXT NOT NULL,
  coverage_status TEXT NOT NULL,
  evidence_count INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(acceptance_run_id) REFERENCES acceptance_runs(id) ON DELETE CASCADE,
  UNIQUE(acceptance_run_id, surface)
);

CREATE TABLE IF NOT EXISTS acceptance_journey_coverage (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  acceptance_run_id TEXT NOT NULL,
  journey TEXT NOT NULL,
  coverage_status TEXT NOT NULL,
  evidence_count INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(acceptance_run_id) REFERENCES acceptance_runs(id) ON DELETE CASCADE,
  UNIQUE(acceptance_run_id, journey)
);
```

## 8. `0007_init_evidence_replay_audit_tables.sql`

```sql
CREATE TABLE IF NOT EXISTS evidence_items (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  run_id TEXT,
  surface TEXT NOT NULL,
  journey TEXT,
  evidence_type TEXT NOT NULL,
  file_path TEXT NOT NULL,
  content_hash TEXT NOT NULL,
  file_size INTEGER,
  captured_at TEXT NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS evidence_links (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  evidence_item_id TEXT NOT NULL,
  linked_object_type TEXT NOT NULL,
  linked_object_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(evidence_item_id) REFERENCES evidence_items(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS replay_items (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  run_id TEXT,
  replay_type TEXT NOT NULL,
  file_path TEXT NOT NULL,
  summary TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id TEXT PRIMARY KEY,
  project_id TEXT,
  event_type TEXT NOT NULL,
  actor_kind TEXT NOT NULL,
  summary TEXT NOT NULL,
  payload_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS workspace_snapshots (
  key TEXT PRIMARY KEY,
  snapshot_json TEXT NOT NULL,
  generated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS project_snapshots (
  project_id TEXT PRIMARY KEY,
  snapshot_json TEXT NOT NULL,
  generated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

## 9. `0008_add_core_indexes.sql`

```sql
CREATE INDEX IF NOT EXISTS idx_projects_status_updated_at
  ON projects(status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_projects_category_status
  ON projects(project_category, status);
CREATE INDEX IF NOT EXISTS idx_projects_production_status_updated_at
  ON projects(production_status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_plan_drafts_project_version
  ON workflow_plan_drafts(project_id, version DESC);
CREATE INDEX IF NOT EXISTS idx_plan_reviews_plan_draft
  ON workflow_plan_review_results(plan_draft_id, review_version DESC);
CREATE INDEX IF NOT EXISTS idx_compiled_plans_project_version
  ON workflow_compiled_plans(project_id, compiled_version DESC);
CREATE INDEX IF NOT EXISTS idx_compiled_tasks_plan_phase
  ON workflow_compiled_tasks(compiled_plan_id, phase);
CREATE INDEX IF NOT EXISTS idx_domain_tasks_project_phase_status
  ON domain_tasks(project_id, phase, status);
CREATE INDEX IF NOT EXISTS idx_domain_tasks_project_updated
  ON domain_tasks(project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_run_bindings_project_status
  ON brain_run_bindings(project_id, run_status, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_run_events_run_seq
  ON run_event_index(run_binding_id, sequence_no);
CREATE INDEX IF NOT EXISTS idx_acceptance_runs_project_status
  ON acceptance_runs(project_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_acceptance_issues_run_severity
  ON acceptance_issues(acceptance_run_id, severity);
CREATE INDEX IF NOT EXISTS idx_evidence_project_captured
  ON evidence_items(project_id, captured_at DESC);
CREATE INDEX IF NOT EXISTS idx_evidence_surface_journey
  ON evidence_items(project_id, surface, journey);
CREATE INDEX IF NOT EXISTS idx_replay_project_time
  ON replay_items(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_project_time
  ON audit_logs(project_id, created_at DESC);
```

## 10. `0009_add_formal_replay_index_tables.sql`

```sql
CREATE TABLE IF NOT EXISTS workflow_replay_index (
  id TEXT PRIMARY KEY,
  replay_id TEXT NOT NULL,
  project_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  domain_task_id TEXT,
  compiled_task_id TEXT,
  event_id TEXT,
  trace_id TEXT,
  span_id TEXT,
  replay_kind TEXT NOT NULL,
  seq_no INTEGER NOT NULL,
  title TEXT NOT NULL,
  summary TEXT,
  file_path TEXT NOT NULL,
  file_ext TEXT,
  mime_type TEXT,
  file_size INTEGER,
  sha256 TEXT,
  source_object_kind TEXT,
  source_object_id TEXT,
  status TEXT NOT NULL DEFAULT 'available',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  UNIQUE(replay_id)
);

CREATE TABLE IF NOT EXISTS workflow_run_log_segments (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  segment_id TEXT NOT NULL,
  stream_kind TEXT NOT NULL,
  seq_no INTEGER NOT NULL,
  file_path TEXT NOT NULL,
  file_size INTEGER,
  sha256 TEXT,
  started_at TEXT,
  ended_at TEXT,
  status TEXT NOT NULL DEFAULT 'available',
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  UNIQUE(run_id, stream_kind, seq_no)
);

CREATE INDEX IF NOT EXISTS idx_workflow_replay_index_run_seq
  ON workflow_replay_index(run_id, seq_no);

CREATE INDEX IF NOT EXISTS idx_workflow_replay_index_project_created
  ON workflow_replay_index(project_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_workflow_replay_index_event_id
  ON workflow_replay_index(event_id);

CREATE INDEX IF NOT EXISTS idx_workflow_replay_index_trace_span
  ON workflow_replay_index(trace_id, span_id);

CREATE INDEX IF NOT EXISTS idx_workflow_replay_index_run_kind_seq
  ON workflow_replay_index(run_id, replay_kind, seq_no);

CREATE INDEX IF NOT EXISTS idx_workflow_run_log_segments_run_stream_seq
  ON workflow_run_log_segments(run_id, stream_kind, seq_no);

CREATE INDEX IF NOT EXISTS idx_workflow_run_log_segments_project_run_created
  ON workflow_run_log_segments(project_id, run_id, created_at DESC);
```

说明：

1. `replay_items` 保留为早期轻量回放产物登记表，不再承担 V3 正式 replay 查询主索引职责。
2. V3 正式回放查询统一基于 `workflow_replay_index + workflow_run_log_segments`。
3. `replay_id` 是回放条目的稳定外部标识，`id` 是数据库主键。

## 11. 实施顺序

1. 写入 migration 文件
2. 计算 checksum
3. 本地空库验证
4. 执行外键检查
