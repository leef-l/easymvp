# EasyMVP V3 GoFrame 包级接口与依赖关系设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Go本地核心服务架构设计](./EasyMVP-V3-Go本地核心服务架构设计.md)
> 关联文档：[EasyMVP-V3-Service与Repository接口分层设计](./EasyMVP-V3-Service与Repository接口分层设计.md)
> 关联文档：[EasyMVP-V3-事务边界与一致性设计](./EasyMVP-V3-事务边界与一致性设计.md)
> 目标：把 `Go + GoFrame v2` 本地核心服务继续细化到包级接口和依赖方向，降低后续实现时的耦合风险。

## 1. 设计结论

GoFrame 本地核心服务必须采用“接口在上层定义、实现向下依赖”的方式组织。

推荐依赖方向：

```text
controller
  ↓
services / orchestrator / queries
  ↓
repositories / runtime ports / tx manager
  ↓
sqlite / brainserve / file storage
```

## 2. 建议上层接口

### 2.1 Commands

建议集中定义：

1. `ProjectCreationService`
2. `PlanCompileService`
3. `TaskRunService`
4. `AcceptanceService`
5. `ManualDecisionService`

### 2.2 Queries

建议集中定义：

1. `WorkspaceHomeQueryService`
2. `ProjectWorkspaceQueryService`
3. `PlanQueryService`
4. `AcceptanceQueryService`
5. `ReplayQueryService`
6. `AuditQueryService`

### 2.3 Ports

建议显式定义：

1. `BrainRuntimeClient`
2. `TxManager`
3. `Clock`
4. `IDGenerator`
5. `FileStore`

## 3. 示例接口

```go
type BrainRuntimeClient interface {
    CreateRun(ctx context.Context, req CreateBrainRunRequest) (CreateBrainRunResult, error)
    GetRun(ctx context.Context, brainRunID string) (BrainRunSnapshot, error)
    CancelRun(ctx context.Context, brainRunID string) error
}

type TxManager interface {
    WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

## 4. Repository 依赖原则

Repository 接口应只暴露：

1. `Get`
2. `List`
3. `Save`
4. `Delete`
5. `Lock-like query` 的等价约束实现

不应暴露：

1. UI 视图对象
2. 外部运行时调用
3. 复杂业务流程

## 5. Query Service 原则

Query Service 可依赖：

1. repository
2. aggregation helpers
3. read-only cache

不应依赖：

1. command service
2. worker manager
3. 外部写入副作用

## 6. Worker 与接口关系

worker 应通过上层接口工作，而不是直接写库。

例如：

1. `run_sync_worker` 调 `RuntimeSyncService`
2. `acceptance_refresh_worker` 调 `AcceptanceRefreshService`
3. `snapshot_refresh_worker` 调 `WorkspaceSnapshotService`

## 7. 包级禁止规则

不允许：

1. `api` 直接 import `storage/sqlite`
2. `aggregation` 直接 import `brainserve`
3. `worker` 直接 import `api`
4. `runtime` 直接修改 UI DTO

## 8. 后续细分专题

1. 每个 package 的 `interfaces.go`
2. command/query 拆分目录规范
3. 单元测试 mock 边界
