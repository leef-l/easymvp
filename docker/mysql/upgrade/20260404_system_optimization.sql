-- =============================================
-- EasyMVP 系统优化升级 SQL
-- 日期: 2026-04-04
-- 内容: P0-1 调度器持久化 + P0-3 资源锁持久化 + P1-6 心跳机制 + P1-9 索引优化
-- =============================================

-- P0-1: 调度器状态持久化 - 项目表加 active_batch_no
ALTER TABLE `mvp_project`
ADD COLUMN `active_batch_no` INT NOT NULL DEFAULT 0 COMMENT '当前活跃批次号（调度器持久化，0=无活跃批次）' AFTER `work_dir`;

-- P0-3: 资源锁持久化 - 任务表加 locked_resources
ALTER TABLE `mvp_task`
ADD COLUMN `locked_resources` JSON DEFAULT NULL COMMENT '任务持有的资源锁（JSON数组，持久化防泄露）' AFTER `affected_resources`;

-- P1-6: 看门狗心跳机制 - 任务表加 heartbeat_at
ALTER TABLE `mvp_task`
ADD COLUMN `heartbeat_at` DATETIME DEFAULT NULL COMMENT '最近心跳时间（执行器定期更新，看门狗检测超时）' AFTER `locked_resources`;

-- P1-9: 关键索引优化
CREATE INDEX idx_mvp_task_project_status ON mvp_task(project_id, status);
CREATE INDEX idx_mvp_task_project_batch ON mvp_task(project_id, batch_no);
CREATE INDEX idx_mvp_task_conversation ON mvp_task(conversation_id);
CREATE INDEX idx_mvp_message_conversation_status ON mvp_message(conversation_id, status);
CREATE INDEX idx_mvp_conversation_project ON mvp_conversation(project_id);
CREATE INDEX idx_mvp_task_log_task ON mvp_task_log(task_id);
