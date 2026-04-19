# EasyMVP V3 RoleResolver 解析规则与优先级设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3分类策略与角色自动解析设计](./EasyMVP-V3分类策略与角色自动解析设计.md)
> 关联文档：[EasyMVP-V3分类Profile结构与版本化规范](./EasyMVP-V3分类Profile结构与版本化规范.md)
> 关联文档：[EasyMVP-V3方案编译模型设计](./EasyMVP-V3方案编译模型设计.md)
> 目标：定义 `RoleResolver` 的输入、解析顺序、覆盖优先级和回退规则，让角色自动解析成为稳定主路径。

## 1. 设计结论

`RoleResolver` 不应是简单映射表，而应是正式解析器。

它的职责是：

1. 根据分类和任务上下文解析默认角色
2. 同步解析对应 brain
3. 在高风险场景补充人工审核要求
4. 输出可解释来源

## 2. 输入输出

### 2.1 输入

建议至少包含：

1. `project_category`
2. `category_profile_version`
3. `phase`
4. `task_kind`
5. `risk_level`
6. `capability_need`
7. `project_role_override`
8. `manual_review_required`

### 2.2 输出

建议至少包含：

1. `role_type`
2. `brain_kind`
3. `required`
4. `manual_review_required`
5. `resolve_source`
6. `resolve_trace_json`

## 3. 解析顺序

建议固定为 5 层：

1. 硬性任务覆盖
2. 项目级 override
3. 分类级 task kind 规则
4. 分类级 phase 规则
5. 全局回退规则

## 4. 优先级规则

### 4.1 硬性任务覆盖

以下情况优先级最高：

1. 任务明确指定 `brain_kind`
2. 任务明确指定 `role_type`
3. 高风险任务必须人工审核

### 4.2 项目级 override

允许项目做轻覆盖，但不允许破坏分类核心 gate。

项目级 override 可影响：

1. `role_type`
2. `brain_kind`
3. `manual_review_required`

项目级 override 不应影响：

1. 核心 acceptance 规则
2. 核心 production gate

### 4.3 分类级 task kind 规则

例如：

1. `frontend_build` 默认前端实现角色
2. `browser_acceptance` 默认浏览器验证角色
3. `plan_review` 默认 `easymvp-brain`

### 4.4 分类级 phase 规则

例如：

1. `Review` 阶段优先 `easymvp-brain`
2. `Execute` 阶段优先 `code-brain`
3. `Acceptance` 阶段优先 `browser-brain` 或 `verifier-brain`

### 4.5 全局回退规则

当以上都未命中时：

1. `Design / Review / Compile` 回退到 `easymvp-brain`
2. `Execute` 回退到 `code-brain`
3. `Acceptance` 回退到 `verifier-brain`

## 5. 风险升级规则

以下情况应自动升级要求：

1. `risk_level = high`
2. 涉及大范围资源
3. 涉及生产级验收关键路径
4. 涉及人工放行前置项

自动升级后可追加：

1. `manual_review_required = true`
2. 更严格的 `brain_kind`
3. 更高优先级的 verifier 角色

## 6. resolve_trace_json

为了保证可解释性，建议输出解析链路。

每项包含：

1. `rule_layer`
2. `matched_rule`
3. `before_role`
4. `after_role`
5. `before_brain`
6. `after_brain`
7. `reason`

## 7. 建议角色目录

第一版建议至少有：

1. `architect_reviewer`
2. `plan_compiler`
3. `code_executor`
4. `browser_verifier`
5. `acceptance_verifier`
6. `release_gate_reviewer`

## 8. 角色到 brain 的映射

建议：

1. `architect_reviewer -> easymvp-brain`
2. `plan_compiler -> easymvp-brain`
3. `code_executor -> code-brain`
4. `browser_verifier -> browser-brain`
5. `acceptance_verifier -> verifier-brain`
6. `release_gate_reviewer -> easymvp-brain`

## 9. 与 CompiledTask 的关系

`RoleResolver` 的输出应写入 `CompiledTask`：

1. `role_type`
2. `brain_kind`
3. `manual_review_required`
4. `resolve_trace_json`

这样工作台和计划页才能解释：

1. 为什么是这个角色
2. 为什么是这个 brain
3. 是否存在项目级覆盖

## 10. 异常与回退

如果解析失败，不应静默兜底到任意角色。

建议：

1. 记录 `resolve_failed`
2. 写入 blocking issue
3. 阻止进入 compile 激活

## 11. 后续细分专题

本专题后续继续拆：

1. `RoleCatalog` 字段定义
2. `project_role_override` schema
3. 风险升级规则表
4. resolve trace 可视化展示
