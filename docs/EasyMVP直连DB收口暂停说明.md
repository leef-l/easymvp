# EasyMVP直连DB收口暂停说明

更新日期：2026-04-11

这份文档记录本轮“禁止直接 DB、统一走 repo / service”改造的暂停点。目的不是总结成绩，而是明确当前还没做完的范围、已经做完的范围，以及恢复工作时应该按什么顺序继续。

## 1. 当前判定

当前状态不是“已完成”，只能算“验收主链阶段性收口”。

已完成的关键链路：

- `workflow.category_resolver`
- `workflow.verification.service`
- `workflow.acceptance.rule_engine`
- `workflow.stage.accept.service`
- `controller/chat/workflow_verification.go`
- `controller/chat/workflow_accept.go`
- `controller/chat/workflow_review.go`
- `controller/chat/workflow.go` 中 `latestWorkflowRunForProject` 公共 helper

这些链路已经完成从编排层/控制层直连表，迁移到 `repo` 的收口。

## 2. 已完成内容

### 2.1 已 repo 化的主链

本轮已经完成以下主链的 repo 收口：

- 分类解析：`CategoryResolver -> ProjectRepo / ProjectCategoryRepo`
- 标准化验证：`verification.Service -> ProjectRepo / ProjectCategoryRepo / DomainTaskRepo / Verification*Repo`
- 验收规则引擎：`acceptance.RuleEngine -> TaskWorkspaceRepo / DomainTaskRepo / StageRunRepo / Verification*Repo`
- 验收阶段编排：`stage.accept.Service -> WorkflowRunRepo / ProjectRepo / StageRunRepo / AcceptRunRepo`
- 控制层验收接口：`workflow_accept.go -> Accept*Repo / WorkflowRunRepo`
- 控制层验证接口：`workflow_verification.go -> ProjectRepo / VerificationIssueRepo / DomainTaskRepo / WorkflowRunRepo`
- 控制层评审接口：`workflow_review.go -> PlanVersionRepo / BlueprintRepo / ReviewIssueRepo / StageRunRepo / StageTaskRepo / WorkflowRunRepo / ProjectRepo`

### 2.2 本轮补充的 repo 能力

为完成上述收口，已经补充或扩展以下仓储能力：

- `ProjectRepo`
  - `GetByID`
  - `BackfillCategoryCodeIfEmpty`
  - `UpdateStatus`
  - `UpdateStatusIfCurrent`
- `WorkflowRunRepo`
  - `GetLatestByProject`
  - `GetLatestByProjectStatuses`
  - `GetLatestByProjectExcludingStatuses`
- `StageRunRepo`
  - `GetStatusByID`
  - `ListCompletedStageTypes`
  - `GetLatestByWorkflowAndType`
- `DomainTaskRepo`
  - `GetByWorkflowAndID`
  - `ListByWorkflowAndStatuses`
  - `ListCompletedByWorkflowAndKinds`
  - `ListByIDs`
  - `FindLatestByWorkflowAndAffectedResourceLike`
  - `MarkFailedForRework`
- `VerificationIssueRepo`
  - `ListOpenByVerificationRunAndIDs`
  - `UpdateDomainTaskID`
  - `MarkReworkRequested`
- `AcceptIssueRepo`
  - `ListOpenByAcceptRunAndIDs`
- `PlanVersionRepo`
  - `ListByProjectStatuses`
  - `GetLatestByProjectStatusAndReviewStatus`
  - `RestoreRejectedForManualApprove`
- `BlueprintRepo`
  - `CountByPlanVersion`
- `ReviewIssueRepo`
  - `CountOpenByStageRunAndSeverity`
  - `ListByStageRun`
  - `ListOpenByStageRunAndIDs`
- `StageTaskRepo`
  - `ListByStageRun`
- `TaskWorkspaceRepo`
  - `ListDeliveriesByWorkflow`

## 3. 验证结果

本轮已经通过的回归：

```bash
go test ./app/mvp/internal/workflow/acceptance ./app/mvp/internal/workflow/verification ./app/mvp/internal/workflow/stage/accept ./app/mvp/internal/workflow/repo
go test ./app/mvp/internal/workflow ./app/mvp/internal/workflow/presetutil
go test ./app/mvp/internal/controller/chat
```

说明：

- 当前暂停点不是“代码未验证”，而是“还有大量未收口区域”
- 以上通过结果覆盖了本轮已经 repo 化的验收、验证、评审控制层链路

## 4. 还没做完的范围

### 4.1 口径说明

`repo` 目录内允许使用 `g.DB()`，因为那本来就是数据访问层。

真正仍然未完成的是：

- `controller`
- `workflow` 编排层
- `stage`
- `orchestrator`
- `autonomy`
- `scheduler`
- 其他被铁律禁止直接访问 DB 的业务层

### 4.2 剩余数量

截至本次暂停：

- `admin-go/app/mvp/internal/controller/chat` 中剩余 `g.DB()`：`178`
- `admin-go/app/mvp/internal/workflow` 中非 `repo` 目录剩余 `g.DB()`：`251`
- 两者合计的禁区剩余 `g.DB()`：`429`

如果把 `repo` 目录也一起算上，则 `workflow + controller/chat` 总量仍有 `565`，但这不是同一口径，不应混淆。

## 5. 未完成高优先级文件

以下是当前最应该继续清理的文件，按剩余 `g.DB()` 数量排序：

1. `admin-go/app/mvp/internal/controller/chat/workflow.go`：`41`
2. `admin-go/app/mvp/internal/controller/chat/feishu_bot.go`：`40`
3. `admin-go/app/mvp/internal/workflow/stage/review/service.go`：`30`
4. `admin-go/app/mvp/internal/controller/chat/workflow_trace.go`：`26`
5. `admin-go/app/mvp/internal/workflow/domain/plan/plan_version_service.go`：`24`
6. `admin-go/app/mvp/internal/workflow/scheduler/domain_task_scheduler.go`：`17`
7. `admin-go/app/mvp/internal/workflow/orchestrator/registry.go`：`17`
8. `admin-go/app/mvp/internal/controller/chat/workflow_system_check.go`：`16`
9. `admin-go/app/mvp/internal/workflow/orchestrator/workflow_service.go`：`15`
10. `admin-go/app/mvp/internal/workflow/orchestrator/stage_service.go`：`15`
11. `admin-go/app/mvp/internal/workflow/stage/rework/service.go`：`13`
12. `admin-go/app/mvp/internal/workflow/stage/execute/service.go`：`13`
13. `admin-go/app/mvp/internal/workflow/stage/complete/service.go`：`11`
14. `admin-go/app/mvp/internal/controller/chat/workflow_timeline.go`：`9`
15. `admin-go/app/mvp/internal/controller/chat/workflow_runtime.go`：`9`
16. `admin-go/app/mvp/internal/controller/chat/workflow_execution.go`：`9`
17. `admin-go/app/mvp/internal/controller/chat/workflow_autonomy.go`：`9`

## 6. 推荐恢复顺序

如果后续继续推进，建议严格按以下顺序恢复：

1. `workflow.go`
   - 原因：这是多个控制器共享的基础入口，继续收口后能带动 `trace / timeline / runtime / execution` 一起降成本
2. `stage/review/service.go`
   - 原因：评审阶段服务仍是主链，且和已收口的 `workflow_review.go` 紧密相关
3. `workflow_runtime.go / workflow_execution.go / workflow_timeline.go / workflow_trace.go`
   - 原因：这是控制台运行时与验收视图主链，属于高频读接口
4. `feishu_bot.go`
   - 原因：文件大、耦合高，但属于外部入口，应该放在主链稳定后再拆
5. `orchestrator/* + scheduler/*`
   - 原因：这是更深层的编排内核，适合在控制层主要入口收口后做系统化治理

## 7. 暂停时的明确结论

当前可以明确下结论的只有这几条：

- “禁止直接 DB”的铁律已经从文档约束，变成了部分主链上的真实代码约束
- 验收、验证、评审三条关键链路已经开始按统一方向收口
- 但整个 EasyMVP 仍然没有达到“所有禁区层全部接口化”的完成态
- 因此此项工作目前不能被判定为完成，只能判定为“进行中且已有稳定暂停点”

## 8. 恢复工作前的检查项

恢复本项工作时，先做以下检查：

1. 重新统计禁区层剩余 `g.DB()` 数量，确认暂停期间是否有新增回流
2. 优先检查 `workflow.go` 是否又新增了新的直连查询
3. 跑一遍：
   - `go test ./app/mvp/internal/controller/chat`
   - `go test ./app/mvp/internal/workflow/acceptance ./app/mvp/internal/workflow/verification ./app/mvp/internal/workflow/stage/accept ./app/mvp/internal/workflow/repo`
4. 再开始下一轮 repo 收口

否则很容易出现“新代码回流直连 DB，旧代码在清理，净进度反而变慢”的情况。
