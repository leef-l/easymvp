# EasyMVP V3 角色目录标准化设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-RoleResolver解析规则与优先级设计](./EasyMVP-V3-RoleResolver解析规则与优先级设计.md)
> 关联文档：[EasyMVP-V3分类策略与角色自动解析设计](./EasyMVP-V3分类策略与角色自动解析设计.md)
> 目标：把 V3 的 `RoleCatalog` 从零散命名升级为标准角色目录，作为 `RoleResolver`、工作台解释和配置覆盖的统一基线。

## 1. 设计结论

V3 不能继续依赖临时角色命名。

第一版必须建立正式 `RoleCatalog`，统一：

1. 角色标识
2. 角色语义
3. 默认 brain
4. 允许出现的阶段
5. 风险升级行为

## 2. RoleCatalog 顶层结构

建议每个角色至少包含：

1. `role_type`
2. `display_name`
3. `description`
4. `default_brain_kind`
5. `allowed_phases`
6. `allowed_task_kinds`
7. `manual_review_capable`
8. `status`

## 3. 第一版标准角色

### 3.1 architect_reviewer

职责：

1. 审核方案
2. 标记 blocking issue
3. 给出 split / override 建议

默认 brain：

1. `easymvp-brain`

### 3.2 plan_compiler

职责：

1. 编译计划
2. 补齐合同
3. 决定正式任务结构

默认 brain：

1. `easymvp-brain`

### 3.3 code_executor

职责：

1. 执行正式代码实现任务
2. 产出交付结果

默认 brain：

1. `code-brain`

### 3.4 browser_verifier

职责：

1. 浏览器路径验证
2. 页面采证

默认 brain：

1. `browser-brain`

### 3.5 acceptance_verifier

职责：

1. 只读验收核验
2. 汇总证据
3. 给出结构化验证结果

默认 brain：

1. `verifier-brain`

### 3.6 release_gate_reviewer

职责：

1. 判断放行前置项
2. 裁决是否需要人工 release

默认 brain：

1. `easymvp-brain`

## 4. 命名规则

### 4.1 统一使用英文 `role_type`

例如：

1. `architect_reviewer`
2. `plan_compiler`
3. `code_executor`

### 4.2 不允许混用临时别名

例如不应再出现：

1. `architect-reviewer`
2. `planCompiler`
3. `codeExec`

## 5. 与 RoleResolver 的关系

`RoleResolver` 只能输出 `RoleCatalog` 中存在的标准角色。

如果命中未知角色：

1. 视为 `resolve_failed`
2. 阻止 compile 激活
3. 写入 blocking issue

## 6. 与 ProjectRoleOverride 的关系

项目级覆盖允许：

1. 改标准角色到另一个标准角色
2. 改默认 brain 到允许的替代 brain

项目级覆盖不允许：

1. 引入目录外角色
2. 绕过高风险人工审核要求

## 7. 与工作台解释的关系

工作台中所有 role 展示都应来源于目录标准项，而不是临时文本。

这样用户才能看懂：

1. 当前角色是什么
2. 为什么命中这个角色
3. 为什么它对应这个 brain

## 8. 建议存储结构

建议表：

1. `workflow_role_catalog`
2. `workflow_role_catalog_versions`

如果第一版不单独建表，也应至少先固化成标准配置文件。

## 9. 后续细分专题

本专题后续继续拆：

1. `RoleCatalog` JSON schema
2. 角色目录版本化规则
3. 角色显示名本地化策略
