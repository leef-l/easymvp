# Accept Stage 自动验收阶段设计文档

## 一、背景与目标

当前 WorkflowRun 主链已经收口为：

`design -> review -> execute -> complete`

但“完成”当前仍然只代表：

1. 任务执行完毕
2. 工作流完成收口
3. 生成统计与总结

这还不足以回答“项目结果是否达标”。因此月 5 不能只在 `CompleteStage` 中补几条检查逻辑，而应一次性引入独立的 `Accept Stage`，让系统在进入 `complete` 前先经过自动验收裁决。

本文档定义一步到位方案，目标是：

1. 在主链中加入独立 `accept` 阶段
2. 支持显式规则 + LLM 裁决的双层验收模型
3. 支持结构化验收问题输出
4. 支持验收失败自动进入 `rework`
5. 为月 6 风险管理、自动重规划预留标准接口

## 二、总体方案

最终阶段模型定义为：

- 主链：`design -> review -> execute -> accept -> complete`
- 旁路：`execute -> rework`
- 旁路：`accept -> rework`
- 回流：`rework -> execute`
- 回流：`rework -> accept`

约束：

1. `accept` 是裁决阶段，不承担真实执行任务。
2. `complete` 只负责收口、总结、归档，不负责是否通过的判定。
3. `rework` 仍然是旁路，不重新放回线性 `stageOrder`。

## 三、核心设计原则

1. 验收必须在 `complete` 之前发生。
2. 验收结果必须可审计、可回放、可人工干预。
3. 验收失败原因必须结构化，不能只存自然语言。
4. LLM 只能参与裁决，不能替代显式规则。
5. 返工输入必须直接来自验收 issue，而不是人工二次翻译。

## 四、状态机设计

### 4.1 新增阶段类型

- `StageAccept = "accept"`

### 4.2 新增工作流状态

- `WorkflowAccepting = "accepting"`

### 4.3 主链顺序

主链顺序调整为：

1. `design`
2. `review`
3. `execute`
4. `accept`
5. `complete`

### 4.4 合法状态迁移

新增或调整以下迁移：

1. `executing -> accepting`
2. `accepting -> completed`
3. `accepting -> reworking`
4. `accepting -> failed`
5. `accepting -> paused`
6. `accepting -> canceled`
7. `reworking -> executing`
8. `reworking -> accepting`

语义约束：

1. `execute` 成功后统一推进到 `accept`
2. `accept passed` 才能进入 `complete`
3. `accept failed` 自动进入 `rework`
4. `rework` 完成后由返工类型决定回流到 `execute` 或 `accept`
5. `accept` 阶段自身若执行异常，工作流可降级到 `failed`

## 五、数据模型

一步到位方案建议新增 4 张表。

### 5.1 `mvp_accept_run`

表示一次完整的自动验收执行。

字段建议：

- `id`
- `workflow_run_id`
- `stage_run_id`
- `project_id`
- `plan_version_id`
- `accept_round`
- `status`：`pending/running/completed/failed/canceled`
- `decision`：`passed/failed/manual_review`
- `score`
- `summary`
- `rules_version`
- `rules_snapshot_ref`
- `created_by`
- `dept_id`
- `started_at`
- `finished_at`
- `created_at`
- `updated_at`
- `deleted_at`

### 5.2 `mvp_accept_issue`

表示一次验收过程中发现的结构化问题。

字段建议：

- `id`
- `accept_run_id`
- `workflow_run_id`
- `project_id`
- `domain_task_id`
- `issue_type`
- `rule_code`
- `severity`
- `title`
- `detail`
- `expected_value`
- `actual_value`
- `suggested_action`
- `resource_ref`
- `status`：`open/resolved/ignored`
- `created_by`
- `dept_id`
- `created_at`
- `updated_at`
- `deleted_at`

说明：

1. `domain_task_id` 表示该 issue 主要关联的任务对象，用于返工定位和前端跳转。
2. `resource_ref` 仍保留，用于附加文件、stage、日志等扩展引用。

### 5.3 `mvp_accept_rule`

表示验收规则定义。

字段建议：

- `id`
- `project_type`
- `rule_code`
- `rule_name`
- `rule_type`
- `scope_type`
- `config_json`
- `enabled`
- `priority`
- `created_at`
- `updated_at`
- `deleted_at`

说明：

1. `rule_type` 区分 `artifact/process/quality`
2. `scope_type` 区分 `project/task/file/stage`

### 5.4 `mvp_accept_evidence`

表示验收时收集到的证据。

字段建议：

- `id`
- `accept_run_id`
- `evidence_type`
- `source_type`
- `source_id`
- `content_ref`
- `summary`
- `created_at`
- `updated_at`
- `deleted_at`

权限约束：

1. `mvp_accept_evidence` 不单独建立业务归属字段。
2. evidence 的可见性必须继承其所属 `accept_run -> workflow_run -> project` 的权限域。
3. 任何 evidence 查询都必须绑定 `accept_run_id` 或其上游业务对象，不允许裸查。

## 六、服务拆分

Accept Stage 不应做成一个“大而全单服务”，而应拆成 5 个服务。

### 6.1 `AcceptStageService`

职责：

1. 启动 `accept` 阶段
2. 创建 `accept_run`
3. 编排证据收集、规则执行、LLM 裁决、决策归并
4. 决定后续走 `complete` 还是 `rework`

### 6.2 `AcceptanceEvidenceCollector`

职责：

1. 收集 `domain_task` 结果
2. 收集 `stage_run.output_ref`
3. 收集 `handoff_record`
4. 收集文件产物、diff、日志、summary
5. 产出标准化 evidence 集合

### 6.3 `AcceptanceRuleEngine`

职责：

1. 加载项目类型对应规则
2. 执行显式规则
3. 输出结构化规则命中结果

### 6.4 `AcceptanceJudge`

职责：

1. 基于 evidence 和规则结果调用 LLM
2. 输出质量判断结论
3. 给出补充说明与建议

### 6.5 `AcceptanceDecisionReducer`

职责：

1. 合并硬规则与 LLM 判断
2. 产出统一决策：
   - `passed`
   - `failed`
   - `manual_review`
3. 生成 `accept_issue`
4. 固化本次裁决快照，确保历史结果可回放

## 七、验收模型

一步到位方案采用三层裁决模型。

### 7.1 第一层：硬规则引擎

用于保证可审计性和稳定性。

第一版建议覆盖：

1. 是否存在 `failed/escalated` 任务
2. 必需产物是否存在
3. 关键文件/目录是否存在
4. 关键阶段是否已完成
5. 关键输出是否为空

### 7.2 第二层：LLM 质量判断

用于处理质量类问题。

第一版建议覆盖：

1. 代码结果是否满足需求
2. 文档是否覆盖要求
3. 创作内容是否达到结构要求
4. 分析结论是否具备完整性和可解释性

### 7.3 第三层：统一裁决

建议规则：

1. 任一阻塞级硬规则失败，直接 `failed`
2. 硬规则通过但质量判断低于阈值，可 `failed`
3. 结果不确定时进入 `manual_review`
4. 全部通过才 `passed`

## 八、规则模板体系

第一版不要把项目类型写死在代码流程里，而应使用模板化规则。

建议内置 4 类模板：

1. `software_dev`
2. `document`
3. `creative`
4. `analysis`

每类模板至少包含：

1. 必需产物规则
2. 过程规则
3. 质量规则
4. 失败返工建议

## 九、和 Rework 的衔接

### 9.1 验收失败后的行为

`accept failed` 后系统必须：

1. 写入 `accept_issue`
2. 生成返工输入包
3. 自动触发 `rework`

### 9.2 返工输入包结构

至少包含：

- `accept_run_id`
- `issue_ids`
- `rule_code`
- `severity`
- `title`
- `detail`
- `expected`
- `actual`
- `suggested_action`
- `resource_refs`

### 9.3 返工回流策略

回流目标由 issue 类型决定：

1. 代码修复类问题：`rework -> execute`
2. 产物补齐类问题：`rework -> accept`
3. 文档质量补强类问题：`rework -> accept`

## 十、和 Complete Stage 的边界

`CompleteStage` 保持纯收口，不承载验收裁决。

`accept` 负责：

1. 判断是否达标
2. 生成结构化问题
3. 决定是否进入返工

`complete` 负责：

1. 汇总指标
2. 生成完成总结
3. 写归档输出
4. 发布完成事件

## 十一、后端模块建议

建议新增目录：

- `admin-go/app/mvp/internal/workflow/stage/accept/`
- `admin-go/app/mvp/internal/workflow/acceptance/`

建议子模块：

1. `stage/accept/service.go`
2. `acceptance/evidence_collector.go`
3. `acceptance/rule_engine.go`
4. `acceptance/judge.go`
5. `acceptance/decision_reducer.go`
6. `acceptance/template_registry.go`

## 十二、API 设计

建议新增以下接口：

1. `POST /workflow/{id}/accept/start`
2. `POST /workflow/{id}/accept/retry`
3. `GET /workflow/{id}/accept/latest`
4. `GET /workflow/{id}/accept/runs`
5. `GET /accept-run/{id}/issues`
6. `GET /accept-run/{id}/evidences`
7. `POST /accept-run/{id}/approve`
8. `POST /accept-run/{id}/reject`

人工兜底能力要求：

1. 支持人工放行
2. 支持人工驳回
3. 支持人工重跑验收

## 十三、前端页面

一步到位版本要求同步提供 `Accept Console`。

至少包含：

1. 验收状态总览
2. 决策结果与评分
3. issue 列表
4. evidence 查看
5. rule 命中详情
6. 人工放行 / 驳回 / 重验

建议页面：

- `views/mvp/workflow/accept.vue`

建议能力：

1. 失败问题按严重级别过滤
2. 可跳转到相关任务、日志、产物
3. 可直接发起返工或重验

## 十四、实施顺序

虽然方案一步到位，但实现上建议压缩为两批，而不是拆成四批。

### 14.1 第一批：可闭环底座

目标：

1. 让 `execute -> accept -> complete`
2. 让 `accept failed -> rework`
3. 让验收结果具备结构化、可追溯、可回放能力

内容：

1. 状态机扩展
2. 数据表迁移
3. `AcceptStageService`
4. `AcceptanceEvidenceCollector`
5. `AcceptanceRuleEngine`
6. `AcceptanceDecisionReducer`
7. `accept_issue` 持久化
8. `execute -> accept -> complete` 主链改造
9. `accept failed -> rework` 闭环
10. 规则版本与规则快照持久化

交付标准：

1. 系统已能在无 LLM judge 的前提下跑通 Accept 主链
2. 验收失败可输出结构化 issue
3. 历史 accept 结果可按 run / issue / evidence / rule snapshot 回放

### 14.2 第二批：智能化与运营化

目标：

1. 在第一批稳定底座上补齐质量判断和运营能力
2. 为月 6 风险管理与自动重规划提供稳定输入

内容：

1. 模板规则体系
2. `AcceptanceJudge`
3. Accept Console
4. 人工放行 / 驳回 / 重验接口
5. 统计与审计
6. 与月 6 风险管理、自动重规划打通

交付标准：

1. 支持按项目类型模板做差异化验收
2. LLM judge 已接入统一裁决链
3. 前端可完整查看 accept run、issue、evidence、decision
4. 人工可介入 accept 阶段而不破坏自动链路

## 十五、验收标准

达到以下条件，才算月 5 完成：

1. `execute` 完成后统一进入 `accept`
2. 项目进入 `complete` 前必须先通过 `accept`
3. 验收失败产生结构化 `accept_issue`
4. 验收失败可自动进入 `rework`
5. `accept` 结果可在前端查看与人工干预
6. 至少支持 `software_dev/document/creative/analysis` 四类模板

## 十六、数据库表结构建议

以下为建议字段草案，供后续升级 SQL 直接落地。

### 16.1 `mvp_accept_run`

```sql
CREATE TABLE `mvp_accept_run` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `workflow_run_id` bigint NOT NULL COMMENT '工作流运行ID',
  `stage_run_id` bigint NOT NULL COMMENT 'accept阶段stage_run_id',
  `project_id` bigint NOT NULL COMMENT '项目ID',
  `plan_version_id` bigint DEFAULT NULL COMMENT '关联方案版本ID',
  `accept_round` int NOT NULL DEFAULT '1' COMMENT '第几轮验收',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/running/completed/failed/canceled',
  `decision` varchar(20) DEFAULT NULL COMMENT 'passed/failed/manual_review',
  `score` decimal(5,2) DEFAULT NULL COMMENT '验收评分',
  `summary` text COMMENT '验收摘要',
  `rules_version` varchar(64) DEFAULT NULL COMMENT '规则版本号',
  `rules_snapshot_ref` longtext COMMENT '规则快照引用或JSON',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `finished_at` datetime DEFAULT NULL COMMENT '结束时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_run_id` (`workflow_run_id`),
  KEY `idx_stage_run_id` (`stage_run_id`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_workflow_round` (`workflow_run_id`,`accept_round`),
  KEY `idx_status` (`status`),
  KEY `idx_decision` (`decision`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动验收运行记录';
```

### 16.2 `mvp_accept_issue`

```sql
CREATE TABLE `mvp_accept_issue` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `accept_run_id` bigint NOT NULL COMMENT '验收运行ID',
  `workflow_run_id` bigint NOT NULL COMMENT '工作流运行ID',
  `project_id` bigint NOT NULL COMMENT '项目ID',
  `domain_task_id` bigint DEFAULT NULL COMMENT '主关联任务ID',
  `issue_type` varchar(32) NOT NULL COMMENT 'artifact/process/quality/risk',
  `rule_code` varchar(64) DEFAULT NULL COMMENT '规则编码',
  `severity` varchar(16) NOT NULL COMMENT 'info/warn/error/blocker',
  `title` varchar(255) NOT NULL COMMENT '问题标题',
  `detail` text COMMENT '问题详情',
  `expected_value` text COMMENT '预期值',
  `actual_value` text COMMENT '实际值',
  `suggested_action` text COMMENT '建议动作',
  `resource_ref` text COMMENT '关联资源引用(JSON)',
  `status` varchar(20) NOT NULL DEFAULT 'open' COMMENT 'open/resolved/ignored',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_accept_run_id` (`accept_run_id`),
  KEY `idx_workflow_run_id` (`workflow_run_id`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_domain_task_id` (`domain_task_id`),
  KEY `idx_rule_code` (`rule_code`),
  KEY `idx_severity` (`severity`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动验收问题';
```

### 16.3 `mvp_accept_rule`

```sql
CREATE TABLE `mvp_accept_rule` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `project_type` varchar(64) NOT NULL COMMENT '项目类型模板',
  `rule_code` varchar(64) NOT NULL COMMENT '规则编码',
  `rule_name` varchar(255) NOT NULL COMMENT '规则名称',
  `rule_type` varchar(32) NOT NULL COMMENT 'artifact/process/quality',
  `scope_type` varchar(32) NOT NULL COMMENT 'project/task/file/stage',
  `config_json` longtext NOT NULL COMMENT '规则配置',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `priority` int NOT NULL DEFAULT '100' COMMENT '优先级',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_type_rule_code` (`project_type`,`rule_code`),
  KEY `idx_rule_type` (`rule_type`),
  KEY `idx_scope_type` (`scope_type`),
  KEY `idx_enabled_priority` (`enabled`,`priority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动验收规则定义';
```

### 16.4 `mvp_accept_evidence`

```sql
CREATE TABLE `mvp_accept_evidence` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `accept_run_id` bigint NOT NULL COMMENT '验收运行ID',
  `evidence_type` varchar(32) NOT NULL COMMENT 'task_output/file/log/diff/stage_output/handoff/summary',
  `source_type` varchar(32) NOT NULL COMMENT 'domain_task/stage_run/file/handoff_record/workflow_run',
  `source_id` bigint DEFAULT NULL COMMENT '来源对象ID',
  `content_ref` longtext COMMENT '证据引用或JSON',
  `summary` text COMMENT '证据摘要',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_accept_run_id` (`accept_run_id`),
  KEY `idx_evidence_type` (`evidence_type`),
  KEY `idx_source_type_source_id` (`source_type`,`source_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动验收证据';
```

## 十七、Go 模块清单

建议按当前 WorkflowRun 目录风格拆分。

### 17.1 目录结构

```text
admin-go/app/mvp/internal/workflow/
  acceptance/
    model.go
    rule_engine.go
    evidence_collector.go
    judge.go
    decision_reducer.go
    template_registry.go
  stage/accept/
    service.go
  repo/
    accept_run_repo.go
    accept_issue_repo.go
    accept_rule_repo.go
    accept_evidence_repo.go
```

### 17.2 核心对象建议

`acceptance/model.go`

- `AcceptDecision`
- `RuleHit`
- `EvidenceItem`
- `JudgeResult`
- `AcceptContext`

`acceptance/rule_engine.go`

- `LoadRules(projectType string) ([]Rule, error)`
- `Evaluate(ctx context.Context, in *AcceptContext) ([]RuleHit, error)`

`acceptance/evidence_collector.go`

- `Collect(ctx context.Context, workflowRunID, acceptRunID int64) ([]EvidenceItem, error)`

`acceptance/judge.go`

- `Judge(ctx context.Context, in *AcceptContext) (*JudgeResult, error)`

`acceptance/decision_reducer.go`

- `Reduce(ctx context.Context, in *AcceptContext, hits []RuleHit, judge *JudgeResult) (*DecisionResult, error)`

`stage/accept/service.go`

- `Start(ctx context.Context, workflowRunID, stageRunID int64) error`
- `Run(ctx context.Context, workflowRunID, stageRunID int64) error`
- `Approve(ctx context.Context, acceptRunID int64, operatorID int64) error`
- `Reject(ctx context.Context, acceptRunID int64, operatorID int64, reason string) error`
- `Retry(ctx context.Context, workflowRunID int64) error`

### 17.3 Repo 层权限预留

Accept 相关 Repo 设计时必须预留数据权限过滤能力，不允许把查询方法设计成永久裸查。

建议统一引入查询作用域对象，例如：

- `DataScope`
- `ProjectScope`
- `WorkflowScope`

Repo 查询方法至少应支持后续注入以下过滤条件：

1. `created_by`
2. `dept_id`
3. `project_id`
4. `workflow_run_id`

目标：

1. 后续接入部门隔离时不需要改表结构
2. 后续接入项目级可见性时不需要重写 Accept 查询接口
3. 前端 issue / evidence / accept run 查询可统一走权限域裁剪

## 十八、API 清单

建议沿用当前 WorkflowRun 控制器风格，放入 `controller/chat/workflow.go` 或拆专属控制器。

### 18.1 运行类接口

1. `POST /workflow/{id}/accept/start`
2. `POST /workflow/{id}/accept/retry`
3. `GET /workflow/{id}/accept/latest`
4. `GET /workflow/{id}/accept/runs`

### 18.2 数据查询接口

1. `GET /accept-run/{id}/issues`
2. `GET /accept-run/{id}/evidences`
3. `GET /accept-run/{id}/summary`
4. `GET /accept-run/{id}/rules`

### 18.3 人工干预接口

1. `POST /accept-run/{id}/approve`
2. `POST /accept-run/{id}/reject`
3. `POST /accept-run/{id}/rerun`
4. `POST /accept-run/{id}/rework`

### 18.5 API 权限预留

Accept API 在第一版即需冻结权限语义，哪怕首版先走弱校验，也不能省略接口层语义定义。

建议至少区分：

1. 查询权限
   - 可查看 `accept_run`
   - 可查看 `accept_issue`
   - 可查看 `accept_evidence`

2. 操作权限
   - 可触发重验
   - 可人工放行
   - 可人工驳回
   - 可强制进入返工

3. 作用域规则
   - 默认按 `project_id` 过滤
   - 默认继承 `dept_id` 数据域
   - evidence 权限继承 `accept_run`，不能独立放开

### 18.4 前端页面清单

1. `views/mvp/workflow/accept.vue`
2. `views/mvp/workflow/accept-issues.vue`
3. `views/mvp/workflow/accept-evidence.vue`

## 十九、升级 SQL 规划

建议按现有数据库升级目录风格拆成独立 SQL 文件。

### 19.1 文件建议

1. `docker/mysql/upgrade/20260408_accept_stage_core.sql`
2. `docker/mysql/upgrade/20260408_accept_stage_seed.sql`

### 19.2 `20260408_accept_stage_core.sql`

建议包含：

1. 新建 `mvp_accept_run`
2. 新建 `mvp_accept_issue`
3. 新建 `mvp_accept_rule`
4. 新建 `mvp_accept_evidence`
5. 为 `mvp_stage_run.stage_type` 扩展 `accept`
6. 为相关索引和唯一键建模

### 19.3 `20260408_accept_stage_seed.sql`

建议包含：

1. 默认模板规则种子数据
2. `software_dev` 基础规则
3. `document` 基础规则
4. `creative` 基础规则
5. `analysis` 基础规则

## 二十、API 请求与响应示例

以下仅定义第一版建议格式，最终以控制器实现为准。

### 20.1 启动验收

`POST /workflow/{id}/accept/start`

请求体：

```json
{
  "force": false
}
```

响应体：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "workflow_run_id": 320100000000000001,
    "stage_run_id": 320100000000000021,
    "accept_run_id": 320100000000000031,
    "status": "running"
  }
}
```

### 20.2 查询最近一次验收

`GET /workflow/{id}/accept/latest`

响应体：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "accept_run_id": 320100000000000031,
    "workflow_run_id": 320100000000000001,
    "status": "completed",
    "decision": "failed",
    "score": 72.5,
    "summary": "核心产物已生成，但验收规则命中 2 个 blocker。",
    "rules_version": "v1.0.0",
    "started_at": "2026-04-08T10:00:00+08:00",
    "finished_at": "2026-04-08T10:03:12+08:00"
  }
}
```

### 20.3 查询验收问题列表

`GET /accept-run/{id}/issues`

响应体：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 320100000000000101,
      "issue_type": "artifact",
      "rule_code": "software.required_file_exists",
      "severity": "blocker",
      "title": "缺少 README.md",
      "detail": "项目根目录未检测到 README.md",
      "expected_value": "存在 README.md",
      "actual_value": "文件不存在",
      "suggested_action": "补充项目说明文档并重新验收",
      "resource_ref": {
        "path": "/workspace/project/README.md"
      },
      "status": "open"
    }
  ]
}
```

### 20.4 人工放行

`POST /accept-run/{id}/approve`

请求体：

```json
{
  "reason": "当前 blocker 已人工确认可接受"
}
```

响应体：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "accept_run_id": 320100000000000031,
    "decision": "passed"
  }
}
```

### 20.5 驳回并进入返工

`POST /accept-run/{id}/rework`

请求体：

```json
{
  "reason": "阻塞问题仍未解决",
  "issue_ids": [320100000000000101, 320100000000000102]
}
```

响应体：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "accept_run_id": 320100000000000031,
    "decision": "failed",
    "next_stage": "rework"
  }
}
```

## 二十一、规则模板样例

建议第一版直接内置模板 JSON，而不是等 DSL。

### 21.1 `software_dev` 模板示例

```json
{
  "project_type": "software_dev",
  "rules": [
    {
      "rule_code": "software.no_failed_tasks",
      "rule_name": "不得存在失败任务",
      "rule_type": "process",
      "scope_type": "project",
      "priority": 10,
      "config": {
        "forbid_status": ["failed", "escalated"]
      }
    },
    {
      "rule_code": "software.required_file_exists",
      "rule_name": "关键文件必须存在",
      "rule_type": "artifact",
      "scope_type": "file",
      "priority": 20,
      "config": {
        "required_files": ["README.md"]
      }
    },
    {
      "rule_code": "software.output_not_empty",
      "rule_name": "关键输出不得为空",
      "rule_type": "artifact",
      "scope_type": "task",
      "priority": 30,
      "config": {
        "task_kinds": ["implement", "refactor", "fix"],
        "require_non_empty_result": true
      }
    }
  ]
}
```

### 21.2 `document` 模板示例

```json
{
  "project_type": "document",
  "rules": [
    {
      "rule_code": "document.required_output_exists",
      "rule_name": "文档产物必须存在",
      "rule_type": "artifact",
      "scope_type": "file",
      "priority": 10,
      "config": {
        "required_extensions": [".md", ".docx"]
      }
    },
    {
      "rule_code": "document.summary_present",
      "rule_name": "必须生成总结",
      "rule_type": "process",
      "scope_type": "stage",
      "priority": 20,
      "config": {
        "required_stage_outputs": ["execute"]
      }
    }
  ]
}
```

### 21.3 `creative` 模板示例

```json
{
  "project_type": "creative",
  "rules": [
    {
      "rule_code": "creative.required_sections_present",
      "rule_name": "核心章节必须齐全",
      "rule_type": "artifact",
      "scope_type": "project",
      "priority": 10,
      "config": {
        "required_sections": ["outline", "main_content", "ending"]
      }
    }
  ]
}
```

### 21.4 `analysis` 模板示例

```json
{
  "project_type": "analysis",
  "rules": [
    {
      "rule_code": "analysis.must_have_conclusion",
      "rule_name": "分析报告必须包含结论",
      "rule_type": "quality",
      "scope_type": "project",
      "priority": 10,
      "config": {
        "required_keywords": ["结论", "建议"]
      }
    }
  ]
}
```

## 二十二、与路线图的关系

本方案用于落实半年路线图中的“第 5 个月：自动验收能力”。

对应关系如下：

1. 自动验收规则模型 -> `mvp_accept_rule + RuleEngine`
2. 项目级 acceptance checklist -> `accept_run + Accept Console`
3. 代码任务自动检查 -> `artifact/process/quality` 规则
4. 文档/创作/分析类自动验收模板 -> 模板化规则体系
5. 验收失败自动进入返工链 -> `accept -> rework`

## 二十三、结论

如果月 5 只做“简版 gate”，后续在月 6 做风险管理和自动重规划时很可能重构 Accept 体系。

因此更合理的一步到位方案是：

1. 引入独立 `Accept Stage`
2. 一次建好验收运行、规则、问题、证据模型
3. 使用“显式规则 + LLM 判断 + 统一裁决”三层模型
4. 让 `accept failed` 直接进入 `rework`
5. 保持 `complete` 为纯收口阶段

这套方案能直接作为月 5 正式实施基线，并为月 6 自治能力提供稳定底座。

## 二十四、实施任务拆解表

以下拆解已经按“开发前可直接排期”标准整理。

### 24.1 第一批：可闭环底座

1. 状态机扩展
   文件：
   - `admin-go/app/mvp/internal/consts/task.go`
   - `admin-go/app/mvp/internal/workflow/orchestrator/transition.go`
   - `admin-go/app/mvp/internal/workflow/orchestrator/stage_service.go`
   - `admin-go/app/mvp/internal/workflow/orchestrator/workflow_service.go`
   目标：
   - 新增全局常量 `consts.WorkflowRunStatusAccepting`
   - 新增 `accept` 阶段与 `accepting` 状态
   - 调整主链为 `design -> review -> execute -> accept -> complete`
   - 在 `workflow_service.go` 的 `activeStatuses` 白名单中加入 `accepting`，保证 `accepting -> paused` 可执行

2. 数据库升级
   文件：
   - `docker/mysql/upgrade/20260408_accept_stage_core.sql`
   - `docker/mysql/upgrade/20260408_accept_stage_seed.sql`
   目标：
   - 建 `accept_run / issue / rule / evidence`
   - 初始化四类模板规则

3. Repo / Model 层
   文件：
   - `admin-go/app/mvp/internal/workflow/repo/accept_run_repo.go`
   - `admin-go/app/mvp/internal/workflow/repo/accept_issue_repo.go`
   - `admin-go/app/mvp/internal/workflow/repo/accept_rule_repo.go`
   - `admin-go/app/mvp/internal/workflow/repo/accept_evidence_repo.go`
   - `admin-go/app/mvp/internal/dao/...`
   - `admin-go/app/mvp/internal/model/entity/...`
   - `admin-go/app/mvp/internal/model/do/...`
   目标：
   - 补齐 DAO/Entity/DO/Repo

4. Accept 核心服务
   文件：
   - `admin-go/app/mvp/internal/workflow/stage/accept/service.go`
   - `admin-go/app/mvp/internal/workflow/acceptance/evidence_collector.go`
   - `admin-go/app/mvp/internal/workflow/acceptance/rule_engine.go`
   - `admin-go/app/mvp/internal/workflow/acceptance/decision_reducer.go`
   - `admin-go/app/mvp/internal/workflow/acceptance/model.go`
   目标：
   - 跑通证据收集、规则执行、统一裁决

5. 主链接线
   文件：
   - `admin-go/app/mvp/internal/workflow/orchestrator/registry.go`
   - `admin-go/app/mvp/internal/workflow/orchestrator/stage_service.go`
   目标：
   - `execute` 完成后通过 `TransitionNext` 进入 `accept`
   - `accept passed -> complete`
   - `accept failed -> rework`
   - 当 `workflow.accept.enabled=false` 时，`TransitionNext` 在 `execute` 后跳过 `accept` 直接进入 `complete`

6. 回流策略
   文件：
   - `admin-go/app/mvp/internal/workflow/stage/rework/service.go`
   - `admin-go/app/mvp/internal/workflow/orchestrator/registry.go`
   目标：
   - 根据 issue 类型决定 `rework -> execute` 或 `rework -> accept`
   - `rework service` 需新增 `SetAcceptTrigger` 回调，不能只保留 `SetExecuteTrigger`

7. 接口骨架
   文件：
   - `admin-go/app/mvp/internal/controller/chat/workflow.go`
   - `admin-go/app/mvp/internal/service/...`
   目标：
   - 查询最新验收
   - 查询 issue / evidence
   - 触发重验

### 24.2 第二批：智能化与运营化

1. 模板规则体系
   文件：
   - `admin-go/app/mvp/internal/workflow/acceptance/template_registry.go`
   目标：
   - 按 `software_dev/document/creative/analysis` 装配规则

2. LLM Judge
   文件：
   - `admin-go/app/mvp/internal/workflow/acceptance/judge.go`
   目标：
   - 引入质量判断与说明生成

3. 前端控制台
   文件：
   - `vue-vben-admin/apps/web-antd/src/views/mvp/workflow/accept.vue`
   - `vue-vben-admin/apps/web-antd/src/api/mvp/workflow/index.ts`
   目标：
   - 展示 accept run / issue / evidence / decision

4. 人工干预
   文件：
   - `admin-go/app/mvp/internal/controller/chat/workflow.go`
   - `vue-vben-admin/apps/web-antd/src/views/mvp/workflow/accept.vue`
   目标：
   - 放行、驳回、重验、强制返工

5. 统计与审计
   文件：
   - `admin-go/app/mvp/internal/workflow/stage/complete/service.go`
   - `vue-vben-admin/apps/web-antd/src/views/mvp/workflow/dashboard.vue`
   目标：
   - 展示 accept 成功率、失败类型、模板命中情况

## 二十五、文件落点清单

### 25.1 后端新增文件

```text
admin-go/app/mvp/internal/workflow/stage/accept/service.go
admin-go/app/mvp/internal/workflow/acceptance/model.go
admin-go/app/mvp/internal/workflow/acceptance/evidence_collector.go
admin-go/app/mvp/internal/workflow/acceptance/rule_engine.go
admin-go/app/mvp/internal/workflow/acceptance/judge.go
admin-go/app/mvp/internal/workflow/acceptance/decision_reducer.go
admin-go/app/mvp/internal/workflow/acceptance/template_registry.go
admin-go/app/mvp/internal/workflow/repo/accept_run_repo.go
admin-go/app/mvp/internal/workflow/repo/accept_issue_repo.go
admin-go/app/mvp/internal/workflow/repo/accept_rule_repo.go
admin-go/app/mvp/internal/workflow/repo/accept_evidence_repo.go
```

### 25.2 后端修改文件

```text
admin-go/app/mvp/internal/workflow/orchestrator/transition.go
admin-go/app/mvp/internal/workflow/orchestrator/stage_service.go
admin-go/app/mvp/internal/workflow/orchestrator/workflow_service.go
admin-go/app/mvp/internal/workflow/orchestrator/registry.go
admin-go/app/mvp/internal/consts/task.go
admin-go/app/mvp/internal/workflow/stage/rework/service.go
admin-go/app/mvp/internal/workflow/stage/complete/service.go
admin-go/app/mvp/internal/controller/chat/workflow.go
```

### 25.3 前端新增文件

```text
vue-vben-admin/apps/web-antd/src/views/mvp/workflow/accept.vue
vue-vben-admin/apps/web-antd/src/components/mvp/workflow/accept-issues.vue
vue-vben-admin/apps/web-antd/src/components/mvp/workflow/accept-evidence.vue
vue-vben-admin/apps/web-antd/src/components/mvp/workflow/accept-decision.vue
```

### 25.4 前端修改文件

```text
vue-vben-admin/apps/web-antd/src/api/mvp/workflow/index.ts
vue-vben-admin/apps/web-antd/src/views/mvp/workflow/dashboard.vue
vue-vben-admin/apps/web-antd/src/router/routes/modules/mvp.ts
```

## 二十六、主链时序设计

### 26.1 成功路径

1. `execute` 所有任务完成
2. `StageService.TransitionNext()` 进入 `accept`
3. `AcceptStageService.Run()` 创建 `accept_run`
4. `EvidenceCollector` 收集证据
5. `RuleEngine` 执行硬规则
6. `Judge` 执行质量判断
7. `DecisionReducer` 合并结果
8. 若 `passed`，推进到 `complete`
9. `CompleteStage.Finalize()` 完成收口

### 26.2 失败返工路径

1. `accept` 产出 `failed`
2. 写入 `accept_issue`
3. 生成返工输入包
4. 触发 `rework`
5. `rework` 完成后根据 issue 类型：
   - 回 `execute`
   - 或回 `accept`

### 26.3 Accept 自身异常降级路径

1. `accept` 进入 `running`
2. 证据收集、规则执行或裁决过程异常
3. `AcceptStageService` 写入失败原因
4. 工作流状态从 `accepting` 降级为 `failed`
5. 前端显示 accept 阶段失败，而不是卡死在 `accepting`

### 26.4 人工介入路径

1. `accept` 产出 `manual_review`
2. 前端展示问题与证据
3. 人工选择：
   - 放行 -> `complete`
   - 驳回 -> `rework`
   - 重验 -> 重新运行 `accept`

## 二十七、联调顺序

建议按以下顺序联调，避免前后端互相阻塞。

1. 先联调状态机和 `accept_run`
2. 再联调 issue / evidence 查询接口
3. 再联调 `accept failed -> rework`
4. 再联调 `accept passed -> complete`
5. 最后联调人工放行 / 驳回 / 重验

## 二十八、测试设计

### 28.1 单元测试

1. `RuleEngine` 规则命中测试
2. `DecisionReducer` 决策归并测试
3. `EvidenceCollector` 证据收集测试
4. `TemplateRegistry` 模板装配测试

### 28.2 集成测试

1. `execute -> accept -> complete`
2. `execute -> accept -> rework -> execute`
3. `execute -> accept -> rework -> accept`
4. `accept -> manual_review -> approve`
5. `accept -> manual_review -> reject`

### 28.3 回归测试

1. `review -> execute` 现有主链不回归
2. `rework` 旁路行为不回归
3. `complete` 总结生成不回归
4. Dashboard / Timeline 状态展示不回归

## 二十九、灰度与回滚设计

### 29.1 灰度策略

1. 第一阶段只对 `workflow_v2` 新项目开启
2. 第二阶段只对 `software_dev` 模板开启自动验收
3. 第三阶段逐步扩到 `document/creative/analysis`

### 29.2 开关建议

建议增加配置开关：

1. `workflow.accept.enabled`
2. `workflow.accept.llm_judge_enabled`
3. `workflow.accept.manual_review_enabled`
4. `workflow.accept.project_types`

默认行为约束：

1. 当 `workflow.accept.enabled=false` 时，`TransitionNext` 在 `execute` 后直接跳过 `accept` 进入 `complete`
2. 此旁路必须集中在状态机推进层处理，不应分散在多个 stage service 中各自判断

### 29.3 回滚策略

若 Accept 主链出现严重问题：

1. 关闭 `workflow.accept.enabled`
2. `execute` 临时恢复直达 `complete`
3. 保留 accept 表数据，不做 destructive rollback
4. 问题修复后重新灰度开启

## 三十、数据权限预留原则

Accept Stage 属于工作流业务域，必须从第一版开始保留数据权限扩展口，而不是等后续计划再返工。

### 30.1 业务对象权限字段

以下表必须保留业务归属字段：

1. `mvp_accept_run`
2. `mvp_accept_issue`

字段要求：

1. `created_by`
2. `dept_id`

语义：

1. 与现有业务表风格保持一致
2. 为后续部门隔离、项目隔离、操作审计提供稳定字段基础

### 30.2 系统级对象权限策略

以下表可不强制增加 `created_by/dept_id`：

1. `mvp_accept_rule`
2. `mvp_accept_evidence`

但必须满足：

1. `mvp_accept_rule` 作为系统级规则定义，由服务层控制可见性
2. `mvp_accept_evidence` 的访问权限必须继承所属 `accept_run/project/workflow_run`
3. 不允许直接通过 evidence 做独立裸查询

### 30.3 Repo 层要求

1. 所有 Accept 查询必须预留 scope 参数或统一权限过滤入口
2. 不允许将 `ListAcceptRuns/ListAcceptIssues/ListAcceptEvidences` 设计成固定全量查询
3. 后续接入更细粒度 RBAC 时，应只扩展过滤逻辑，不改数据模型

### 30.4 API 层要求

1. `accept_run` 查询默认按项目范围过滤
2. `accept_issue` 查询默认按项目和部门范围过滤
3. `accept_evidence` 查询必须绑定上游 `accept_run_id`
4. 人工放行、驳回、重验必须预留独立操作权限点

### 30.5 冻结要求

在进入实施开发前，以下要求必须冻结：

1. `accept_run / accept_issue` 保留 `created_by / dept_id`
2. evidence 可见性继承规则写入实现约束
3. Repo 查询必须支持数据权限扩展
4. API 权限语义必须冻结

## 三十一、角色分工建议

### 30.1 后端负责人

负责：

1. 状态机与主链改造
2. Repo / Service / API
3. `accept -> rework -> 回流`

### 30.2 前端负责人

负责：

1. Accept Console
2. issue / evidence 视图
3. 人工干预操作

### 30.3 规则负责人

负责：

1. 四类模板规则设计
2. 规则种子数据
3. 验收阈值和 issue 分级

### 30.4 测试负责人

负责：

1. 联调路径设计
2. 灰度验证
3. 回滚预案演练

## 三十二、冻结前检查清单

在进入开发前，必须冻结以下内容：

1. `accept` 是否进入主链：必须冻结
2. 四张表字段：必须冻结
3. `passed/failed/manual_review` 三类决策语义：必须冻结
4. `accept failed -> rework` 行为：必须冻结
5. `rework -> execute/accept` 回流规则：必须冻结
6. 第一批规则模板范围：必须冻结
7. 前端首版页面范围：必须冻结

## 三十三、开发启动条件

达到以下条件后即可正式进入实施开发：

1. 本文档冻结
2. 升级 SQL 文件名与建表方案确认
3. API 路径与返回结构确认
4. 第一批范围冻结
5. 前后端负责人和测试负责人明确
6. 灰度开关策略确认
