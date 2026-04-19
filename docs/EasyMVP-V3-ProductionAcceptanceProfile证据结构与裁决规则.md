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

按当前钱学森总纲口径，这里的“真正完成”还必须再满足两层前提：

1. `Verification Contract` 中的 `required_checks / required_evidence / preferred_channel` 已被结构化兑现
2. `CompletionVerdict.completed = true`，不能用 `production_passed = true` 直接偷换成完成

换句话说：

- `ProductionAcceptanceProfile` 解决“这类项目最终要覆盖什么”
- `Verification Contract` 解决“这项任务当下要怎么验”
- `CompletionVerdict` 解决“现在能不能真正 completed”

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
10. `preferred_verification_channel`
11. `fallback_verification_channels`

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

补充口径：

- 当前阶段允许 `ci_result` 对应 `github_actions` 产出的远端验证证据
- 但文档语义必须明确：`github_actions` 是当前替代通道，不是长期最终验证环境
- 长期目标仍然是 `high_spec_remote`

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

并且建议补一条：

5. 关键验证项的 `VerificationResult` 不存在未收口的 `channel_unavailable / verification_conflict / manual_review_required`

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

这里要特别防止一个旧误解：

- `released_by_human = true` 不等于系统已经业务完成
- 只有与 `CompletionVerdict.completed = true` 联合成立时，才表示 EasyMVP 闭环真正收口

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
14. `channel`
15. `missing_evidence_count`
16. `failed_check_count`
17. `decision`

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

按当前 EasyMVP 页面口径，还应明确：

5. `AcceptancePage` 不能只看 `AcceptanceRun`
6. 页面必须同时展示合同要求、实际执行结果、证据缺口和完成裁决
7. 若当前通道为 `github_actions`，页面上必须展示“替代通道”说明

## 12. 与当前总纲的冲突处理

如果本文件和最新总纲文档发生冲突，统一以下列文档为准：

1. [EasyMVP-Verification-Contract统一设计](./钱学森总纲设计/EasyMVP-Verification-Contract统一设计.md)
2. [EasyMVP-对象级字段清单](./钱学森总纲设计/EasyMVP-对象级字段清单.md)
3. [EasyMVP-页面读取与展示清单](./钱学森总纲设计/EasyMVP-页面读取与展示清单.md)
4. [EasyMVP-闭环状态机补充说明](./钱学森总纲设计/EasyMVP-闭环状态机补充说明.md)

## 13. 后续细分专题

本专题后续继续拆：

1. `Evidence` 存储与 retention 策略
2. `journey` 标准库
3. 不同分类 gate 样例
4. 人工放行动作记录模型
