# EasyMVP V3 Service 与 Repository 方法清单终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Service与Repository接口分层设计](./EasyMVP-V3-Service与Repository接口分层设计.md)
> 关联文档：[EasyMVP-V3-Go包级接口与依赖关系设计](./EasyMVP-V3-Go包级接口与依赖关系设计.md)
> 目标：把 V3 首批 GoFrame 服务层和仓储层的方法面一次列清，供 `internal/service` 与 `internal/dao/repository` 直接落接口。

## 1. 设计结论

这份文档只解决一个问题：

> 首批实现到底要有哪些 service 方法和 repository 方法。

## 2. Service 终稿

### 2.1 `SystemService`

1. `Health(ctx context.Context) (*v1.HealthRes, error)`

### 2.2 `WorkspaceQueryService`

1. `GetHomeView(ctx context.Context) (*workspacev1.HomeViewRes, error)`
2. `GetProjectWorkspaceView(ctx context.Context, projectID string) (*projectsv1.ProjectWorkspaceViewRes, error)`

### 2.3 `ProjectCommandService`

1. `CreateProject(ctx context.Context, req CreateProjectCommand) (*CommandResult, error)`
2. `GetProjectWorkspaceView(ctx context.Context, projectID string) (*projectsv1.ProjectWorkspaceViewRes, error)`

### 2.4 `PlanService`

1. `GetPlanView(ctx context.Context, projectID string) (*planv1.PlanViewRes, error)`
2. `CompilePlan(ctx context.Context, req CompilePlanCommand) (*CommandResult, error)`

### 2.5 `AcceptanceService`

1. `GetAcceptanceView(ctx context.Context, projectID string) (*acceptancev1.AcceptanceViewRes, error)`
2. `StartAcceptance(ctx context.Context, req StartAcceptanceCommand) (*CommandResult, error)`

### 2.6 `ManualDecisionService`

1. `ApplyDecision(ctx context.Context, req ApplyManualDecisionCommand) (*CommandResult, error)`

## 3. Repository 终稿

### 3.1 `ProjectRepository`

1. `GetByID(ctx context.Context, id string) (*entity.Project, error)`
2. `ListActive(ctx context.Context) ([]*entity.Project, error)`
3. `Insert(ctx context.Context, data *do.Project) error`
4. `UpdateStatus(ctx context.Context, id string, status string, productionStatus string) error`

### 3.2 `ProjectProfileRepository`

1. `Insert(ctx context.Context, data *do.ProjectProfile) error`
2. `GetByProjectID(ctx context.Context, projectID string) (*entity.ProjectProfile, error)`

### 3.3 `ProjectWorkspaceRepository`

1. `Insert(ctx context.Context, data *do.ProjectWorkspace) error`
2. `GetByProjectID(ctx context.Context, projectID string) (*entity.ProjectWorkspace, error)`

### 3.4 `PlanDraftRepository`

1. `Insert(ctx context.Context, data *do.WorkflowPlanDraft) error`
2. `GetCurrentByProjectID(ctx context.Context, projectID string) (*entity.WorkflowPlanDraft, error)`
3. `ListByProjectID(ctx context.Context, projectID string) ([]*entity.WorkflowPlanDraft, error)`

### 3.5 `PlanReviewRepository`

1. `Insert(ctx context.Context, data *do.WorkflowPlanReviewResult) error`
2. `GetLatestByDraftID(ctx context.Context, draftID string) (*entity.WorkflowPlanReviewResult, error)`

### 3.6 `CompiledPlanRepository`

1. `Insert(ctx context.Context, data *do.WorkflowCompiledPlan) error`
2. `GetCurrentByProjectID(ctx context.Context, projectID string) (*entity.WorkflowCompiledPlan, error)`

### 3.7 `CompiledTaskRepository`

1. `BatchInsert(ctx context.Context, list []*do.WorkflowCompiledTask) error`
2. `ListByCompiledPlanID(ctx context.Context, compiledPlanID string) ([]*entity.WorkflowCompiledTask, error)`

### 3.8 `DomainTaskRepository`

1. `BatchInsert(ctx context.Context, list []*do.DomainTask) error`
2. `ListByProjectID(ctx context.Context, projectID string) ([]*entity.DomainTask, error)`
3. `UpdateStatus(ctx context.Context, taskID string, status string) error`

### 3.9 `BrainRunBindingRepository`

1. `Insert(ctx context.Context, data *do.BrainRunBinding) error`
2. `GetByBrainRunID(ctx context.Context, brainRunID string) (*entity.BrainRunBinding, error)`
3. `UpdateStatus(ctx context.Context, id string, status string) error`

### 3.10 `AcceptanceRunRepository`

1. `Insert(ctx context.Context, data *do.AcceptanceRun) error`
2. `GetLatestByProjectID(ctx context.Context, projectID string) (*entity.AcceptanceRun, error)`

补充说明：

- 这里保留 `AcceptanceRunRepository` 代表历史实现资产和当期读写现实
- 但按当前钱学森总纲，后续不应把 `AcceptanceRun` 继续当成最终完成语义中心
- 更准确的方向是让 `AcceptanceRun` 与 `VerificationResult / CompletionVerdict` 并存，并逐步让后两者承担验证与完成裁决语义

### 3.11 `AcceptanceIssueRepository`

1. `BatchInsert(ctx context.Context, list []*do.AcceptanceIssue) error`
2. `ListByRunID(ctx context.Context, runID string) ([]*entity.AcceptanceIssue, error)`

### 3.12 `EvidenceRepository`

1. `ListByProjectID(ctx context.Context, projectID string) ([]*entity.EvidenceItem, error)`
2. `ListByRunID(ctx context.Context, runID string) ([]*entity.EvidenceItem, error)`

补充说明：

- `Evidence` 的读取不应只服务 `AcceptanceRun`
- 还应逐步支持 `VerificationResult`、`RuntimeEscalation`、`FaultSummary` 等闭环对象的页面消费

### 3.13 `ReplayRepository`

1. `ListByProjectID(ctx context.Context, projectID string) ([]*entity.ReplayItem, error)`

### 3.14 `AuditRepository`

1. `Insert(ctx context.Context, data *do.AuditLog) error`
2. `ListByProjectID(ctx context.Context, projectID string, limit int) ([]*entity.AuditLog, error)`

## 4. 方法实现优先级

### P0

1. `SystemService.Health`
2. `ProjectCommandService.CreateProject`
3. `WorkspaceQueryService.GetHomeView`
4. `WorkspaceQueryService.GetProjectWorkspaceView`

### P1

1. `PlanService.GetPlanView`
2. `PlanService.CompilePlan`
3. `AcceptanceService.GetAcceptanceView`
4. `AcceptanceService.StartAcceptance`

### P2

1. `ManualDecisionService.ApplyDecision`
2. `RuntimeSyncService.*`
3. `DiagnosticsService.*`

## 5. 实施约束

1. 所有写方法必须经过 service
2. repository 不返回 UI DTO
3. query service 可以聚合多个 repository 读取
4. command service 应持有事务边界
