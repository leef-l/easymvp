CREATE TABLE IF NOT EXISTS `mvp_workflow_transition_log` (
  `id` BIGINT UNSIGNED NOT NULL COMMENT '雪花ID',
  `workflow_run_id` BIGINT UNSIGNED NOT NULL COMMENT '所属工作流运行ID',
  `scope` VARCHAR(32) NOT NULL DEFAULT 'workflow' COMMENT '转移范围: workflow/stage',
  `from_state` VARCHAR(32) NOT NULL COMMENT '源状态',
  `to_state` VARCHAR(32) NOT NULL COMMENT '目标状态',
  `guard_result` JSON DEFAULT NULL COMMENT '守卫求值结果',
  `duration_ms` BIGINT DEFAULT NULL COMMENT '状态停留时长(ms)',
  `actor` VARCHAR(64) NOT NULL DEFAULT 'system' COMMENT '操作者: user_id或system',
  `reason` VARCHAR(500) DEFAULT NULL COMMENT '转移原因',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `dept_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '部门ID',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_time` (`workflow_run_id`, `created_at`),
  KEY `idx_datascope` (`dept_id`, `created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流状态转移日志';
