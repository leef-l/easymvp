-- L4 Phase B：策略函数 — Planner + 6 策略 + Actuator 效果跟踪
-- 新增策略独立灰度开关 + action_outcome 效果跟踪表

-- 1. 策略函数灰度配置项
INSERT INTO `mvp_config` (`config_key`, `config_value`, `config_type`, `category`, `description`, `created_at`, `updated_at`)
VALUES
  -- 总开关
  ('workflow.autonomy.strategy_enabled', '0', 'int', 'autonomy', '策略函数总开关(0关/1开)', NOW(), NOW()),
  -- 6 个策略独立开关（默认开启，由总开关控制是否生效）
  ('workflow.autonomy.adaptive_retry_enabled', '1', 'int', 'autonomy', '自适应重试策略(0关/1开)', NOW(), NOW()),
  ('workflow.autonomy.engine_selection_enabled', '1', 'int', 'autonomy', '执行器选择策略(0关/1开)', NOW(), NOW()),
  ('workflow.autonomy.batch_adjust_enabled', '1', 'int', 'autonomy', '批次调整策略(0关/1开)', NOW(), NOW()),
  ('workflow.autonomy.cost_guard_enabled', '1', 'int', 'autonomy', '成本守卫策略(0关/1开)', NOW(), NOW()),
  ('workflow.autonomy.proactive_replan_enabled', '1', 'int', 'autonomy', '主动重规划策略(0关/1开)', NOW(), NOW()),
  ('workflow.autonomy.quality_gate_enabled', '1', 'int', 'autonomy', '质量门策略(0关/1开)', NOW(), NOW())
ON DUPLICATE KEY UPDATE `updated_at` = NOW();

-- 2. mvp_action_outcome: 策略效果跟踪表（Actuator 使用）
CREATE TABLE IF NOT EXISTS `mvp_action_outcome` (
  `id`              bigint unsigned NOT NULL COMMENT '雪花ID',
  `action_id`       bigint unsigned NOT NULL COMMENT '关联 mvp_decision_action.id',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '关联 workflow_run',
  `project_id`      bigint unsigned NOT NULL COMMENT '关联 project',
  `strategy_name`   varchar(64) NOT NULL DEFAULT '' COMMENT '策略名称',
  `action_type`     varchar(64) NOT NULL DEFAULT '' COMMENT '动作类型',
  `decision_level`  varchar(8) NOT NULL DEFAULT '' COMMENT '决策级别(A/B/C)',
  `sit_before`      json DEFAULT NULL COMMENT '执行前态势摘要(JSON)',
  `sit_after`       json DEFAULT NULL COMMENT '执行后态势摘要(JSON)',
  `effective`       varchar(16) NOT NULL DEFAULT 'unknown' COMMENT '效果: positive/negative/neutral/unknown',
  `effect_score`    decimal(5,3) NOT NULL DEFAULT 0 COMMENT '效果得分(-1~1)',
  `eval_delay_ms`   bigint NOT NULL DEFAULT 0 COMMENT '评估延迟(毫秒)',
  `created_by`      bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`         bigint unsigned NOT NULL DEFAULT 0,
  `created_at`      datetime DEFAULT NULL,
  `deleted_at`      datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_action` (`action_id`),
  KEY `idx_workflow_run` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_strategy` (`strategy_name`),
  KEY `idx_effective` (`effective`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='策略效果跟踪';
