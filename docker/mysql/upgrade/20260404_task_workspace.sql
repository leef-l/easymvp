-- ============================================================
-- M4.5: mvp_task_workspace 任务工作空间表
-- 用于 Git Worktree 任务级环境隔离
-- 幂等: 使用 IF NOT EXISTS
-- ============================================================

CREATE TABLE IF NOT EXISTS `mvp_task_workspace` (
  `id`              bigint unsigned  NOT NULL                COMMENT '雪花ID',
  `task_id`         bigint unsigned  NOT NULL                COMMENT '任务ID(domain_task或mvp_task)',
  `workflow_run_id` bigint unsigned  DEFAULT NULL            COMMENT '所属工作流运行ID',
  `project_id`      bigint unsigned  NOT NULL                COMMENT '项目ID',
  `workspace_type`  varchar(32)      NOT NULL DEFAULT 'git_worktree' COMMENT '工作空间类型: git_worktree',
  `workspace_path`  varchar(500)     NOT NULL                COMMENT '工作空间绝对路径',
  `base_ref`        varchar(255)     DEFAULT NULL            COMMENT '基线引用(commit hash/branch)',
  `status`          varchar(32)      NOT NULL DEFAULT 'creating' COMMENT '状态: creating/ready/running/completed/failed/canceled',
  `cleanup_status`  varchar(32)      NOT NULL DEFAULT 'pending'  COMMENT '清理状态: pending/done/retained/failed',
  `diff_summary`    longtext         DEFAULT NULL            COMMENT '变更摘要(diff统计)',
  `error_message`   text             DEFAULT NULL            COMMENT '错误信息',
  `created_at`      datetime         NOT NULL                COMMENT '创建时间',
  `updated_at`      datetime         NOT NULL                COMMENT '更新时间',
  `deleted_at`      datetime         DEFAULT NULL            COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task` (`task_id`),
  KEY `idx_project_status` (`project_id`, `status`),
  KEY `idx_workflow` (`workflow_run_id`),
  KEY `idx_cleanup` (`cleanup_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务工作空间(Git Worktree隔离)';
