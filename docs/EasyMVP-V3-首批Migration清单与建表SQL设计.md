# EasyMVP V3 首批 Migration 清单与建表 SQL 设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-SQLite初始化与Migration设计](./EasyMVP-V3-SQLite初始化与Migration设计.md)
> 关联文档：[EasyMVP-V3-数据库Schema总设计](./EasyMVP-V3-数据库Schema总设计.md)
> 关联文档：[EasyMVP-V3-数据库索引与查询优化设计](./EasyMVP-V3-数据库索引与查询优化设计.md)
> 目标：把 V3 首批数据库落地所需的 migration 顺序、建表范围和 SQL 结构基线一次明确。

## 1. 设计结论

V3 首批 migration 建议控制在 8 个文件内完成核心落地。

推荐顺序：

1. `0001_init_system_tables.sql`
2. `0002_init_project_tables.sql`
3. `0003_init_plan_tables.sql`
4. `0004_init_task_tables.sql`
5. `0005_init_runtime_tables.sql`
6. `0006_init_acceptance_tables.sql`
7. `0007_init_evidence_replay_audit_tables.sql`
8. `0008_add_core_indexes.sql`

这样做的原则是：

1. 先表族
2. 后索引
3. 核心事实先落
4. 视图缓存最后补

## 2. 首批 migration 范围

### 2.1 `0001_init_system_tables.sql`

建议创建：

1. `schema_migrations`
2. `settings`
3. `diagnostic_records`

### 2.2 `0002_init_project_tables.sql`

建议创建：

1. `projects`
2. `project_profiles`
3. `project_workspaces`

### 2.3 `0003_init_plan_tables.sql`

建议创建：

1. `workflow_plan_drafts`
2. `workflow_plan_review_results`
3. `workflow_compiled_plans`
4. `workflow_compiled_tasks`

### 2.4 `0004_init_task_tables.sql`

建议创建：

1. `domain_tasks`
2. `task_dependencies`
3. `task_manual_gates`

### 2.5 `0005_init_runtime_tables.sql`

建议创建：

1. `brain_run_bindings`
2. `run_checkpoints`
3. `run_event_index`

### 2.6 `0006_init_acceptance_tables.sql`

建议创建：

1. `acceptance_runs`
2. `acceptance_issues`
3. `acceptance_judgements`
4. `acceptance_surface_coverage`
5. `acceptance_journey_coverage`

### 2.7 `0007_init_evidence_replay_audit_tables.sql`

建议创建：

1. `evidence_items`
2. `evidence_links`
3. `replay_items`
4. `audit_logs`
5. `workspace_snapshots`
6. `project_snapshots`

### 2.8 `0008_add_core_indexes.sql`

建议统一创建高频索引。

## 3. 建表 SQL 风格基线

建议统一规则：

1. 表名使用 `snake_case`
2. 时间字段统一 `created_at / updated_at / finished_at`
3. 主键统一 `TEXT`
4. 布尔字段使用 `INTEGER`
5. JSON 字段统一以 `_json` 结尾

字段命名补充：

1. `brain_kind` 统一表示归一化后的运行目标标识
2. 不在 migration/DDL 注释里把它写成原始工具名或任意“执行脑类型”

## 4. 关键表 SQL 基线

### 4.1 `projects`

```sql
CREATE TABLE projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  project_category TEXT NOT NULL,
  goal_summary TEXT NOT NULL,
  status TEXT NOT NULL,
  production_status TEXT NOT NULL,
  workspace_root TEXT NOT NULL,
  current_plan_draft_id TEXT,
  current_compiled_plan_id TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
```

### 4.2 `workflow_plan_drafts`

```sql
CREATE TABLE workflow_plan_drafts (
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
  FOREIGN KEY(project_id) REFERENCES projects(id),
  UNIQUE(project_id, version)
);
```

### 4.3 `workflow_plan_review_results`

```sql
CREATE TABLE workflow_plan_review_results (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  plan_draft_id TEXT NOT NULL,
  review_version INTEGER NOT NULL,
  review_run_id TEXT,
  decision TEXT NOT NULL,
  blocking_issue_count INTEGER NOT NULL,
  advisory_issue_count INTEGER NOT NULL,
  issues_json TEXT NOT NULL,
  split_suggestions_json TEXT,
  override_suggestions_json TEXT,
  status TEXT NOT NULL,
  reviewed_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id),
  FOREIGN KEY(plan_draft_id) REFERENCES workflow_plan_drafts(id),
  UNIQUE(plan_draft_id, review_version)
);
```

### 4.4 `workflow_compiled_plans`

```sql
CREATE TABLE workflow_compiled_plans (
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
  FOREIGN KEY(project_id) REFERENCES projects(id),
  FOREIGN KEY(plan_draft_id) REFERENCES workflow_plan_drafts(id),
  FOREIGN KEY(plan_review_result_id) REFERENCES workflow_plan_review_results(id),
  UNIQUE(project_id, compiled_version)
);
```

### 4.5 `domain_tasks`

字段语义说明：

1. `brain_kind` 为 canonical runtime identifier
2. 由 `RoleResolver` / compiler / runtime adapter 共同收敛，不是自由输入名

```sql
CREATE TABLE domain_tasks (
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
  FOREIGN KEY(project_id) REFERENCES projects(id),
  FOREIGN KEY(source_compiled_plan_id) REFERENCES workflow_compiled_plans(id),
  FOREIGN KEY(source_compiled_task_id) REFERENCES workflow_compiled_tasks(id)
);
```

### 4.6 `brain_run_bindings`

字段语义说明：

1. `brain_kind` 为运行绑定层使用的 canonical runtime identifier
2. 不表示领域脑职责，也不要求与底层原始工具名一一对应

```sql
CREATE TABLE brain_run_bindings (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  task_id TEXT NOT NULL,
  brain_kind TEXT NOT NULL,
  brain_run_id TEXT NOT NULL,
  run_status TEXT NOT NULL,
  started_at TEXT NOT NULL,
  finished_at TEXT,
  last_sync_at TEXT,
  FOREIGN KEY(project_id) REFERENCES projects(id),
  FOREIGN KEY(task_id) REFERENCES domain_tasks(id),
  UNIQUE(brain_run_id)
);
```

### 4.7 `acceptance_runs`

说明：

1. `acceptance_runs` 本身不保存底层执行脑原始信息
2. 相关运行目标归属应通过 `task_id -> domain_tasks / brain_run_bindings` 追溯

```sql
CREATE TABLE acceptance_runs (
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
  FOREIGN KEY(project_id) REFERENCES projects(id),
  FOREIGN KEY(task_id) REFERENCES domain_tasks(id)
);
```

### 4.8 `evidence_items`

```sql
CREATE TABLE evidence_items (
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
  FOREIGN KEY(project_id) REFERENCES projects(id)
);
```

### 4.9 `audit_logs`

```sql
CREATE TABLE audit_logs (
  id TEXT PRIMARY KEY,
  project_id TEXT,
  event_type TEXT NOT NULL,
  actor_kind TEXT NOT NULL,
  summary TEXT NOT NULL,
  payload_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id)
);
```

## 5. 首批索引 SQL 范围

建议 `0008_add_core_indexes.sql` 至少创建：

1. `projects` 状态与更新时间索引
2. `workflow_plan_drafts` 项目版本索引
3. `workflow_compiled_plans` 项目版本索引
4. `domain_tasks` 项目阶段状态索引
5. `brain_run_bindings` 任务 / 项目状态索引
6. `acceptance_runs` 项目状态索引
7. `evidence_items` 项目时间与 surface/journey 索引
8. `audit_logs` 项目时间索引

## 6. 首批不进入 migration 的对象

首批可以暂缓：

1. 大型快照缓存优化表
2. 低频设置扩展表
3. 复杂全文检索表
4. 统计预聚合表

## 7. 后续细分专题

1. 每个 migration 的完整 SQL 文件正文
2. seed 数据策略
3. rollback 与 repair 脚本策略
