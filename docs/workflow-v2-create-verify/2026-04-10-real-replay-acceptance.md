# 2026-04-10 Workflow V2 真实回放验收

> 历史说明：本文是 2026-04-10 的专项留档，文中真实服务回放、API 推进与本机环境描述只代表当时的验收证据，不代表当前仓库允许的正式验证入口。现行测试与编译验收统一只认 GitHub Actions。

## 1. 验收目标

在真实 `system + mvp + MySQL + Redis` 服务上，完成至少一条 Workflow V2 项目链路回放，并确认：

- `create-project -> design -> review -> execute -> accept -> complete` 能真实走通
- Redis event stream consumer 真实启动并消费
- 显式失败快路径、`retry_due / escalate_due / rework`、人工接管口都能在真实运行数据里看到
- 项目最终可收敛到 `completed`

## 2. 验收环境

- 日期：`2026-04-10`
- system 服务：`http://127.0.0.1:9000`
- mvp 服务：`http://127.0.0.1:9002`
- Redis：本机真实 Redis，consumer=`codex-replay`
- 工作目录：`/www/wwwroot/project/easymvp/test-workspaces/workflow-v2-replay-20260410-183400/repo`

本次真实项目：

- `projectID = 318000616905379840`
- `conversationID = 318000616993460224`
- `workflowRunID = 318000617047986176`

## 3. 推进方式

本次仍采用受控推进，避免把结果绑定到外部实现执行器：

1. 创建真实 Workflow V2 项目
2. 向架构师对话注入可解析的 JSON 任务清单
3. 调用 `parse-tasks`、`confirm-plan`、`manual-approve`
4. 观察真实 `execute -> accept -> complete`
5. 用真实 API 回放人工接管、`force-stage`、`accept-approve`
6. 用 `project-status / accept-status / completion-summary / stage-history / timeline / system-check` 留证

## 4. 真实运行结果

### 4.1 主链已真实完成

第一次真实推进中，这条 workflow 已在时间线里完成：

- `design` 完成：`2026-04-10 18:35:01`
- `review` 完成：`2026-04-10 18:35:10`
- `execute` 完成：`2026-04-10 18:36:21`
- `accept` 完成：`2026-04-10 18:36:25`
- `complete` 完成：`2026-04-10 18:36:25`

也就是说，单就“主链能否在真实服务上走通”这件事，本次验收已经成立。

### 4.2 回放中额外打到了失败与返工链

在第一次完成后，我继续对已完成任务做了受控人工接管，导致该 run 被重新拉回：

- `execute` 强制重启
- `task.retry_due`
- `task.escalate_due`
- `rework`

这不是新的主链缺陷，而是一次真实的“已完成后再次人工干预”场景。它反而补充验证了：

- 显式失败会秒级进入 `retry_due / escalate_due`
- `rework` 会真实生成失败分析任务
- `force-stage accept`
- `accept-approve`

这些人工与恢复入口在真实服务上都能工作。

### 4.3 最终收敛结果

最终通过 `force-stage accept` + `accept-approve` 收口，并清理掉残留的 pending 失败分析任务后，最终结果为：

- `project-status`
  - `workflowStatus = completed`
  - `currentStage = complete`
  - `progressPercent = 100`
- `accept-status`
  - `acceptRound = 3`
  - `status = completed`
  - `decision = passed`
  - `score = 100`
- `completion-summary`
  - `totalTasks = 4`
  - `completedTasks = 4`
  - `failedTasks = 0`
  - `skippedTasks = 3`
  - `successRate = 1`
  - `finishedAt = 2026-04-10T18:40:43+08:00`

## 5. Redis 与运行态证据

最终 `system-check` 返回：

- `workflow_event_stream = ok`
- `workflow_event_consumer = ok`
- `consumer_started = true`
- `pending = 0`
- `lag = 0`
- `last_consume = 2026-04-10 18:40:43`
- `last_ack = 2026-04-10 18:40:43`

说明这次回放不是只靠本地内存快路径，Redis event stream consumer 也处于真实运行态。

## 6. 验收结论

这次回放完成后，可以把“真实 workflow 回放验收仍待补齐”从专项结论中删除。

理由很直接：

- 至少一条真实 Workflow V2 项目链路已经走通到 `completed`
- 真实 Redis consumer 在回放期间处于健康消费状态
- 时间线里真实出现了 `task.retry_due / task.escalate_due / rework / accept / complete`
- 人工接管和强制收敛入口也经过了真实数据验证

从专项收口口径看，剩下的不再是“机制未闭环”，而是后续常规回归维护。
