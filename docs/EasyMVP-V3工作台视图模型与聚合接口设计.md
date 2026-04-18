# EasyMVP V3 工作台视图模型与聚合接口设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
> 关联文档：[EasyMVP-V3方案编译模型设计](./EasyMVP-V3方案编译模型设计.md)
> 关联文档：[EasyMVP-V3分类策略与角色自动解析设计](./EasyMVP-V3分类策略与角色自动解析设计.md)
> 关联文档：[EasyMVP-V3专精大脑接入计划](./EasyMVP-V3专精大脑接入计划.md)
> 关联文档：[EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3生产级分类验收体系设计.md)
> 关联文档：[EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md)
> 目标：定义 V3 核心业务对象如何聚合为前端可消费的工作台视图对象，并明确 Workspace / Plan / Acceptance 三类页面的接口边界。

## 1. 设计结论

V3 页面设计不能直接面向底层业务对象逐个取数。

正确做法是增加一层稳定的视图模型层：

1. 后端继续维护业务对象和状态机
2. 视图聚合层负责把业务对象组合成页面对象
3. 前端只消费聚合后的页面模型

这层的作用不是替代业务模型，而是：

1. 降低前端对内部状态机的耦合
2. 让实时工作台具备稳定数据边界
3. 让页面和 V3 主链路保持一一映射

## 2. 为什么需要这一层

如果没有视图模型层，前端会直接面对：

1. `PlanDraft`
2. `PlanReviewResult`
3. `CompiledPlan`
4. `DomainTask`
5. `DeliveryContract`
6. `VerificationContract`
7. `brain-v3` Run
8. `AcceptanceRun`

这会带来 4 个问题：

1. 页面需要理解过多领域细节
2. 实时刷新边界会非常混乱
3. UI 会反过来绑死后端内部结构
4. 工作台很容易退化为“对象详情页拼接”

所以 V3 必须定义正式的聚合视图对象。

## 3. 视图模型层的定位

V3 的对象分三层看：

### 3.1 领域对象层

由 Workflow Orchestrator 持有：

1. `ProjectCategory`
2. `CategoryProfile`
3. `PlanDraft`
4. `PlanReviewResult`
5. `CompiledPlan`
6. `DomainTask`
7. `DeliveryContract`
8. `VerificationContract`
9. `ProductionAcceptanceProfile`
10. `AcceptanceRun`

### 3.2 运行时对象层

由 `brain-v3` 提供：

1. `run_id`
2. `status`
3. `logs`
4. `replay`
5. `cancel`
6. `resume`

### 3.3 视图对象层

由 EasyMVP 聚合接口提供：

1. `WorkspaceView`
2. `PlanView`
3. `AcceptanceView`
4. `ProjectSnapshot`
5. `StageProgress`
6. `LiveEvent`
7. `ActionInboxItem`
8. `AcceptanceCoverage`

## 4. 聚合设计原则

### 4.1 页面消费聚合对象，不直接拼底层对象

前端默认不应直接自行组装：

1. `CompiledPlan + DomainTask + run logs`
2. `AcceptanceRun + Evidence + blocker`
3. `PlanDraft + Review diff + role resolve`

这些应由后端先做归一化。

### 4.2 聚合对象按页面语义组织

不要按数据库或服务边界组织接口。

应按页面语义组织：

1. `Workspace` 看当前推进态
2. `Plan` 看决策与编译态
3. `Acceptance` 看生产级裁决态

### 4.3 实时对象和静态对象分离

以下内容变化频繁：

1. `current_stage`
2. `active_run`
3. `LiveEvent`
4. `ActionInbox`
5. `AcceptanceCoverage`

以下内容相对稳定：

1. `PlanDraft`
2. `PlanReviewResult`
3. `CompiledPlan`
4. `ProductionAcceptanceProfile`

因此接口层应区分：

1. 快照接口
2. 增量流接口
3. 详情接口

### 4.4 聚合对象必须可回溯到来源对象

每个视图对象都应保留来源引用，避免页面层变成黑盒。

至少应支持：

1. `source_object_kind`
2. `source_object_id`
3. `source_run_id`
4. `source_task_id`

## 5. Workspace 的视图模型

`Workspace` 是 V3 的实时主页面。

它的根对象应定义为 `WorkspaceView`。

### 5.1 WorkspaceView

建议包含：

1. `project`
2. `snapshot`
3. `stage_progress`
4. `active_runs`
5. `live_events`
6. `action_inbox`
7. `acceptance_coverage`
8. `quick_actions`

### 5.2 project

用于展示稳定项目头信息。

建议包含：

1. `project_id`
2. `project_name`
3. `project_category`
4. `workspace_path`
5. `created_at`
6. `updated_at`

### 5.3 snapshot

用于支持顶部总状态条。

建议包含：

1. `current_stage`
2. `overall_progress`
3. `current_run_status`
4. `blocker_count`
5. `waiting_action_count`
6. `production_readiness`
7. `final_status_hint`

### 5.4 stage_progress

用于支撑阶段流。

建议为数组，每项包含：

1. `stage_name`
2. `status`
3. `started_at`
4. `ended_at`
5. `duration_ms`
6. `active_item_title`
7. `blocker_count`
8. `linked_view`

`linked_view` 只允许跳到：

1. `Plan`
2. `Workspace` 内部展开
3. `Acceptance`

### 5.5 active_runs

用于显示当前活跃 brain 运行态。

建议每项包含：

1. `run_id`
2. `brain_kind`
3. `task_id`
4. `task_name`
5. `status`
6. `started_at`
7. `last_event_at`
8. `progress_text`

### 5.6 live_events

用于支撑中部实时事件流。

建议每项包含：

1. `event_id`
2. `event_type`
3. `severity`
4. `summary`
5. `created_at`
6. `source_brain`
7. `source_task_id`
8. `source_run_id`
9. `needs_attention`
10. `deep_link`

### 5.7 action_inbox

用于支撑右侧待处理问题区。

建议每项包含：

1. `item_id`
2. `kind`
3. `severity`
4. `blocking`
5. `title`
6. `summary`
7. `recommended_action`
8. `action_button_text`
9. `action_target`
10. `source_object_kind`
11. `source_object_id`

### 5.8 acceptance_coverage

用于支撑底部覆盖矩阵。

建议每项包含：

1. `surface`
2. `journey`
3. `evidence_type`
4. `coverage_status`
5. `blocking_gap`
6. `last_verified_at`
7. `detail_target`

### 5.9 quick_actions

用于顶部或右侧快捷动作。

建议只允许 5 类：

1. `view_plan`
2. `open_blockers`
3. `resume_run`
4. `start_acceptance`
5. `manual_release`

## 6. Plan 的视图模型

`Plan` 页不是任务管理页，而是决策解释页。

它的根对象应定义为 `PlanView`。

### 6.1 PlanView

建议包含：

1. `project`
2. `plan_draft`
3. `review_result`
4. `compiled_plan`
5. `compile_diffs`
6. `task_projection`
7. `role_resolution_summary`

### 6.2 plan_draft

用于展示原始设计输入。

建议包含：

1. `plan_id`
2. `version`
3. `summary`
4. `goals`
5. `raw_tasks`
6. `created_at`

### 6.3 review_result

用于展示 review 阶段输出。

建议包含：

1. `review_id`
2. `decision`
3. `blocking_issues`
4. `advisory_issues`
5. `split_suggestions`
6. `override_suggestions`
7. `reviewed_at`

### 6.4 compiled_plan

用于展示最终正式计划。

建议包含：

1. `compiled_plan_id`
2. `version`
3. `task_count`
4. `risk_summary`
5. `generated_at`

### 6.5 compile_diffs

用于解释编译前后差异。

建议每项包含：

1. `diff_kind`
2. `before_label`
3. `after_label`
4. `reason`
5. `source_review_issue_id`

### 6.6 task_projection

用于说明 `CompiledPlan` 如何落成 `DomainTask`。

建议每项包含：

1. `task_id`
2. `task_name`
3. `role_type`
4. `brain_kind`
5. `risk_level`
6. `delivery_summary`
7. `verification_summary`
8. `affected_resources`

### 6.7 role_resolution_summary

用于说明角色与 brain 的来源。

建议每项包含：

1. `task_id`
2. `role_type`
3. `brain_kind`
4. `resolve_source`
5. `category_profile_source`
6. `project_override_applied`

## 7. Acceptance 的视图模型

`Acceptance` 页是最终裁决页。

它的根对象应定义为 `AcceptanceView`。

### 7.1 AcceptanceView

建议包含：

1. `project`
2. `acceptance_profile`
3. `acceptance_run`
4. `coverage`
5. `evidence_cards`
6. `final_judgement`
7. `release_gate`

### 7.2 acceptance_profile

用于展示按分类生成的正式验收框架。

建议包含：

1. `project_category`
2. `required_surfaces`
3. `required_journeys`
4. `required_evidence`
5. `required_brains`

### 7.3 acceptance_run

用于展示当前验收执行态。

建议包含：

1. `acceptance_run_id`
2. `status`
3. `started_at`
4. `ended_at`
5. `blocking_issue_count`
6. `warning_count`

### 7.4 coverage

用于展示验收覆盖状态。

建议每项包含：

1. `surface`
2. `journey`
3. `status`
4. `evidence_count`
5. `blocking_gap`

### 7.5 evidence_cards

用于展示证据。

建议每项包含：

1. `evidence_id`
2. `evidence_type`
3. `title`
4. `summary`
5. `generated_at`
6. `source_brain`
7. `preview_target`

### 7.6 final_judgement

用于展示最终裁决。

建议包含：

1. `functional_passed`
2. `production_passed`
3. `manual_release_required`
4. `released_by_human`
5. `current_blocking_reason`

### 7.7 release_gate

用于展示最后放行动作。

建议包含：

1. `can_release`
2. `requires_manual_release`
3. `release_button_text`
4. `release_action_target`

## 8. 视图对象与来源对象映射

### 8.1 WorkspaceView 来源

主要来源：

1. `DomainTask`
2. `brain-v3 Run`
3. `DeliveryContract` 执行结果
4. `VerificationContract` 执行结果
5. `AcceptanceRun`

### 8.2 PlanView 来源

主要来源：

1. `PlanDraft`
2. `PlanReviewResult`
3. `CompiledPlan`
4. `CategoryProfile`
5. `RoleResolver` 结果

### 8.3 AcceptanceView 来源

主要来源：

1. `ProductionAcceptanceProfile`
2. `AcceptanceRun`
3. `Evidence`
4. `browser-brain / verifier-brain / easymvp-brain` 结果

## 9. 接口边界建议

V3 第一版建议只开放 3 类页面接口和 1 类事件流接口。

### 9.1 页面快照接口

建议：

1. `GET /api/v3/projects/{id}/workspace-view`
2. `GET /api/v3/projects/{id}/plan-view`
3. `GET /api/v3/projects/{id}/acceptance-view`

### 9.2 事件流接口

建议：

1. `GET /api/v3/projects/{id}/live-events`

第一版可以先用短轮询或 SSE。

### 9.3 操作接口

建议操作接口不要直接暴露底层对象修改语义。

应面向动作语义，例如：

1. `POST /api/v3/projects/{id}/actions/resume-run`
2. `POST /api/v3/projects/{id}/actions/start-acceptance`
3. `POST /api/v3/projects/{id}/actions/manual-release`
4. `POST /api/v3/projects/{id}/actions/resolve-blocker`

## 10. 刷新策略建议

### 10.1 Workspace

建议：

1. 快照短轮询
2. 事件流增量刷新
3. 待处理项立即刷新

### 10.2 Plan

建议：

1. 以快照拉取为主
2. 编译结果变更后局部刷新

### 10.3 Acceptance

建议：

1. 验收运行中增量刷新
2. 最终裁决变化立即刷新

## 11. 页面跳转规则

### 11.1 Workspace -> Plan

在以下情况下可跳：

1. 点击阶段流中的 `Design / Review / Compile`
2. 点击阻塞项中的方案问题
3. 点击事件流中的编译变更事件

### 11.2 Workspace -> Acceptance

在以下情况下可跳：

1. 点击阶段流中的 `Acceptance`
2. 点击覆盖矩阵
3. 点击缺失证据卡片

### 11.3 Plan -> Workspace

在以下情况下可跳：

1. 点击任务投影项
2. 点击某任务当前运行状态

## 12. 不该怎么做

不应该：

1. 让前端自己拼一堆底层对象
2. 让页面直接依赖数据库表结构
3. 把 `brain-v3` logs 原样塞进主页面
4. 把验收页做成 issue 列表页
5. 把工作台做成对象详情页集合

## 13. 与后续页面设计的关系

本专题完成后，页面设计应按以下顺序继续推进：

1. `Workspace` 详细页面设计
2. `Plan` 详细页面设计
3. `Acceptance` 详细页面设计
4. SSE / polling 事件流设计

页面详细设计必须基于本页定义的视图对象展开，不应重新发明数据结构。
