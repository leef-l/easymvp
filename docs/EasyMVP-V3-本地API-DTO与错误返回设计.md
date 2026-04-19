# EasyMVP V3 本地 API DTO 与错误返回设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-API路由分组与命令查询边界设计](./EasyMVP-V3-API路由分组与命令查询边界设计.md)
> 关联文档：[EasyMVP-V3-错误码与诊断分级设计](./EasyMVP-V3-错误码与诊断分级设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：把 `/api/v3/*` 在 `GoFrame v2` 下的请求、响应和错误返回结构统一成实现可直接使用的 DTO 基线。

## 1. 设计结论

V3 本地 API 需要统一三类 DTO：

1. `query response`
2. `command request / result`
3. `error envelope`

原则：

1. 查询返回页面视图对象
2. 命令返回结果对象，不返回整页快照
3. 错误返回必须结构化
4. 查询与命令 DTO 只暴露 EasyMVP 自己的稳定字段，不直接暴露 `brain-v3` 原始工具协议字段

## 2. 通用响应结构

### 2.1 查询响应

建议：

```json
{
  "data": {},
  "meta": {
    "request_id": "req_xxx",
    "generated_at": "2026-04-19T12:00:00Z"
  }
}
```

### 2.2 命令响应

建议：

```json
{
  "result": {
    "command_id": "cmd_xxx",
    "accepted": true,
    "resource_id": "proj_xxx",
    "next_action": "open_project_workspace"
  },
  "meta": {
    "request_id": "req_xxx"
  }
}
```

### 2.3 错误响应

建议：

```json
{
  "error": {
    "code": "PRJ_002",
    "level": "error",
    "scope": "project_creation",
    "message": "项目工作区初始化失败",
    "debug_message": "mkdir workspace failed",
    "recovery_hint": "请检查目标路径权限后重试"
  },
  "meta": {
    "request_id": "req_xxx"
  }
}
```

## 3. 核心命令 DTO

### 3.1 `POST /api/v3/projects`

请求：

```json
{
  "name": "EasyMVP V3",
  "project_category": "web_app",
  "goal_summary": "设计并推进 V3",
  "workspace_root": "/path/to/workspace",
  "repo_root": "/path/to/repo"
}
```

结果：

```json
{
  "command_id": "cmd_xxx",
  "accepted": true,
  "resource_id": "proj_xxx",
  "next_action": "open_project_workspace"
}
```

### 3.2 `POST /api/v3/projects/{id}/plan/compile`

请求：

```json
{
  "plan_draft_id": "pd_xxx",
  "force_recompile": false
}
```

结果：

```json
{
  "command_id": "cmd_xxx",
  "accepted": true,
  "resource_id": "cp_xxx",
  "next_action": "refresh_plan_view"
}
```

### 3.3 `POST /api/v3/tasks/{id}/runs`

请求：

```json
{
  "brain_kind": "code",
  "resume_if_exists": true
}
```

说明：

1. `brain_kind` 是 EasyMVP 归一化后的运行时目标标识
2. 它通常由 `RoleResolver` / compile / runtime adapter 共同约束
3. 不表示 `easymvp-brain` 能力，也不等于底层原始工具名

结果：

```json
{
  "command_id": "cmd_xxx",
  "accepted": true,
  "resource_id": "runbind_xxx",
  "next_action": "watch_runtime_events"
}
```

### 3.4 `POST /api/v3/projects/{id}/acceptance-runs`

请求：

```json
{
  "profile_version": "v1",
  "mode": "production"
}
```

结果：

```json
{
  "command_id": "cmd_xxx",
  "accepted": true,
  "resource_id": "acc_xxx",
  "next_action": "open_acceptance_view"
}
```

### 3.5 `POST /api/v3/manual-decisions`

请求：

```json
{
  "project_id": "proj_xxx",
  "decision_kind": "release_gate_approve",
  "target_id": "acc_xxx",
  "comment": "人工确认放行"
}
```

## 4. 核心查询 DTO

### 4.1 `GET /api/v3/workspace/home-view`

`data` 应返回：

1. `summary`
2. `active_projects`
3. `need_attention`
4. `recent_activity`
5. `release_readiness`

### 4.2 `GET /api/v3/projects/{id}/workspace-view`

`data` 应返回：

1. `project_snapshot`
2. `stage_progress`
3. `live_activity`
4. `action_inbox`
5. `acceptance_coverage`

### 4.3 `GET /api/v3/projects/{id}/plan-view`

`data` 应返回：

1. `draft`
2. `review`
3. `compiled`
4. `task_projection`
5. `diff_summary`

### 4.4 `GET /api/v3/projects/{id}/acceptance-view`

`data` 应返回：

1. `acceptance_run`
2. `coverage_matrix`
3. `evidence_cards`
4. `release_gate`

补充约束：

1. 查询 DTO 只返回归一化后的页面对象
2. 不返回 `tools/list` / `tools/call`、原始 `content[]`、原始 run payload 或内置脑原始错误形态
3. `issues`
4. `evidence_cards`
5. `release_gate`

## 5. 错误返回规则

必须保证：

1. 所有 4xx / 5xx 都返回统一 `error envelope`
2. `code` 与文档中的错误码一致
3. `message` 面向用户
4. `debug_message` 面向日志和诊断

## 6. 字段命名规则

建议：

1. API JSON 统一 `snake_case`
2. 时间统一 RFC3339 字符串
3. 布尔字段用显式 `true / false`

## 7. 后续细分专题

1. 每个 handler 的完整 DTO 清单
2. 事件流 payload DTO
3. TypeScript client types 生成策略
