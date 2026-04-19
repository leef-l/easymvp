# EasyMVP V3 实时事件流推送机制设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 关联文档：[EasyMVP-V3-brain-serve接口接入与Run生命周期映射](./EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md)
> 目标：定义 Workspace 所需的实时事件流推送机制、事件归一化规则和刷新优先级。

## 1. 设计结论

Workspace 的实时感必须来自结构化事件流，而不是轮询大对象。

## 2. 事件来源

1. `PlanReviewResult`
2. `CompiledPlan`
3. `workflow_brain_run_events`
4. `AcceptanceRun`
5. `ActionInbox`

## 3. 事件归一化

每条事件建议包含：

1. `event_id`
2. `event_type`
3. `summary`
4. `severity`
5. `created_at`
6. `source_object_kind`
7. `source_object_id`

## 4. 推送机制

第一版建议：

1. SSE 优先
2. 短轮询兜底

## 5. 优先级

高优先级事件：

1. blocker
2. manual release
3. run failed

普通优先级事件：

1. task started
2. task succeeded
3. plan compiled

## 6. 后续细分专题

本专题后续继续拆：

1. SSE payload schema
2. reconnect 策略
3. backpressure 策略
