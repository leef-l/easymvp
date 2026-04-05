-- L3.5 自治中台第一批：决策与闸门底座
-- 新建 4 张核心表

-- 1. mvp_policy_rule: 自治策略规则
CREATE TABLE IF NOT EXISTS `mvp_policy_rule` (
  `id`                     bigint unsigned NOT NULL COMMENT '雪花ID',
  `rule_code`              varchar(64)     NOT NULL COMMENT '规则编码(唯一)',
  `rule_name`              varchar(128)    NOT NULL COMMENT '规则名称',
  `decision_type`          varchar(32)     NOT NULL COMMENT '决策动作类型',
  `decision_level`         char(1)         NOT NULL COMMENT '决策等级: A/B/C',
  `trigger_source`         varchar(64)     NOT NULL COMMENT '触发源事件类型',
  `project_family`         varchar(32)     DEFAULT NULL COMMENT '适用项目家族(NULL=全局)',
  `project_category_code`  varchar(64)     DEFAULT NULL COMMENT '适用项目分类(NULL=全局)',
  `config_json`            json            NOT NULL COMMENT '规则配置(条件/阈值)',
  `enabled`                tinyint         NOT NULL DEFAULT 1 COMMENT '是否启用',
  `priority`               int             NOT NULL DEFAULT 100 COMMENT '优先级(越小越优先)',
  `created_by`             bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`                bigint unsigned NOT NULL DEFAULT 0,
  `created_at`             datetime        DEFAULT NULL,
  `updated_at`             datetime        DEFAULT NULL,
  `deleted_at`             datetime        DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_rule_code` (`rule_code`),
  KEY `idx_trigger` (`trigger_source`, `enabled`),
  KEY `idx_level` (`decision_level`),
  KEY `idx_family_cat` (`project_family`, `project_category_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='自治策略规则';

-- 2. mvp_risk_gate_rule: 风险闸门规则
CREATE TABLE IF NOT EXISTS `mvp_risk_gate_rule` (
  `id`                     bigint unsigned NOT NULL COMMENT '雪花ID',
  `gate_code`              varchar(64)     NOT NULL COMMENT '闸门编码(唯一)',
  `gate_name`              varchar(128)    NOT NULL COMMENT '闸门名称',
  `gate_type`              varchar(32)     NOT NULL COMMENT '闸门类型: permission/quality/cost/runtime',
  `project_family`         varchar(32)     DEFAULT NULL COMMENT '适用项目家族(NULL=全局)',
  `project_category_code`  varchar(64)     DEFAULT NULL COMMENT '适用项目分类(NULL=全局)',
  `trigger_expression`     json            NOT NULL COMMENT '触发表达式(JSON规则)',
  `block_action`           varchar(64)     NOT NULL COMMENT '命中后禁止的动作',
  `fallback_action`        varchar(64)     DEFAULT NULL COMMENT '命中后降级动作',
  `enabled`                tinyint         NOT NULL DEFAULT 1 COMMENT '是否启用',
  `priority`               int             NOT NULL DEFAULT 100 COMMENT '优先级(越小越优先)',
  `created_by`             bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`                bigint unsigned NOT NULL DEFAULT 0,
  `created_at`             datetime        DEFAULT NULL,
  `updated_at`             datetime        DEFAULT NULL,
  `deleted_at`             datetime        DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_gate_code` (`gate_code`),
  KEY `idx_gate_type` (`gate_type`, `enabled`),
  KEY `idx_family_cat` (`project_family`, `project_category_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='风险闸门规则';

-- 3. mvp_decision_action: 统一决策动作记录（审计链核心表）
CREATE TABLE IF NOT EXISTS `mvp_decision_action` (
  `id`                bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id`   bigint unsigned NOT NULL,
  `project_id`        bigint unsigned NOT NULL,
  `stage_run_id`      bigint unsigned DEFAULT NULL COMMENT '关联阶段运行ID',
  `domain_task_id`    bigint unsigned DEFAULT NULL COMMENT '关联领域任务ID',
  `decision_type`     varchar(32)     NOT NULL COMMENT '动作类型',
  `decision_level`    char(1)         NOT NULL COMMENT '决策等级: A/B/C',
  `trigger_source`    varchar(64)     NOT NULL COMMENT '触发源事件类型',
  `trigger_context`   json            DEFAULT NULL COMMENT '触发上下文',
  `matched_rule_id`   bigint unsigned DEFAULT NULL COMMENT '匹配的策略规则ID',
  `matched_gate_ids`  json            DEFAULT NULL COMMENT '命中的闸门ID列表',
  `action_type`       varchar(64)     NOT NULL DEFAULT '' COMMENT '策略匹配的动作类型(闸门降级后为fallback)',
  `recommendation`    json            DEFAULT NULL COMMENT '系统建议(闸门命中时含original_action)',
  `final_action`      varchar(64)     DEFAULT NULL COMMENT '最终实际执行的动作',
  `action_status`     varchar(16)     NOT NULL DEFAULT 'pending' COMMENT 'pending/auto_executed/waiting_human/approved/rejected/failed',
  `auto_executable`   tinyint         NOT NULL DEFAULT 0 COMMENT '是否可自动执行',
  `human_required`    tinyint         NOT NULL DEFAULT 0 COMMENT '是否需要人工',
  `executed_at`       datetime        DEFAULT NULL COMMENT '实际执行时间',
  `result`            json            DEFAULT NULL COMMENT '执行结果',
  `created_by`        bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`           bigint unsigned NOT NULL DEFAULT 0,
  `created_at`        datetime        DEFAULT NULL,
  `updated_at`        datetime        DEFAULT NULL,
  `deleted_at`        datetime        DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_status` (`action_status`),
  KEY `idx_type` (`decision_type`),
  KEY `idx_trigger` (`trigger_source`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='自治决策动作记录';

-- 4. mvp_human_checkpoint: 人工介入节点
CREATE TABLE IF NOT EXISTS `mvp_human_checkpoint` (
  `id`                   bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id`      bigint unsigned NOT NULL,
  `project_id`           bigint unsigned NOT NULL,
  `decision_action_id`   bigint unsigned NOT NULL COMMENT '关联的决策动作ID',
  `checkpoint_type`      varchar(32)     NOT NULL COMMENT '节点类型: manual_review/approval/escalation',
  `title`                varchar(256)    NOT NULL COMMENT '标题',
  `description`          text            DEFAULT NULL COMMENT '详细描述',
  `status`               varchar(16)     NOT NULL DEFAULT 'open' COMMENT 'open/handled/expired/canceled',
  `assigned_to`          bigint unsigned DEFAULT NULL COMMENT '指派给谁',
  `handled_by`           bigint unsigned DEFAULT NULL COMMENT '实际处理人',
  `handle_action`        varchar(32)     DEFAULT NULL COMMENT '处理动作: approve/reject/retry/rework/override',
  `handle_reason`        text            DEFAULT NULL COMMENT '处理理由',
  `handled_at`           datetime        DEFAULT NULL COMMENT '处理时间',
  `expires_at`           datetime        DEFAULT NULL COMMENT '过期时间',
  `created_by`           bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`              bigint unsigned NOT NULL DEFAULT 0,
  `created_at`           datetime        DEFAULT NULL,
  `updated_at`           datetime        DEFAULT NULL,
  `deleted_at`           datetime        DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_decision_action` (`decision_action_id`),
  KEY `idx_status` (`status`),
  KEY `idx_assigned` (`assigned_to`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='人工介入节点';
