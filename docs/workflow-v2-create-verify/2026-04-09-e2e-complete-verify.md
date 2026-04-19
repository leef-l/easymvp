# 2026-04-09 Workflow V2 全链路完成验证

## 1. 验证目标

在真实后端服务上验证 Workflow V2 从 `create-project` 开始，经过 `design -> review -> execute -> accept -> complete`，直到项目状态进入 `completed` 的整条链路；同时记录真实问题、完成修复并进行复测。

## 2. 验证环境

- 日期：2026-04-09
- 服务：`admin-go/app/mvp/main.go`
- 接口前缀：`http://127.0.0.1:9002/api/mvp`
- 验证方式：直接调用后端 API，不经过 `web-antd`
- 工作目录：`/www/wwwroot/project/easymvp/test-workspaces/workflow-v2-e2e-verify/repo`
- 项目类型：`software_dev`

## 3. 验证策略

为了在当前服务器上验证完整主链，同时避免依赖外部 AI 执行器，本次采用受控方式推进：

1. 通过 `POST /workflow/create-project` 创建真实项目
2. 向项目会话写入一条符合 `parse-tasks` 期望格式的架构师 JSON 回复
3. 调用 `parse-tasks` 与 `confirm-plan` 进入 `review`
4. 调用 `manual-approve` 进入 `execute`
5. 将项目执行模式切换为仅用于验证的 `disabled_for_e2e`
6. 用 `skip-task` 推进执行阶段，验证阶段流转、时间线、汇总统计与验收收口
7. 轮询 `project-status / accept-status / completion-summary / timeline`，直到工作流进入 `completed`

这样可以验证 Workflow V2 的阶段编排、状态聚合、时间显示和验收收口，而不触发 `web-antd`、`validate.sh` 或外部执行器负载。

## 4. 首次全链路运行

首次真实运行项目：

- `projectID = 317591057413967872`
- `conversationID = 317591057460105216`
- `workflowRunID = 317591057472688128`

链路推进结果：

- `create-project` 成功
- `parse-tasks` 成功解析 2 个蓝图
- `confirm-plan` 成功进入 `review`
- `manual-approve` 成功进入 `execute`
- `skip-task` 可以推动阶段前进
- `accept` 自动触发
- 工作流最终进入 `complete`

也正是在这次真实运行中，暴露出 4 个接口层问题。

## 5. 发现的问题

### 5.1 `review-status.stageTasks` 时间未做本地化

首次运行样例见 `e2e-initial-review-status-broken.json`：

- `startedAt = 2026-04-09 07:27:14`
- 但同一项目的其他接口返回本地时间 `2026-04-09 15:27:14`

结论：

- `review-status` 里的 `stageTasks` 直接返回了数据库 UTC 原值
- 和 `timeline / project-trace / stage-history` 的时间口径不一致

### 5.2 `skip-task` 完成时间被错误偏移 8 小时

首次运行样例见 `e2e-initial-execution-status-broken.json`：

- 任务 `startedAt = 2026-04-09 15:27:34`
- 同一任务 `completedAt = 2026-04-09 23:27:59`

结论：

- `skip-task` 写入 `completed_at` 时直接用了数据库 `NOW()`
- 之后接口层又按 UTC 数据再次做本地化，导致跳过任务出现二次偏移

### 5.3 `completion-summary` 统计和时间口径异常

首次运行样例见 `e2e-initial-completion-summary-broken.json`：

- `skippedTasks = 0`，但实际两条任务都通过 `skip-task` 收口
- `avgTaskDuration = 8h0m`
- `startedAt = 2026-04-09T07:26:49+08:00`
- `finishedAt = 2026-04-09T07:28:04+08:00`

结论：

- 完成汇总没有单独统计 `result = skipped`
- 工作流起止时间直接按数据库原值拼装，返回给前端时产生错位

### 5.4 时间线里验收阶段缺少明确标签

首次运行中，`accept` 阶段启动事件的文案只是通用的“阶段已启动”，无法和 `review / execute` 阶段保持一致，不利于前端直接展示。

## 6. 修复内容

### 6.1 审核阶段时间统一走归一化

文件：

- `admin-go/app/mvp/internal/controller/chat/workflow_review.go`

处理：

- 新增统一构造函数，给 `ReviewStatus` 和 `ReviewIssues` 的时间字段都走 `normalizeDBUTCGTime`

### 6.2 `skip-task` 改为写入应用层当前时间

文件：

- `admin-go/app/mvp/internal/controller/chat/workflow.go`

处理：

- 把 `completed_at / updated_at` 从 `gdb.Raw("NOW()")` 改为 `gtime.Now()`
- 避免数据库本地时间再被接口层按 UTC 二次转换

### 6.3 完成汇总补齐 `skipped` 统计与时间归一化

文件：

- `admin-go/app/mvp/internal/workflow/stage/complete/service.go`

处理：

- 补 `skippedTasks` 聚合
- 统一归一化 `started_at / finished_at`
- 修正平均耗时口径

### 6.4 验收阶段时间线标签补全

文件：

- `admin-go/app/mvp/internal/controller/chat/workflow_timeline.go`

处理：

- 为 `accept` 阶段补齐 `验收阶段已启动` 标签

### 6.5 回归测试

文件：

- `admin-go/app/mvp/internal/controller/chat/workflow_time_test.go`
- `admin-go/app/mvp/internal/workflow/stage/complete/service_test.go`

处理：

- 补 `review-status` 时间归一化测试
- 补验收阶段标签测试
- 补完成汇总 `skipped` 统计与时间归一化测试

## 7. 修复后复测

修复后重新完整跑了一次链路，复测项目：

- `projectID = 317593202297147392`
- `conversationID = 317593202334896128`
- `workflowRunID = 317593202347479040`

关键复测结果：

- `e2e-rerun-review-status-fixed.json`
  - `stageTasks.startedAt = 2026-04-09 15:35:20`
- `e2e-rerun-execution-status-fixed.json`
  - `skip-task` 后 `completedAt = 2026-04-09 15:35:22`
- `e2e-rerun-timeline-final.json`
  - 时间线文案已显示 `验收阶段已启动`
- `e2e-rerun-accept-status-completed.json`
  - `status = completed`
  - `decision = passed`
  - `score = 91`
- `e2e-rerun-completion-summary-completed.json`
  - `skippedTasks = 2`
  - `avgTaskDuration = 1s`
  - `startedAt = 2026-04-09T15:35:20+08:00`
  - `finishedAt = 2026-04-09T15:35:32+08:00`

最终 `project-status` 返回：

- `workflowStatus = completed`
- `currentStage = complete`
- `progressPercent = 100`

说明这次修复已经通过真实主链复测，而不是仅靠单元测试。

## 8. 验证边界

本次没有执行以下内容：

- `web-antd` 编译或运行
- `test-workspaces/validate.sh`
- guard 脚本
