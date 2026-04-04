-- ============================================================
-- WorkflowRun 状态语义校准
-- 1. workflow_run 状态从扁平(pending/running)改为阶段化(designing/reviewing/executing/reworking)
-- 2. 新增 status_before_pause 列（暂停前状态，恢复时回退）
-- 3. 迁移已有数据
-- ============================================================

-- 新增 status_before_pause 列
SET @db = DATABASE();
SET @tbl = 'mvp_workflow_run';
SET @col = 'status_before_pause';
SET @sql_check = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=@db AND TABLE_NAME=@tbl AND COLUMN_NAME=@col) = 0,
  'ALTER TABLE `mvp_workflow_run` ADD COLUMN `status_before_pause` varchar(32) DEFAULT NULL COMMENT ''暂停前的阶段状态（恢复时回退）'' AFTER `pause_reason`',
  'SELECT 1'
));
PREPARE stmt FROM @sql_check;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 迁移已有数据：pending → designing, running → 根据 current_stage 推断
UPDATE `mvp_workflow_run`
SET `status` = 'designing'
WHERE `status` = 'pending' AND `deleted_at` IS NULL;

UPDATE `mvp_workflow_run`
SET `status` = CASE `current_stage`
    WHEN 'design' THEN 'designing'
    WHEN 'review' THEN 'reviewing'
    WHEN 'execute' THEN 'executing'
    WHEN 'rework' THEN 'reworking'
    WHEN 'complete' THEN 'completed'
    ELSE 'designing'
END
WHERE `status` = 'running' AND `deleted_at` IS NULL;

-- 更新 status 列注释
ALTER TABLE `mvp_workflow_run` MODIFY COLUMN `status` varchar(32) NOT NULL COMMENT '状态: designing/reviewing/executing/reworking/paused/completed/failed/canceled';
