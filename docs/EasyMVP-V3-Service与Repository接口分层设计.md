# EasyMVP V3 Service 与 Repository 接口分层设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-实现架构与模块拆分设计](./EasyMVP-V3-实现架构与模块拆分设计.md)
> 关联文档：[EasyMVP-V3-代码目录结构与模块归属建议](./EasyMVP-V3-代码目录结构与模块归属建议.md)
> 目标：定义 V3 中 Go Service 层和 Repository 层的职责分层、接口边界与典型接口集合，保证实现时不会把业务编排、聚合投影、存储访问混写。

## 1. 设计结论

V3 必须明确：

1. Service 负责业务编排
2. Repository 负责持久化边界
3. Aggregation 负责页面投影

三者不能混。

## 2. Service 层建议

建议至少存在以下 Go Service：

1. `ProjectCreationService`
2. `ProjectService`
3. `PlanService`
4. `CompileService`
5. `TaskRunService`
6. `AcceptanceService`
7. `ManualDecisionService`
8. `WorkspaceViewService`
9. `DiagnosticsService`

## 3. Repository 层建议

建议至少存在：

1. `ProjectRepository`
2. `PlanRepository`
3. `TaskRepository`
4. `RunBindingRepository`
5. `AcceptanceRepository`
6. `EvidenceRepository`
7. `ReplayRepository`
8. `AuditRepository`
9. `SettingsRepository`

## 4. Service 只应做什么

1. 业务命令编排
2. 状态推进
3. 子系统调用
4. 事务边界控制
5. 领域事件发出

## 5. Repository 只应做什么

1. 查询
2. 保存
3. 删除
4. 索引访问
5. 按事务上下文执行

## 6. Aggregation 只应做什么

1. 读取领域对象
2. 聚合页面快照
3. 映射卡片、时间线、覆盖矩阵、待处理项

不能做：

1. 推进领域状态
2. 改写主对象

## 7. Go 接口风格建议

建议接口风格如下：

```go
type ProjectCreationService interface {
    CreateProject(ctx context.Context, cmd CreateProjectCommand) (CreateProjectResult, error)
    RetryCreation(ctx context.Context, projectID string) error
}

type ProjectRepository interface {
    GetByID(ctx context.Context, id string) (*Project, error)
    Save(ctx context.Context, project *Project) error
    ListActive(ctx context.Context, filter ActiveProjectFilter) ([]Project, error)
}
```

说明：

1. UI 不直接依赖这些接口
2. HTTP handler 调用 Service
3. Service 通过 Repository 落库

## 8. 事务边界建议

事务应主要落在 Service 层。

例如：

1. `ProjectCreationService.CreateProject`
2. `CompileService.CompilePlan`
3. `AcceptanceService.StartAcceptance`
4. `ManualDecisionService.ApplyDecision`

Repository 不主动决定大事务。

## 9. 禁止

1. Repository 做业务裁决
2. Service 直接拼页面卡片
3. UI 直接调 Repository
4. Handler 直接拼复杂事务
5. Aggregation 绕过 Service 修改对象
