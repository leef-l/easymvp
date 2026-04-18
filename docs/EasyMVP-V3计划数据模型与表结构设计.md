# EasyMVP V3 计划数据模型与表结构设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3方案编译模型设计](./EasyMVP-V3方案编译模型设计.md)
> 关联文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：把 `PlanDraft -> PlanReviewResult -> CompiledPlan -> DomainTask` 这条主链路落成稳定的数据模型与表结构基线。

## 1. 设计结论

V3 的计划链路必须先稳定四类对象：

1. `PlanDraft`
2. `PlanReviewResult`
3. `CompiledPlan`
4. `CompiledTask`

其中：

- `PlanDraft` 表示原始方案输入
- `PlanReviewResult` 表示结构化审核结论
- `CompiledPlan` 表示正式执行计划
- `CompiledTask` 表示 `CompiledPlan` 中的任务投影结果

`DomainTask` 可以由 `CompiledTask` 投影生成，但不应反向替代 `CompiledTask`。

## 2. 要解决的问题

本专题主要解决：

1. 计划对象字段不稳定，导致前端和执行层接口混乱
2. review 结果无法结构化沉淀
3. compile 阶段前后差异无法追踪
4. rework 无法复用初始计划链路
5. 任务来源无法可靠回溯到 plan 版本

## 3. 核心对象关系

```text
Project
  └─ PlanDraft(version)
       └─ PlanReviewResult(review_version)
            └─ CompiledPlan(compiled_version)
                 └─ CompiledTask(task_key)
                      └─ DomainTask(task_id)
```

要求：

1. 一个项目允许多个 `PlanDraft` 版本
2. 一个 `PlanDraft` 允许多个 review 结果，但只能有一个当前生效 review
3. 一个 `PlanReviewResult` 只能产出一个当前生效 `CompiledPlan`
4. 一个 `CompiledPlan` 可投影多个 `CompiledTask`
5. 每个 `DomainTask` 必须能追溯到来源 `CompiledTask`

## 4. PlanDraft

### 4.1 语义

`PlanDraft` 是 architect 或 `easymvp-brain` 生成的原始计划草案。

它允许不完美，但不允许直接执行。

### 4.2 建议字段

建议至少包含：

1. `id`
2. `project_id`
3. `version`
4. `source_kind`
5. `source_run_id`
6. `project_category`
7. `goal_summary`
8. `input_requirements_json`
9. `draft_tasks_json`
10. `status`
11. `created_by`
12. `created_at`
13. `updated_at`

### 4.3 状态

建议：

1. `draft_created`
2. `review_pending`
3. `reviewed`
4. `superseded`
5. `archived`

## 5. PlanReviewResult

### 5.1 语义

`PlanReviewResult` 是 `easymvp-brain` 对 `PlanDraft` 的结构化审核结果。

### 5.2 建议字段

建议至少包含：

1. `id`
2. `project_id`
3. `plan_draft_id`
4. `review_version`
5. `review_run_id`
6. `decision`
7. `blocking_issue_count`
8. `advisory_issue_count`
9. `issues_json`
10. `split_suggestions_json`
11. `override_suggestions_json`
12. `status`
13. `reviewed_at`

### 5.3 decision

建议：

1. `rejected`
2. `revise_required`
3. `approved_for_compile`

### 5.4 issues_json 结构

建议每项包含：

1. `issue_id`
2. `severity`
3. `issue_kind`
4. `summary`
5. `affected_task_key`
6. `recommended_action`
7. `blocking`

## 6. CompiledPlan

### 6.1 语义

`CompiledPlan` 是经过 review 通过后生成的正式执行计划。

### 6.2 建议字段

建议至少包含：

1. `id`
2. `project_id`
3. `plan_draft_id`
4. `plan_review_result_id`
5. `compiled_version`
6. `compile_run_id`
7. `project_category`
8. `status`
9. `risk_summary_json`
10. `compile_diff_json`
11. `generated_at`
12. `activated_at`

### 6.3 状态

建议：

1. `compiled`
2. `activated`
3. `superseded`
4. `archived`

## 7. CompiledTask

### 7.1 语义

`CompiledTask` 是 `CompiledPlan` 中的正式任务单元。

它比 `DomainTask` 更上游，负责稳定承载编译结果。

### 7.2 建议字段

建议至少包含：

1. `id`
2. `compiled_plan_id`
3. `task_key`
4. `name`
5. `description`
6. `phase`
7. `task_kind`
8. `role_type`
9. `brain_kind`
10. `risk_level`
11. `affected_resources_json`
12. `delivery_contract_json`
13. `verification_contract_json`
14. `manual_review_required`
15. `depends_on_task_keys_json`
16. `status`

### 7.3 状态

建议：

1. `compiled`
2. `projected`
3. `executing`
4. `verified`
5. `failed`
6. `superseded`

## 8. DomainTask 投影规则

`DomainTask` 应由 `CompiledTask` 投影生成。

必须保留以下来源字段：

1. `source_compiled_plan_id`
2. `source_compiled_task_id`
3. `source_task_key`
4. `compiled_version`

这样才能保证：

1. 执行链路可回溯
2. 返工可追根
3. 工作台可解释“为什么是这个任务”

## 9. 建议表结构

### 9.1 `workflow_plan_drafts`

建议核心列：

1. `id`
2. `project_id`
3. `version`
4. `source_kind`
5. `source_run_id`
6. `project_category`
7. `goal_summary`
8. `input_requirements_json`
9. `draft_tasks_json`
10. `status`
11. `created_by`
12. `created_at`
13. `updated_at`

唯一约束建议：

1. `uniq(project_id, version)`

### 9.2 `workflow_plan_review_results`

建议核心列：

1. `id`
2. `project_id`
3. `plan_draft_id`
4. `review_version`
5. `review_run_id`
6. `decision`
7. `blocking_issue_count`
8. `advisory_issue_count`
9. `issues_json`
10. `split_suggestions_json`
11. `override_suggestions_json`
12. `status`
13. `reviewed_at`

唯一约束建议：

1. `uniq(plan_draft_id, review_version)`

### 9.3 `workflow_compiled_plans`

建议核心列：

1. `id`
2. `project_id`
3. `plan_draft_id`
4. `plan_review_result_id`
5. `compiled_version`
6. `compile_run_id`
7. `project_category`
8. `status`
9. `risk_summary_json`
10. `compile_diff_json`
11. `generated_at`
12. `activated_at`

唯一约束建议：

1. `uniq(project_id, compiled_version)`

### 9.4 `workflow_compiled_tasks`

建议核心列：

1. `id`
2. `compiled_plan_id`
3. `task_key`
4. `name`
5. `description`
6. `phase`
7. `task_kind`
8. `role_type`
9. `brain_kind`
10. `risk_level`
11. `affected_resources_json`
12. `delivery_contract_json`
13. `verification_contract_json`
14. `manual_review_required`
15. `depends_on_task_keys_json`
16. `status`

唯一约束建议：

1. `uniq(compiled_plan_id, task_key)`

## 10. 版本化规则

### 10.1 PlanDraft

以下情况必须升新版本：

1. 用户目标变化
2. 项目分类变化
3. 主要任务集合变化
4. rework 进入重新设计

### 10.2 PlanReviewResult

以下情况必须升新 review 版本：

1. review prompt 或规则版本变化
2. `PlanDraft` 内容变化
3. blocking issue 被重新评估

### 10.3 CompiledPlan

以下情况必须升新编译版本：

1. 任务拆分变化
2. role / brain 变化
3. delivery / verification 合同变化
4. 风险等级变化

## 11. 与工作台的关系

这些对象最终会进入：

1. `PlanView`
2. `WorkspaceView`
3. `LiveEvent`

因此要求：

1. 每个对象都要有稳定 `id`
2. 每个对象都要有稳定 `version`
3. 每个对象都要可回溯到来源 run

## 12. 后续细分专题

本专题后续继续拆：

1. `draft_tasks_json` schema
2. `issues_json` schema
3. `compile_diff_json` schema
4. `CompiledTask -> DomainTask` 投影策略
