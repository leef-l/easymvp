# EasyMVP V3 方案编译模型设计

> 更新时间：2026-04-19  
> 上游文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)  
> 目标：把“架构师方案”升级为“可审核、可编译、可执行的正式对象”。

## 1. 设计结论

V3 不再允许架构师输出直接落成 `DomainTask`。

必须先经过：

`PlanDraft -> PlanReviewResult -> CompiledPlan`

只有 `CompiledPlan` 才能投影生成正式任务。

## 2. 三个核心对象

### 2.1 PlanDraft

来源：architect  
性质：原始方案草案  
特点：

- 允许不完美
- 允许存在待修问题
- 不允许直接执行

### 2.2 PlanReviewResult

来源：`easymvp-brain`  
性质：结构化审核结果  
作用：

- 标出 blocking/advisory 问题
- 给出 split/drop/override 建议
- 决定是否允许进入 compile

### 2.3 CompiledPlan

来源：compiler  
性质：正式执行计划  
作用：

- 给出最终任务集
- 给出 delivery / verification 合同
- 给出最终 role 和 execution brain

## 3. 编译器职责

编译器必须统一完成：

1. 资源范围归一化
2. 大任务拆分
3. execution mode / brain 解析
4. role 自动补全
5. delivery contract 生成
6. verification contract 生成
7. 风险等级和人工审核要求补齐

## 4. 必须挡住的问题

方案编译阶段必须前置挡住：

1. 范围明显过大的任务
2. 资源范围不精确的任务
3. 当前环境不可用的执行方式
4. 没有交付定义的任务
5. 没有验证出口的任务
6. 明显会造成 rework 膨胀的任务

## 5. 输出要求

`CompiledPlan` 中每个任务至少包含：

1. `name`
2. `role_type`
3. `brain_kind`
4. `affected_resources`
5. `delivery_contract`
6. `verification_contract`
7. `risk_level`

## 6. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3计划数据模型与表结构设计](./EasyMVP-V3计划数据模型与表结构设计.md)
2. `PlanDraft` 字段设计
3. `PlanReviewResult` schema
4. `CompiledPlan` schema
5. compiler pipeline
6. rework 与 repair draft 并轨设计
