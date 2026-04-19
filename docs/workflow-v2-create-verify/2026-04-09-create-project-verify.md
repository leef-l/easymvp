# 2026-04-09 Workflow V2 创建项目验证

## 1. 验证目标

在真实后端服务上验证 Workflow V2 `create-project` 主链是否可以成功创建项目、创建会话并启动阶段化工作流，同时记录接口问题并及时修复。

## 2. 验证环境

- 日期：2026-04-09
- 服务：`admin-go/app/mvp/main.go`
- 接口前缀：`http://127.0.0.1:9002/api/mvp`
- 验证方式：直接调用后端 API，不经过 `web-antd`
- 工作目录：`/www/wwwroot/project/easymvp/test-workspaces/workflow-v2-create-verify/repo`

## 3. 请求与结果

### 3.1 创建项目请求

- 接口：`POST /workflow/create-project`
- 请求体：见 `create-project-request.json`

### 3.2 首次真实返回

接口返回成功：

```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "projectID": "317586673791668224",
    "conversationID": "317586673875554304",
    "workflowRunID": "317586673909108736"
  }
}
```

后续又补查了：

- `GET /workflow/project-status`
- `GET /workflow/timeline`
- `GET /workflow/stage-history`
- `GET /workflow/project-trace`

确认项目已进入 `workflow_v2 / design`，不是仅有项目记录落库。

### 3.3 修复后复测样例

修复后重新创建了项目 `workflow-v2-create-verify-20260409-fix`，成功响应见 `create-project-response-success.json`。

## 4. 发现的问题

### 4.1 `stage-history` 时间与其他接口不一致

同一条创建链路下：

- `timeline` 与 `project-trace` 显示时间为 `2026-04-09 15:09:24`
- `stage-history` 显示 `startedAt` 为 `2026-04-09 07:09:24`

结论：

- `stage-history` 直接返回了数据库 UTC 原值
- `timeline` / `project-trace` 已做 `UTC -> 本地时区` 归一化
- 导致同一项目在不同 API 中出现 8 小时时差，前端会把阶段起始时间显示早 8 小时

## 5. 修复内容

代码修复：

- `admin-go/app/mvp/internal/controller/chat/workflow_timeline.go`
  - 给 `StageHistory` 统一改为复用时间归一化
  - 同步修正 `ReworkStatus` 内的 `started_at / created_at`，避免返工链路出现同类时区漂移

测试补充：

- `admin-go/app/mvp/internal/controller/chat/workflow_time_test.go`
  - 新增 `TestBuildStageHistoryItemNormalizesTimes`

## 6. 复测结果

修复后重新启动 `admin-go` 并复测：

- `POST /workflow/create-project` 成功
- `GET /workflow/project-status` 返回 `workflowStatus=designing`、`currentStage=design`
- `GET /workflow/stage-history` 与 `GET /workflow/project-trace` 时间口径一致，均返回 `2026-04-09 15:15:43`
- `go test ./app/mvp/internal/controller/chat` 通过

## 7. 本次验证边界

本次没有执行以下内容：

- `web-antd` 编译或运行
- `test-workspaces/validate.sh`
- guard 脚本
