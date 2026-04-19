# EasyMVP V3 easymvp-brain Manifest 与 Tool Schema 设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-easymvp-brain职责边界与输入输出合同设计](./EasyMVP-V3-easymvp-brain职责边界与输入输出合同设计.md)
> 关联文档：[EasyMVP-V3专精大脑接入计划](./EasyMVP-V3专精大脑接入计划.md)
> 目标：给出 `easymvp-brain` 第一版 Manifest、能力列表和 Tool Schema 设计基线。

## 1. Manifest 设计结论

第一版 `easymvp-brain` 应尽量小而专。

它不追求全能，只覆盖 5 类领域任务。

## 2. Manifest 建议

```json
{
  "schema_version": 1,
  "kind": "easymvp",
  "name": "EasyMVP Brain",
  "brain_version": "0.1.0",
  "description": "Workflow domain brain for review, compile, repair, acceptance mapping and completion adjudication",
  "capabilities": [
    "workflow.plan.review",
    "workflow.plan.compile",
    "workflow.repair.design",
    "workflow.acceptance.map",
    "workflow.task.adjudicate"
  ],
  "runtime": {
    "type": "native"
  }
}
```

## 3. Tool Schema 设计原则

### 3.1 只暴露领域工具

第一版建议工具只包括：

1. `plan_review_tool`
2. `plan_compile_tool`
3. `repair_design_tool`
4. `acceptance_mapping_tool`
5. `completion_adjudication_tool`

### 3.2 只收结构化输入

每个工具输入必须直接吃结构化对象，而不是散乱文本。

### 3.3 输出必须稳定

每个工具输出都应带：

1. `schema_version`
2. `source_input_id`
3. `result_json`

## 4. 工具定义

### 4.1 plan_review_tool

输入：

1. `plan_draft`
2. `category_profile`

输出：

1. `plan_review_result`

### 4.2 plan_compile_tool

输入：

1. `plan_draft`
2. `plan_review_result`
3. `category_profile`

输出：

1. `compiled_plan`

### 4.3 repair_design_tool

输入：

1. `failed_task_context`
2. `failure_reason`
3. `original_contracts`

输出：

1. `repair_plan_draft`

### 4.4 acceptance_mapping_tool

输入：

1. `project_category`
2. `category_profile`
3. `artifact_summary`

输出：

1. `production_acceptance_profile`

### 4.5 completion_adjudication_tool

输入：

1. `executor_result`
2. `delivery_result`
3. `verification_result`

输出：

1. `completion_adjudication_result`

## 5. 后续细分专题

本专题后续继续拆：

1. 每个 tool 的 JSON schema
2. capability 到 tool 的映射矩阵
3. audit 字段规范
