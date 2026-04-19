# EasyMVP V3 数据库 Schema 总设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3单机版存储架构设计](./EasyMVP-V3单机版存储架构设计.md)
> 关联文档：[EasyMVP-V3计划数据模型与表结构设计](./EasyMVP-V3计划数据模型与表结构设计.md)
> 关联文档：[EasyMVP-V3-Evidence索引表结构设计](./EasyMVP-V3-Evidence索引表结构设计.md)
> 关联文档：[EasyMVP-V3-Replay索引表结构设计](./EasyMVP-V3-Replay索引表结构设计.md)
> 关联文档：[EasyMVP-V3-SQLite初始化与Migration设计](./EasyMVP-V3-SQLite初始化与Migration设计.md)
> 目标：把 V3 本地 SQLite 的整体 schema 做成一份统一总图，明确表族、主键、外键和核心查询面。

## 1. 设计结论

V3 本地库建议按 9 个表族设计：

1. project
2. plan
3. task
4. runtime
5. acceptance
6. evidence
7. replay_audit
8. view_cache
9. system

原则：

1. 主业务事实入正式表
2. 大文件只入索引
3. 页面快照可缓存但不可替代事实表

## 2. 表族总览

### 2.1 project

建议包含：

1. `projects`
2. `project_profiles`
3. `project_workspaces`

### 2.2 plan

建议包含：

1. `workflow_plan_drafts`
2. `workflow_plan_review_results`
3. `workflow_compiled_plans`
4. `workflow_compiled_tasks`

### 2.3 task

建议包含：

1. `domain_tasks`
2. `task_dependencies`
3. `task_manual_gates`

### 2.4 runtime

建议包含：

1. `brain_run_bindings`
2. `run_checkpoints`
3. `run_event_index`

### 2.5 acceptance

建议包含：

1. `acceptance_runs`
2. `acceptance_issues`
3. `acceptance_judgements`
4. `acceptance_surface_coverage`
5. `acceptance_journey_coverage`

### 2.6 evidence

建议包含：

1. `evidence_items`
2. `evidence_links`

### 2.7 replay_audit

建议包含：

1. `replay_items`
2. `audit_logs`

### 2.8 view_cache

建议包含：

1. `workspace_snapshots`
2. `project_snapshots`

### 2.9 system

建议包含：

1. `settings`
2. `schema_migrations`
3. `diagnostic_records`

## 3. 主键建议

推荐：

1. 所有主业务表使用文本型 `id`
2. `schema_migrations.version` 使用整数主键
3. 关系表可用复合唯一键，不必全部单独增 surrogate key

## 4. 外键主线

```text
projects
  ├─ workflow_plan_drafts
  │   └─ workflow_plan_review_results
  │       └─ workflow_compiled_plans
  │           └─ workflow_compiled_tasks
  │               └─ domain_tasks
  │                   ├─ brain_run_bindings
  │                   ├─ acceptance_runs
  │                   ├─ evidence_links
  │                   └─ replay_items
  └─ audit_logs
```

## 5. 推荐核心表清单

### 5.1 `projects`

核心列建议：

1. `id`
2. `name`
3. `project_category`
4. `goal_summary`
5. `status`
6. `production_status`
7. `workspace_root`
8. `created_at`
9. `updated_at`

### 5.2 `domain_tasks`

核心列建议：

1. `id`
2. `project_id`
3. `source_compiled_plan_id`
4. `source_compiled_task_id`
5. `name`
6. `phase`
7. `status`
8. `brain_kind`
9. `role_type`
10. `risk_level`

字段语义补充：

1. `brain_kind` 存归一化后的能力归属或运行目标标识
2. 该值由 `RoleResolver` / compiler / runtime adapter 共同收敛
3. 不等于原始工具名，也不表示 `easymvp-brain` 扩权

### 5.3 `brain_run_bindings`

核心列建议：

1. `id`
2. `project_id`
3. `task_id`
4. `brain_kind`
5. `brain_run_id`
6. `run_status`
7. `started_at`
8. `finished_at`

字段语义补充：

1. `brain_kind` 仍是 canonical runtime identifier
2. 它用于追踪运行目标归属，不是页面或领域层的执行脑选择器

### 5.4 `acceptance_runs`

核心列建议：

1. `id`
2. `project_id`
3. `task_id`
4. `profile_version`
5. `status`
6. `functional_status`
7. `production_status`
8. `manual_release_required`
9. `created_at`
10. `finished_at`

说明：

1. `acceptance_runs` 自身不承载底层执行脑原始信息
2. 如需追踪执行来源，应通过关联的 `domain_tasks` / `brain_run_bindings` 读取归一化标识

### 5.5 `evidence_items`

核心列建议：

1. `id`
2. `project_id`
3. `run_id`
4. `surface`
5. `journey`
6. `evidence_type`
7. `file_path`
8. `content_hash`
9. `captured_at`

### 5.6 `audit_logs`

核心列建议：

1. `id`
2. `project_id`
3. `event_type`
4. `actor_kind`
5. `summary`
6. `payload_json`
7. `created_at`

## 6. 页面查询面要求

Schema 设计必须覆盖 5 个高频查询面：

1. Workspace 首页项目概览
2. Project Workspace 时间线与待处理
3. Plan 页面版本链路
4. Acceptance 页面覆盖矩阵
5. Replay / Audit 详情抽屉

补充约束：

1. 页面查询面读取的 `brain_kind` / `source_brain` 都应来自归一化字段
2. 不直接从 schema 层把原始运行时协议字段暴露成页面 DTO

## 7. JSON 字段约束

允许使用 JSON，但必须克制。

适合放 JSON 的字段：

1. `risk_summary_json`
2. `issues_json`
3. `delivery_contract_json`
4. `verification_contract_json`
5. `payload_json`

不适合放 JSON 的字段：

1. 高频筛选状态
2. 高频排序时间
3. 高频 join 关系

## 8. 版本化原则

需要版本化的对象：

1. `PlanDraft`
2. `PlanReviewResult`
3. `CompiledPlan`
4. `AcceptanceProfile`
5. `CategoryProfile`

这些对象的版本引用应保留在事实表中，不能只通过文件推断。

## 9. 后续细分专题

1. `projects / domain_tasks / acceptance_runs` 完整建表 SQL
2. 索引清单
3. 删除与归档策略
