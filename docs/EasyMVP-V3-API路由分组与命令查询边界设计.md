# EasyMVP V3 API 路由分组与命令查询边界设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-实现架构与模块拆分设计](./EasyMVP-V3-实现架构与模块拆分设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 关联文档：[EasyMVP-V3-本地API与IPC适配设计](./EasyMVP-V3-本地API与IPC适配设计.md)
> 目标：定义 V3 本地 Go 服务的 API 路由分组、查询接口与命令接口边界，避免实现时接口风格混乱。

## 1. 设计结论

V3 API 必须按“查询”和“命令”分开。

这些 API 由本地 `GoFrame v2` 服务提供，并默认只绑定回环地址。

## 2. 服务边界

首版建议：

1. 服务监听 `127.0.0.1`
2. 不默认对局域网暴露
3. Electron Renderer 通过本地 client 调用

说明：

1. 这是单机工作台的本地 API
2. 不是公网 SaaS API

## 3. 路由分组建议

建议至少分：

1. `/api/v3/workspace/*`
2. `/api/v3/projects/*`
3. `/api/v3/plan/*`
4. `/api/v3/acceptance/*`
5. `/api/v3/runtime/*`
6. `/api/v3/replay/*`
7. `/api/v3/settings/*`
8. `/api/v3/system/*`

## 4. 查询接口

查询接口负责：

1. 页面快照
2. 详情抽屉
3. 索引列表
4. 事件流读取
5. 健康与诊断读取
6. runtime binding 读取
7. runtime binding event 读取
8. runtime run detail 读取

查询接口原则：

1. 不推进状态机
2. 不落关键业务写
3. 尽量返回聚合后的页面对象

## 5. 命令接口

命令接口负责：

1. create project
2. retry project creation
3. compile plan
4. start task run
5. start acceptance
6. apply manual decision
7. retry failed worker action
8. sync runtime run binding
9. resume runtime run binding
10. cancel runtime run binding

命令接口原则：

1. 只返回 command result
2. 页面完整刷新通过 query 获取
3. 状态变化通过 event stream 感知

## 6. 示例

### 6.1 查询

1. `GET /api/v3/workspace/home-view`
2. `GET /api/v3/projects/{id}/workspace-view`
3. `GET /api/v3/projects/{id}/plan-view`
4. `GET /api/v3/projects/{id}/acceptance-view`
5. `GET /api/v3/runtime/healthz`
6. `GET /api/v3/runtime-runs/{binding_id}`
7. `GET /api/v3/runtime-runs/{binding_id}/detail`
8. `GET /api/v3/runtime-runs/{binding_id}/events`
9. `GET /api/v3/projects/{project_id}/runs/{run_id}/replay-summary`
10. `GET /api/v3/projects/{project_id}/runs/{run_id}/replay-timeline`
11. `GET /api/v3/projects/{project_id}/runs/{run_id}/replay-items/{replay_id}`
12. `GET /api/v3/projects/{project_id}/runs/{run_id}/replay-items/{replay_id}/raw`
13. `GET /api/v3/projects/{project_id}/runs/{run_id}/log-segments`
14. `GET /api/v3/projects/{project_id}/runs/{run_id}/log-segments/{segment_id}/raw`
8. `GET /api/v3/runtime-runs/{binding_id}/events`

### 6.2 命令

1. `POST /api/v3/projects`
2. `POST /api/v3/projects/{id}/plan/compile`
3. `POST /api/v3/projects/{id}/runtime-runs`
4. `POST /api/v3/projects/{id}/acceptance-runs`
5. `POST /api/v3/manual-decisions`
6. `POST /api/v3/runtime-runs/{binding_id}/sync`
7. `POST /api/v3/runtime-runs/{binding_id}/resume`
8. `DELETE /api/v3/runtime-runs/{binding_id}`

## 7. Handler 到 Service 的边界

推荐调用关系：

```text
GoFrame Controller
  → validate request
  → call Service
  → map result
```

不要：

1. 在 Controller 里写业务流程
2. 在 Controller 里直接调多个 Repository 拼事务

## 8. 页面刷新策略

规则：

1. 命令接口完成后，不直接返回整页快照
2. 页面优先重新拉取 query snapshot
3. 同时订阅 event stream 获取实时变化

## 9. 后续细分专题

1. HTTP handler 命名规范
2. 错误码与诊断码体系
3. 事件流接口协议
