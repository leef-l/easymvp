# EasyMVP V3 brain serve 接口接入与 Run 生命周期映射

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3专精大脑接入计划](./EasyMVP-V3专精大脑接入计划.md)
> 关联文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：定义 EasyMVP 如何通过 `brain serve` 接入 `brain-v3`，并把 Run 生命周期稳定映射到 V3 工作流对象和工作台状态。

## 1. 设计结论

V3 首版不应把 `brain-v3` 当 shell 执行器接入。

正确接法是：

1. EasyMVP 调用 `brain serve`
2. 使用 HTTP Run API 获取 `run_id`
3. 持续同步 `status / logs / replay`
4. 把 Run 生命周期映射到 `DomainTask` 和工作台事件流

## 2. 对接边界

### 2.1 EasyMVP 负责

1. 业务对象
2. 状态机
3. 任务合同
4. 人工动作
5. 页面聚合接口

### 2.2 brain-v3 负责

1. Run 生命周期
2. 多脑调度
3. logs / replay
4. tool / sandbox / workdir policy

## 3. 建议接入对象

建议在 EasyMVP 内新增 `BrainRunBinding` 概念。

它用于绑定：

1. `project_id`
2. `domain_task_id`
3. `compiled_task_id`
4. `brain_kind`
5. `run_id`
6. `status`

## 4. Run 生命周期映射

### 4.1 brain-v3 原始状态

第一版按最小集合考虑：

1. `queued`
2. `running`
3. `succeeded`
4. `failed`
5. `cancelled`

### 4.2 EasyMVP 映射状态

建议映射为：

1. `run_pending`
2. `run_active`
3. `run_succeeded`
4. `run_failed`
5. `run_cancelled`

### 4.3 DomainTask 状态关系

必须注意：

1. `run_succeeded` 不等于 `completed`
2. `completed` 需要叠加 delivery / verification 结果

## 5. 建议接口流程

### 5.1 创建 Run

流程：

1. EasyMVP 准备 `brain_kind`
2. EasyMVP 准备 prompt / input
3. 调用 `brain serve` 创建 run
4. 记录 `run_id`
5. 建立 `BrainRunBinding`

### 5.2 查询状态

流程：

1. 根据 `run_id` 拉取状态
2. 同步 `status`
3. 生成 `LiveEvent`
4. 更新 `WorkspaceView.active_runs`

### 5.3 拉取日志

流程：

1. 根据 `run_id` 拉取 logs
2. 归一化成结构化事件
3. 将高价值事件推送到 `LiveEvent`

### 5.4 取消或恢复

流程：

1. EasyMVP 发动作接口
2. 调用 `brain serve cancel / resume`
3. 更新本地绑定状态

## 6. 建议本地表

建议至少有：

1. `workflow_brain_run_bindings`
2. `workflow_brain_run_events`

### 6.1 `workflow_brain_run_bindings`

建议核心列：

1. `id`
2. `project_id`
3. `domain_task_id`
4. `compiled_task_id`
5. `brain_kind`
6. `run_id`
7. `status`
8. `started_at`
9. `ended_at`
10. `last_synced_at`

### 6.2 `workflow_brain_run_events`

建议核心列：

1. `id`
2. `project_id`
3. `run_id`
4. `event_type`
5. `summary`
6. `severity`
7. `payload_json`
8. `created_at`

## 7. 与工作台的映射

这些运行时对象会进入：

1. `WorkspaceView.active_runs`
2. `WorkspaceView.live_events`
3. `ProjectSnapshot.current_run_status`

所以对接层必须保证：

1. `run_id` 稳定
2. `status` 可增量同步
3. logs 可提炼成结构化事件

## 8. 错误处理原则

### 8.1 远端不可用

建议：

1. 标记 `run_sync_failed`
2. 进入 `ActionInbox`
3. 不静默吞掉

### 8.2 logs 拉取失败

建议：

1. 不影响已知状态同步
2. 记录 warning
3. 延迟重试

### 8.3 resume 失败

建议：

1. 生成 blocker
2. 提示人工介入

## 9. 后续细分专题

本专题后续继续拆：

1. `brain serve` 请求 schema
2. 日志归一化规则
3. replay 到工作台的展示映射
4. 多脑并发运行约束
