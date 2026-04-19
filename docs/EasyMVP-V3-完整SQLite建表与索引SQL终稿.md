# EasyMVP V3 完整 SQLite 建表与索引 SQL 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-首批Migration清单与建表SQL设计](./EasyMVP-V3-首批Migration清单与建表SQL设计.md)
> 关联文档：[EasyMVP-V3-数据库Schema总设计](./EasyMVP-V3-数据库Schema总设计.md)
> 关联文档：[EasyMVP-V3-数据库索引与查询优化设计](./EasyMVP-V3-数据库索引与查询优化设计.md)
> 目标：给出 V3 首批本地 SQLite 可直接落地的完整建表与索引 SQL 终稿，作为 migration 实现正文的上游依据。

## 1. 设计结论

这份文档不再描述“应该有哪些表”，而是直接给出首批可用的 SQL 终稿基线。

原则：

1. 优先保证主链路完整
2. 优先保证本地单机可维护
3. 允许后续加列，不先做过度扩展

## 2. PRAGMA 基线

```sql
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
PRAGMA synchronous = NORMAL;
```

## 3. 系统表

### 3.1 `schema_migrations`

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
```

### 3.2 `settings`

```sql
CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value_json TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
```

### 3.3 `diagnostic_records`

```sql
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

## 4. 项目表

### 4.1 `projects`

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
```

### 4.2 `project_profiles`

```sql
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
```

### 4.3 `project_workspaces`

```sql
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

## 5. 计划表

### 5.1 `workflow_plan_drafts`

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
```

### 5.2 `workflow_plan_review_results`

```sql
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
```

### 5.3 `workflow_compiled_plans`

```sql
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
```

### 5.4 `workflow_compiled_tasks`

字段语义说明：

1. `brain_kind` 为 canonical runtime identifier
2. 它表示归一化后的执行目标归属，不表示领域脑扩张，也不要求等于原始工具名

```sql
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

## 6. 任务与运行时表

### 6.1 `domain_tasks`

字段语义说明：

1. `brain_kind` 延续编译结果中的归一化运行目标标识
2. 页面和领域层应通过该 canonical 值解释任务归属，而不是反推底层协议细节

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
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(source_compiled_plan_id) REFERENCES workflow_compiled_plans(id),
  FOREIGN KEY(source_compiled_task_id) REFERENCES workflow_compiled_tasks(id)
);
```

### 6.2 `task_dependencies`

```sql
CREATE TABLE IF NOT EXISTS task_dependencies (
  task_id TEXT NOT NULL,
  depends_on_task_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY(task_id, depends_on_task_id),
  FOREIGN KEY(task_id) REFERENCES domain_tasks(id) ON DELETE CASCADE,
  FOREIGN KEY(depends_on_task_id) REFERENCES domain_tasks(id) ON DELETE CASCADE
);
```

### 6.3 `task_manual_gates`

```sql
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

### 6.4 `brain_run_bindings`

字段语义说明：

1. `brain_kind` 存运行绑定的 canonical runtime identifier
2. 不在这里记录原始工具名或“执行脑类型”自由文本

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
```

### 6.5 `run_checkpoints`

```sql
CREATE TABLE IF NOT EXISTS run_checkpoints (
  id TEXT PRIMARY KEY,
  run_binding_id TEXT NOT NULL,
  checkpoint_type TEXT NOT NULL,
  payload_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(run_binding_id) REFERENCES brain_run_bindings(id) ON DELETE CASCADE
);
```

### 6.6 `run_event_index`

```sql
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

## 7. 验收表

### 7.1 `acceptance_runs`

说明：

1. `acceptance_runs` 不直接存底层执行脑原始信息
2. 相关来源归属应通过上游任务与运行绑定的归一化字段追溯

补充说明：

- 这里的 SQL 可以继续保留为当前 SQLite 首版现实结构
- 但按当前钱学森总纲，`production_status` 不应再被实现者误读为“最终完成状态”
- 更准确的方向是保留本表，同时逐步引入 `verification_results / completion_verdicts` 一类结构化对象

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
```

### 7.2 `acceptance_issues`

```sql
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
```

### 7.3 `acceptance_judgements`

```sql
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
```

### 7.4 `acceptance_surface_coverage`

```sql
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
```

### 7.5 `acceptance_journey_coverage`

```sql
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

## 8. 证据、回放、审计、快照表

### 8.1 `evidence_items`

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
```

### 8.2 `evidence_links`

```sql
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
```

### 8.3 `replay_items`

```sql
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
```

说明：

1. `replay_items` 只保留为早期兼容产物登记表。
2. 正式回放检索、时间线、日志分段查询不再直接依赖此表。

### 8.4 `audit_logs`

```sql
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
```

### 8.5 `workspace_snapshots`

```sql
CREATE TABLE IF NOT EXISTS workspace_snapshots (
  key TEXT PRIMARY KEY,
  snapshot_json TEXT NOT NULL,
  generated_at TEXT NOT NULL
);
```

### 8.6 `project_snapshots`

```sql
CREATE TABLE IF NOT EXISTS project_snapshots (
  project_id TEXT PRIMARY KEY,
  snapshot_json TEXT NOT NULL,
  generated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

### 8.7 `workflow_replay_index`

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
```

### 8.8 `workflow_run_log_segments`

```sql
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
```

## 9. 索引终稿

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

## 10. 推荐实现顺序

建议先写这 4 个 migration：

1. 系统与项目表
2. 计划与任务表
3. 运行时与验收表
4. 证据、回放、审计与索引

## 11. 后续细分专题

1. 每个 SQL 文件的 checksum 生成规则
2. 默认 seed 数据
3. 数据修复脚本模板
