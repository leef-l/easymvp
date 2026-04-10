# Watchdog 双保险实施计划

## 1. 当前结论

截至 `2026-04-10`，Watchdog 双保险不是“从零开始的待设计项”，而是已经完成主体落地，并已完成专项验收，当前进入回归维护阶段。

当前结论分三层：

- 已落地：`watchdog` lease 判死、失败快路径、事件持久化、Redis Stream consumer、pending reclaim、Redis 降级、system-check 运行态观测、服务启动恢复
- 已补齐：`task.completed` completion 恢复闭环
- 已补齐：真实 workflow 回放验收

换句话说，当前不该再把这份文档写成“要不要做双保险”，而应该写成“已经做到哪一步，以及如何保持回归稳定”。

## 2. 当前代码基线

### 2.1 Watchdog 慢路径：已落地

已落地代码：

- `admin-go/app/mvp/internal/workflow/watchdog/watchdog.go`
- `admin-go/app/mvp/internal/workflow/watchdog/watchdog_test.go`
- `admin-go/app/mvp/internal/engine/config.go`
- `admin-go/app/mvp/internal/controller/chat/workflow_system_check.go`

当前行为：

- `watchdog.New()` 读取 `watchdog.check_interval`、`watchdog.heartbeat_timeout_seconds`、`watchdog.max_stale_count`、`watchdog.max_retries`
- `checkRunningTasks()` 已按 `heartbeat_at / started_at + lease timeout` 判死，`max_stale_count` 只保留为兼容日志语义，不再作为主判定条件
- `checkFailedTasks()` 会扫描活跃工作流中的 `failed` 任务并执行 `retry / escalate`
- `Snapshot()` 已暴露运行态，`workflow_system_check.go` 已展示 `check / lease / max_retries / timeout / retry / escalate`
- `watchdog_test.go` 已覆盖 lease timeout、时间解析、snapshot 等基础逻辑

这意味着 Track C 的“Phase 1 收紧”主体已经落库，不再是待办。

### 2.2 进程内快路径：已落地大部分

已落地代码：

- `admin-go/app/mvp/internal/workflow/scheduler/domain_task_scheduler.go`
- `admin-go/app/mvp/internal/workflow/orchestrator/registry.go`
- `admin-go/app/mvp/internal/workflow/orchestrator/event_wiring.go`
- `admin-go/app/mvp/internal/workflow/event/publisher.go`
- `admin-go/app/mvp/internal/workflow/event/metadata.go`

当前行为：

- `OnTaskFailed()` 会先发布 `task.failed`
- `registry.go` 中的 `SetFailureCallback()` 会把显式失败立刻转换成 `task.retry_due` 或 `task.escalate_due`
- `handleTaskRetryDueEvent()` 会把任务从 `failed -> pending`，再发 `task.retried` 和 `scheduler.wakeup`
- `handleTaskEscalateDueEvent()` 会把任务从 `failed -> escalated`，再触发 `rework`
- `event.EnsureMetadata()` 已统一生成 `event_id / created_at / attempt / idempotency_key`
- 当前 recovery handler 真实覆盖的事件为：`scheduler.wakeup / task.completed / task.retry_due / task.escalate_due`

这意味着显式失败路径已经不再单纯依赖 watchdog 轮询，快路径已存在。

### 2.3 Redis Stream 与降级：主体已落地

已落地代码：

- `admin-go/app/mvp/internal/workflow/eventstream/config.go`
- `admin-go/app/mvp/internal/workflow/eventstream/bridge.go`
- `admin-go/app/mvp/internal/workflow/eventstream/consumer.go`
- `admin-go/app/mvp/internal/workflow/eventstream/bridge_test.go`
- `admin-go/app/mvp/internal/workflow/eventstream/consumer_test.go`
- `admin-go/app/mvp/internal/worker/workflow_event_worker.go`
- `admin-go/app/mvp/internal/cmd/cmd.go`

当前行为：

- producer 已通过 `Bridge.Publish()` 将事件写入 Redis Stream
- consumer 已实现 `XGROUP CREATE`、`XREADGROUP`、`XPENDING`、`XCLAIM`、`XACK`
- `pending reclaim` 已具备基础实现
- `safeGetWorkflowEventRedis()` 会在 Redis 不可用或认证失败时自动退化
- `workflow_event_worker` 已存在，且在 `consumer_enabled` 打开、consumer 成功创建时，`cmd.go` 会实际拉起 consumer
- `workflow.event_stream.redis_required=false` 时，Redis 故障不会阻塞主流程

这意味着 Track D2 的基础设施不是“待新增目录”，而是已经可运行。

### 2.4 启动恢复：已落地，闭环已补齐

已落地代码：

- `admin-go/app/mvp/internal/workflow/orchestrator/recovery.go`
- `admin-go/app/mvp/internal/workflow/runtime/runtime.go`
- `admin-go/app/mvp/internal/cmd/cmd.go`

当前行为：

- 服务启动会执行 `orchestrator.Init()`
- 随后执行 `RecoverActiveWorkflows()`
- 对 `execute / rework` 中的活跃 workflow 会重新绑定执行器并重启 scheduler
- 如事件流 consumer 已配置启用，`cmd.go` 也会启动 `workflow_event_worker`

这已经满足“重启后恢复运行态”的基础要求；而“重启后最后一步阶段推进”的闭环也已在本轮补齐。

## 3. 本轮收口结果与剩余缺口

### 3.1 已补齐：`task.completed` 跨进程恢复闭环

截至 `2026-04-10`，这个 P0 缺口已经补齐。

本轮已落地：

- `OnTaskCompleted()` 会发布 `task.completed`
- 随后会经共享 `ReconcileWorkflowProgress()` 统一执行 `scheduleOnce()` 和 `checkAllDone()`
- `event_wiring.go` 已将 `task.completed` 纳入 recovery handler，consumer 恢复后会走同一条 reconcile 链
- `RecoverActiveWorkflows()` 在恢复 execute 阶段时会先做完成态判断；若已全部完成则直接补 reconcile 并跳过 scheduler 重启，否则在启动后再补一次 reconcile
- workflow 恢复回调在 execute 场景也会补一次 reconcile
- `domain_task_scheduler_test.go`、`event_wiring_test.go` 与 `service_test.go` 已补最小单测覆盖 reconcile helper、completion recovery dedupe 和 execute completion plan 关键分支

这意味着原先这个 crash window：

1. 任务已经从 `running -> completed`
2. 进程在执行 `checkAllDone()` 之前崩溃
3. 服务重启后只会恢复调度轮询
4. 如果此时该 workflow 已无 `pending/running/failed` 任务，scheduler 会持续空轮询
5. execute 阶段可能停留在“任务都完成了，但阶段没推进”的卡态

已经被启动恢复链补住，不再是当前未闭环项。

### 3.2 已补齐：system-check 已展示 lag / pending / worker 健康

本轮新增：

- `Consumer.Snapshot()` 会暴露 `pending / lag / reclaim / last_consume / last_ack / worker heartbeat / started_at`
- `workflow_system_check.go` 已将 bridge 健康与 consumer 运行态拆开显示
- `workflow_system_check_test.go` 已新增 consumer 摘要单测
- `consumer_test.go` 已覆盖 snapshot / heartbeat 行为

当前 system-check 已能直接区分：

- “stream 已启用但 Redis 降级”
- “consumer 已创建但未启动”
- “consumer 已启动且正在消费”

### 3.3 已补齐：durable idempotency 已落地

截至 `2026-04-10`，这部分已经完成从“内存字段”到“DB metadata + durable ledger”的收口。

本轮已完成：

- `event_id`
- `idempotency_key`
- `attempt`
- `Publisher.Emit()` 已把这些字段写入 `mvp_workflow_event`
- `000011_workflow_event_durable_idempotency` 已补 `mvp_workflow_event` 新列并创建 `mvp_workflow_event_ledger`
- `docker/mysql` 与 `codegen/sql` 的 `init/schema` 快照已同步到 durable ledger 版本，避免新环境初始化与增量 migration 漂移
- `workflow.publish` 与 `workflow.recovery.*` 已共用 durable ledger
- `workflow_system_check.go` 已新增 `Workflow 事件幂等账本` 检查项
- 若 migration 未执行，publisher / recovery handler 会告警并回退到非 durable 模式

结论：

- durable 幂等与跨进程消费账本已经是现有代码能力
- 剩余工作转为 migration rollout、集成测试和真实环境验收

### 3.4 P1：事件恢复链缺少针对崩溃场景的测试

现有测试主要覆盖：

- `watchdog` 时间与 snapshot
- `eventstream` 的 `XADD / XREADGROUP / XPENDING / XCLAIM / XACK`
- `domain_task_scheduler` 的 reconcile helper
- `event_wiring` 的 `task.completed` 去重单测
- `recovery` 的 execute 启动 reconcile / 跳过重启单测
- `eventstream.Consumer` 的 runtime snapshot / heartbeat
- `workflow_system_check` 的 consumer 摘要

但仍缺：

- `handleTaskRetryDueEvent()` 幂等测试
- `handleTaskEscalateDueEvent()` 幂等测试
- “任务已 completed，但进程在 `checkAllDone()` 前崩溃”的集成恢复测试
- “Redis 不可用 -> 降级 -> 恢复后继续消费”的集成测试

这部分不补，双保险只能算“有实现”，还不能算“已验证”。

### 3.5 已补齐：真实 Redis 与真实 workflow 回放验收

截至 `2026-04-10 18:40:43`，这部分已经完成验收。

真实回放记录见：

- `docs/workflow-v2-create-verify/2026-04-10-real-replay-acceptance.md`

本次真实验收已确认：

- Redis event stream consumer 真实启动并消费
- 至少一条真实 workflow run 已完成到 `completed`
- 时间线中真实出现 `task.retry_due / task.escalate_due / rework`
- 人工接管口 `force-stage accept`、`accept-approve` 可用

## 4. 剩余收口计划

### Track W1：补齐完成态恢复闭环

状态：`已完成`

本轮已完成：

- 引入共享的 `ReconcileWorkflowProgress(workflowRunID)`
- `OnTaskCompleted()` 与 `RecoverActiveWorkflows()` 共用这条 reconcile 逻辑
- execute 恢复场景下，如果已全部完成则直接跳过 scheduler 重启
- 已补 `reconcile helper`、`event_wiring`、`recovery` 单测

验收标准：

- 最后一个任务完成后进程崩溃，重启后 workflow 仍能推进到下一阶段
- 不出现重复推进、重复完成阶段

代码落点：

- `admin-go/app/mvp/internal/workflow/scheduler/domain_task_scheduler.go`
  入口函数：`OnTaskCompleted()`
- `admin-go/app/mvp/internal/workflow/orchestrator/recovery.go`
  入口函数：`RecoverActiveWorkflows()`
- `admin-go/app/mvp/internal/workflow/orchestrator/event_wiring.go`
  入口函数：`task.completed` 对应 recovery handler

本轮补齐：

1. completion crash recovery 的集成级演练
2. Redis 不可用 / 恢复 / pending reclaim 的真实 Redis 集成验证
3. 重复投递 / 重复消费的 durable 去重集成验证

### Track W2：补齐运行态观测

状态：`已完成`

本轮已完成：

- stream pending 数
- stream reclaim 数
- consumer 最近 ack 时间
- consumer 最近成功消费时间
- worker 存活心跳

验收标准：

- system-check 可直接看出“启用了但没消费”与“消费正常”
- Redis 挂掉、恢复、pending 积压时都能被看见

### Track W3：补完验证与灰度

目标：从“实现存在”升级到“已验证可依赖”。

状态：`已完成`

本轮已完成：

- 真实 workflow 回放验收
- 真实 Redis consumer 运行态验收
- 专项文档回填

## 5. 事件覆盖矩阵

| 事件 | 当前是否进入 recovery handler | 当前用途 |
| --- | --- | --- |
| `scheduler.wakeup` | 是 | 唤醒调度扫描 |
| `task.retry_due` | 是 | 失败任务转回 `pending` 并重试 |
| `task.escalate_due` | 是 | 失败任务升级并触发 rework |
| `task.failed` | 否 | 记录失败，由本地失败回调继续派生恢复动作 |
| `task.completed` | 是 | 完成后触发 shared reconcile，补偿阶段推进 |

这张矩阵要明确保留，避免把“事件已落地”误读成“所有事件都已纳入跨进程恢复闭环”。

## 6. 当前状态判定

按当前事实，双保险状态应定义为：

- `watchdog` Phase 1：已完成
- 显式失败快路径：已完成
- Redis Stream producer / consumer / reclaim：已完成主体
- Redis 降级：已完成主体
- system-check 运行态观测：已完成
- 最后一个任务完成后的跨进程恢复：已完成
- 运行态监控面：已完成
- 真实环境验收：已完成（Redis / recovery 集成与真实 workflow 回放均已完成）

因此，当前最准确的描述不是“Watchdog 双保险尚未开始”，而是：

`Watchdog 双保险主干已落地，completion 恢复闭环、stream 运行态观测、durable idempotency、Redis 降级/恢复/reclaim 集成验证与真实 workflow 回放验收均已补齐。`

## 7. 验证矩阵

| 场景 | 触发方式 | 预期结果 | 通过信号 |
| --- | --- | --- | --- |
| 显式失败快路径 | 执行器直接失败 | 秒级进入 `retry / rework` | 状态迁移与事件日志一致 |
| 静默卡死慢路径 | 停止心跳但不写失败 | `<= 90s` 内被 watchdog 判死 | watchdog snapshot 与任务状态一致 |
| completion crash recovery | 最后一个任务完成后立刻杀进程 | 重启后仍推进 execute/rework 后续阶段 | 不出现空轮询卡态 |
| Redis 降级 | Redis 不可用 | 主流程不阻塞，降级为“本地快路径 + watchdog” | system-check 显示 degraded |
| Redis 恢复 | Redis 恢复可用 | pending/backlog 可继续消费 | system-check 恢复 healthy，pending 下降 |
| consumer reclaim | kill 当前 consumer | pending 被其他 consumer reclaim | 不出现重复重试、重复升级 |

## 8. 交付物定义

每个 Track W1/W2/W3 完成时，至少要有：

- 代码改动
- 对应单测或集成测试
- system-check 或日志可观测项
- 一条回放/演练记录
- 文档回填

没有这些交付物，就只能算“实现了一部分”，不能算“专项完成”。

## 9. 完成判定

当且仅当以下条件全部满足时，才可判定“Watchdog 双保险实施计划完成”：

1. 显式失败到 `retry / rework` 仍保持秒级响应
2. 静默卡死检测时间稳定在 `<= 90s`
3. 最后一个任务完成后即使进程崩溃，重启后仍能自动推进阶段
4. Redis 不可用时主流程继续运行
5. Redis 恢复后 pending 事件可继续消费
6. system-check 能显示 stream 健康、pending、consumer 存活信息
7. 相关单测与集成测试补齐
8. 至少一条真实 workflow 回放验证通过
