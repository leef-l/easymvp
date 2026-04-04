# WorkflowRun 后端模块拆分与目录重构设计文档

> 文档定位：`WorkflowRun阶段化工作流引擎重构架构设计文档` 的后端落地设计
>
> 目标：将当前 `engine` 聚合实现，拆分为清晰的工作流内核、阶段服务、任务调度、事件总线和兼容层。

---

## 一、现状问题

当前 `admin-go/app/mvp/internal/engine/` 同时承载：

- 流程编排
- 审核
- 调度
- 执行
- Bug 闭环
- SSE
- 上下文压缩
- 配置
- 路径校验

这带来几个问题：

1. 模块边界不清晰
2. 状态流转分散
3. 新旧架构难并存
4. 审核链和执行链耦合过深
5. 很难按阶段做灰度和替换

---

## 二、重构目标

### 2.1 核心目标

将现有 `engine` 拆分成以下层级：

1. 运行时内核层
2. 阶段服务层
3. 领域任务层
4. 事件与通知层
5. 兼容适配层

### 2.2 设计原则

1. 目录结构必须体现运行实体边界。
2. 工作流阶段不能再直接耦合具体任务执行器实现。
3. 新架构与旧架构必须可并存。
4. 所有状态流转必须通过统一 service。

---

## 三、目标目录结构

建议新增目录结构如下：

```text
admin-go/app/mvp/internal/
  workflow/
    runtime/
      runtime.go
      manager.go
      registry.go
    orchestrator/
      workflow_service.go
      stage_service.go
      transition.go
    stage/
      design/
        service.go
        parser.go
        plan_builder.go
      review/
        service.go
        precheck.go
        auditor.go
        coordinator.go
        summary.go
      execute/
        service.go
        instantiate.go
      rework/
        service.go
        bug_analysis.go
        failure_analysis.go
        dispatch.go
      complete/
        service.go
        summary.go
    scheduler/
      stage_scheduler.go
      domain_task_scheduler.go
      resource_lock.go
      dependency_graph.go
    domain/
      plan/
        plan_version_service.go
        blueprint_service.go
      task/
        task_service.go
        task_state_machine.go
        handoff_service.go
      review/
        issue_service.go
        decision_service.go
    event/
      bus.go
      publisher.go
      workflow_events.go
      sse_bridge.go
    repo/
      workflow_run_repo.go
      stage_run_repo.go
      plan_version_repo.go
      blueprint_repo.go
      domain_task_repo.go
      review_issue_repo.go
    compat/
      legacy_gateway.go
      project_status_adapter.go
      task_adapter.go
```

旧 `engine/` 保留，但逐步只作为兼容入口。

---

## 四、模块职责

### 4.1 `workflow/runtime`

职责：

- 管理 `workflow_run` 级别 runtime
- 创建 `context`
- 暂停、恢复、取消
- runtime registry

核心对象：

```go
type WorkflowRuntime struct {
    WorkflowRunID int64
    ProjectID     int64
    Ctx           context.Context
    Cancel        context.CancelFunc
}
```

### 4.2 `workflow/orchestrator`

职责：

- 驱动工作流阶段切换
- 启动和结束 `stage_run`
- 检查阶段前置条件
- 处理阶段失败与回退

这是整个新内核的"总协调器"。

### 4.3 `workflow/stage/design`

职责：

- 管理架构师对话产物
- 生成 `plan_version`
- 解析任务蓝图

### 4.4 `workflow/stage/review`

职责：

- 驱动审核阶段
- 创建 `stage_task`
- 聚合 `review_issue`
- 得出审核结论

### 4.5 `workflow/stage/execute`

职责：

- 将审核通过的 `task_blueprint` 实例化为 `domain_task`
- 启动执行阶段

### 4.6 `workflow/stage/rework`

职责：

- 统一承接 bug 修复与失败升级
- 生成返工相关任务
- 回到执行阶段

### 4.7 `workflow/stage/complete`

职责：

- 总结
- 归档
- 指标计算

### 4.8 `workflow/scheduler`

拆分为：

- `stage_scheduler.go`
- `domain_task_scheduler.go`

其中：

- `stage_scheduler` 只调阶段任务
- `domain_task_scheduler` 只调执行任务

### 4.9 `workflow/domain`

职责：

- 承载领域级 service
- 所有状态机和领域规则都在这里
- 避免直接在 orchestrator 中散落业务判断

### 4.10 `workflow/event`

职责：

- 统一事件发布
- 统一 SSE 桥接
- 输出 workflow/stage/task/review 事件

### 4.11 `workflow/repo`

职责：

- 新架构专用仓储层
- 不与旧 DAO 混写

### 4.12 `workflow/compat`

职责：

- 新旧项目兼容
- 将旧 DTO 适配到新结构
- 降低 API 切换成本

---

## 五、核心服务设计

### 5.1 WorkflowService

关键职责：

- `CreateRun(projectID)`
- `StartDesign(workflowRunID)`
- `SubmitPlan(workflowRunID, planVersionID)`
- `Pause(workflowRunID)`
- `Resume(workflowRunID)`
- `Cancel(workflowRunID)`

### 5.2 StageService

关键职责：

- `StartStage(workflowRunID, stageType)`
- `CompleteStage(stageRunID)`
- `FailStage(stageRunID, reason)`
- `TransitionNext(workflowRunID)`

### 5.3 PlanVersionService

关键职责：

- `CreateFromArchitectReply(...)`
- `SupersedePreviousVersions(projectID)`
- `SubmitForReview(planVersionID)`
- `Approve(planVersionID)`
- `Reject(planVersionID)`

### 5.4 ReviewIssueService

关键职责：

- `BatchCreateIssues(...)`
- `ListByPlanVersion(planVersionID)`
- `ResolveIssue(issueID)`
- `SummarizeIssues(planVersionID)`

### 5.5 DomainTaskService

关键职责：

- `InstantiateFromBlueprint(planVersionID)`
- `Retry(taskID)`
- `Skip(taskID)`
- `Escalate(taskID)`
- `DispatchRework(taskID)`

---

## 六、状态流转统一入口

### 6.1 原则

不允许再在各个模块里直接 `Update(status=...)`。

必须提供三套统一入口：

1. `UpdateWorkflowStatus()`
2. `UpdateStageStatus()`
3. `UpdateDomainTaskStatus()`

### 6.2 作用

- 统一校验状态迁移
- 统一写事件
- 统一 CAS
- 统一告警

---

## 七、兼容层设计

### 7.1 为什么要兼容层

因为现有前端和 API 已经围绕旧模型构建，不能一次全部推翻。

### 7.2 LegacyGateway

作用：

- 读取 `project.engine_version`
- 决定调用旧链路还是新链路

### 7.3 ProjectStatusAdapter

作用：

- 将 `workflow_run + stage_run + domain_task` 聚合成旧的项目状态 DTO

### 7.4 TaskAdapter

作用：

- 将 `domain_task` 适配成旧任务列表响应

---

## 八、事件总线与 SSE 桥接

### 8.1 目标

将当前"消息流"和"项目状态变化"分离。

### 8.2 事件类型

建议定义：

- `workflow.created`
- `workflow.status_changed`
- `stage.started`
- `stage.completed`
- `stage.failed`
- `plan_version.created`
- `plan_version.submitted`
- `review.issue_created`
- `review.approved`
- `review.rejected`
- `task.created`
- `task.started`
- `task.completed`
- `task.failed`
- `task.escalated`

### 8.3 SSE 分层

建议三个 SSE 入口：

1. `message stream`
2. `workflow event stream`
3. `conversation event stream`

---

## 九、旧 engine 的处理策略

### 9.1 旧 `engine/` 不立即删除

过渡期：

- 保留 `engine/`
- 旧项目继续使用
- 新项目从 controller 层路由到 `workflow/`

### 9.2 逐步瘦身

旧 `engine` 逐步移除：

- 审核逻辑
- 流程编排
- 新项目调度入口

最终只保留：

- 兼容桥接
- 存量项目维护

---

## 十、实施顺序

### 10.1 第一步：建新目录与接口骨架

目标：

- 完成目录搭建
- 定义 service 接口
- 定义 repo 接口

### 10.2 第二步：接入 runtime + orchestrator

目标：

- `workflow_run` 创建
- `stage_run` 切换
- runtime registry

### 10.3 第三步：设计与审核模块切换

目标：

- `plan_version`
- `task_blueprint`
- `review_issue`
- `stage_task`

### 10.4 第四步：执行阶段切换

目标：

- `domain_task`
- 新任务调度器
- 新看门狗

### 10.5 第五步：事件与前端

目标：

- workflow event SSE
- 状态聚合接口
- 新控制台页面

---

## 十一、测试策略

### 11.1 单元测试

重点覆盖：

- 状态机
- orchestrator 转移规则
- review decision
- blueprint -> domain_task 实例化

### 11.2 集成测试

重点覆盖：

- 新项目完整生命周期
- 审核通过
- 审核拒绝
- 执行失败返工
- 暂停恢复

### 11.3 回归测试

重点覆盖：

- legacy 项目不受影响
- 新旧项目列表同页展示正常
- 聊天页消息流不受影响

---

## 十二、结论

后端模块拆分的重点不是"把代码挪目录"，而是把运行职责真正拆开：

1. runtime 负责运行时
2. orchestrator 负责流程推进
3. stage 负责阶段实现
4. domain 负责业务规则
5. scheduler 负责调度
6. event 负责事件
7. compat 负责新旧并存

只有这样，WorkflowRun 架构才不会变成"旧 engine 上再套一层皮"。
