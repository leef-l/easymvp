# EasyMVP V3 创建初始化事件流接口设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-创建项目流程与页面设计](./EasyMVP-V3-创建项目流程与页面设计.md)
> 关联文档：[EasyMVP-V3-实时事件流推送机制设计](./EasyMVP-V3-实时事件流推送机制设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 关联文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 关联文档：[EasyMVP-V3-本地目录与项目工作区规范](./EasyMVP-V3-本地目录与项目工作区规范.md)
> 目标：定义创建项目过程中“初始化进度 + 实时事件 + 创建结果”的接口边界、事件 schema 和前端消费规则。

## 1. 设计结论

创建项目不能只有一个同步 `POST` 返回结果。

V3 必须把“创建项目”拆成：

1. 创建命令接口
2. 创建初始化快照接口
3. 创建初始化事件流接口
4. 创建完成后的项目工作台跳转接口

这样才能满足：

1. 创建过程中可实时展示
2. 创建失败时可准确定位到步骤
3. 创建完成后无缝切换到 `Project Workspace`
4. 页面刷新或弹层关闭后可恢复初始化状态

## 2. 为什么不能只靠同步接口

如果创建项目只返回：

1. 成功
2. 失败

那前端会失去这几类关键信息：

1. 当前初始化进行到哪一步
2. 是否已经创建了项目主记录
3. 是否已经准备好本地目录
4. 是哪一环失败
5. 是否可以直接进入工作台

而 V3 的创建过程本身已经是一个小型工作流，所以必须被事件化。

## 3. 接口总览

建议创建链路至少暴露以下接口：

1. `POST /api/v3/projects`
2. `GET /api/v3/project-creations/{creation_id}`
3. `GET /api/v3/project-creations/{creation_id}/events`
4. `GET /api/v3/project-creations/{creation_id}/result`

## 4. 创建命令接口

### 4.1 接口定义

`POST /api/v3/projects`

### 4.2 请求体

建议请求体如下：

```json
{
  "project_name": "My Web App",
  "project_category": "web_app",
  "goal_prompt": "做一个本地优先的客户管理工具",
  "workspace_source": "existing_repository",
  "repo_path_or_local_path": "/Users/demo/projects/my-web-app"
}
```

### 4.3 返回体

建议返回：

```json
{
  "creation_id": "pc_01",
  "status": "creating",
  "project_id": null,
  "snapshot_url": "/api/v3/project-creations/pc_01",
  "events_url": "/api/v3/project-creations/pc_01/events",
  "result_url": "/api/v3/project-creations/pc_01/result"
}
```

### 4.4 设计原则

返回的重点不是完整项目对象，而是一个可持续追踪的 `creation_id`。

## 5. 创建快照接口

### 5.1 接口定义

`GET /api/v3/project-creations/{creation_id}`

### 5.2 作用

用于回答：

1. 现在到哪一步了
2. 哪些步骤已完成
3. 是否失败
4. 当前是否可进入项目工作台

### 5.3 建议返回对象

建议定义 `ProjectCreationSnapshot`：

```json
{
  "creation_id": "pc_01",
  "status": "initializing_workspace",
  "project_id": "proj_01",
  "project_name": "My Web App",
  "project_category": "web_app",
  "started_at": "2026-04-19T10:00:00Z",
  "updated_at": "2026-04-19T10:00:03Z",
  "current_step": "initializing_workspace",
  "progress_percent": 42,
  "steps": [
    {
      "step_key": "creating",
      "status": "done"
    },
    {
      "step_key": "initializing_workspace",
      "status": "active"
    },
    {
      "step_key": "binding_profiles",
      "status": "pending"
    },
    {
      "step_key": "bootstrapping_plan",
      "status": "pending"
    }
  ],
  "latest_summary": "正在扫描本地工作区并建立初始索引",
  "can_open_workspace": true,
  "workspace_url": "/projects/proj_01/workspace"
}
```

## 6. 创建事件流接口

### 6.1 接口定义

`GET /api/v3/project-creations/{creation_id}/events`

### 6.2 推送方式

第一版建议采用 SSE。

理由：

1. 创建过程是服务端单向推送
2. 前端只需要实时接收进度和失败信息
3. 与工作台主事件机制保持一致

### 6.3 事件通道

建议使用逻辑通道：

1. `project_creation`

### 6.4 事件命名

建议至少包括：

1. `creation.started`
2. `creation.step_changed`
3. `creation.step_succeeded`
4. `creation.step_failed`
5. `creation.project_bound`
6. `creation.workspace_ready`
7. `creation.plan_bootstrapped`
8. `creation.completed`
9. `creation.failed`

## 7. 事件 schema

建议所有创建事件统一使用 `ProjectCreationEvent`：

```json
{
  "event_id": "pce_1001",
  "event_seq": 12,
  "event_type": "creation.step_changed",
  "creation_id": "pc_01",
  "project_id": "proj_01",
  "severity": "info",
  "created_at": "2026-04-19T10:00:02Z",
  "step_key": "initializing_workspace",
  "summary": "开始初始化本地工作区",
  "needs_attention": false,
  "can_retry": false,
  "deep_link": "/projects/proj_01/workspace"
}
```

### 7.1 关键字段

1. `creation_id`
2. `project_id`
3. `step_key`
4. `event_seq`
5. `summary`
6. `needs_attention`
7. `can_retry`

### 7.2 可选扩展字段

建议按步骤补充：

1. `path_check_result`
2. `profile_binding_result`
3. `plan_bootstrap_result`
4. `error_code`
5. `error_message`
6. `suggested_action`

## 8. 步骤定义与映射

创建初始化事件必须和创建状态机严格对齐。

建议固定步骤：

1. `creating`
2. `initializing_workspace`
3. `binding_profiles`
4. `bootstrapping_plan`

### 8.1 `creating`

表示：

1. 写入项目主记录
2. 生成 `project_id`
3. 创建初始项目元数据

### 8.2 `initializing_workspace`

表示：

1. 检查路径
2. 绑定工作区根目录
3. 建立初始目录结构
4. 建立索引和忽略规则

### 8.3 `binding_profiles`

表示：

1. 绑定 `ProjectCategory`
2. 命中 `CategoryProfile`
3. 绑定默认验收框架
4. 生成初始执行偏好上下文

### 8.4 `bootstrapping_plan`

表示：

1. 创建初始 `PlanDraft`
2. 或创建 `PlanBootstrapIntent`
3. 准备将项目推进到 `Design / Review` 主链路

## 9. 创建结果接口

### 9.1 接口定义

`GET /api/v3/project-creations/{creation_id}/result`

### 9.2 作用

用于在创建结束后提供明确落点。

### 9.3 成功返回

```json
{
  "creation_id": "pc_01",
  "status": "created",
  "project_id": "proj_01",
  "workspace_url": "/projects/proj_01/workspace",
  "plan_url": "/projects/proj_01/plan",
  "current_stage": "design",
  "initial_action_hint": "查看初始计划草案"
}
```

### 9.4 失败返回

```json
{
  "creation_id": "pc_01",
  "status": "create_failed",
  "failed_step": "binding_profiles",
  "error_code": "profile_binding_failed",
  "error_message": "默认分类策略绑定失败",
  "suggested_action": "检查分类 profile 配置后重试",
  "can_retry": true
}
```

## 10. 前端消费规则

### 10.1 弹层打开时

前端流程建议为：

1. 提交 `POST /api/v3/projects`
2. 拿到 `creation_id`
3. 先拉取一次创建快照
4. 再建立 SSE 连接

### 10.2 收到事件时

前端应：

1. 追加到初始化事件流时间线
2. 同步更新当前步骤和进度条
3. 遇到 `creation.project_bound` 后允许提前进入工作台
4. 遇到 `creation.completed` 后自动提供“进入工作台”

### 10.3 页面刷新时

若用户刷新或关闭弹层后重新回来，前端应：

1. 通过 `creation_id` 重新拉快照
2. 再次连接事件流
3. 以 `Last-Event-ID` 做续传

## 11. 与 Project Workspace 的衔接

创建过程不能和单项目工作台割裂。

建议规则如下：

1. 当 `project_id` 已生成后，`Project Workspace` 就可以被访问
2. 但此时页面顶部显示 `Project is being prepared`
3. `Live Activity` 优先显示创建初始化事件
4. `Action Inbox` 只显示初始化失败和待确认项

也就是说：

创建初始化事件本身就是项目最早期的实时活动流。

## 12. 与 Workspace Home 的衔接

项目一旦进入 `creation.project_bound`，首页就应当能够看见这个项目。

建议展示原则：

1. `Running Projects` 中显示新项目卡片
2. `current_stage` 显示为 `Design`
3. `active_task_or_run` 显示为 `Preparing project`
4. `Recent Activity` 中可见最新初始化事件

## 13. 失败与恢复策略

### 13.1 可重试失败

例如：

1. `runtime_not_ready`
2. `profile_binding_failed`
3. `plan_bootstrap_failed`

这类失败建议允许：

1. 修正后重试当前步骤
2. 保留已成功的前置结果

### 13.2 不可直接重试失败

例如：

1. `workspace_conflict`
2. `path_not_accessible`

这类失败建议要求用户先修改输入，再重新发起创建。

### 13.3 回滚可见性

如果系统回滚了部分结果，快照和事件必须明确说明：

1. 哪部分已回滚
2. 哪部分仍保留
3. 当前是否已占用 `project_id`

## 14. 刷新与幂等规则

### 14.1 幂等

同一 `creation_id` 下：

1. `event_id` 必须幂等
2. `event_seq` 必须单调递增

### 14.2 快照优先

前端不能完全依赖事件流重建状态。

正确做法：

1. 快照作为当前真相
2. 事件流作为实时增量

### 14.3 过期处理

当创建记录已过期或清理后，接口应返回明确状态：

1. `expired`
2. `not_found`

而不是让前端无限重连。

## 15. 不该怎么做

创建初始化接口不应该：

1. 只返回一个长时间挂起的同步 HTTP 请求
2. 把所有日志直接原样推给前端
3. 没有 `creation_id`
4. 没有快照接口
5. 创建完成后无法明确跳去哪里

## 16. 后续细分专题

本专题后续继续拆：

1. 创建初始化事件表结构设计
2. 创建初始化弹层线框图设计
3. 创建失败恢复与回滚策略设计
4. 创建完成后首次进入 `Project Workspace` 引导态设计
