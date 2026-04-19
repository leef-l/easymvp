# EasyMVP V3 分类策略与角色自动解析设计

> 更新时间：2026-04-19  
> 上游文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)  
> 目标：让项目分类成为正式输入，让角色自动解析成为默认主路径。

## 1. 设计结论

V3 不应该为每一种项目分类单独做一套脑或一套工作流。

正确做法是：

- 一个通用 `easymvp-brain`
- 一套 `CategoryProfile`
- 一套 `RoleResolver`

分类决定策略差异，脑负责通用 workflow intelligence。

## 2. CategoryProfile

`CategoryProfile` 是项目分类的正式策略对象。

它至少负责 5 类差异：

1. `planning_policy`
2. `role_policy`
3. `delivery_policy`
4. `verification_policy`
5. `acceptance_policy`

这样未来扩更多项目分类时，主要扩的是 profile，而不是重写主系统。

## 3. 角色系统的重构方向

旧的角色配置方式太重。  
V3 要改成四层：

1. `RoleCatalog`
2. `CategoryRoleProfile`
3. `ProjectRoleOverride`
4. `RoleResolver`

原则：

- 系统先按分类自动给默认角色
- 项目级配置只做覆盖
- 所有任务都通过 `RoleResolver` 统一拿角色

## 4. RoleResolver 输入输出

输入建议至少包括：

1. `project_category`
2. `phase`
3. `task_kind`
4. `risk_level`
5. `capability_need`

输出至少包括：

1. `role_type`
2. `brain_kind`
3. `required`
4. `source`

## 5. V3 想解决的问题

这条主线主要解决：

1. 新项目不能因为没手工配角色就跑不起来
2. 不同分类能自动切换默认角色班底
3. 角色配置从重资产变成轻覆盖
4. role 最终应映射到 brain，而不是只映射到 execution_mode

## 6. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3分类Profile结构与版本化规范](./EasyMVP-V3分类Profile结构与版本化规范.md)
2. [EasyMVP-V3-RoleResolver解析规则与优先级设计](./EasyMVP-V3-RoleResolver解析规则与优先级设计.md)
3. `RoleCatalog` 标准角色目录
4. `CategoryRoleProfile` 设计
5. `ProjectRoleOverride` 规则
6. `RoleResolver` 优先级和回退策略
