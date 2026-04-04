-- ============================================================
-- WorkflowRun M1: 执行表 (第二批 3 表)
-- 执行顺序: 在 20260406_workflow_run_core.sql 之后
-- 幂等: 全部使用 IF NOT EXISTS
-- ============================================================

-- -----------------------------------------------------------
-- 1. mvp_domain_task  领域任务(新执行模型的核心任务表)
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS `mvp_domain_task` (
  `id`                bigint unsigned  NOT NULL                COMMENT '雪花ID',
  `workflow_run_id`   bigint unsigned  NOT NULL                COMMENT '所属工作流运行ID',
  `stage_run_id`      bigint unsigned  NOT NULL                COMMENT '所属阶段运行ID',
  `plan_version_id`   bigint unsigned  DEFAULT NULL            COMMENT '来源计划版本ID',
  `blueprint_id`      bigint unsigned  DEFAULT NULL            COMMENT '来源蓝图ID',
  `parent_task_id`    bigint unsigned  DEFAULT NULL            COMMENT '父任务ID',
  `source_task_id`    bigint unsigned  DEFAULT NULL            COMMENT '来源任务ID(链路追踪)',
  `root_task_id`      bigint unsigned  DEFAULT NULL            COMMENT '根任务ID(链路追踪)',
  `task_kind`         varchar(32)      NOT NULL                COMMENT '任务种类: implement/audit/bug_analysis/failure_analysis',
  `name`              varchar(255)     NOT NULL                COMMENT '任务名称',
  `description`       text             NOT NULL                COMMENT '任务描述',
  `role_type`         varchar(32)      NOT NULL                COMMENT '角色类型',
  `role_level`        varchar(32)      NOT NULL                COMMENT '角色等级',
  `execution_mode`    varchar(32)      NOT NULL                COMMENT '执行方式: chat/aider/openhands',
  `status`            varchar(32)      NOT NULL                COMMENT '状态: pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched',
  `conversation_id`   bigint unsigned  DEFAULT NULL            COMMENT '关联对话ID',
  `model_id`          bigint unsigned  DEFAULT NULL            COMMENT '使用的AI模型ID',
  `batch_no`          int              NOT NULL                COMMENT '批次号',
  `sort`              int              NOT NULL DEFAULT 0      COMMENT '排序',
  `retry_count`       int              NOT NULL DEFAULT 0      COMMENT '重试次数',
  `affected_resources` json            DEFAULT NULL            COMMENT '影响资源列表(JSON)',
  `locked_resources`  json             DEFAULT NULL            COMMENT '锁定资源列表(JSON)',
  `result`            longtext         DEFAULT NULL            COMMENT '执行结果',
  `context_summary`   text             DEFAULT NULL            COMMENT '上下文摘要',
  `heartbeat_at`      datetime         DEFAULT NULL            COMMENT '心跳时间',
  `started_at`        datetime         DEFAULT NULL            COMMENT '开始时间',
  `completed_at`      datetime         DEFAULT NULL            COMMENT '完成时间',
  `created_at`        datetime         NOT NULL                COMMENT '创建时间',
  `updated_at`        datetime         NOT NULL                COMMENT '更新时间',
  `deleted_at`        datetime         DEFAULT NULL            COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_status` (`workflow_run_id`, `status`),
  KEY `idx_workflow_batch` (`workflow_run_id`, `batch_no`, `sort`),
  KEY `idx_stage_status` (`stage_run_id`, `status`),
  KEY `idx_root_task` (`root_task_id`),
  KEY `idx_source_task` (`source_task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='领域任务';

-- -----------------------------------------------------------
-- 2. mvp_task_resource_lock  任务资源锁
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS `mvp_task_resource_lock` (
  `id`              bigint unsigned  NOT NULL                COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned  NOT NULL                COMMENT '所属工作流运行ID',
  `task_id`         bigint unsigned  NOT NULL                COMMENT '持锁任务ID',
  `resource_path`   varchar(500)     NOT NULL                COMMENT '资源路径',
  `lock_status`     varchar(32)      NOT NULL                COMMENT '锁状态: held/released/leaked',
  `locked_at`       datetime         DEFAULT NULL            COMMENT '加锁时间',
  `released_at`     datetime         DEFAULT NULL            COMMENT '释放时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_resource` (`task_id`, `resource_path`),
  KEY `idx_workflow_resource` (`workflow_run_id`, `resource_path`, `lock_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务资源锁';

-- -----------------------------------------------------------
-- 3. mvp_handoff_record  交接记录(任务间交接/升级/返工)
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS `mvp_handoff_record` (
  `id`              bigint unsigned  NOT NULL                COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned  NOT NULL                COMMENT '所属工作流运行ID',
  `from_task_id`    bigint unsigned  DEFAULT NULL            COMMENT '来源任务ID',
  `to_task_id`      bigint unsigned  DEFAULT NULL            COMMENT '目标任务ID',
  `handoff_type`    varchar(32)      NOT NULL                COMMENT '交接类型: bug_fix/failure_escalation/rework/audit',
  `reason`          text             DEFAULT NULL            COMMENT '交接原因',
  `payload`         json             DEFAULT NULL            COMMENT '交接载荷(JSON)',
  `created_at`      datetime         NOT NULL                COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_type` (`workflow_run_id`, `handoff_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交接记录';
