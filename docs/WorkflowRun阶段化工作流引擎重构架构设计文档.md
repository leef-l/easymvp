# WorkflowRun 阶段化工作流引擎重构架构设计文档

> 更新日期：2026-04-08
>
> 说明：文件名保留为“重构架构设计文档”，但本文内容已经切换为当前 `Workflow V2` 的落地实现说明。

## 1. 总览

当前 EasyMVP 的正式执行主链已经是阶段化工作流：

```text
design -> review -> execute -> accept -> complete
                         \-> rework -> execute / accept
```

其中：

- `project` 负责业务容器和展示信息
- `workflow_run` 是正式运行实例
- `stage_run` 是阶段实例
- `plan_version` / `task_blueprint` 负责设计产物
- `domain_task` 负责执行期任务

虽然创建项目 API 里仍保留 `engineVersion` 兼容字段，但前端主流程、数据库和服务编排都已经围绕 `workflow_v2` 展开。

## 2. 运行实体

当前实现里的主要实体关系如下：

```text
Project
  └── WorkflowRun
        ├── StageRun(design)
        │     ├── Conversation / Message
        │     └── PlanVersion -> TaskBlueprint
        ├── StageRun(review)
        │     ├── StageTask
        │     └── ReviewIssue
        ├── StageRun(execute)
        │     └── DomainTask -> TaskWorkspace / TaskLog
        ├── StageRun(accept)
        │     └── AcceptRun -> AcceptIssue / AcceptEvidence
        ├── StageRun(rework)
        │     └── FailureAnalysis / HandoffRecord
        └── StageRun(complete)
```

当前数据库中的关键表：

- `mvp_workflow_run`
- `mvp_stage_run`
- `mvp_plan_version`
- `mvp_task_blueprint`
- `mvp_domain_task`
- `mvp_review_issue`
- `mvp_accept_run`
- `mvp_handoff_record`
- `mvp_workflow_event`
- `mvp_task_workspace`

## 3. 阶段链路

### 3.1 设计阶段 `design`

入口：

- 新建项目
- 与架构师对话
- 手动检查拆分 / 解析任务

当前行为：

- 架构师回复可被解析成 `plan_version`
- `plan_version` 继续展开为 `task_blueprint`
- 项目仍可多轮对话修订，不会直接进入执行

### 3.2 审核阶段 `review`

当用户确认方案后，工作流进入审核阶段。当前实现会落地：

- 预检
- 审计员审核
- 协调员优化
- `mvp_review_issue` 问题列表

审核阶段只有在 `error=0` 且 `warning=0` 时才会触发执行阶段。
如果仍存在 warning，系统会把完整问题清单回发给架构师对话，并接受两类修订结果：

- 完整 `{"tasks": [...]}`：生成新的 `plan_version`
- 局部 `{"task_patches": [...]}`：直接回写当前 `task_blueprint`

如果架构师因为输出过长而分段回复，系统只会在上一条系统提示里明确约定了 `[AUTO_CONTINUE_NEXT]` 续传标记时，才根据该标记自动继续索取后续分段；普通自然语言“请继续”不会触发自动续取。收齐后仍会把同一轮回复合并后再解析。

### 3.3 执行阶段 `execute`

执行阶段由 `stage/execute/service.go` 负责：

1. 把 `task_blueprint` 实例化为 `mvp_domain_task`
2. 启动 `DomainTaskScheduler`
3. 根据 `execution_mode` 分发到统一执行器注册表
4. 写入任务状态、结果、日志和事件

执行期已经具备：

- 依赖和批次调度
- 资源锁
- 心跳与看门狗
- 写仓任务的 `git worktree` 隔离

### 3.4 验收阶段 `accept`

执行阶段结束后会进入自动验收：

1. 创建 `accept_run`
2. 收集证据
3. 运行规则引擎
4. 通过 `DecisionReducer` 汇总裁决
5. 决定直接完成、进入返工，或保持人工验收

当前验收链路同时支持：

- 规则评估
- LLM Judge 灰度注入
- 人工放行 / 驳回 / 重验

### 3.5 返工阶段 `rework`

返工阶段负责围绕失败任务或验收失败结果进行补救，主要能力包括：

- 失败分析任务
- 重派发
- 返工完成后恢复执行或再次进入验收

### 3.6 完成阶段 `complete`

完成阶段负责关闭工作流和汇总结果。当前前端的工作流仪表盘、时间线和通知能力都依赖这条闭环链路。

## 4. 编排与状态推进

当前阶段推进不是散落在控制器里，而是在 `internal/workflow/orchestrator/registry.go` 集中装配：

- `review -> execute`
- `accept -> complete`
- `accept -> rework`
- `rework -> execute`
- `rework -> accept`

也就是说，阶段推进的关键规则已经从“文档方案”变成“代码回调注册”。

工作流常见状态：

- `designing`
- `reviewing`
- `executing`
- `accepting`
- `reworking`
- `paused`
- `completed`
- `failed`
- `canceled`

阶段常见状态：

- `pending`
- `running`
- `completed`
- `failed`
- `skipped`

## 5. 观测、事件与前端

当前 Workflow V2 的运行数据会同时进入多个面向：

- `mvp_workflow_event`：时间线和事件流
- `mvp_task_log`：任务日志
- SSE：对话流输出
- 前端工作流页：审核、执行、验收、返工、自治、时间线

对应前端页面位于：

- `views/mvp/workflow/dashboard.vue`
- `views/mvp/workflow/review.vue`
- `views/mvp/workflow/execution.vue`
- `views/mvp/workflow/accept.vue`
- `views/mvp/workflow/rework.vue`
- `views/mvp/workflow/autonomy.vue`
- `views/mvp/workflow/timeline.vue`

## 6. 自治介入点

当前自治模块已经不是独立方案稿，而是工作流主链上的实际扩展点：

- `DecisionCenter`
- `PolicyEngine`
- `RiskGate`
- `Sensor`
- `ObjectiveService`
- `Planner`
- `Actuator`
- `MetaObserver`
- `Learner`

实际介入的触发点包括：

- `accept.passed`
- `accept.failed`
- `accept.manual_review`
- `rework.completed`

对应前端页还提供：

- `objective.vue`
- `situation.vue`
- `meta-cognition.vue`

## 7. 当前边界

更新本文档时确认的几个边界如下：

- `engineVersion` 兼容字段仍存在，但不再是文档主线
- `internal/engine` 仍保留一部分对话、模型解析和兼容辅助逻辑，未完全剥离
- 旧的迁移 Phase、数据库灰度计划和旧链切换说明已不再保留在仓库文档中
