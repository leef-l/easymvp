# WorkflowRun 前端页面与事件协议设计文档

> 文档定位：`WorkflowRun阶段化工作流引擎重构架构设计文档` 的前端实施文档
>
> 目标：定义新工作流架构下的页面结构、交互模型、状态展示模型、SSE 事件协议与前后端契约。

---

## 一、设计目标

当前前端的主要问题是：

1. 聊天页承担了过多流程语义。
2. 项目页只能粗粒度展示状态，无法表达阶段过程。
3. 审核、执行、返工、计划版本没有独立视图。
4. SSE 主要围绕 message 流，缺少 workflow/stage 级别事件通道。

本次前端重构目标：

1. 将"会话视图"和"流程视图"彻底分开。
2. 围绕 `workflow_run / stage_run / plan_version / domain_task` 重建页面结构。
3. 引入统一事件协议，支撑实时状态刷新。
4. 保证新旧项目在过渡期可共存。

---

## 二、目标页面结构

### 2.1 总体信息架构

建议将现有 `mvp` 前端页面结构扩展为：

```text
views/mvp/
  project/
    index.vue
    modules/
      workflow-run-drawer.vue
  workflow/
    dashboard.vue
    timeline.vue
  plan/
    design.vue
    version-list.vue
    blueprint-table.vue
    diff-panel.vue
  review/
    workspace.vue
    issue-list.vue
    stage-task-panel.vue
    summary-card.vue
  execution/
    console.vue
    batch-board.vue
    resource-lock-panel.vue
    task-chain-drawer.vue
  rework/
    workspace.vue
  chat/
    index.vue
    conversation.vue
```

### 2.2 页面角色划分

#### 项目页

职责：

- 显示项目列表
- 显示当前活跃 `workflow_run`
- 提供进入设计、审核、执行控制台的入口

#### Workflow Dashboard

职责：

- 展示当前工作流概况
- 显示当前阶段、总体进度、最近事件、关键风险

#### Plan Design

职责：

- 展示架构师对话
- 展示 `plan_version`
- 展示任务蓝图
- 展示版本 diff

#### Review Workspace

职责：

- 展示 `review stage`
- 展示 precheck / auditor / coordinator 的阶段任务结果
- 展示 `review_issue`
- 触发重新审核、人工通过、人工驳回

#### Execution Console

职责：

- 展示执行批次
- 展示运行中任务
- 展示资源锁
- 展示任务链与返工链

#### Rework Workspace

职责：

- 展示返工链
- 展示 bug/failure 分析与派发情况

#### Chat

职责：

- 只负责消息会话
- 不再承担流程主控

---

## 三、路由设计

### 3.1 建议路由

```text
/mvp/project
/mvp/workflow/dashboard?projectId=...&workflowRunId=...
/mvp/workflow/timeline?workflowRunId=...
/mvp/plan/design?workflowRunId=...&planVersionId=...
/mvp/review/workspace?workflowRunId=...&stageRunId=...
/mvp/execution/console?workflowRunId=...&stageRunId=...
/mvp/rework/workspace?workflowRunId=...&stageRunId=...
/mvp/chat?conversationId=...
```

### 3.2 路由进入规则

1. 进入项目页后，优先显示活跃工作流摘要。
2. 点击"继续设计"进入 plan design。
3. 点击"审核工作台"进入 review workspace。
4. 点击"执行控制台"进入 execution console。
5. 聊天按钮只打开 conversation，不再驱动流程状态切换。

---

## 四、页面级数据模型

### 4.1 项目列表项

前端 DTO 建议：

```ts
interface ProjectWorkflowSummary {
  projectId: string;
  projectName: string;
  engineVersion: 'legacy' | 'workflow_v2';
  activeWorkflowRunId?: string;
  workflowStatus?: string;
  currentStage?: string;
  progressPercent?: number;
  latestEventAt?: string;
  riskLevel?: 'low' | 'medium' | 'high';
}
```

### 4.2 Workflow 概览 DTO

```ts
interface WorkflowRunDetail {
  workflowRunId: string;
  projectId: string;
  status: string;
  currentStage: string;
  currentStageRunId?: string;
  activePlanVersionId?: string;
  startedAt?: string;
  finishedAt?: string;
  stageProgress: {
    design: string;
    review: string;
    execute: string;
    rework: string;
    complete: string;
  };
  metrics: {
    totalTasks: number;
    completedTasks: number;
    failedTasks: number;
    reworkRounds: number;
  };
}
```

### 4.3 Review Issue DTO

```ts
interface ReviewIssueItem {
  id: string;
  severity: 'error' | 'warning' | 'info';
  issueCode: string;
  issueType: string;
  taskName?: string;
  sourceRole: string;
  message: string;
  suggestion?: string;
  status: string;
}
```

### 4.4 Domain Task DTO

```ts
interface DomainTaskItem {
  id: string;
  taskKind: string;
  name: string;
  roleType: string;
  roleLevel: string;
  batchNo: number;
  status: string;
  retryCount: number;
  sourceTaskId?: string;
  rootTaskId?: string;
  affectedResources: string[];
}
```

---

## 五、页面设计细节

### 5.1 项目列表页

#### 列设计建议

- 项目名称
- 引擎版本
- 当前工作流状态
- 当前阶段
- 总体进度
- 风险等级
- 最近活跃时间
- 操作

#### 操作建议

- 查看工作流
- 进入设计
- 进入审核
- 进入执行
- 暂停
- 恢复
- 取消

### 5.2 Workflow Dashboard

#### 区块建议

1. 顶部状态卡
2. 阶段时间线
3. 关键指标卡
4. 最近事件流
5. 风险与告警面板

#### 交互建议

- 当前阶段卡片可直接跳转到对应工作台
- 最近事件可过滤 `stage/task/review`

### 5.3 Plan Design

#### 左侧

- 架构师会话列表
- 计划版本列表

#### 中间

- 当前版本任务蓝图表

#### 右侧

- 版本摘要
- 与上一版本 diff
- 提交审核按钮

#### 蓝图表列建议

- 名称
- 角色
- 批次
- 依赖
- 影响资源
- 当前蓝图状态

### 5.4 Review Workspace

#### 顶部

- 审核阶段状态
- 审核结论
- 审核耗时

#### 主体

1. 阶段任务卡片区
2. Issue 列表区
3. 版本修订建议区

#### 阶段任务卡片

- `precheck`
- `auditor_review`
- `coordinator_optimize`
- `review_summary`

卡片展示：

- 状态
- 开始时间
- 结束时间
- 耗时
- 错误信息
- 降级说明

#### Issue 列表交互

- severity 筛选
- issue_type 筛选
- taskName 搜索
- 解决状态筛选

### 5.5 Execution Console

#### 布局建议

1. 批次泳道
2. 运行队列
3. 资源锁面板
4. 任务链追踪抽屉

#### 批次泳道

按 `batch_no` 分组展示：

- pending
- running
- completed
- failed

#### 资源锁面板

展示：

- 资源路径
- 占用任务
- 锁定时长
- 锁状态

#### 任务链抽屉

展示：

- root task
- audit task
- bug_analysis
- failure_analysis
- handoff timeline

### 5.6 Rework Workspace

#### 展示重点

- 当前返工阶段状态
- 当前返工链
- 最近一次架构师分析结果
- 最近一次派发记录

---

## 六、组件设计建议

### 6.1 可复用组件

建议新增：

- `WorkflowStatusTag`
- `StageTimeline`
- `ReviewIssueTable`
- `PlanVersionSwitcher`
- `BlueprintTable`
- `ResourceLockList`
- `TaskChainTree`
- `EventFeed`

### 6.2 组件职责边界

1. 列表组件只负责展示。
2. 状态推进由页面级 action 触发。
3. SSE 事件接入应集中在 page-level composable。

---

## 七、前端状态管理设计

### 7.1 Store 划分

建议新增：

- `useWorkflowStore`
- `usePlanStore`
- `useReviewStore`
- `useExecutionStore`
- `useWorkflowEventStore`

### 7.2 Store 职责

#### useWorkflowStore

- 当前工作流详情
- 当前阶段
- 顶层状态切换

#### usePlanStore

- planVersion 列表
- 当前版本详情
- 蓝图列表

#### useReviewStore

- review issue 列表
- stage_task 状态
- 审核结论

#### useExecutionStore

- task list
- batch board
- resource lock

#### useWorkflowEventStore

- 最近事件流
- SSE 连接状态
- 事件去重

---

## 八、事件协议设计

### 8.1 事件流分层

建议定义三类事件流：

1. Workflow Event Stream
2. Conversation Event Stream
3. Message Stream

### 8.2 Workflow Event Stream

接口：

`GET /mvp/workflow/events?workflowRunID=...`

事件载荷格式：

```json
{
  "event_id": "123",
  "event_type": "stage.started",
  "workflow_run_id": "1001",
  "stage_run_id": "2001",
  "entity_type": "stage_run",
  "entity_id": "2001",
  "created_at": "2026-04-04 21:00:00",
  "payload": {
    "stage_type": "review",
    "status": "running"
  }
}
```

### 8.3 Conversation Event Stream

接口：

`GET /mvp/conversation/events?conversationID=...`

用途：

- 会话级元事件
- 对话被系统接管
- 会话绑定到 plan version / task / stage

### 8.4 Message Stream

接口：

`GET /mvp/chat/sse?messageID=...`

仍保留，用于纯消息流式输出。

### 8.5 事件类型规范

#### Workflow 级

- `workflow.created`
- `workflow.status_changed`
- `workflow.paused`
- `workflow.resumed`
- `workflow.canceled`
- `workflow.completed`

#### Stage 级

- `stage.created`
- `stage.started`
- `stage.completed`
- `stage.failed`
- `stage.skipped`

#### Plan 级

- `plan_version.created`
- `plan_version.submitted`
- `plan_version.approved`
- `plan_version.rejected`
- `plan_version.superseded`

#### Review 级

- `review.issue_created`
- `review.issue_resolved`
- `review.decision_ready`

#### Task 级

- `task.created`
- `task.started`
- `task.completed`
- `task.failed`
- `task.escalated`
- `task.retried`
- `task.skipped`

#### Resource 级

- `resource.locked`
- `resource.released`
- `resource.lock_leaked`

---

## 九、SSE 解析规则

### 9.1 协议格式

继续使用标准 SSE：

```text
event: workflow.status_changed
data: {...}

event: stage.started
data: {...}
```

### 9.2 前端解析要求

前端必须：

1. 同时解析 `event:` 和 `data:`
2. 保留 `event_type`
3. 保留 `event_id`
4. 去重处理
5. 支持断线重连

### 9.3 重连机制

建议支持：

- `Last-Event-ID`
- 重连后按 `event_id > lastSeenId` 补拉

### 9.4 事件去重

Store 层维护：

- `lastEventIDByStream`
- 最近 200 条事件 id 集合

---

## 十、前后端接口契约

### 10.1 页面初始化加载

Workflow Dashboard 初始化建议并行请求：

1. workflow detail
2. stage list
3. recent events
4. metrics summary

### 10.2 Review Workspace 初始化加载

1. stage detail
2. stage_task list
3. review issue list
4. review summary

### 10.3 Execution Console 初始化加载

1. domain task list
2. batch board
3. resource lock list
4. recent task events

---

## 十一、兼容期页面策略

### 11.1 旧项目

旧项目仍显示旧页面入口：

- 项目状态
- 旧任务列表
- 旧聊天页

### 11.2 新项目

新项目显示新页面入口：

- Workflow Dashboard
- Plan Design
- Review Workspace
- Execution Console

### 11.3 兼容提示

项目页建议显式显示：

- `Legacy Engine`
- `Workflow V2`

避免用户混淆。

---

## 十二、实施顺序

### 12.1 第一阶段

- 项目页接入 `engineVersion`
- Workflow Dashboard 基础版
- 新 SSE workflow event 通道

### 12.2 第二阶段

- Plan Design 页面
- Review Workspace 页面
- review issue 表格与阶段任务卡片

### 12.3 第三阶段

- Execution Console
- Resource Lock 面板
- Task Chain 抽屉

### 12.4 第四阶段

- Rework Workspace
- Timeline 页
- Event Feed 高级筛选

---

## 十三、验收标准

### 13.1 可视化验收

1. 新项目能完整看到 workflow/stage 信息
2. 审核问题不再依赖 description 拼接展示
3. 执行过程可实时显示任务状态变化
4. 返工链可视化可追溯

### 13.2 协议验收

1. workflow SSE 能稳定重连
2. 事件不重复
3. 页面刷新后可通过补拉恢复状态

---

## 十四、结论

前端重构的关键不是"多几个页面"，而是把系统观察视角从"消息中心"切到"工作流中心"。

一旦页面结构和事件协议完成重构，后端新工作流内核的价值才能真正被用户感知。
