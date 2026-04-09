# Watchdog 双保险实施计划

## 1. 目标

把当前以 `watchdog` 轮询为主的恢复机制，升级为：

- 事件驱动快路径：任务显式完成、失败、可重试时，秒级触发调度和恢复
- watchdog 兜底慢路径：任务静默卡死、进程崩溃、事件丢失时，仍能检测并恢复

最终形态不是“消息队列替代 watchdog”，而是“消息驱动 + watchdog 双保险”。

## 2. 当前问题

当前 V2 链路的恢复存在几个明确痛点：

1. `watchdog` 检测慢  
   当前默认 `checkInterval=120s`、`maxStaleCount=3`，卡死任务最坏要数分钟后才会被处理。

2. 显式失败也要等轮询  
   任务已经明确失败、完成、可升级时，没有统一的持久化快路径把状态变化立刻交给调度器。

3. 单点退化明显  
   当前依赖本地调度回调和 DB 状态回写；如果执行器回写异常或进程抖动，容易出现“假 running / 假挂起”。

4. Redis 当前未就绪  
   日志里已经有 `NOAUTH Authentication required`，说明 Redis 还不是稳定依赖，因此新方案必须支持“Redis 不可用时自动降级”，不能让主流程被 MQ 绑死。

## 3. 范围

本计划只覆盖工作流执行与恢复链路，不改业务流程定义本身。

涉及代码范围：

- `admin-go/app/mvp/internal/workflow/watchdog/watchdog.go`
- `admin-go/app/mvp/internal/workflow/scheduler/domain_task_scheduler.go`
- `admin-go/app/mvp/internal/workflow/orchestrator/registry.go`
- `admin-go/app/mvp/internal/workflow/event/publisher.go`
- `admin-go/app/mvp/internal/controller/chat/workflow_system_check.go`
- `admin-go/app/mvp/internal/worker/`
- `admin-go/app/mvp/manifest/config/config.yaml`

必要时可新增：

- `admin-go/app/mvp/internal/workflow/eventstream/`
- `admin-go/app/mvp/internal/worker/workflow_event_worker.go`

## 4. 目标架构

### 4.1 快路径

任务状态发生明确变化时，立刻发事件：

- `task.completed`
- `task.failed`
- `task.retry_due`
- `task.escalate_due`
- `scheduler.wakeup`

处理原则：

- 当前进程内先走本地回调，保证最低延迟
- 再写持久化事件
- Redis Stream 可用时再投递到 Stream，由独立 consumer 处理

### 4.2 慢路径

`watchdog` 继续保留，但职责收窄为：

- 检测静默卡死
- 检测心跳过期
- 检测事件丢失后的兜底恢复
- 检测进程崩溃、consumer 异常后的补偿

### 4.3 降级策略

Redis 不可用时：

- 不阻塞主流程
- 不阻塞执行器状态回写
- 自动退化为“本地回调 + watchdog”
- 系统检查明确显示“事件流已降级”

## 5. 设计原则

1. 快路径和慢路径职责分离  
   显式状态变化走快路径，静默异常走慢路径。

2. Redis 不是硬依赖  
   Redis 故障不能让工作流停摆。

3. 幂等优先  
   重复投递、重复消费、重复回调不能造成重复重试或重复升级。

4. 兼容现有流程  
   先增量接入，不重写整个 scheduler。

5. 可观测  
   每一步都要能看到事件是否发出、是否消费、是否降级。

## 6. 实施方案

### Phase 1: 收紧现有 watchdog

目标：在不引入 MQ 的前提下，先把“卡死恢复太慢”的问题显著缓解。

要做的事：

1. 把心跳判断改成 lease 模型  
   不再采用“超时后还要累计多轮 stale count 才判死”的模式。  
   规则改为：超过最后心跳阈值，直接认定失联。

2. 下调默认检查参数  
   建议：
   - `watchdog.check_interval`: `15~30s`
   - `watchdog.max_stale_count`: 仅保留兼容字段，内部不再作为主要判死依据
   - `watchdog.max_retries`: 保持现状或按项目类型调优

3. 修正看门狗失败回调语义  
   让 `watchdog` 失败处理只负责“补偿失败”，不再和执行器本地显式失败处理混用。

4. 增加基础指标/日志  
   至少新增：
   - watchdog 判死耗时
   - watchdog 补偿次数
   - 假 running 修复次数

验收标准：

- 静默卡死从现在的分钟级下降到 `60~90s`
- 不引入新的重复重试

### Phase 2: 补进进程内快路径

目标：任务显式完成/失败时，立刻唤醒调度，不等 watchdog。

要做的事：

1. 收敛任务完成/失败事件入口  
   在 scheduler / executor 完成回写后统一发本地事件。

2. 保留现有回调接口，但加幂等保护  
   至少对这些动作做去重：
   - 同任务重复失败
   - 同任务重复完成
   - 同一重试轮次重复触发

3. 将 `scheduler.wakeup` 标准化  
   避免不同代码路径各自手搓唤醒逻辑。

验收标准：

- 显式失败到重试/返工延迟小于 `3s`
- 完成到下游任务唤醒小于 `2s`

### Phase 3: 引入 Redis Stream 持久化事件流

目标：把快路径从“仅当前进程有效”升级为“跨进程可恢复”。

建议选型：

- Redis Stream
- 使用 Consumer Group
- 不使用普通 Redis List

原因：

- Stream 更适合 ACK
- 支持 pending 重领
- 更适合事件重放与幂等

事件流建议：

- Stream 名：`easymvp:workflow:task-events`
- 事件类型：
  - `task.completed`
  - `task.failed`
  - `task.retry_due`
  - `task.escalate_due`
  - `scheduler.wakeup`

消息字段建议：

- `event_id`
- `event_type`
- `workflow_run_id`
- `task_id`
- `stage_run_id`
- `attempt`
- `created_at`
- `payload_json`

验收标准：

- Redis 正常时，显式状态变化可跨进程秒级传播
- Redis 重启后可继续消费未 ACK 事件

### Phase 4: 独立 consumer 与降级机制

目标：把事件处理做成独立 worker，并保证 Redis 挂了也能退化运行。

要做的事：

1. 新增独立 worker  
   负责消费 Stream 并调用：
   - scheduler 唤醒
   - retry
   - escalate
   - rework 触发

2. 增加 pending reclaim 逻辑  
   consumer 崩溃后，其他 consumer 可重新认领超时消息。

3. 增加降级开关  
   Redis 不可用时：
   - 记录 warning
   - 停止 Stream 投递
   - 回退本地回调 + watchdog

4. 系统检查增加状态项  
   明确显示：
   - Stream 可用 / 已降级
   - consumer lag
   - pending 数量

验收标准：

- Redis 故障时主链不阻塞
- 恢复后 Stream 消费自动恢复

### Phase 5: 观测、灰度、回滚

目标：把双保险接入生产前可控化。

要做的事：

1. 配置开关
   - `workflow.event_stream.enabled`
   - `workflow.event_stream.redis_required`
   - `workflow.event_stream.consumer_enabled`

2. 灰度顺序
   - 先开 Phase 1 watchdog 收紧
   - 再开本地快路径
   - 再开 Stream producer
   - 最后开 Stream consumer

3. 回滚策略
   - 关掉 `event_stream.enabled`
   - 保留本地回调和 watchdog
   - 不需要回滚数据库结构即可恢复旧行为

## 7. 幂等要求

以下动作必须幂等：

- 同一个 `task.failed` 事件重复投递
- 同一个 `task.completed` 事件重复投递
- 同一个 `scheduler.wakeup` 事件重复投递
- consumer 崩溃后的重复认领

建议幂等键：

- `workflow_run_id`
- `task_id`
- `event_type`
- `attempt`

如果需要持久化去重，可新增一张轻量事件消费表，或复用现有事件表加状态字段。

## 8. 配置建议

建议新增配置项：

```yaml
workflow:
  event_stream:
    enabled: false
    stream_name: easymvp:workflow:task-events
    consumer_group: easymvp-workflow
    consumer_name: ${HOSTNAME}
    block_ms: 5000
    reclaim_idle_seconds: 60
    redis_required: false
    consumer_enabled: false

watchdog:
  check_interval: 20
  heartbeat_timeout_seconds: 60
  max_retries: 3
```

## 9. 验收标准

必须满足以下目标才算做完：

1. 显式失败到重试/返工延迟 `< 3s`
2. 任务完成到下游唤醒 `< 2s`
3. 静默卡死检测时间 `<= 90s`
4. Redis 挂掉时主流程仍可继续
5. 不出现重复重试、重复返工、重复升级
6. 系统检查能看到事件流状态
7. 至少有一条真实项目工作流验证通过

## 10. 建议的 Agent 执行拆分

如果要并行做，建议拆成三个独立任务：

### Agent A：Watchdog 收紧

负责：

- `watchdog.go`
- `config.yaml`
- `workflow_system_check.go`
- 对应单测

交付：

- lease 判死模型
- 新配置项
- watchdog 指标/日志

### Agent B：事件流基础设施

负责：

- `workflow/event/`
- 新 `eventstream/`
- Redis Stream producer/consumer
- 幂等处理

交付：

- Stream 生产/消费链路
- Redis 降级逻辑
- 对应单测

### Agent C：调度接线与灰度

负责：

- `domain_task_scheduler.go`
- `registry.go`
- 回调接线
- 灰度开关

交付：

- 本地快路径统一入口
- event_stream 与 scheduler 集成
- 一条真实工作流回放验证

## 11. 实施顺序建议

推荐顺序：

1. 先做 Phase 1  
   这是最小风险、最快见效的部分。

2. 再做 Phase 2  
   让显式失败/完成先变快。

3. 最后做 Phase 3-4  
   引入 Redis Stream 和独立 worker。

不要一上来同时重写 watchdog 和 scheduler 主流程，风险太大。

## 12. 明确不做的事

当前这版计划不包含：

- 用 MQ 替代整个 scheduler
- 引入 Kafka / RabbitMQ 等新基础设施
- 把 Redis 设为强依赖
- 改写工作流业务阶段语义

## 13. 交付定义

当以下条件同时满足时，可判定“watchdog 双保险计划完成”：

- 代码合并
- 配置上线
- 系统检查可见
- 单测/集成测试通过
- 至少一条真实项目从失败/完成到恢复/唤醒的链路验证通过
- Redis 不可用时验证过降级生效
