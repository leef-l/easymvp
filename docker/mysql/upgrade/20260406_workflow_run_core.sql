-- ============================================================
-- WorkflowRun M1: 核心表 (第一批 7 表 + 旧表兼容字段)
-- 执行顺序: 先于 20260407_workflow_run_execution.sql
-- 幂等: 全部使用 IF NOT EXISTS / IF NOT EXISTS 列检测
-- ============================================================

-- -----------------------------------------------------------
-- 1. mvp_workflow_run  工作流运行实例
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS `mvp_workflow_run` (
  `id`                     bigint unsigned  NOT NULL                COMMENT '雪花ID',
  `project_id`             bigint unsigned  NOT NULL                COMMENT '所属项目ID',
  `run_no`                 int              NOT NULL                COMMENT '项目内运行序号(从1递增)',
  `status`                 varchar(32)      NOT NULL                COMMENT '状态: pending/running/paused/completed/canceled',
  `current_stage`          varchar(32)      NOT NULL                COMMENT '当前阶段: design/review/execute/rework/complete',
  `current_stage_run_id`   bigint unsigned  DEFAULT NULL            COMMENT '当前阶段运行ID',
  `active_plan_version_id` bigint unsigned  DEFAULT NULL            COMMENT '当前活跃计划版本ID',
  `pause_reason`           varchar(500)     DEFAULT NULL            COMMENT '暂停原因',
  `cancel_reason`          varchar(500)     DEFAULT NULL            COMMENT '取消原因',
  `runtime_token`          varchar(64)      DEFAULT NULL            COMMENT '运行时令牌(防重入)',
  `started_at`             datetime         DEFAULT NULL            COMMENT '开始时间',
  `finished_at`            datetime         DEFAULT NULL            COMMENT '结束时间',
  `created_by`             bigint unsigned  DEFAULT 0               COMMENT '创建人ID',
  `dept_id`                bigint unsigned  DEFAULT 0               COMMENT '所属部门ID',
  `created_at`             datetime         NOT NULL                COMMENT '创建时间',
  `updated_at`             datetime         NOT NULL                COMMENT '更新时间',
  `deleted_at`             datetime         DEFAULT NULL            COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_run_no` (`project_id`, `run_no`),
  KEY `idx_project_status` (`project_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流运行实例';

-- -----------------------------------------------------------
-- 2. mvp_stage_run  阶段运行实例
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS `mvp_stage_run` (
  `id`              bigint unsigned  NOT NULL                COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned  NOT NULL                COMMENT '所属工作流运行ID',
  `stage_type`      varchar(32)      NOT NULL                COMMENT '阶段类型: design/review/execute/rework/complete',
  `stage_no`        int              NOT NULL                COMMENT '同类型阶段序号(支持多轮)',
  `status`          varchar(32)      NOT NULL                COMMENT '状态: pending/running/completed/failed/skipped',
  `input_ref`       json             DEFAULT NULL            COMMENT '阶段输入引用(JSON)',
  `output_ref`      json             DEFAULT NULL            COMMENT '阶段输出引用(JSON)',
  `decision`        json             DEFAULT NULL            COMMENT '阶段决策结果(JSON)',
  `error_message`   text             DEFAULT NULL            COMMENT '错误信息',
  `started_at`      datetime         DEFAULT NULL            COMMENT '开始时间',
  `finished_at`     datetime         DEFAULT NULL            COMMENT '结束时间',
  `created_at`      datetime         NOT NULL                COMMENT '创建时间',
  `updated_at`      datetime         NOT NULL                COMMENT '更新时间',
  `deleted_at`      datetime         DEFAULT NULL            COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_stage` (`workflow_run_id`, `stage_type`, `stage_no`),
  KEY `idx_workflow_status` (`workflow_run_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='阶段运行实例';

-- -----------------------------------------------------------
-- 3. mvp_plan_version  计划版本
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS `mvp_plan_version` (
  `id`                      bigint unsigned  NOT NULL                COMMENT '雪花ID',
  `project_id`              bigint unsigned  NOT NULL                COMMENT '所属项目ID',
  `workflow_run_id`         bigint unsigned  NOT NULL                COMMENT '所属工作流运行ID',
  `version_no`              int              NOT NULL                COMMENT '版本号(项目内递增)',
  `source_conversation_id`  bigint unsigned  DEFAULT NULL            COMMENT '来源对话ID',
  `source_message_id`       bigint unsigned  DEFAULT NULL            COMMENT '来源消息ID',
  `status`                  varchar(32)      NOT NULL                COMMENT '版本状态: draft/active/superseded',
  `review_status`           varchar(32)      NOT NULL                COMMENT '审核状态: pending/approved/rejected',
  `summary`                 text             DEFAULT NULL            COMMENT '版本摘要',
  `diff_summary`            text             DEFAULT NULL            COMMENT '与上一版本的差异摘要',
  `approved_at`             datetime         DEFAULT NULL            COMMENT '审核通过时间',
  `rejected_at`             datetime         DEFAULT NULL            COMMENT '审核驳回时间',
  `created_at`              datetime         NOT NULL                COMMENT '创建时间',
  `updated_at`              datetime         NOT NULL                COMMENT '更新时间',
  `deleted_at`              datetime         DEFAULT NULL            COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_version` (`project_id`, `version_no`),
  KEY `idx_workflow_status` (`workflow_run_id`, `status`, `review_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计划版本';

-- -----------------------------------------------------------
-- 4. mvp_task_blueprint  任务蓝图
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS `mvp_task_blueprint` (
  `id`                       bigint unsigned  NOT NULL                COMMENT '雪花ID',
  `plan_version_id`          bigint unsigned  NOT NULL                COMMENT '所属计划版本ID',
  `parent_blueprint_id`      bigint unsigned  DEFAULT NULL            COMMENT '父蓝图ID(支持层级)',
  `name`                     varchar(255)     NOT NULL                COMMENT '任务名称',
  `description`              text             NOT NULL                COMMENT '任务描述',
  `role_type`                varchar(32)      NOT NULL                COMMENT '角色类型: architect/implementer/auditor/coordinator',
  `role_level`               varchar(32)      NOT NULL                COMMENT '角色等级: lite/pro/max',
  `batch_no`                 int              NOT NULL                COMMENT '批次号',
  `sort`                     int              NOT NULL DEFAULT 0      COMMENT '排序',
  `affected_resources`       json             DEFAULT NULL            COMMENT '影响资源列表(JSON)',
  `depends_on_blueprint_ids` json             DEFAULT NULL            COMMENT '依赖蓝图ID列表(JSON)',
  `blueprint_status`         varchar(32)      NOT NULL                COMMENT '蓝图状态: draft/confirmed/superseded',
  `created_at`               datetime         NOT NULL                COMMENT '创建时间',
  `updated_at`               datetime         NOT NULL                COMMENT '更新时间',
  `deleted_at`               datetime         DEFAULT NULL            COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_plan_batch` (`plan_version_id`, `batch_no`, `sort`),
  KEY `idx_plan_status` (`plan_version_id`, `blueprint_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务蓝图';

-- -----------------------------------------------------------
-- 5. mvp_review_issue  审核问题
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS `mvp_review_issue` (
  `id`              bigint unsigned  NOT NULL                COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned  NOT NULL                COMMENT '所属工作流运行ID',
  `stage_run_id`    bigint unsigned  NOT NULL                COMMENT '所属阶段运行ID',
  `plan_version_id` bigint unsigned  NOT NULL                COMMENT '所属计划版本ID',
  `blueprint_id`    bigint unsigned  DEFAULT NULL            COMMENT '关联蓝图ID',
  `severity`        varchar(16)      NOT NULL                COMMENT '严重级别: error/warning/info',
  `issue_code`      varchar(64)      NOT NULL                COMMENT '问题代码',
  `issue_type`      varchar(32)      NOT NULL                COMMENT '问题类型',
  `source_role`     varchar(32)      NOT NULL                COMMENT '发现角色',
  `task_name`       varchar(255)     DEFAULT NULL            COMMENT '关联任务名',
  `message`         text             NOT NULL                COMMENT '问题描述',
  `suggestion`      text             DEFAULT NULL            COMMENT '修复建议',
  `status`          varchar(32)      NOT NULL                COMMENT '状态: open/resolved/ignored',
  `resolved_at`     datetime         DEFAULT NULL            COMMENT '解决时间',
  `created_at`      datetime         NOT NULL                COMMENT '创建时间',
  `updated_at`      datetime         NOT NULL                COMMENT '更新时间',
  `deleted_at`      datetime         DEFAULT NULL            COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_plan_issue` (`plan_version_id`, `severity`, `status`),
  KEY `idx_blueprint_issue` (`blueprint_id`, `severity`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审核问题';

-- -----------------------------------------------------------
-- 6. mvp_stage_task  阶段任务(审核/完结等阶段内的子任务)
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS `mvp_stage_task` (
  `id`              bigint unsigned  NOT NULL                COMMENT '雪花ID',
  `stage_run_id`    bigint unsigned  NOT NULL                COMMENT '所属阶段运行ID',
  `task_type`       varchar(32)      NOT NULL                COMMENT '任务类型: precheck/auditor_review/coordinator_optimize/review_summary',
  `role_type`       varchar(32)      NOT NULL                COMMENT '执行角色',
  `status`          varchar(32)      NOT NULL                COMMENT '状态: pending/running/completed/failed/skipped',
  `input_payload`   json             DEFAULT NULL            COMMENT '输入载荷(JSON)',
  `output_payload`  json             DEFAULT NULL            COMMENT '输出载荷(JSON)',
  `error_message`   text             DEFAULT NULL            COMMENT '错误信息',
  `started_at`      datetime         DEFAULT NULL            COMMENT '开始时间',
  `completed_at`    datetime         DEFAULT NULL            COMMENT '完成时间',
  `created_at`      datetime         NOT NULL                COMMENT '创建时间',
  `updated_at`      datetime         NOT NULL                COMMENT '更新时间',
  `deleted_at`      datetime         DEFAULT NULL            COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_stage_type` (`stage_run_id`, `task_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='阶段任务';

-- -----------------------------------------------------------
-- 7. mvp_workflow_event  工作流事件(事件溯源/审计)
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS `mvp_workflow_event` (
  `id`              bigint unsigned  NOT NULL                COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned  NOT NULL                COMMENT '所属工作流运行ID',
  `stage_run_id`    bigint unsigned  DEFAULT NULL            COMMENT '关联阶段运行ID',
  `entity_type`     varchar(32)      NOT NULL                COMMENT '实体类型: workflow_run/stage_run/plan_version/domain_task/review_issue',
  `entity_id`       bigint unsigned  DEFAULT NULL            COMMENT '实体ID',
  `event_type`      varchar(64)      NOT NULL                COMMENT '事件类型: workflow.created/stage.started/task.completed等',
  `payload`         json             DEFAULT NULL            COMMENT '事件载荷(JSON)',
  `created_at`      datetime         NOT NULL                COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_event` (`workflow_run_id`, `created_at`),
  KEY `idx_entity_event` (`entity_type`, `entity_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流事件';

-- -----------------------------------------------------------
-- 8. 旧表兼容字段: mvp_project 新增 engine_version + active_workflow_run_id
-- -----------------------------------------------------------
-- 幂等检测: 仅在列不存在时添加
SET @db_name = DATABASE();

SELECT COUNT(*) INTO @col_exists
FROM information_schema.columns
WHERE table_schema = @db_name AND table_name = 'mvp_project' AND column_name = 'engine_version';

SET @add_engine_version = IF(@col_exists = 0,
  'ALTER TABLE `mvp_project` ADD COLUMN `engine_version` varchar(32) DEFAULT ''legacy'' COMMENT ''执行引擎版本: legacy/workflow_v2'' AFTER `active_batch_no`',
  'SELECT ''engine_version already exists''');
PREPARE stmt FROM @add_engine_version;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @col_exists2
FROM information_schema.columns
WHERE table_schema = @db_name AND table_name = 'mvp_project' AND column_name = 'active_workflow_run_id';

SET @add_workflow_run_id = IF(@col_exists2 = 0,
  'ALTER TABLE `mvp_project` ADD COLUMN `active_workflow_run_id` bigint unsigned DEFAULT NULL COMMENT ''当前活跃工作流运行ID'' AFTER `engine_version`',
  'SELECT ''active_workflow_run_id already exists''');
PREPARE stmt FROM @add_workflow_run_id;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
