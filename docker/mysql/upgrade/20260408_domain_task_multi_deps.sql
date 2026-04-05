-- ============================================================
-- 为 mvp_domain_task 新增多依赖字段
-- 背景: parent_task_id 只能表达单依赖，多依赖蓝图需要完整依赖列表
-- 幂等: 使用 IF NOT EXISTS (PROCEDURE 方式)
-- ============================================================

DELIMITER //
CREATE PROCEDURE IF NOT EXISTS add_depends_on_task_ids()
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = DATABASE()
          AND table_name   = 'mvp_domain_task'
          AND column_name  = 'depends_on_task_ids'
    ) THEN
        ALTER TABLE `mvp_domain_task`
            ADD COLUMN `depends_on_task_ids` json DEFAULT NULL COMMENT '依赖任务ID列表(JSON数组)' AFTER `parent_task_id`;
    END IF;
END //
DELIMITER ;

CALL add_depends_on_task_ids();
DROP PROCEDURE IF EXISTS add_depends_on_task_ids;
