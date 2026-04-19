# EasyMVP V3 GoFrame v2 本地核心服务架构设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-实现架构与模块拆分设计](./EasyMVP-V3-实现架构与模块拆分设计.md)
> 关联文档：[EasyMVP-V3-代码目录结构与模块归属建议](./EasyMVP-V3-代码目录结构与模块归属建议.md)
> 关联文档：[EasyMVP-V3-Service与Repository接口分层设计](./EasyMVP-V3-Service与Repository接口分层设计.md)
> 关联文档：[EasyMVP-V3-单机版启动时序与进程内调用链设计](./EasyMVP-V3-单机版启动时序与进程内调用链设计.md)
> 目标：把 V3 的 `Go + GoFrame v2` 本地核心服务细化到可直接落模块、落 package、落接口装配的程度。

## 1. 设计结论

V3 的 GoFrame v2 本地核心服务不是简单的本地 API 代理，而是：

> 一个承接工作流编排、计划编译、运行时接入、验收裁决、视图聚合和后台调度的本地业务内核。

建议正式拆成 8 个一级模块：

1. `app`
2. `api`
3. `orchestrator`
4. `plan`
5. `acceptance`
6. `runtime`
7. `aggregation`
8. `storage`

另配两个支撑模块：

9. `worker`
10. `diagnostics`

## 2. 模块总览

```text
main.go
  ↓
internal/cmd
  ↓
internal/logic
  ├─ orchestrator
  ├─ plan
  ├─ acceptance
  ├─ runtime
  ├─ aggregation
  ├─ worker
  └─ diagnostics
internal/service
internal/controller
internal/dao
internal/model
```

## 3. 各模块职责

### 3.1 `app`

负责：

1. 配置读取
2. 依赖装配
3. 服务启动
4. 生命周期管理
5. 健康检查注册

### 3.2 `api`

负责：

1. HTTP 路由分组
2. request / response mapping
3. 参数校验
4. command/query handler 绑定
5. 错误码映射

在 GoFrame v2 下，推荐由：

1. `controller`
2. `service`
3. `model`

三层配合完成。

### 3.3 `orchestrator`

负责：

1. 项目主状态机
2. 计划到任务的推进
3. 命令事务边界
4. 跨子系统协调
5. 人工动作落地

### 3.4 `plan`

负责：

1. `PlanDraft`
2. `PlanReviewResult`
3. `CompiledPlan`
4. `CompiledTask`
5. compiler pipeline
6. `RoleResolver`
7. category strategy resolve

### 3.5 `acceptance`

负责：

1. acceptance profile resolve
2. `AcceptanceRun`
3. issue / judgement
4. coverage 计算
5. release gate 判定

补充说明：

- 这里保留 `AcceptanceRun` 代表当前实现与历史设计现实
- 按当前钱学森总纲，`acceptance` 子系统后续不应只围绕 `AcceptanceRun` 组织最终语义
- 更准确的方向是逐步补齐 `VerificationResult / CompletionVerdict`，并让它们承担验证结果与完成裁决

### 3.6 `runtime`

负责：

1. `brain serve` client
2. `BrainRunBinding`
3. run status sync
4. logs / replay / checkpoints normalize
5. runtime event mapping

边界补充：

1. `runtime` 可以理解 `brain-v3` 原始 `tools/list` / `tools/call`、`content[]` 与 `completed / failed / unsupported / denied`
2. 但它对外只能输出 EasyMVP 归一化后的 runtime / verification / fault / evidence 摘要
3. 不把内置脑原始 payload 直接透传给 `plan`、`acceptance` 或 `aggregation`

### 3.7 `aggregation`

负责：

1. `WorkspaceHomeView`
2. `ProjectWorkspaceView`
3. `PlanView`
4. `AcceptanceView`
5. `LiveEvent`
6. `ActionInboxItem`

边界补充：

1. `aggregation` 只消费领域对象与归一化后的运行时摘要
2. `source_brain` 一类字段只表示来源归属，不表示页面直接拥有执行脑选择能力
3. `aggregation` 不直接依赖 `brain-v3` 内置脑原始工具协议

### 3.8 `storage`

负责：

1. SQLite connection
2. migrations
3. repository impl
4. file storage mapping
5. transaction helpers

### 3.9 `worker`

负责：

1. run sync jobs
2. evidence index jobs
3. replay index jobs
4. acceptance refresh jobs
5. snapshot refresh jobs

### 3.10 `diagnostics`

负责：

1. system health snapshot
2. startup diagnostics
3. migration failure details
4. worker failure records
5. user-facing recovery hints

## 4. 推荐 package 结构

```text
internal/
  cmd/
    cmd.go
  controller/
    workspace/
    projects/
    plan/
    acceptance/
    settings/
    system/
  service/
    workspace.go
    projects.go
    plan.go
    acceptance.go
    runtime.go
  logic/
    orchestrator/
    plan/
    acceptance/
    runtime/
    aggregation/
    worker/
    diagnostics/
  dao/
  model/
    entity/
    do/
  packed/
manifest/
resource/
hack/
```

## 5. 服务装配顺序

建议装配顺序：

1. `config`
2. `logger`
3. `g.Meta` / router registration
4. `sqlite`
5. `migrations`
6. `dao / repositories`
7. `runtime client`
8. `plan services`
9. `acceptance services`
10. `orchestrator services`
11. `aggregation services`
12. `worker manager`
13. `GoFrame HTTP server`

装配原则补充：

1. `runtime client` 在装配时就承担原始协议吸收与归一化边界
2. 上层服务默认只看归一化 DTO，不重复理解 `brain-v3` 原始协议细节

## 6. 核心接口建议

建议至少明确以下接口：

1. `ProjectCreationService`
2. `ProjectCommandService`
3. `PlanCommandService`
4. `AcceptanceCommandService`
5. `WorkspaceQueryService`
6. `PlanQueryService`
7. `AcceptanceQueryService`
8. `RuntimeSyncService`
9. `WorkerManager`
10. `TxManager`

## 7. 重要边界

必须保证：

1. `api` 不直接调 repository
2. `aggregation` 不改领域状态
3. `worker` 不绕过 service 改关键状态
4. `runtime` 不替代 orchestrator 做业务裁决
5. `storage` 不承载业务规则

## 8. 可直接开工的第一批 package

建议先落：

1. `storage/sqlite`
2. `storage/repositories`
3. `api/system`
4. `api/projects`
5. `orchestrator`
6. `aggregation/workspace`
7. `runtime/brainserve`
8. `worker/manager`

## 9. 后续细分专题

1. package 级接口清单
2. handler request/response DTO 清单
3. runtime client 重试与超时规范
