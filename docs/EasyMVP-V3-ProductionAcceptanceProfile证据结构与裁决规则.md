# EasyMVP V3 ProductionAcceptanceProfile 证据结构与裁决规则

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3生产级分类验收体系设计.md)
> 关联文档：[EasyMVP-V3分类Profile结构与版本化规范](./EasyMVP-V3分类Profile结构与版本化规范.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：把 `ProductionAcceptanceProfile`、`Evidence`、最终裁决状态落成结构化规则，作为 V3 验收主线的正式基线。

## 1. 设计结论

V3 的最终验收目标必须是结构化裁决，而不是自然语言总结。

正式状态建议固定为：

1. `functional_passed`
2. `production_passed`
3. `manual_release_required`
4. `released_by_human`

其中只有：

1. `production_passed = true`
2. 且必要时 `released_by_human = true`

才算真正完成。

## 2. ProductionAcceptanceProfile 顶层结构

建议至少包含：

1. `profile_id`
2. `project_category`
3. `profile_version`
4. `required_surfaces`
5. `required_journeys`
6. `required_evidence`
7. `production_gates`
8. `manual_release_rules`
9. `status`

## 3. required_surfaces

建议按分类定义，例如：

1. `user_frontend`
2. `admin_backend`
3. `api_backend`
4. `game_runtime`
5. `editor_runtime`

## 4. required_journeys

建议每项包含：

1. `journey_id`
2. `surface`
3. `name`
4. `blocking`
5. `required_evidence_types`

## 5. required_evidence

建议每项包含：

1. `evidence_type`
2. `source_brain`
3. `required`
4. `blocking`
5. `retention_policy`

### 5.1 evidence_type 示例

1. `browser_screenshot`
2. `browser_trace`
3. `build_artifact`
4. `ci_result`
5. `runtime_log`
6. `export_output`
7. `verification_report`

## 6. Evidence 结构

建议每条证据至少包含：

1. `evidence_id`
2. `acceptance_run_id`
3. `surface`
4. `journey_id`
5. `evidence_type`
6. `source_brain`
7. `summary`
8. `artifact_uri`
9. `generated_at`
10. `status`

### 6.1 status

建议：

1. `collected`
2. `validated`
3. `rejected`

## 7. production_gates

建议至少包含三类：

1. `functional_gate`
2. `operability_gate`
3. `release_readiness_gate`

必要时增加：

4. `recovery_readiness_gate`

## 8. 裁决规则

### 8.1 functional_passed

满足以下条件：

1. 所有 blocking journeys 通过
2. 所有 blocking issues 已关闭
3. 基础验证证据齐全

### 8.2 production_passed

在 `functional_passed = true` 基础上，再满足：

1. required surfaces 全覆盖
2. required evidence 全覆盖
3. release readiness gate 通过
4. recovery readiness gate 通过或明确豁免

### 8.3 manual_release_required

以下情况任一命中应设为 true：

1. 分类 profile 指定必须人工放行
2. 高风险项目
3. 关键上线变更
4. 人工确认规则命中

### 8.4 released_by_human

仅在：

1. `manual_release_required = true`
2. 且明确执行人工放行动作后

才可置为 true。

## 9. AcceptanceRun 建议字段

建议至少包含：

1. `id`
2. `project_id`
3. `production_acceptance_profile_id`
4. `profile_version`
5. `status`
6. `functional_passed`
7. `production_passed`
8. `manual_release_required`
9. `released_by_human`
10. `blocking_issue_count`
11. `warning_count`
12. `started_at`
13. `ended_at`

## 10. 分类差异样例

### 10.1 Web

强调：

1. 用户端真实操作
2. 后台端真实操作
3. CI 证据
4. 状态读写一致

### 10.2 Game

强调：

1. 运行时启动证据
2. 核心玩法循环证据
3. 引擎运行日志

### 10.3 Video Editing

强调：

1. 编辑器运行态
2. 导入/编辑/导出主路径
3. 导出产物

## 11. 与工作台的关系

这些对象最终进入：

1. `AcceptanceView.coverage`
2. `AcceptanceView.evidence_cards`
3. `AcceptanceView.final_judgement`
4. `Workspace.acceptance_coverage`

## 12. 后续细分专题

本专题后续继续拆：

1. `Evidence` 存储与 retention 策略
2. `journey` 标准库
3. 不同分类 gate 样例
4. 人工放行动作记录模型
