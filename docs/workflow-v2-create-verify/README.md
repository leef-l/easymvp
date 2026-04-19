# Workflow V2 验证记录

本目录专门记录 Workflow V2 的真实接口验证、问题复现、修复结论和复测结果，避免信息继续散落在总进度文档里。

当前文件：

- `2026-04-09-create-project-verify.md`：`create-project` 创建链路真实验证说明、问题、修复与复测
- `2026-04-09-e2e-complete-verify.md`：从创建项目到 `completed` 的全链路真实验证、问题记录、修复与复测
- `create-project-request.json`：创建项目验证请求体样例
- `create-project-response-success.json`：创建项目验证修复后成功响应样例
- `project-status-response-success.json`：创建项目验证项目状态响应样例
- `stage-history-response-success.json`：创建项目验证阶段历史响应样例
- `project-trace-response-success.json`：创建项目验证项目轨迹响应样例
- `e2e-create-project-request.json`：全链路验证复测请求体样例
- `e2e-create-project-response-success.json`：全链路验证复测创建成功响应样例
- `e2e-initial-review-status-broken.json`：首次全链路运行中 `review-status` 时间异常样例
- `e2e-initial-execution-status-broken.json`：首次全链路运行中 `skip-task` 时间异常样例
- `e2e-initial-completion-summary-broken.json`：首次全链路运行中完成汇总异常样例
- `e2e-rerun-review-status-fixed.json`：修复后 `review-status` 复测样例
- `e2e-rerun-execution-status-fixed.json`：修复后 `execution-status` 复测样例
- `e2e-rerun-accept-status-completed.json`：修复后验收完成样例
- `e2e-rerun-completion-summary-completed.json`：修复后完成汇总样例
- `e2e-rerun-timeline-final.json`：修复后时间线复测样例

说明：

- 鉴权 `Authorization` 头没有落盘，避免把临时 JWT 写入仓库文档。
- 全链路验证为了避免外部执行器依赖，使用受控方式写入架构师 JSON 回复，并将项目执行模式切到仅用于验证的 `disabled_for_e2e`，再通过 `skip-task` 推进执行阶段。
- 本次验证未运行 `web-antd`，也未执行 `test-workspaces/validate.sh` 或 guard 脚本。
