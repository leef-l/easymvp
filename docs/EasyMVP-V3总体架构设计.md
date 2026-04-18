# EasyMVP V3 总体架构设计

> 更新时间：2026-04-19  
> 上游文档：[EasyMVP-V3文档总纲](./EasyMVP-V3文档总纲.md)  
> 目标：给出 EasyMVP V3 的顶层系统边界、核心对象、主链路和分层职责。

## 1. 设计结论

EasyMVP V3 的顶层结构分为三层：

1. EasyMVP Workflow Orchestrator
2. EasyMVP Domain Brain（`easymvp-brain`）
3. Brain Runtime Base（`brain-v3`）

其中：

- EasyMVP 主仓继续负责工作流状态机和业务对象
- `easymvp-brain` 负责 EasyMVP 领域认知
- `brain-v3` 负责运行时、工具、Run 生命周期和多脑协作

## 2. V3 要解决的根因

V2 的主要根因有：

1. 架构师输出直接落任务，缺方案编译层
2. review / execute / rework / accept 的决策逻辑散落
3. 项目分类没有成为系统主输入
4. 角色配置过重，自动解析能力不足
5. 执行器只是一层命令分发，没有统一 Run 生命周期
6. 验收不够分类化，也不够生产级

V3 的设计必须正面解决这 6 类问题。

## 3. 核心对象

V3 第一批必须稳定下来的对象：

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

## 4. 顶层链路

V3 的主链路：

```text
用户需求
  ↓
PlanDraft
  ↓
PlanReviewResult
  ↓
CompiledPlan
  ↓
DomainTask
  ↓
Executor Run / Delivery / Verification
  ↓
Production Acceptance
  ↓
Complete
```

与 V2 最大的区别是：

- 方案不能直接进入 execute
- 验收目标不是普通通过，而是生产级通过

## 5. 三层分工

### 5.1 EasyMVP Workflow Orchestrator

负责：

- 项目、阶段、任务、验收状态机
- 所有领域对象持久化
- 人工介入与控制 API
- 管理 `brain-v3` Run 调用

### 5.2 EasyMVP Domain Brain

负责：

- 方案审核
- 方案编译
- rework 重构
- 验收规则映射
- 最终业务裁决

### 5.3 Brain Runtime Base

负责：

- Run 生命周期
- 工具治理
- sidecar / 多脑调度
- replay / logs / resume / cancel
- sandbox / file policy / workdir policy

## 6. 总体原则

V3 的设计原则：

1. 方案先审核、再编译、再执行
2. 分类先决定策略，再影响任务
3. 角色默认自动解析，人工配置只做覆盖
4. 执行期只做执行，不临时发明业务规则
5. 验收按分类、按 surface、按 journey、按证据判定
6. 最终目标是 `production_ready`

## 7. 相关子文档

- 方案编译： [EasyMVP-V3方案编译模型设计](./EasyMVP-V3方案编译模型设计.md)
- 计划数据： [EasyMVP-V3计划数据模型与表结构设计](./EasyMVP-V3计划数据模型与表结构设计.md)
- 分类与角色： [EasyMVP-V3分类策略与角色自动解析设计](./EasyMVP-V3分类策略与角色自动解析设计.md)
- 分类 profile： [EasyMVP-V3分类Profile结构与版本化规范](./EasyMVP-V3分类Profile结构与版本化规范.md)
- 角色解析： [EasyMVP-V3-RoleResolver解析规则与优先级设计](./EasyMVP-V3-RoleResolver解析规则与优先级设计.md)
- 执行底座： [EasyMVP-V3专精大脑接入计划](./EasyMVP-V3专精大脑接入计划.md)
- Run 接入： [EasyMVP-V3-brain-serve接口接入与Run生命周期映射](./EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md)
- 验收体系： [EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3生产级分类验收体系设计.md)
- 裁决规则： [EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则](./EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md)
- 产品形态： [EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md)
- 视图模型： [EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
