-- 任务链路建模增强 V2：新增 task_kind / source_task_id / root_task_id

ALTER TABLE `mvp_task`
  ADD COLUMN `task_kind` varchar(32) NULL DEFAULT NULL
    COMMENT '任务记录类型：implement/audit/bug_analysis/failure_analysis'
    AFTER `role_level`,
  ADD COLUMN `source_task_id` bigint unsigned NULL DEFAULT NULL
    COMMENT '直接来源任务ID，原始任务为NULL'
    AFTER `task_kind`,
  ADD COLUMN `root_task_id` bigint unsigned NULL DEFAULT NULL
    COMMENT '所属主链根任务ID'
    AFTER `source_task_id`;

ALTER TABLE `mvp_task`
  ADD INDEX `idx_root_task` (`root_task_id`),
  ADD INDEX `idx_source_task` (`source_task_id`),
  ADD INDEX `idx_project_kind` (`project_id`, `task_kind`);
