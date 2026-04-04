-- 流水线架构重构：数据库变更
-- 执行时间：2026-04-04

-- 1. 项目状态新增 reviewing（VARCHAR 无需 ALTER 类型，更新注释）
ALTER TABLE mvp_project MODIFY COLUMN status varchar(20) NOT NULL DEFAULT 'designing'
  COMMENT '项目状态: designing/reviewing/running/paused/completed';

-- 2. 任务状态补充 draft（更新注释）
ALTER TABLE mvp_task MODIFY COLUMN status varchar(20) NOT NULL DEFAULT 'pending'
  COMMENT '任务状态: draft/pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched/submit_error';

-- 3. 统一 collation（mvp_role_preset 从 general_ci → unicode_ci）
ALTER TABLE mvp_role_preset CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 4. 清理重复索引
ALTER TABLE mvp_task DROP INDEX idx_project;
ALTER TABLE mvp_task DROP INDEX idx_mvp_task_project_status;
ALTER TABLE mvp_task DROP INDEX idx_mvp_task_project_batch;
ALTER TABLE mvp_task DROP INDEX idx_mvp_task_conversation;
ALTER TABLE mvp_conversation DROP INDEX idx_mvp_conversation_project;
ALTER TABLE mvp_task_log DROP INDEX idx_mvp_task_log_task;
ALTER TABLE mvp_task_dependency DROP INDEX idx_task;
ALTER TABLE mvp_task_dependency DROP INDEX idx_task_depends;

-- 5. 添加缺失的联合索引
ALTER TABLE mvp_project_role ADD INDEX idx_project_role_level (project_id, role_type, role_level);
ALTER TABLE mvp_role_preset ADD INDEX idx_project_category (project_category);

-- 6. 添加审核超时配置
INSERT INTO mvp_config (config_key, config_value, config_type, category, description, created_at, updated_at)
VALUES ('review.timeout_seconds', '300', 'int', 'engine', '方案审核阶段超时时间（秒），超时跳过AI审核', NOW(), NOW());

INSERT INTO mvp_config (config_key, config_value, config_type, category, description, created_at, updated_at)
VALUES ('review.auto_fix_batch', '1', 'int', 'engine', '预检时是否自动修正batch_no不合理的问题（1=是）', NOW(), NOW());
