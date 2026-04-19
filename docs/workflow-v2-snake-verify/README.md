# Workflow V2 贪吃蛇验证记录

本目录记录 2026-04-09 对 React + GoFrame v2 贪吃蛇样例项目的重建、项目级验证和 Workflow V2 全链路完成验证材料。

说明：本目录中的命令与“已通过”描述均为 2026-04-09 的历史验证记录。自 2026-04-13 起，仓库现行铁律已经切换为“测试与编译统一只走 GitHub Actions”，因此下文中的本机 `go test` / `npm run test` / `npm run build` 只作为历史证据保留，不再是当前执行入口。

核心结论：

- 样例项目目录：`/www/wwwroot/project/easymvp/test-workspaces/workflow-v2-snake-verify/repo`
- Workflow V2 项目 `317610689273270272` 已完成
- `project-status` 已返回 `workflowStatus=completed`、`currentStage=complete`
- 验收已通过，`score=91`
- 历史代码验证记录：`go test ./...`、`npm run test -- --run`、`npm run lint`、`npm run build`

主要文件：

- `2026-04-09-workflow-v2-snake-verify.md`：本次重建、验证、问题与处理总报告
- `create-project-request.json`：创建项目请求体
- `create-project-response.json`：创建项目返回的项目、会话、工作流 ID
- `review-status-approved.json`：审核阶段状态
- `review-issues-warning.json`：审核阶段 warning 样例
- `execution-status-completed.json`：执行阶段完成态
- `accept-status-completed.json`：验收完成态
- `completion-summary-completed.json`：完成汇总
- `project-status-completed.json`：最终项目状态
- `stage-history-completed.json`：阶段历史
- `timeline-completed.json`：时间线
- `project-trace-completed.json`：项目轨迹总览
- `verification-status.json`：项目代码验证结果
- `backend-health.json`：后端健康检查样例
- `backend-config.json`：后端配置接口样例
- `backend-score-submit-invalid.json`：错误字段名请求样例
- `backend-score-submit.json`：正确提交分数样例
- `backend-scores.json`：排行榜查询样例
- `frontend-proxy-config.json`：前端代理访问后端配置样例
