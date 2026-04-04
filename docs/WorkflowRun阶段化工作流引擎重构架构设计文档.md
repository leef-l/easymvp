# WorkflowRun 阶段化工作流引擎重构架构设计文档

> 文档定位：EasyMVP 下一阶段核心架构蓝图
>
> 目标：将当前“项目 + 任务调度 + AI 角色协作”系统，升级为“可多阶段运行、可审核、可回放、可恢复、可统计”的正式工作流引擎。
>
> 适用范围：后端工作流内核、数据库模型、调度器、前端流程页、运行时事件总线、迁移方案。

---

## 一、背景与设计目标

### 1.1 当前系统的阶段

当前 EasyMVP 已经具备以下能力：

- 用户与架构师对话拆分任务
- 任务草稿生成与确认
- 审核阶段 `reviewing`
- 执行阶段 `running`
- 审计、Bug 闭环、失败升级
- 资源锁、看门狗、SSE 实时推送

这说明系统已经从“单轮对话驱动的 AI 助手”演进为“多角色协作的自动化开发系统”。

### 1.2 当前系统的根本矛盾

当前代码库的主要问题，不再是单个功能缺失，而是架构层级不够清晰：

1. `project`、`task`、`conversation`、`message` 被直接拼成了执行系统，缺少正式的“运行实例”模型。
2. 审核阶段虽然存在，但仍偏函数式，不属于统一调度体系。
3. 任务草稿、审核问题、执行任务、返工任务混在少量表和状态字段里，模型边界不清。
4. 架构师多轮修订任务时，仍偏“覆盖重建”，不适合长期协作。
5. 统计、回放、暂停、恢复、重试、流程可视化等能力缺少统一运行实体支撑。

### 1.3 本次重构目标

本次重构不是补丁式优化，而是一步到位地引入正式工作流内核。

目标如下：

1. 引入 `WorkflowRun` 作为项目执行实例。
2. 引入 `StageRun` 作为阶段运行实例。
3. 将设计、审核、执行、返工、完成纳入统一阶段模型。
4. 将“任务蓝图”和“执行任务”分离。
5. 将审核问题从任务描述中剥离，独立建模。
6. 将调度器分层，区分“阶段调度”和“领域任务调度”。
7. 将上下文取消、暂停恢复、SSE 事件、日志、统计统一到运行实体。
8. 保持可灰度迁移，不一次性推翻存量项目。

---

## 二、目标架构总览

### 2.1 核心架构图

```text
Project
  └── WorkflowRun（一次完整执行实例）
        ├── StageRun: Design
        │     ├── Architect Conversation
        │     └── PlanVersion
        ├── StageRun: Review
        │     ├── PrecheckTask
        │     ├── AuditorReviewTask
        │     ├── CoordinatorOptimizeTask
        │     └── ReviewIssue / ReviewDecision
        ├── StageRun: Execute
        │     ├── DomainTask（Implement/Audit）
        │     ├── Scheduler / Resource Lock
        │     └── TaskResult / ContextSummary
        ├── StageRun: Rework
        │     ├── BugAnalysisTask
        │     ├── FailureAnalysisTask
        │     └── Re-dispatch
        └── StageRun: Complete
              └── Summary / Metrics / Archive
```

### 2.2 设计原则

1. `Project` 是业务容器，不是执行状态机本体。
2. `WorkflowRun` 是唯一正式运行状态来源。
3. `StageRun` 是流程控制的最小单位。
4. `PlanVersion` 负责设计产物版本化。
5. `DomainTask` 只承载审核通过后的执行任务。
6. 审核问题、失败问题、返工原因必须独立建模。
7. 统一使用事件驱动 + 状态机 + 可取消运行时模型。

---

## 三、领域模型重构

### 3.1 现有模型问题

当前主要模型为：

- `mvp_project`
- `mvp_task`
- `mvp_conversation`
- `mvp_message`
- `mvp_task_log`

这套模型的问题是：

1. 缺少“运行实例”。
2. 任务既承担设计草稿，又承担执行任务，又承担返工任务。
3. 审核问题没有正式实体。
4. 多轮设计修订没有版本模型。

### 3.2 目标模型分层

重构后分为五层：

1. 业务层：`project`
2. 运行层：`workflow_run`、`stage_run`
3. 计划层：`plan_version`、`task_blueprint`
4. 执行层：`domain_task`
5. 问题层：`review_issue`、`task_issue`、`handoff_record`

### 3.3 模型职责

#### 3.3.1 Project

职责：

- 保存业务元数据
- 保存默认配置
- 不再承担完整执行状态机

#### 3.3.2 WorkflowRun

职责：

- 表示项目的一次完整执行
- 绑定项目级运行时上下文
- 管理当前阶段
- 支持暂停、恢复、取消、完成、失败

#### 3.3.3 StageRun

职责：

- 表示一次阶段执行
- 保存阶段输入、输出、耗时、结论
- 支持阶段级重试、跳过、回退

#### 3.3.4 PlanVersion

职责：

- 保存架构师某一轮任务规划结果
- 支持审核前后状态变化
- 保留版本 diff、来源消息、审核结论

#### 3.3.5 TaskBlueprint

职责：

- 表示尚未执行的任务蓝图
- 来自某个 `PlanVersion`
- 仅在审核通过后实例化为执行任务

#### 3.3.6 DomainTask

职责：

- 表示执行期真正参与调度的任务
- 包含 implement/audit/bug_analysis/failure_analysis 等
- 统一资源锁、重试、审计、返工链

#### 3.3.7 ReviewIssue

职责：

- 保存审核问题
- 与任务描述分离
- 支持 severity、issue_code、来源角色、解决状态

---

## 四、数据库设计

### 4.1 保留表

保留但语义调整：

- `mvp_project`
- `mvp_conversation`
- `mvp_message`
- `mvp_task_log`
- `mvp_project_role`
- `mvp_role_preset`

### 4.2 新增表总览

建议新增以下表：

1. `mvp_workflow_run`
2. `mvp_stage_run`
3. `mvp_plan_version`
4. `mvp_task_blueprint`
5. `mvp_domain_task`
6. `mvp_review_issue`
7. `mvp_stage_task`
8. `mvp_handoff_record`
9. `mvp_workflow_event`
10. `mvp_task_resource_lock`

### 4.3 `mvp_workflow_run`

字段建议：

```sql
CREATE TABLE mvp_workflow_run (
  id bigint unsigned NOT NULL,
  project_id bigint unsigned NOT NULL,
  run_no int NOT NULL,
  status varchar(32) NOT NULL,
  current_stage varchar(32) NOT NULL,
  current_stage_run_id bigint unsigned DEFAULT NULL,
  active_plan_version_id bigint unsigned DEFAULT NULL,
  pause_reason varchar(500) DEFAULT NULL,
  cancel_reason varchar(500) DEFAULT NULL,
  runtime_token varchar(64) DEFAULT NULL,
  started_at datetime DEFAULT NULL,
  finished_at datetime DEFAULT NULL,
  created_by bigint unsigned DEFAULT 0,
  dept_id bigint unsigned DEFAULT 0,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_project_run_no (project_id, run_no),
  KEY idx_project_status (project_id, status)
);
```

状态建议：

- `designing`
- `reviewing`
- `executing`
- `reworking`
- `paused`
- `completed`
- `failed`
- `canceled`

### 4.4 `mvp_stage_run`

字段建议：

```sql
CREATE TABLE mvp_stage_run (
  id bigint unsigned NOT NULL,
  workflow_run_id bigint unsigned NOT NULL,
  stage_type varchar(32) NOT NULL,
  stage_no int NOT NULL,
  status varchar(32) NOT NULL,
  input_ref json DEFAULT NULL,
  output_ref json DEFAULT NULL,
  decision json DEFAULT NULL,
  started_at datetime DEFAULT NULL,
  finished_at datetime DEFAULT NULL,
  error_message text DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_workflow_stage (workflow_run_id, stage_type, stage_no),
  KEY idx_workflow_status (workflow_run_id, status)
);
```

`stage_type`：

- `design`
- `review`
- `execute`
- `rework`
- `complete`

### 4.5 `mvp_plan_version`

字段建议：

```sql
CREATE TABLE mvp_plan_version (
  id bigint unsigned NOT NULL,
  project_id bigint unsigned NOT NULL,
  workflow_run_id bigint unsigned NOT NULL,
  version_no int NOT NULL,
  source_conversation_id bigint unsigned DEFAULT NULL,
  source_message_id bigint unsigned DEFAULT NULL,
  status varchar(32) NOT NULL,
  review_status varchar(32) NOT NULL,
  summary text DEFAULT NULL,
  diff_summary text DEFAULT NULL,
  approved_at datetime DEFAULT NULL,
  rejected_at datetime DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_project_version (project_id, version_no),
  KEY idx_workflow_status (workflow_run_id, status, review_status)
);
```

`status`：

- `draft`
- `submitted`
- `approved`
- `rejected`
- `superseded`

### 4.6 `mvp_task_blueprint`

字段建议：

```sql
CREATE TABLE mvp_task_blueprint (
  id bigint unsigned NOT NULL,
  plan_version_id bigint unsigned NOT NULL,
  parent_blueprint_id bigint unsigned DEFAULT NULL,
  name varchar(255) NOT NULL,
  description text NOT NULL,
  role_type varchar(32) NOT NULL,
  role_level varchar(32) NOT NULL,
  batch_no int NOT NULL,
  sort int NOT NULL DEFAULT 0,
  affected_resources json DEFAULT NULL,
  depends_on_blueprint_ids json DEFAULT NULL,
  blueprint_status varchar(32) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_plan_batch (plan_version_id, batch_no, sort),
  KEY idx_plan_status (plan_version_id, blueprint_status)
);
```

### 4.7 `mvp_domain_task`

这是未来替代现有 `mvp_task` 的正式执行任务表。

字段建议：

```sql
CREATE TABLE mvp_domain_task (
  id bigint unsigned NOT NULL,
  workflow_run_id bigint unsigned NOT NULL,
  stage_run_id bigint unsigned NOT NULL,
  plan_version_id bigint unsigned DEFAULT NULL,
  blueprint_id bigint unsigned DEFAULT NULL,
  parent_task_id bigint unsigned DEFAULT NULL,
  source_task_id bigint unsigned DEFAULT NULL,
  root_task_id bigint unsigned DEFAULT NULL,
  task_kind varchar(32) NOT NULL,
  name varchar(255) NOT NULL,
  description text NOT NULL,
  role_type varchar(32) NOT NULL,
  role_level varchar(32) NOT NULL,
  execution_mode varchar(32) NOT NULL,
  status varchar(32) NOT NULL,
  conversation_id bigint unsigned DEFAULT NULL,
  model_id bigint unsigned DEFAULT NULL,
  batch_no int NOT NULL,
  sort int NOT NULL DEFAULT 0,
  retry_count int NOT NULL DEFAULT 0,
  affected_resources json DEFAULT NULL,
  locked_resources json DEFAULT NULL,
  result longtext DEFAULT NULL,
  context_summary text DEFAULT NULL,
  heartbeat_at datetime DEFAULT NULL,
  started_at datetime DEFAULT NULL,
  completed_at datetime DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_workflow_status (workflow_run_id, status),
  KEY idx_workflow_batch (workflow_run_id, batch_no, sort),
  KEY idx_root_task (root_task_id),
  KEY idx_source_task (source_task_id)
);
```

### 4.8 `mvp_review_issue`

字段建议：

```sql
CREATE TABLE mvp_review_issue (
  id bigint unsigned NOT NULL,
  workflow_run_id bigint unsigned NOT NULL,
  stage_run_id bigint unsigned NOT NULL,
  plan_version_id bigint unsigned NOT NULL,
  blueprint_id bigint unsigned DEFAULT NULL,
  severity varchar(16) NOT NULL,
  issue_code varchar(64) NOT NULL,
  issue_type varchar(32) NOT NULL,
  source_role varchar(32) NOT NULL,
  task_name varchar(255) DEFAULT NULL,
  message text NOT NULL,
  suggestion text DEFAULT NULL,
  status varchar(32) NOT NULL,
  resolved_at datetime DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_plan_issue (plan_version_id, severity, status),
  KEY idx_blueprint_issue (blueprint_id, severity, status)
);
```

### 4.9 `mvp_stage_task`

用于承载阶段内部系统任务。

```sql
CREATE TABLE mvp_stage_task (
  id bigint unsigned NOT NULL,
  stage_run_id bigint unsigned NOT NULL,
  task_type varchar(32) NOT NULL,
  role_type varchar(32) NOT NULL,
  status varchar(32) NOT NULL,
  input_payload json DEFAULT NULL,
  output_payload json DEFAULT NULL,
  error_message text DEFAULT NULL,
  started_at datetime DEFAULT NULL,
  completed_at datetime DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_stage_type (stage_run_id, task_type, status)
);
```

`task_type`：

- `precheck`
- `auditor_review`
- `coordinator_optimize`
- `review_summary`
- `final_summary`

### 4.10 `mvp_handoff_record`

用于记录跨角色交接。

```sql
CREATE TABLE mvp_handoff_record (
  id bigint unsigned NOT NULL,
  workflow_run_id bigint unsigned NOT NULL,
  from_task_id bigint unsigned DEFAULT NULL,
  to_task_id bigint unsigned DEFAULT NULL,
  handoff_type varchar(32) NOT NULL,
  reason text DEFAULT NULL,
  payload json DEFAULT NULL,
  created_at datetime NOT NULL,
  PRIMARY KEY (id),
  KEY idx_workflow_type (workflow_run_id, handoff_type)
);
```

### 4.11 `mvp_workflow_event`

统一事件表，用于回放与前端状态流。

```sql
CREATE TABLE mvp_workflow_event (
  id bigint unsigned NOT NULL,
  workflow_run_id bigint unsigned NOT NULL,
  stage_run_id bigint unsigned DEFAULT NULL,
  entity_type varchar(32) NOT NULL,
  entity_id bigint unsigned DEFAULT NULL,
  event_type varchar(64) NOT NULL,
  payload json DEFAULT NULL,
  created_at datetime NOT NULL,
  PRIMARY KEY (id),
  KEY idx_workflow_event (workflow_run_id, created_at),
  KEY idx_entity_event (entity_type, entity_id, created_at)
);
```

---

## 五、状态机设计

### 5.1 Project 状态

项目只保留业务壳状态：

- `idle`
- `active`
- `archived`

### 5.2 WorkflowRun 状态机

```text
designing
  -> reviewing
  -> executing
  -> reworking
  -> executing
  -> completed

任意阶段
  -> paused
  -> canceled
  -> failed
```

规则：

1. `designing -> reviewing` 仅当存在 `draft plan_version`
2. `reviewing -> executing` 仅当审核通过并成功实例化执行任务
3. `executing -> reworking` 当审计失败或失败升级进入返工阶段
4. `reworking -> executing` 仅当返工计划完成派发
5. `paused` 可恢复到上一个活动阶段

### 5.3 StageRun 状态机

- `pending`
- `running`
- `succeeded`
- `failed`
- `canceled`
- `skipped`

### 5.4 DomainTask 状态机

```text
draft -> pending -> running -> completed -> auditing -> completed
                    |            |           |
                    v            v           v
                  failed     escalated   bug_found
                    |            |           |
                    +-------> rework/dispatched ------> pending
```

说明：

1. `draft` 只存在于蓝图或实例化前阶段
2. 审核通过后生成的 `domain_task` 默认从 `pending` 开始
3. `bug_found` 与 `escalated` 分属不同返工路径

---

## 六、执行流设计

### 6.1 设计阶段

输入：

- 架构师对话
- 项目描述
- 当前历史 plan version

输出：

- 新的 `plan_version`
- 一组 `task_blueprint`

流程：

1. 用户与架构师继续对话
2. 架构师输出任务规划 JSON
3. 解析后生成新的 `plan_version`
4. 将 `task_blueprint` 绑定到该版本
5. 将旧版本标记为 `superseded` 或保留可对比

### 6.2 审核阶段

审核阶段是正式 `StageRun(review)`，内部包含多个 `stage_task`。

#### 6.2.1 PrecheckTask

职责：

- 任务名检查
- 描述质量检查
- 依赖合法性
- 批次顺序预检
- 路径格式检查
- 角色覆盖检查
- 资源冲突预检

产出：

- 一组 `review_issue`
- 可选自动修正建议

#### 6.2.2 AuditorReviewTask

职责：

- 审核任务粒度与可执行性
- 审核依赖合理性
- 审核角色分配合理性
- 审核遗漏模块

产出：

- `review_issue`
- 审核结论

#### 6.2.3 CoordinatorOptimizeTask

职责：

- 批次调整
- 并行度建议
- 风险评估

产出：

- 优化后的批次方案

#### 6.2.4 ReviewSummaryTask

职责：

- 汇总 precheck / auditor / coordinator 结果
- 给出最终审核结论

结论：

- `approved`
- `rejected`
- `approved_with_warnings`

### 6.3 执行阶段

审核通过后：

1. 从 `task_blueprint` 实例化 `domain_task`
2. 建立依赖图
3. 分配到 `execute` 阶段
4. `DomainTaskScheduler` 启动调度

### 6.4 返工阶段

返工阶段不再只是零散的 bugloop，而是正式 `StageRun(rework)`。

包含：

- `bug_analysis`
- `failure_analysis`
- `repair_dispatch`

返工完成后，回到新的 `execute` 阶段。

### 6.5 完成阶段

完成阶段负责：

- 汇总阶段数据
- 输出最终总结
- 持久化统计指标
- 标记 `workflow_run.completed`

---

## 七、调度器重构

### 7.1 从单调度器升级为双调度器

#### 7.1.1 StageScheduler

职责：

- 驱动 `StageRun`
- 决定下一个阶段
- 管理阶段暂停/恢复/取消
- 管理阶段内 `stage_task`

#### 7.1.2 DomainTaskScheduler

职责：

- 只处理 `domain_task`
- 批次门控
- 资源锁
- 执行并发
- 看门狗
- 闭环返工触发

### 7.2 运行时上下文模型

建议统一为：

```go
type WorkflowRuntime struct {
    WorkflowRunID int64
    Ctx           context.Context
    Cancel        context.CancelFunc
}
```

每个 `workflow_run` 对应一个 runtime。

所有阶段任务、执行任务、返工任务都继承这个 runtime。

### 7.3 资源锁模型

资源锁从“任务字段”升级为“锁表 + 内存缓存”双写。

建议保留 `locked_resources` 作为缓存，同时新增：

- `mvp_task_resource_lock`

用于：

- 排查锁泄漏
- 多进程恢复
- 管理员手工干预

---

## 八、审核阶段任务化设计

### 8.1 为什么不能继续函数式实现

函数式审核有以下问题：

1. 难以暂停恢复
2. 无法统一统计耗时
3. 无法细粒度重试
4. 无法独立可视化显示审核过程
5. 无法沉淀阶段日志与问题记录

### 8.2 审核任务化方案

审核阶段内部固定包含以下 `stage_task`：

1. `precheck`
2. `auditor_review`
3. `coordinator_optimize`
4. `review_summary`

执行顺序：

```text
precheck
  -> auditor_review
  -> coordinator_optimize
  -> review_summary
```

异常处理：

- `precheck` 失败：可直接结束为 `rejected`
- `auditor_review` 超时：进入 `warning + degraded mode`
- `coordinator_optimize` 失败：允许降级通过
- `review_summary` 失败：阶段失败，不得推进

---

## 九、计划版本化设计

### 9.1 目标

不再让架构师重新拆任务时“覆盖旧任务”，而是生成新计划版本。

### 9.2 规则

1. 每次有效拆分都生成新 `plan_version`
2. 每个 `plan_version` 拥有独立 `task_blueprint`
3. 审核只针对单一 `plan_version`
4. 执行阶段只实例化 `approved plan_version`
5. 新版本出现后，旧未通过版本标记 `superseded`

### 9.3 价值

- 支持版本对比
- 支持回滚到上个方案
- 支持统计“架构师修订次数”
- 支持前端展示方案演进历史

---

## 十、前端重构方案

### 10.1 页面结构调整

当前以前端聊天页承担了过多流程语义。

建议拆为：

1. 项目概览页
2. 计划设计页
3. 审核工作台
4. 执行控制台
5. 阶段日志页
6. 任务链路页

### 10.2 项目概览页

展示：

- 当前 `workflow_run`
- 当前阶段
- 阶段进度
- 最近事件
- 风险提示

### 10.3 计划设计页

展示：

- 架构师对话
- plan version 列表
- 当前版本任务蓝图
- 版本 diff

### 10.4 审核工作台

展示：

- precheck 结果
- auditor review 结果
- coordinator optimize 结果
- review issue 列表
- 最终审核结论

### 10.5 执行控制台

展示：

- 批次
- 运行中任务
- 资源锁
- 返工链
- 看门狗告警

### 10.6 SSE / 事件总线前端协议

建议前端不再只订阅 message SSE，而是按实体订阅：

- `/workflow/events?workflowRunID=...`
- `/conversation/events?conversationID=...`
- `/message/sse?messageID=...`

事件类型建议：

- `workflow.status_changed`
- `stage.started`
- `stage.completed`
- `review.issue_created`
- `task.started`
- `task.completed`
- `task.failed`
- `task.escalated`

---

## 十一、API 设计

### 11.1 新增 API

#### WorkflowRun

- `POST /workflow-run/create`
- `GET /workflow-run/detail`
- `POST /workflow-run/pause`
- `POST /workflow-run/resume`
- `POST /workflow-run/cancel`

#### PlanVersion

- `GET /plan-version/list`
- `GET /plan-version/detail`
- `POST /plan-version/submit`
- `POST /plan-version/approve`
- `POST /plan-version/reject`

#### Review

- `GET /review/summary`
- `GET /review/issues`
- `POST /review/retry-stage-task`

#### Stage

- `GET /stage-run/list`
- `GET /stage-run/detail`

#### DomainTask

- `GET /domain-task/list`
- `GET /domain-task/detail`
- `POST /domain-task/retry`
- `POST /domain-task/skip`

### 11.2 兼容 API

旧 API 可在过渡期保留，但内部逐步改为代理到新模型。

例如：

- `/workflow/project-status` -> 读取 `workflow_run`
- `/workflow/confirm-plan` -> 触发 plan submit + review stage

---

## 十二、可观测性设计

### 12.1 日志体系

日志分三类：

1. 运行日志
2. 审核日志
3. 任务日志

所有关键动作写 `mvp_workflow_event`。

### 12.2 指标体系

建议沉淀以下指标：

- 平均设计修订次数
- 审核通过率
- 审核平均耗时
- 平均任务返工次数
- implementer 失败率
- auditor bug 发现率
- coordinator 批次调整命中率
- 项目完成总耗时

### 12.3 告警体系

建议加入：

- `workflow paused too long`
- `review timeout`
- `task lock leaked`
- `handoff rounds exceeded`
- `stage transition failed`

---

## 十三、迁移方案

### 13.1 迁移原则

架构要激进，发布必须渐进。

### 13.2 迁移阶段

#### Phase 1：新增结构，不切主链

- 新增表
- 新增 DAO/Entity
- 新增事件模型
- 保持旧 `mvp_task` 主执行链不变

#### Phase 2：新项目走新架构

- 仅新建项目走 `workflow_run + plan_version + domain_task`
- 老项目继续走旧路径

#### Phase 3：审核链先切新架构

- 设计阶段继续兼容旧聊天
- 审核阶段切到 `stage_run + stage_task`
- 执行阶段仍可暂时映射回旧 `mvp_task`

#### Phase 4：执行阶段切新任务表

- 新项目全面切 `mvp_domain_task`
- 旧 `mvp_task` 退为兼容层

#### Phase 5：移除旧执行内核

- 停用旧 `mvp_task` 主执行路径
- 保留历史查询兼容

### 13.3 双读单写策略

过渡期建议：

- 新建项目：单写新结构
- 旧项目：继续读写旧结构
- 管理端查询：双读聚合

---

## 十四、实施计划

### 14.1 第一阶段：地基

目标：

- 建表
- 建 DAO
- 建 `workflow_run` / `stage_run` 基础服务
- 建统一事件模型

交付：

- 数据库升级 SQL
- 后端基础模型
- 最小 API

### 14.2 第二阶段：设计与审核阶段

目标：

- plan version
- task blueprint
- review stage task
- review issue

交付：

- 新设计页
- 新审核页
- review runtime

### 14.3 第三阶段：执行阶段迁移

目标：

- domain_task
- new scheduler
- resource lock table
- watchdog 迁移

### 14.4 第四阶段：返工与汇总

目标：

- rework stage
- handoff record
- final summary stage

### 14.5 第五阶段：前端统一控制台

目标：

- workflow dashboard
- execution console
- stage timeline

---

## 十五、风险与对策

### 15.1 风险：模型过多导致实现复杂度暴涨

对策：

- 分阶段落地
- 保持每阶段单一职责
- API 层做兼容代理

### 15.2 风险：新旧任务表并存导致查询复杂

对策：

- 只在过渡期双轨
- 新项目直接全走新结构

### 15.3 风险：前端重构成本高

对策：

- 先新增视图，不立即移除旧页
- 聊天页保留为会话页，不承担流程主控

### 15.4 风险：审核任务化后延迟上升

对策：

- 审核阶段任务是系统任务，不走重资源调度
- 支持降级模式
- 支持阶段级超时与快速失败

---

## 十六、最终结论

本次重构的目标，不是简单把审计员/协调员塞进现有任务表，而是把 EasyMVP 升级为正式的阶段化工作流平台：

- `Project` 负责业务容器
- `WorkflowRun` 负责一次完整执行
- `StageRun` 负责阶段推进
- `PlanVersion` 负责设计产物版本化
- `TaskBlueprint` 负责审核前任务蓝图
- `DomainTask` 负责审核后执行任务
- `ReviewIssue` 负责审核问题独立建模

这是一次高强度、一步到位的架构升级，但它解决的是系统未来 1 到 2 个大版本的可持续性问题。

如果后续需要继续落地，建议下一步直接基于本设计文档补以下实施文档：

1. 数据库升级设计文档
2. 后端模块拆分与目录结构设计文档
3. 前端页面与事件协议设计文档
4. 迁移实施清单与里程碑文档
