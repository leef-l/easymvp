# 2026-04-09 Workflow V2 贪吃蛇项目重建与全链路验证

## 1. 验证目标

在当前服务器上重新创建一个 React + GoFrame v2 的贪吃蛇样例项目，完成项目级代码验证，并把 Workflow V2 从 `create-project` 推进到 `design -> review -> execute -> accept -> complete`，同时记录过程中出现的问题与处理方式。

## 2. 验证对象

- 样例项目目录：`/www/wwwroot/project/easymvp/test-workspaces/workflow-v2-snake-verify/repo`
- Workflow V2 项目名：`workflow-v2-snake-verify-20260409-recreate`
- `projectID = 317610689273270272`
- `conversationID = 317610689323601920`
- `workflowRunID = 317610689336184832`
- 接口地址：`http://127.0.0.1:9002/api/mvp`
- 验证日期：2026-04-09

## 3. 最终结果

- 样例项目已完成并保留在目标工作目录
- Workflow V2 已完成
- `project-status` 返回 `workflowStatus=completed`
- `project-status` 返回 `currentStage=complete`
- `accept-status` 返回 `decision=passed`、`score=91`
- `completion-summary` 返回 `totalTasks=7`、`completedTasks=7`、`skippedTasks=7`

对应材料：

- `project-status-completed.json`
- `accept-status-completed.json`
- `completion-summary-completed.json`
- `stage-history-completed.json`
- `timeline-completed.json`

## 4. 样例项目实现摘要

前端：

- 使用 React + TypeScript 构建贪吃蛇界面
- 提供键盘与按钮控制
- 集成后端健康检查、配置和排行榜接口

后端：

- 使用 GoFrame v2 提供 HTTP API
- 提供 `GET /api/health`
- 提供 `GET /api/game/config`
- 提供 `GET /api/game/scores`
- 提供 `POST /api/game/scores`
- 使用文件方式持久化排行榜

## 5. Workflow V2 推进方式

本次执行链不是依赖外部 AI 执行器完成代码生成，而是采用受控推进方式验证工作流主链：

1. 创建 Workflow V2 项目
2. 向会话注入可解析的架构师任务 JSON
3. 调用 `parse-tasks`
4. 调用 `confirm-plan`
5. 通过 `manual-approve` 进入 `execute`
6. 在执行阶段把任务切到 `disabled_for_e2e` 并通过 `update-domain-task + skip-task` 收口
7. 自动进入 `accept`
8. 自动进入 `complete`

这样验证的是阶段编排、状态流转、验收收口和轨迹接口，而不是外部执行器本身。

## 6. 代码验证结果

本次重新执行的验证项：

- 后端：`go test ./...` 通过
- 前端：`npm run test -- --run` 通过
- 前端：`npm run lint` 通过
- 前端：`npm run build` 通过
- 运行态：后端健康接口通过
- 运行态：后端配置接口通过
- 运行态：前端代理访问后端配置接口通过
- 运行态：排行榜提交流程通过

对应材料：

- `verification-status.json`
- `backend-health.json`
- `backend-config.json`
- `backend-score-submit.json`
- `backend-scores.json`
- `frontend-proxy-config.json`

## 7. 发现的问题与处理

### 7.1 React CLI 初始化路径调整

问题：

- 传统 `create-react-app` 已不适合作为当前初始化方案

处理：

- 改用 Vite 创建 React + TypeScript 工程
- 在文档中按“React CLI 初始化结果”为 Vite 脚手架记录

### 7.2 GoFrame 默认 hello 模板不适合目标架构

问题：

- `gf init` 生成的 hello 示例结构不适合当前贪吃蛇接口设计

处理：

- 删除 hello 示例控制器与 API
- 重新按 `api / controller / service / logic / model` 结构实现游戏域接口

### 7.3 后端默认端口冲突

问题：

- 机器上的 `:8000` 已被占用

处理：

- 后端改到 `:18080`
- 前端代理同步改到 `http://127.0.0.1:18080`

### 7.4 Workflow V2 在中断后停留在 execute

问题：

- 本次会话中途被打断，工作流停在 `execute`

处理：

- 重新接管运行中的 domain task
- 对运行中的任务使用 `update-domain-task` 重置为 `pending`
- 对受控验证任务使用 `skip-task` 推进阶段
- 最终进入 `accept -> complete`

证据：

- `execution-status-completed.json`
- `timeline-completed.json`
- `stage-history-completed.json`

### 7.5 分数提交请求字段名误用

问题：

- 首次运行态验证时，请求体用了 `player`
- 后端实际要求字段名为 `playerName`

处理：

- 修正验证请求体为 `playerName`
- 重新提交分数并确认排行榜返回正常

证据：

- `backend-score-submit-invalid.json`
- `backend-score-submit.json`
- `backend-scores.json`

### 7.6 Review 阶段存在 3 条 `missing_role` warning

问题：

- 方案中 3 个任务使用了 `implementer/max`
- 创建项目时没有选择对应项目角色预设

结果：

- warning 不阻塞本次受控验证完成
- 但这说明如果后续要做真实自动执行，应在创建项目时补 `implementer/max` 的项目角色配置

证据：

- `review-status-approved.json`
- `review-issues-warning.json`

### 7.7 初始化规划规则过宽，未显式要求优先脚手架

问题：

- 当用户只给出“React CLI + GoFrame v2”这类明确技术栈时，旧版运行时提示词只要求“批次 1 做初始化 / 基础框架”
- 它没有强制架构师优先输出“官方 CLI / 模板仓库 / 文档工具 / 现成仓库”的快速起骨架方案
- 同时，当前编码类项目很多默认 `implementer` 预设仍使用 `auto`，执行器命令模板本身也只是通用外壳，不能单独保证一定走脚手架而不是手写

影响：

- 系统可能生成过宽的根初始化任务
- 初始化阶段容易把前端、后端、文档等无关目录混在一个任务里
- 仅凭旧提示词，无法保证实现者会优先用官方脚手架，而不是手写样板代码

处理：

- 在运行时架构师提示词中追加“脚手架优先”规则：如果用户已明确技术栈、官方 CLI、模板仓库、文档工具或现成仓库，批次 1 必须优先规划对应的骨架初始化任务
- 要求前端、后端、基础设施等独立根目录拆成独立初始化任务并尽量并行执行
- 要求 `affected_resources` 只能写相对路径，禁止混入命令、树形结构或说明文字
- 在运行时实现者提示词中追加补充规则：如果任务描述已明确脚手架或模板来源，优先使用官方工具快速初始化，只有明显不适用时才允许手写骨架

结果：

- 后续同类项目在设计阶段会更明确地区分“快速起骨架”和“业务实现”两类任务
- 初始化任务的并行度会更高，资源边界也更清晰
- “是否优先使用脚手架”会由运行时提示词和任务描述共同约束，而不再完全依赖模型自行猜测

## 8. 验证结论

这次贪吃蛇项目已经完成，且不是只有代码目录存在，而是同时满足下面三件事：

- 样例项目代码已完成并通过重新验证
- Workflow V2 已真实走完到 `completed`
- `docs/workflow-v2-snake-verify/` 已保留本次请求、响应、问题和验证材料

当前仍需注意的一点是：`implementer/max` 角色 warning 只是被记录，没有在这次已完成项目上回头重跑 review；如果后续要把这套样例用于“真实自动执行”验证，应该在创建项目时补齐角色配置，而不是继续依赖受控跳过模式。
