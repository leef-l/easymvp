# WorkflowRun 数据库升级与迁移实施设计文档

> 文档定位：`WorkflowRun阶段化工作流引擎重构架构设计文档` 的数据库落地方案
>
> 目标：给出可执行的建表、索引、迁移、兼容、灰度、回滚设计，支撑新工作流内核分阶段上线。

---

## 一、实施目标

本次数据库升级目标不是简单加几个字段，而是引入一套新执行模型，同时保证：

1. 新架构可独立运行。
2. 旧项目可继续使用旧表。
3. 新旧结构在过渡期可并存。
4. 升级脚本可重复执行、可灰度、可回滚。

---

## 二、迁移原则

### 2.1 基本原则

1. 先加表，不删旧表。
2. 先加读路径，再加写路径。
3. 新项目走新结构，旧项目不强迁。
4. 所有升级 SQL 必须幂等。
5. 所有核心新表先不加外键，避免上线阶段被数据顺序卡死。

### 2.2 为什么不直接改造 `mvp_task`

原因：

1. 当前 `mvp_task` 已承载草稿、执行、返工、分析、审计等多重语义。
2. 直接在旧表上硬扩展会把新架构继续拖回旧语义。
3. 新旧项目双轨运行期需要明显边界。

结论：

- 新执行模型单独建表
- `mvp_task` 作为旧路径兼容层

---

## 三、阶段化表结构落地策略

### 3.1 第一批必须上线的表

第一阶段必须落地以下表：

1. `mvp_workflow_run`
2. `mvp_stage_run`
3. `mvp_plan_version`
4. `mvp_task_blueprint`
5. `mvp_review_issue`
6. `mvp_stage_task`
7. `mvp_workflow_event`

### 3.2 第二批表

执行阶段切换时落地：

1. `mvp_domain_task`
2. `mvp_task_resource_lock`
3. `mvp_handoff_record`

### 3.3 旧表保留策略

继续保留：

- `mvp_project`
- `mvp_task`
- `mvp_task_dependency`
- `mvp_conversation`
- `mvp_message`
- `mvp_task_log`

---

## 四、具体 SQL 设计

### 4.1 升级文件命名建议

建议新增升级文件：

- `20260406_workflow_run_core.sql`
- `20260406_workflow_run_indexes.sql`
- `20260407_workflow_run_execution.sql`
- `20260407_workflow_run_backfill.sql`

### 4.2 `20260406_workflow_run_core.sql`

建议内容：

```sql
CREATE TABLE IF NOT EXISTS `mvp_workflow_run` (
  `id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned NOT NULL,
  `run_no` int NOT NULL,
  `status` varchar(32) NOT NULL,
  `current_stage` varchar(32) NOT NULL,
  `current_stage_run_id` bigint unsigned DEFAULT NULL,
  `active_plan_version_id` bigint unsigned DEFAULT NULL,
  `pause_reason` varchar(500) DEFAULT NULL,
  `cancel_reason` varchar(500) DEFAULT NULL,
  `runtime_token` varchar(64) DEFAULT NULL,
  `started_at` datetime DEFAULT NULL,
  `finished_at` datetime DEFAULT NULL,
  `created_by` bigint unsigned DEFAULT 0,
  `dept_id` bigint unsigned DEFAULT 0,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_run_no` (`project_id`, `run_no`),
  KEY `idx_project_status` (`project_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mvp_stage_run` (
  `id` bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `stage_type` varchar(32) NOT NULL,
  `stage_no` int NOT NULL,
  `status` varchar(32) NOT NULL,
  `input_ref` json DEFAULT NULL,
  `output_ref` json DEFAULT NULL,
  `decision` json DEFAULT NULL,
  `error_message` text DEFAULT NULL,
  `started_at` datetime DEFAULT NULL,
  `finished_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow_stage` (`workflow_run_id`, `stage_type`, `stage_no`),
  KEY `idx_workflow_status` (`workflow_run_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mvp_plan_version` (
  `id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `version_no` int NOT NULL,
  `source_conversation_id` bigint unsigned DEFAULT NULL,
  `source_message_id` bigint unsigned DEFAULT NULL,
  `status` varchar(32) NOT NULL,
  `review_status` varchar(32) NOT NULL,
  `summary` text DEFAULT NULL,
  `diff_summary` text DEFAULT NULL,
  `approved_at` datetime DEFAULT NULL,
  `rejected_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_version` (`project_id`, `version_no`),
  KEY `idx_workflow_status` (`workflow_run_id`, `status`, `review_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mvp_task_blueprint` (
  `id` bigint unsigned NOT NULL,
  `plan_version_id` bigint unsigned NOT NULL,
  `parent_blueprint_id` bigint unsigned DEFAULT NULL,
  `name` varchar(255) NOT NULL,
  `description` text NOT NULL,
  `role_type` varchar(32) NOT NULL,
  `role_level` varchar(32) NOT NULL,
  `batch_no` int NOT NULL,
  `sort` int NOT NULL DEFAULT 0,
  `affected_resources` json DEFAULT NULL,
  `depends_on_blueprint_ids` json DEFAULT NULL,
  `blueprint_status` varchar(32) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_plan_batch` (`plan_version_id`, `batch_no`, `sort`),
  KEY `idx_plan_status` (`plan_version_id`, `blueprint_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mvp_review_issue` (
  `id` bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `stage_run_id` bigint unsigned NOT NULL,
  `plan_version_id` bigint unsigned NOT NULL,
  `blueprint_id` bigint unsigned DEFAULT NULL,
  `severity` varchar(16) NOT NULL,
  `issue_code` varchar(64) NOT NULL,
  `issue_type` varchar(32) NOT NULL,
  `source_role` varchar(32) NOT NULL,
  `task_name` varchar(255) DEFAULT NULL,
  `message` text NOT NULL,
  `suggestion` text DEFAULT NULL,
  `status` varchar(32) NOT NULL,
  `resolved_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_plan_issue` (`plan_version_id`, `severity`, `status`),
  KEY `idx_blueprint_issue` (`blueprint_id`, `severity`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mvp_stage_task` (
  `id` bigint unsigned NOT NULL,
  `stage_run_id` bigint unsigned NOT NULL,
  `task_type` varchar(32) NOT NULL,
  `role_type` varchar(32) NOT NULL,
  `status` varchar(32) NOT NULL,
  `input_payload` json DEFAULT NULL,
  `output_payload` json DEFAULT NULL,
  `error_message` text DEFAULT NULL,
  `started_at` datetime DEFAULT NULL,
  `completed_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_stage_type` (`stage_run_id`, `task_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mvp_workflow_event` (
  `id` bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `stage_run_id` bigint unsigned DEFAULT NULL,
  `entity_type` varchar(32) NOT NULL,
  `entity_id` bigint unsigned DEFAULT NULL,
  `event_type` varchar(64) NOT NULL,
  `payload` json DEFAULT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow_event` (`workflow_run_id`, `created_at`),
  KEY `idx_entity_event` (`entity_type`, `entity_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.3 `20260407_workflow_run_execution.sql`

```sql
CREATE TABLE IF NOT EXISTS `mvp_domain_task` (
  `id` bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `stage_run_id` bigint unsigned NOT NULL,
  `plan_version_id` bigint unsigned DEFAULT NULL,
  `blueprint_id` bigint unsigned DEFAULT NULL,
  `parent_task_id` bigint unsigned DEFAULT NULL,
  `source_task_id` bigint unsigned DEFAULT NULL,
  `root_task_id` bigint unsigned DEFAULT NULL,
  `task_kind` varchar(32) NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` text NOT NULL,
  `role_type` varchar(32) NOT NULL,
  `role_level` varchar(32) NOT NULL,
  `execution_mode` varchar(32) NOT NULL,
  `status` varchar(32) NOT NULL,
  `conversation_id` bigint unsigned DEFAULT NULL,
  `model_id` bigint unsigned DEFAULT NULL,
  `batch_no` int NOT NULL,
  `sort` int NOT NULL DEFAULT 0,
  `retry_count` int NOT NULL DEFAULT 0,
  `affected_resources` json DEFAULT NULL,
  `locked_resources` json DEFAULT NULL,
  `result` longtext DEFAULT NULL,
  `context_summary` text DEFAULT NULL,
  `heartbeat_at` datetime DEFAULT NULL,
  `started_at` datetime DEFAULT NULL,
  `completed_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow_status` (`workflow_run_id`, `status`),
  KEY `idx_workflow_batch` (`workflow_run_id`, `batch_no`, `sort`),
  KEY `idx_root_task` (`root_task_id`),
  KEY `idx_source_task` (`source_task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mvp_task_resource_lock` (
  `id` bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `task_id` bigint unsigned NOT NULL,
  `resource_path` varchar(500) NOT NULL,
  `lock_status` varchar(32) NOT NULL,
  `locked_at` datetime DEFAULT NULL,
  `released_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_resource` (`task_id`, `resource_path`),
  KEY `idx_workflow_resource` (`workflow_run_id`, `resource_path`, `lock_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mvp_handoff_record` (
  `id` bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `from_task_id` bigint unsigned DEFAULT NULL,
  `to_task_id` bigint unsigned DEFAULT NULL,
  `handoff_type` varchar(32) NOT NULL,
  `reason` text DEFAULT NULL,
  `payload` json DEFAULT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow_type` (`workflow_run_id`, `handoff_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

## 五、旧表兼容与新增字段

### 5.1 `mvp_project`

保留，但建议新增：

- `active_workflow_run_id`
- `engine_version`

SQL：

```sql
ALTER TABLE `mvp_project`
  ADD COLUMN `active_workflow_run_id` bigint unsigned DEFAULT NULL COMMENT '当前活跃工作流运行ID',
  ADD COLUMN `engine_version` varchar(32) DEFAULT 'legacy' COMMENT '执行引擎版本：legacy/workflow_v2';
```

### 5.2 `mvp_conversation`

建议新增：

- `workflow_run_id`
- `stage_run_id`
- `plan_version_id`
- `domain_task_id`

目的是让会话归属从"项目/任务"升级为"工作流/阶段/计划/任务"。

### 5.3 `mvp_message`

建议新增：

- `workflow_run_id`
- `stage_run_id`
- `entity_type`
- `entity_id`

便于消息层统一接入事件体系。

---

## 六、数据写入策略

### 6.1 Phase 1

新表已存在，但不写业务数据。

### 6.2 Phase 2

新项目创建时：

- `mvp_project.engine_version = workflow_v2`
- 创建 `mvp_workflow_run`
- 创建 `design stage_run`

旧项目仍保留：

- `engine_version = legacy`

### 6.3 Phase 3

设计与审核链写新表：

- 对话产生 `plan_version`
- 任务草稿写入 `task_blueprint`
- 审核结果写入 `review_issue`

执行链仍可临时映射回 `mvp_task`

### 6.4 Phase 4

执行链切换到 `mvp_domain_task`

---

## 七、回填策略

### 7.1 不做全量历史强迁移

原因：

1. 历史数据语义不完整
2. 旧任务表和新任务表语义不同
3. 强迁移成本高且收益有限

### 7.2 可选回填内容

只建议回填以下弱一致数据：

1. 为活跃项目创建默认 `workflow_run`
2. 将当前项目状态映射到 `workflow_run.status`
3. 将当前架构师最后一次拆分结果生成初始 `plan_version`

### 7.3 不建议回填的内容

- 历史 `review_issue`
- 历史 `stage_run`
- 历史返工链完整关系

这些宁可从新架构启用后重新沉淀。

---

## 八、读写兼容策略

### 8.1 路由策略

接口读取根据 `project.engine_version` 决定：

- `legacy`：读旧表
- `workflow_v2`：读新表

### 8.2 写入策略

过渡期避免双写主链，采用单写：

- 旧项目只写旧表
- 新项目只写新表

### 8.3 聚合查询兼容

项目列表页可做聚合适配：

- 总进度
- 当前状态
- 任务数
- 最近活跃时间

按 `engine_version` 分别查询，再统一转成前端 DTO。

---

## 九、升级执行顺序

### 9.1 数据库顺序

1. 核心表
2. 索引
3. 兼容字段
4. 执行表
5. 回填脚本

### 9.2 应用发布顺序

1. 发布只认识新表但不使用新表的版本
2. 执行 SQL
3. 发布支持 `engine_version` 分流的版本
4. 小流量新项目切到 `workflow_v2`
5. 扩大灰度

---

## 十、回滚方案

### 10.1 原则

由于采用新增表 + 新项目分流，回滚相对简单。

### 10.2 回滚方式

如果新架构有问题：

1. 停止新项目创建 `workflow_v2`
2. 将新项目入口降级回 `legacy`
3. 不删除已建新表
4. 保留数据等待修复后继续使用

### 10.3 不建议回滚的动作

- 不建议删除新表
- 不建议把新项目数据强写回旧表

这样只会制造更大不一致。

---

## 十一、实施清单

### 11.1 SQL 清单

- 新增核心表 SQL
- 新增执行表 SQL
- 旧表兼容字段 SQL
- 索引 SQL
- 初始化脚本

### 11.2 后端配套

- DAO 生成
- Entity/DO/DAO 常量
- 服务层分流
- 兼容 DTO

### 11.3 验证清单

- 新项目建表链路
- 新项目设计阶段写入
- 新项目审核阶段写入
- 新项目执行阶段写入
- 旧项目查询不受影响
- 项目列表混合展示正常

---

## 十二、结论

数据库升级必须坚持"新旧双轨、单写分流、增量迁移"的策略。

本次实施最重要的不是表数量，而是明确：

1. 新运行模型绝不继续污染旧 `mvp_task`
2. 新项目直接进入新结构
3. 历史项目不强迁
4. 读写边界清晰

这四点守住了，后端和前端重构才不会失控。
