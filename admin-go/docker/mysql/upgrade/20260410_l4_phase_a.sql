-- L4 Phase A：地基 — 态势快照 + 目标层 + Operator 角色
-- 新建 situation_snapshot 表 + 插入配置项 + Operator 预设

-- 1. mvp_situation_snapshot: 态势快照（Sensor 定时采集）
CREATE TABLE IF NOT EXISTS `mvp_situation_snapshot` (
  `id`              bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '关联 workflow_run',
  `project_id`      bigint unsigned NOT NULL COMMENT '关联 project',
  `snapshot_data`   json NOT NULL COMMENT '态势数据(JSON): progress/health/resource/trend',
  `anomaly_signals` json DEFAULT NULL COMMENT '异常信号列表(JSON)',
  `created_by`      bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`         bigint unsigned NOT NULL DEFAULT 0,
  `created_at`      datetime DEFAULT NULL,
  `deleted_at`      datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow_run` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='态势快照';

-- 2. mvp_project 新增目标层字段（幂等：先检查列是否存在）
SET @col_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='mvp_project' AND COLUMN_NAME='objective_json');
SET @sql = IF(@col_exists=0, 'ALTER TABLE `mvp_project` ADD COLUMN `objective_json` json DEFAULT NULL COMMENT ''项目目标约束(JSON)'' AFTER `global_context`', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 3. mvp_workflow_run 新增目标层统计字段（幂等）
SET @col_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='mvp_workflow_run' AND COLUMN_NAME='tokens_consumed');
SET @sql = IF(@col_exists=0, 'ALTER TABLE `mvp_workflow_run` ADD COLUMN `tokens_consumed` bigint NOT NULL DEFAULT 0 COMMENT ''已消耗Token总量'' AFTER `status`', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='mvp_workflow_run' AND COLUMN_NAME='replan_count');
SET @sql = IF(@col_exists=0, 'ALTER TABLE `mvp_workflow_run` ADD COLUMN `replan_count` int NOT NULL DEFAULT 0 COMMENT ''重规划次数'' AFTER `tokens_consumed`', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 4. 自治配置项
INSERT INTO `mvp_config` (`config_key`, `config_value`, `config_type`, `category`, `description`, `created_at`, `updated_at`)
VALUES
  -- Sensor
  ('workflow.autonomy.patrol_interval', '60', 'int', 'autonomy', '态势巡检间隔(秒)', NOW(), NOW()),
  ('workflow.autonomy.patrol_enabled', '0', 'int', 'autonomy', '定时巡检开关(0关/1开)', NOW(), NOW()),
  -- 目标层
  ('workflow.autonomy.objective_enabled', '0', 'int', 'autonomy', '目标层准入控制开关(0关/1开)', NOW(), NOW()),
  ('workflow.autonomy.default_token_budget', '0', 'int', 'autonomy', '默认Token预算(0=不限)', NOW(), NOW()),
  ('workflow.autonomy.default_time_budget_hours', '0', 'string', 'autonomy', '默认时间预算(小时,0=不限)', NOW(), NOW()),
  ('workflow.autonomy.default_risk_tolerance', 'balanced', 'string', 'autonomy', '默认风险容忍度(conservative/balanced/aggressive)', NOW(), NOW()),
  ('workflow.autonomy.default_autonomy_level', 'supervised', 'string', 'autonomy', '默认自治级别(manual/supervised/autonomous)', NOW(), NOW())
ON DUPLICATE KEY UPDATE `updated_at` = NOW();

-- 5. Operator 角色预设（按现有项目分类插入）
-- 使用确定性 ID：基于 category_code 的 CRC32 生成唯一值，避免 RAND() 重复风险
-- COALESCE 兜底：若该分类无 coordinator 预设，取全局任意默认模型
INSERT INTO `mvp_role_preset` (`id`, `project_category`, `role_type`, `role_level`, `model_id`, `system_prompt`, `sort`, `is_default`, `created_by`, `dept_id`, `created_at`, `updated_at`)
SELECT
  7260410000000000000 + CRC32(pc.`category_code`),
  pc.`category_code`,
  'operator',
  'pro',
  COALESCE(
    (SELECT `model_id` FROM `mvp_role_preset` WHERE `project_category` = pc.`category_code` AND `role_type` = 'coordinator' AND `is_default` = 1 AND `deleted_at` IS NULL LIMIT 1),
    (SELECT `model_id` FROM `mvp_role_preset` WHERE `project_category` = pc.`category_code` AND `is_default` = 1 AND `deleted_at` IS NULL LIMIT 1),
    0
  ),
  '你是一名运维专家（Operator），负责故障恢复、变更风险评估和环境管理。

核心职责：
1. 故障诊断：分析任务失败的根因，区分瞬态错误（网络超时、API限流）、结构性错误（方案缺陷）和致命错误（环境损坏）
2. 恢复策略：根据故障类型制定恢复方案 — 瞬态→重试、结构性→返工、致命→暂停+人工介入
3. 变更风险评估：评估代码变更对系统稳定性的影响
4. 环境健康：监控执行环境状态，确保工作目录、依赖、权限正常

绝对禁区：
- 不做业务设计或需求判断
- 不修改业务代码
- 不做质量评审

输出格式：
- 故障分析必须包含：根因分类、置信度(0-1)、推荐动作、风险等级
- 恢复方案必须包含：步骤列表、回滚方案、预期效果',
  50,
  1,
  0,
  0,
  NOW(),
  NOW()
FROM `mvp_project_category` pc
WHERE pc.`deleted_at` IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM `mvp_role_preset` rp
    WHERE rp.`project_category` = pc.`category_code`
      AND rp.`role_type` = 'operator'
      AND rp.`deleted_at` IS NULL
  );
