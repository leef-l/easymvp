-- L4 Phase D：元认知层 — Observer + Assessor + Tuner + Learner
-- 新建 observation_record / assessment_result / tune_recommendation / learning_record 表 + 灰度开关

-- 1. mvp_observation_record: 决策观测记录（Observer 写入）
CREATE TABLE IF NOT EXISTS `mvp_observation_record` (
  `id`               bigint unsigned NOT NULL COMMENT '雪花ID',
  `decision_action_id` bigint unsigned NOT NULL COMMENT '关联 mvp_decision_action.id',
  `workflow_run_id`  bigint unsigned NOT NULL COMMENT '关联 workflow_run',
  `project_id`       bigint unsigned NOT NULL COMMENT '关联 project',
  `decision_type`    varchar(64) NOT NULL DEFAULT '' COMMENT '决策类型: policy_match/strategy:xxx/objective_guard',
  `trigger_source`   varchar(64) NOT NULL DEFAULT '' COMMENT '触发源',
  `decision_level`   varchar(8) NOT NULL DEFAULT '' COMMENT '决策级别 A/B/C',
  `action_type`      varchar(64) NOT NULL DEFAULT '' COMMENT '动作类型',
  `input_snapshot`   json DEFAULT NULL COMMENT '决策输入快照(JSON)',
  `output_snapshot`  json DEFAULT NULL COMMENT '决策输出快照(JSON)',
  `meta_snapshot`    json DEFAULT NULL COMMENT 'DecisionMeta 快照(JSON)',
  `outcome`          varchar(16) NOT NULL DEFAULT 'pending' COMMENT '结果: success/failure/neutral/pending',
  `effect_score`     decimal(5,3) NOT NULL DEFAULT 0 COMMENT '效果得分(-1~1)',
  `human_override`   tinyint NOT NULL DEFAULT 0 COMMENT '人工是否干预(0否/1是)',
  `override_reason`  varchar(512) NOT NULL DEFAULT '' COMMENT '人工干预原因',
  `signal_weight`    decimal(3,1) NOT NULL DEFAULT 0 COMMENT '学习信号权重(0-1)',
  `created_by`       bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`          bigint unsigned NOT NULL DEFAULT 0,
  `created_at`       datetime DEFAULT NULL,
  `deleted_at`       datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_decision_action` (`decision_action_id`),
  KEY `idx_workflow_run` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_decision_type` (`decision_type`),
  KEY `idx_outcome` (`outcome`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='决策观测记录';

-- 2. mvp_assessment_result: 系统评估结果（Assessor 定期生成）
CREATE TABLE IF NOT EXISTS `mvp_assessment_result` (
  `id`                  bigint unsigned NOT NULL COMMENT '雪花ID',
  `project_id`          bigint unsigned NOT NULL DEFAULT 0 COMMENT '关联 project(0=全局)',
  `period_start`        datetime NOT NULL COMMENT '评估周期开始',
  `period_end`          datetime NOT NULL COMMENT '评估周期结束',
  `sample_count`        int NOT NULL DEFAULT 0 COMMENT '样本数量',
  `policy_accuracy`     decimal(5,3) NOT NULL DEFAULT 0 COMMENT '策略准确率(0~1)',
  `gate_false_positive` decimal(5,3) NOT NULL DEFAULT 0 COMMENT '闸门误报率(0~1)',
  `gate_false_negative` decimal(5,3) NOT NULL DEFAULT 0 COMMENT '闸门漏报率(0~1)',
  `human_override_rate` decimal(5,3) NOT NULL DEFAULT 0 COMMENT '人工干预率(0~1)',
  `match_accuracy`      decimal(5,3) NOT NULL DEFAULT 0 COMMENT '匹配准确率(0~1)',
  `cost_efficiency`     decimal(5,3) NOT NULL DEFAULT 0 COMMENT '成本效率(0~1)',
  `drifts`              json DEFAULT NULL COMMENT '参数偏差列表(JSON)',
  `summary`             text COMMENT '评估摘要',
  `created_by`          bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`             bigint unsigned NOT NULL DEFAULT 0,
  `created_at`          datetime DEFAULT NULL,
  `deleted_at`          datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_period` (`period_start`, `period_end`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='系统评估结果';

-- 3. mvp_tune_recommendation: 调参建议（Tuner 生成）
CREATE TABLE IF NOT EXISTS `mvp_tune_recommendation` (
  `id`              bigint unsigned NOT NULL COMMENT '雪花ID',
  `assessment_id`   bigint unsigned NOT NULL DEFAULT 0 COMMENT '关联评估结果',
  `project_id`      bigint unsigned NOT NULL DEFAULT 0 COMMENT '关联 project(0=全局)',
  `parameter`       varchar(128) NOT NULL DEFAULT '' COMMENT '参数名',
  `current_value`   varchar(256) NOT NULL DEFAULT '' COMMENT '当前值',
  `suggested_value` varchar(256) NOT NULL DEFAULT '' COMMENT '建议值',
  `direction`       varchar(16) NOT NULL DEFAULT '' COMMENT 'conservative/aggressive',
  `reasoning`       text COMMENT '建议理由',
  `confidence`      decimal(5,3) NOT NULL DEFAULT 0 COMMENT '置信度(0~1)',
  `auto_applicable` tinyint NOT NULL DEFAULT 0 COMMENT '是否可自动应用(0否/1是)',
  `risk_level`      varchar(16) NOT NULL DEFAULT 'low' COMMENT 'low/medium/high',
  `status`          varchar(16) NOT NULL DEFAULT 'pending' COMMENT 'pending/applied/rejected/expired',
  `applied_at`      datetime DEFAULT NULL COMMENT '应用时间',
  `applied_by`      bigint unsigned NOT NULL DEFAULT 0 COMMENT '应用人',
  `created_by`      bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`         bigint unsigned NOT NULL DEFAULT 0,
  `created_at`      datetime DEFAULT NULL,
  `deleted_at`      datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_assessment` (`assessment_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_parameter` (`parameter`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='调参建议';

-- 4. mvp_learning_record: EMA 学习记录（Learner 维护）
CREATE TABLE IF NOT EXISTS `mvp_learning_record` (
  `id`              bigint unsigned NOT NULL COMMENT '雪花ID',
  `metric_key`      varchar(128) NOT NULL DEFAULT '' COMMENT '指标名称(如 strategy.cost_guard.accuracy)',
  `project_id`      bigint unsigned NOT NULL DEFAULT 0 COMMENT '关联 project(0=全局)',
  `ema_value`       decimal(10,6) NOT NULL DEFAULT 0 COMMENT 'EMA 当前值',
  `raw_value`       decimal(10,6) NOT NULL DEFAULT 0 COMMENT '最新原始值',
  `sample_count`    int NOT NULL DEFAULT 0 COMMENT '累计样本数',
  `last_updated`    datetime NOT NULL COMMENT '最后更新时间',
  `decay_factor`    decimal(5,3) NOT NULL DEFAULT 0.900 COMMENT 'EMA 衰减因子',
  `created_by`      bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`         bigint unsigned NOT NULL DEFAULT 0,
  `created_at`      datetime DEFAULT NULL,
  `deleted_at`      datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_metric_project` (`metric_key`, `project_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_last_updated` (`last_updated`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='EMA学习记录';

-- 5. 元认知灰度配置项
INSERT INTO `mvp_config` (`config_key`, `config_value`, `config_type`, `category`, `description`, `created_at`, `updated_at`)
VALUES
  ('workflow.autonomy.meta_cognition_enabled', '0', 'int', 'autonomy', '元认知观测开关(0关/1开)', NOW(), NOW()),
  ('workflow.autonomy.meta_auto_tune_enabled', '0', 'int', 'autonomy', '自动校准开关-保守方向(0关/1开)', NOW(), NOW()),
  ('workflow.autonomy.learner_enabled', '0', 'int', 'autonomy', 'EMA学习开关(0关/1开)', NOW(), NOW())
ON DUPLICATE KEY UPDATE `updated_at` = NOW();
