# EasyMVP V3 方案编译模型设计

> 更新时间：2026-04-19  
> 上游文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)  
> 目标：把“架构师方案”升级为“可审核、可编译、可执行的正式对象”。

## 1. 设计结论

V3 不再允许架构师输出直接落成 `DomainTask`。

必须先经过：

`PlanDraft -> PlanReviewResult -> CompiledPlan`

只有 `CompiledPlan` 才能投影生成正式任务。

按当前钱学森总纲口径，这里的“正式任务”应进一步理解为：

1. EasyMVP 闭环里的 `CompiledTask`，而不是仅停留在旧表述里的宽泛 `DomainTask`
2. `CompiledTask` 必须在编译阶段就补齐 `brain_kind / delivery_contract_json / verification_contract_json`
3. 编译阶段必须为后续 `executing -> accepting -> reworking -> completed` 提供可闭环的结构化起点

如果本文件与以下文档发生冲突，统一以后者为准：

1. [easymvp-brain-输入输出契约](./钱学森总纲设计/easymvp-brain-输入输出契约.md)
2. [EasyMVP-对象级字段清单](./钱学森总纲设计/EasyMVP-对象级字段清单.md)
3. [EasyMVP-闭环状态机补充说明](./钱学森总纲设计/EasyMVP-闭环状态机补充说明.md)

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

补充说明：

- 在当前 EasyMVP 口径下，`CompiledPlan` 的核心职责不是“把草案排成任务列表”这么简单，而是输出一组可直接执行、可直接验证、可直接裁决的 `CompiledTask`
- 因此它的价值不只在“compile 成功”，还在于后续任何 `CompletionVerdict / RepairPlanDraft / RuntimeEscalation` 都必须能回溯到编译结果

## 3. 编译器职责

编译器必须统一完成：

1. 资源范围归一化
2. 大任务拆分
3. execution mode / brain 解析
4. role 自动补全
5. delivery contract 生成
6. verification contract 生成
7. 风险等级和人工审核要求补齐

按当前总纲，需要再补 4 个约束：

8. 明确任务是否需要 `manual_review_required`
9. 明确推荐验证通道与可回退通道，不能把验证环境留给执行时临时猜测
10. 对无法进入闭环的任务直接挡在编译阶段，而不是推给执行阶段兜底
11. 生成足以支撑页面展示和最终裁决的最小对象骨架

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

按当前 EasyMVP 对象级基线，建议把这里进一步收口为：

1. `compiled_task_id`
2. `name`
3. `role_type`
4. `brain_kind`
5. `risk_level`
6. `delivery_contract_json`
7. `verification_contract_json`
8. `manual_review_required`
9. `resolve_trace_json`
10. `depends_on_task_ids`

原因很直接：

- `role_type` 解决“谁来做”
- `brain_kind` 解决“走哪条脑路由”
- `delivery_contract_json` 解决“交付什么”
- `verification_contract_json` 解决“怎么验”
- `manual_review_required` 解决“哪里必须停住”

如果缺这些字段，后面的执行、验收、返工都会退化为自然语言兜底。

## 6. 与闭环状态机的关系

方案编译不是孤立步骤，它是闭环状态机的第一道硬闸门。

编译结果至少要支持后续 4 类推进判断：

1. 是否允许从 `reviewing` 进入 `executing`
2. 执行完成后如何进入 `accepting`
3. 验证失败后如何进入 `reworking`
4. 最终 `completed` 是否有足够对象链路可回溯

所以这里的正确理解应当是：

- 编译阶段提前消灭不明确性
- 执行阶段处理真实产出
- 验收阶段处理验证与裁决
- 返工阶段处理合同修正与任务重编译

## 7. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3计划数据模型与表结构设计](./EasyMVP-V3计划数据模型与表结构设计.md)
2. `PlanDraft` 字段设计
3. `PlanReviewResult` schema
4. `CompiledPlan` schema
5. compiler pipeline
6. rework 与 repair draft 并轨设计
