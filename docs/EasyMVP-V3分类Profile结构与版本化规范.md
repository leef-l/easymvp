# EasyMVP V3 分类Profile结构与版本化规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3分类策略与角色自动解析设计](./EasyMVP-V3分类策略与角色自动解析设计.md)
> 关联文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
> 关联文档：[EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3生产级分类验收体系设计.md)
> 目标：把 `CategoryProfile` 定义为 V3 的正式策略对象，并给出结构、继承边界和版本化规则。

## 1. 设计结论

`CategoryProfile` 是 V3 分类驱动设计的核心。

它不是一组散落配置，而是单个版本化策略对象，至少统一承载：

1. `planning_policy`
2. `role_policy`
3. `delivery_policy`
4. `verification_policy`
5. `acceptance_policy`

## 2. 为什么必须独立成对象

如果没有正式 `CategoryProfile`，V3 会重新退回：

1. 各分类逻辑散落代码
2. 执行期临时判断策略
3. 角色解析和验收规则脱节
4. 新分类扩展成本过高

## 3. CategoryProfile 顶层结构

建议顶层包含：

1. `profile_id`
2. `profile_name`
3. `project_category`
4. `version`
5. `base_profile`
6. `planning_policy`
7. `role_policy`
8. `delivery_policy`
9. `verification_policy`
10. `acceptance_policy`
11. `status`
12. `created_at`
13. `updated_at`

## 4. planning_policy

建议至少包含：

1. `plan_style`
2. `task_split_limits`
3. `allowed_task_kinds`
4. `required_plan_sections`
5. `high_risk_task_rules`

### 4.1 plan_style

例如：

1. `web_app`
2. `game_runtime`
3. `editor_pipeline`

### 4.2 task_split_limits

建议包含：

1. `max_resources_per_task`
2. `max_scope_level`
3. `must_split_when_blocking_review`

## 5. role_policy

建议至少包含：

1. `default_roles_by_phase`
2. `default_roles_by_task_kind`
3. `preferred_brains_by_role`
4. `risk_overrides`
5. `manual_review_roles`

## 6. delivery_policy

建议至少包含：

1. `required_delivery_kinds`
2. `required_artifacts`
3. `must_define_scope`
4. `must_define_output`
5. `allow_empty_delivery`

## 7. verification_policy

建议至少包含：

1. `required_check_kinds`
2. `browser_required`
3. `ci_required`
4. `runtime_proof_required`
5. `verification_brains`

## 8. acceptance_policy

建议至少包含：

1. `required_surfaces`
2. `required_journeys`
3. `required_evidence`
4. `production_gate_rules`
5. `manual_release_rules`

## 9. 继承与覆盖规则

`CategoryProfile` 建议支持“基础 profile + 分类 profile”两层。

### 9.1 可继承部分

1. 通用 planning 默认值
2. 通用 role 默认值
3. 通用 evidence 默认值

### 9.2 必须显式定义部分

1. `project_category`
2. `required_surfaces`
3. `required_journeys`
4. `production_gate_rules`

### 9.3 不允许项目级直接改写部分

1. 分类本身
2. 核心 production gate
3. 核心 acceptance evidence 结构

## 10. 版本化规则

### 10.1 升版本触发条件

以下变化必须升新版本：

1. required surfaces 变化
2. required journeys 变化
3. default role 或 brain 变化
4. production gate 变化
5. 核心 verification 规则变化

### 10.2 不强制升版本的变化

以下变化可走补丁版本：

1. 文案说明
2. 非阻塞建议项
3. 默认排序

### 10.3 历史运行绑定

项目一旦进入 `CompiledPlan` 或 `AcceptanceRun`，必须记录当时使用的：

1. `category_profile_id`
2. `category_profile_version`

不允许运行中隐式漂移到新 profile。

## 11. 建议存储结构

建议表：

1. `workflow_category_profiles`
2. `workflow_category_profile_versions`

`workflow_category_profiles` 负责逻辑 identity。

`workflow_category_profile_versions` 负责版本快照。

## 12. 分类样例

### 12.1 Web

核心特点：

1. `required_surfaces` 包含 `user_frontend` 和 `admin_backend`
2. `browser_required = true`
3. `ci_required = true`

### 12.2 Game

核心特点：

1. `required_surfaces` 包含 `game_runtime`
2. `runtime_proof_required = true`
3. `required_journeys` 强调玩法循环

### 12.3 Video Editing

核心特点：

1. `required_surfaces` 包含 `editor_runtime`
2. `required_journeys` 强调导入/编辑/导出
3. `required_evidence` 强调导出产物

## 13. 与其他对象的关系

`CategoryProfile` 会直接影响：

1. `PlanDraft` 审核标准
2. `CompiledPlan` 拆分规则
3. `RoleResolver` 默认结果
4. `VerificationContract`
5. `ProductionAcceptanceProfile`

## 14. 后续细分专题

本专题后续继续拆：

1. `planning_policy` json schema
2. `role_policy` json schema
3. `acceptance_policy` json schema
4. 分类 profile 示例库
