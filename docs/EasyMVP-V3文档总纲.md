# EasyMVP V3 文档总纲

> 更新时间：2026-04-19  
> 目标：建立 EasyMVP V3 的完整文档体系，形成“总纲 -> 核心子系统 -> 细项专题”的稳定结构，支撑后续架构设计、研发实施、评审和交接。

## 1. V3 的一句话定义

EasyMVP V3 不是“再接更多执行器”的版本，而是：

> 以 `brain-v3` 为执行底座、以 `easymvp-brain` 为领域大脑、以“方案编译 + 分类驱动 + 生产级验收”为主线的工作流系统。

## 2. V3 的顶层目标

V3 必须同时满足以下 6 个目标：

1. 方案阶段完成大部分根因治理，而不是在 execute / rework / accept 阶段补救
2. 项目分类成为正式输入，驱动计划、角色、执行、验收差异
3. 角色自动解析成为默认主路径，人工配置只做覆盖
4. `brain-v3` 成为执行运行时底座，而不是普通 shell 执行器
5. 验收按项目分类进入“生产级验收”，而不是通用 accept 检查
6. 文档体系足够完整，能支持长期迭代、交接和专项扩展

## 3. 文档层级

V3 文档体系分三层：

### 3.1 总纲层

用于解释：

- V3 为什么要做
- V3 解决什么根因
- V3 的系统边界
- V3 的总体设计原则
- 各子系统之间的关系

当前文档：

- [EasyMVP-V3文档总纲](./EasyMVP-V3文档总纲.md)
- [EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)

### 3.2 核心子系统层

用于解释 V3 的四条核心主线：

1. 方案编译主线
2. 分类与角色主线
3. 专精大脑与执行底座主线
4. 生产级验收主线

当前文档：

- [EasyMVP-V3方案编译模型设计](./EasyMVP-V3方案编译模型设计.md)
- [EasyMVP-V3分类策略与角色自动解析设计](./EasyMVP-V3分类策略与角色自动解析设计.md)
- [EasyMVP-V3专精大脑接入计划](./EasyMVP-V3专精大脑接入计划.md)
- [EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3生产级分类验收体系设计.md)
- [EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md)

### 3.3 细项专题层

后续每个核心子系统再继续拆成专题文档，例如：

1. `PlanDraft / PlanReview / CompiledPlan` 表结构设计
2. `easymvp-brain` Manifest / Tool Schema / Prompt 设计
3. `CategoryProfile` 规范与版本化策略
4. `RoleCatalog / RoleResolver` 实现细节
5. `ProductionAcceptanceProfile` 规范与证据 Schema
6. `brain serve` 接口适配与 Run 生命周期接入细则
7. V2 -> V3 迁移计划与兼容策略
8. 单机版实时工作台产品形态与页面设计
9. 工作台视图模型与聚合接口设计
10. `Workspace / Plan / Acceptance` 详细页面设计
11. `easymvp-brain` 职责边界、Manifest、Tool Schema、Prompt 设计
12. V2 -> V3 迁移、实时事件流、回放与审计设计
13. 分类 profile 示例库与角色目录标准化

这些文档当前未全部展开，但必须围绕本总纲继续拆，不允许平铺散写。

## 4. V3 的四条核心主线

### 4.1 方案编译主线

V2 的问题之一，是架构师输出直接落成任务。

V3 要改成：

`PlanDraft -> PlanReviewResult -> CompiledPlan -> DomainTask`

方案先审核、再编译、再落任务。

对应文档：

- [EasyMVP-V3方案编译模型设计](./EasyMVP-V3方案编译模型设计.md)

### 4.2 分类与角色主线

V2 的问题之二，是项目分类和角色配置都太重、太散。

V3 要改成：

- `CategoryProfile` 驱动计划/验收差异
- `RoleCatalog + CategoryRoleProfile + ProjectRoleOverride + RoleResolver`

角色自动解析成为主路径，人工配置只做覆盖。

对应文档：

- [EasyMVP-V3分类策略与角色自动解析设计](./EasyMVP-V3分类策略与角色自动解析设计.md)

### 4.3 专精大脑与执行底座主线

V2 的问题之三，是执行器只是命令分发，缺统一底座。

V3 要改成：

- `brain-v3` 提供 Run / Tool / Sidecar / Replay / Policy 底座
- `easymvp-brain` 提供领域思考能力

对应文档：

- [EasyMVP-V3专精大脑接入计划](./EasyMVP-V3专精大脑接入计划.md)

### 4.4 生产级验收主线

V2 的问题之四，是 accept 更像通用验收，不够“生产级”。

V3 要改成：

- 验收按项目分类驱动
- 验收按 surface/journey/evidence 编排
- 最终判定目标是 `production_ready`

对应文档：

- [EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3生产级分类验收体系设计.md)
- [EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md)
- [EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)

## 5. 文档编写规则

后续 V3 文档必须遵守以下规则：

1. 先总纲，再子系统，再细项，不允许直接落零散专题而没有挂靠主线
2. 每篇文档必须说明：
   - 目标
   - 要解决的问题
   - 上游依赖
   - 输出对象
   - 与其他文档的关系
3. 每个稳定对象都要明确：
   - 输入
   - 输出
   - 状态
   - 版本化策略
4. 每个设计都要区分：
   - 通用部分
   - 分类差异部分
5. 所有“生产级”定义必须可落成结构化规则，不允许只停留在自然语言描述

## 6. 推荐阅读顺序

第一次进入 V3 设计的人，按这个顺序读：

1. [EasyMVP-V3文档总纲](./EasyMVP-V3文档总纲.md)
2. [EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
3. [EasyMVP-V3专精大脑接入计划](./EasyMVP-V3专精大脑接入计划.md)
4. [EasyMVP-V3方案编译模型设计](./EasyMVP-V3方案编译模型设计.md)
5. [EasyMVP-V3分类策略与角色自动解析设计](./EasyMVP-V3分类策略与角色自动解析设计.md)
6. [EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3生产级分类验收体系设计.md)

## 7. 后续要继续补的专题

建议下一批拆出的细项专题如下：

1. `PlanDraft / PlanReviewResult / CompiledPlan` 数据模型与表结构设计
2. `easymvp-brain` Manifest / Tool Schema / Prompt 设计
3. `CategoryProfile` 结构与版本化规范
4. `RoleResolver` 解析规则与覆盖优先级
5. `ProductionAcceptanceProfile` 证据结构与裁决规则
6. `brain serve` 接口接入与 Run 生命周期映射
7. V2 -> V3 迁移与兼容策略
8. 单机版实时工作台页面结构与实时数据模型
9. 工作台聚合视图对象与页面接口

## 8. 文档优先级

为了避免平均用力，V3 文档应按 `P0 / P1 / P2` 推进。

### 8.1 P0：必须先稳定

这些文档直接决定主链路、对象边界和接口边界：

1. [EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
2. [EasyMVP-V3方案编译模型设计](./EasyMVP-V3方案编译模型设计.md)
3. [EasyMVP-V3计划数据模型与表结构设计](./EasyMVP-V3计划数据模型与表结构设计.md)
4. [EasyMVP-V3分类策略与角色自动解析设计](./EasyMVP-V3分类策略与角色自动解析设计.md)
5. [EasyMVP-V3分类Profile结构与版本化规范](./EasyMVP-V3分类Profile结构与版本化规范.md)
6. [EasyMVP-V3-RoleResolver解析规则与优先级设计](./EasyMVP-V3-RoleResolver解析规则与优先级设计.md)
7. [EasyMVP-V3专精大脑接入计划](./EasyMVP-V3专精大脑接入计划.md)
8. [EasyMVP-V3-brain-serve接口接入与Run生命周期映射](./EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md)
9. [EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3生产级分类验收体系设计.md)
10. [EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则](./EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md)
11. [EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)

### 8.2 P1：紧接着落地

这些文档主要承接产品形态和前端实施：

1. [EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md)
2. [EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
3. [EasyMVP-V3-Plan详细页面设计](./EasyMVP-V3-Plan详细页面设计.md)
4. [EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
5. [EasyMVP-V3-easymvp-brain职责边界与输入输出合同设计](./EasyMVP-V3-easymvp-brain职责边界与输入输出合同设计.md)
6. [EasyMVP-V3-easymvp-brain-Manifest与ToolSchema设计](./EasyMVP-V3-easymvp-brain-Manifest与ToolSchema设计.md)
7. [EasyMVP-V3-easymvp-brain-Prompt设计](./EasyMVP-V3-easymvp-brain-Prompt设计.md)

### 8.3 P2：工程化与迁移

这些文档在主链路稳定后推进：

1. [EasyMVP-V3-V2到V3迁移与兼容策略](./EasyMVP-V3-V2到V3迁移与兼容策略.md)
2. [EasyMVP-V3-实时事件流推送机制设计](./EasyMVP-V3-实时事件流推送机制设计.md)
3. [EasyMVP-V3-回放与审计展示设计](./EasyMVP-V3-回放与审计展示设计.md)
4. [EasyMVP-V3-分类Profile示例库](./EasyMVP-V3-分类Profile示例库.md)
5. [EasyMVP-V3-角色目录标准化设计](./EasyMVP-V3-角色目录标准化设计.md)

## 9. 当前状态

当前 V3 文档体系已经完成：

1. 总纲入口
2. 总体架构
3. 方案编译主线
4. 分类与角色主线
5. 专精大脑接入主线
6. 生产级分类验收主线
7. 实时工作台产品形态专题
8. 工作台视图模型专题
9. P0 计划数据模型专题
10. P0 分类 profile 规范专题
11. P0 RoleResolver 专题
12. P0 ProductionAcceptanceProfile 专题
13. P0 `brain serve` 生命周期映射专题
14. P1 页面级详细设计专题
15. P1 `easymvp-brain` 合同与 Manifest 专题
16. P2 迁移与工程化专题
17. P2 分类示例库与角色目录专题

下一步开始进入细项专题展开。
