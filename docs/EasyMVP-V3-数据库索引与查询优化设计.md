# EasyMVP V3 数据库索引与查询优化设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-数据库Schema总设计](./EasyMVP-V3-数据库Schema总设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 关联文档：[EasyMVP-V3-Workspace首页聚合接口Schema设计](./EasyMVP-V3-Workspace首页聚合接口Schema设计.md)
> 目标：为 V3 的 SQLite 表设计明确首批索引策略，确保工作台、计划页、验收页和回放页的核心查询可稳定支撑。

## 1. 设计结论

索引设计必须围绕页面查询面来做，而不是围绕“表本身看起来完整”来做。

首批索引应优先覆盖：

1. 活跃项目列表
2. 项目阶段与状态筛选
3. 计划版本链查询
4. 任务状态与阶段查询
5. run 进度和时间线查询
6. 验收覆盖查询
7. evidence / replay / audit 明细查询

## 2. 核心原则

1. 高频筛选字段必须单列索引
2. 高频排序字段优先联合索引
3. 不依赖 JSON 全表扫描做主查询
4. 页面视图延迟优先于写入极致最小化

## 3. 建议索引清单

### 3.1 `projects`

建议：

1. `idx_projects_status_updated_at(status, updated_at desc)`
2. `idx_projects_category_status(project_category, status)`
3. `idx_projects_production_status_updated_at(production_status, updated_at desc)`

### 3.2 `workflow_plan_drafts`

建议：

1. `idx_plan_drafts_project_version(project_id, version desc)`
2. `idx_plan_drafts_project_status(project_id, status)`

### 3.3 `workflow_plan_review_results`

建议：

1. `idx_plan_reviews_plan_draft(plan_draft_id, review_version desc)`
2. `idx_plan_reviews_project_decision(project_id, decision)`

### 3.4 `workflow_compiled_plans`

建议：

1. `idx_compiled_plans_project_version(project_id, compiled_version desc)`
2. `idx_compiled_plans_project_status(project_id, status)`

### 3.5 `workflow_compiled_tasks`

建议：

1. `idx_compiled_tasks_plan_phase(compiled_plan_id, phase)`
2. `idx_compiled_tasks_plan_status(compiled_plan_id, status)`
3. `idx_compiled_tasks_plan_risk(compiled_plan_id, risk_level)`

### 3.6 `domain_tasks`

建议：

1. `idx_domain_tasks_project_status(project_id, status)`
2. `idx_domain_tasks_project_phase_status(project_id, phase, status)`
3. `idx_domain_tasks_project_updated(project_id, updated_at desc)`
4. `idx_domain_tasks_source_compiled(source_compiled_task_id)`

### 3.7 `brain_run_bindings`

建议：

1. `idx_run_bindings_task(task_id, started_at desc)`
2. `idx_run_bindings_project_status(project_id, run_status, started_at desc)`
3. `uniq_brain_run_id(brain_run_id)`

### 3.8 `run_event_index`

建议：

1. `idx_run_events_run_seq(run_binding_id, sequence_no)`
2. `idx_run_events_project_created(project_id, created_at desc)`

### 3.9 `acceptance_runs`

建议：

1. `idx_acceptance_runs_project_created(project_id, created_at desc)`
2. `idx_acceptance_runs_project_status(project_id, status, created_at desc)`
3. `idx_acceptance_runs_project_production(project_id, production_status, created_at desc)`

补充说明：

- 这些索引建议继续保留，因为它们服务当前现实表结构
- 但按当前钱学森总纲，不应让 `production_status` 查询路径继续被误读为“最终完成状态查询”
- 后续若补出 `verification_results / completion_verdicts`，索引设计也应同步向那组结构化对象扩展

### 3.10 `acceptance_issues`

建议：

1. `idx_acceptance_issues_run_severity(acceptance_run_id, severity)`
2. `idx_acceptance_issues_project_blocking(project_id, blocking, severity)`

### 3.11 `evidence_items`

建议：

1. `idx_evidence_project_captured(project_id, captured_at desc)`
2. `idx_evidence_run_type(run_id, evidence_type, captured_at desc)`
3. `idx_evidence_surface_journey(project_id, surface, journey)`
4. `uniq_evidence_hash(content_hash)`

### 3.12 `replay_items`

建议：

1. `idx_replay_project_time(project_id, created_at desc)`
2. `idx_replay_run(run_id, created_at desc)`

### 3.13 `audit_logs`

建议：

1. `idx_audit_project_time(project_id, created_at desc)`
2. `idx_audit_event_type(event_type, created_at desc)`
3. `idx_audit_actor_type(actor_kind, created_at desc)`

## 4. 页面对应关系

### 4.1 Workspace 首页

依赖：

1. `projects`
2. `domain_tasks`
3. `acceptance_runs`
4. `audit_logs`

因此索引重点是：

1. `status`
2. `production_status`
3. `updated_at`

补充说明：

- 这里的 `production_status` 仍可作为当前现实查询字段存在
- 但按当前钱学森总纲，首页壳层不应把它误读成最终完成状态
- 后续如果补出 `completion_verdicts` 一类结构化对象，首页查询重点也应逐步转向 `decision / completed / manual checkpoint / escalation`

### 4.2 Project Workspace

依赖：

1. `domain_tasks`
2. `brain_run_bindings`
3. `acceptance_runs`
4. `audit_logs`

因此索引重点是：

1. `project_id`
2. `phase`
3. `status`
4. `created_at`

补充说明：

- 单项目工作台当前可以继续读取 `acceptance_runs`
- 但按当前总纲，它的真正主解释链应逐步补到 `VerificationResult / CompletionVerdict / RuntimeEscalation`

### 4.3 Plan 页面

依赖：

1. `workflow_plan_drafts`
2. `workflow_plan_review_results`
3. `workflow_compiled_plans`
4. `workflow_compiled_tasks`

因此索引重点是：

1. `project_id`
2. `version`
3. `compiled_version`

### 4.4 Acceptance 页面

依赖：

1. `acceptance_runs`
2. `acceptance_issues`
3. `evidence_items`

因此索引重点是：

1. `project_id`
2. `status`
3. `severity`
4. `surface`
5. `journey`

补充说明：

- 这是旧 acceptance 首版查询现实
- 后续如果补出 `verification_results` 或与合同要求直接对齐的结构化对象，索引设计也应同步扩展

## 5. 不建议的索引方式

不建议：

1. 所有字段都建索引
2. 在低频表上过度建联合索引
3. 把 JSON 提取查询当主路径
4. 先拍脑袋加索引，不对照页面查询

## 6. 后续细分专题

1. 首批 migration 的索引 SQL
2. 热点查询 explain 基线
3. 快照缓存与事实查询分工
