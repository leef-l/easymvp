ALTER TABLE `mvp_workflow_event`
  ADD COLUMN `event_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '事件元ID' AFTER `id`,
  ADD COLUMN `attempt` int NOT NULL DEFAULT 1 COMMENT '事件尝试次数' AFTER `event_type`,
  ADD COLUMN `idempotency_key` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '幂等键' AFTER `attempt`;

UPDATE `mvp_workflow_event`
SET
  `event_id` = COALESCE(NULLIF(`event_id`, ''), CAST(`id` AS CHAR)),
  `attempt` = CASE WHEN `attempt` < 1 THEN 1 ELSE `attempt` END,
  `idempotency_key` = COALESCE(NULLIF(`idempotency_key`, ''), CONCAT('legacy:', `id`))
WHERE `event_id` IS NULL
   OR `event_id` = ''
   OR `idempotency_key` IS NULL
   OR `idempotency_key` = ''
   OR `attempt` < 1;

ALTER TABLE `mvp_workflow_event`
  MODIFY COLUMN `event_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '事件元ID',
  MODIFY COLUMN `idempotency_key` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '幂等键',
  ADD UNIQUE KEY `uk_workflow_event_event_id` (`event_id`),
  ADD UNIQUE KEY `uk_workflow_event_idempotency_key` (`idempotency_key`);

CREATE TABLE `mvp_workflow_event_ledger` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `scope` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '账本作用域: workflow.publish / workflow.recovery.*',
  `event_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '事件元ID',
  `idempotency_key` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '幂等键',
  `workflow_run_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '所属工作流运行ID',
  `event_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '事件类型',
  `attempt` int NOT NULL DEFAULT '1' COMMENT '处理尝试次数',
  `status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '处理状态: handling/handled/failed',
  `last_error` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '最后一次错误',
  `handled_at` datetime DEFAULT NULL COMMENT '处理完成时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_workflow_event_ledger_scope_key` (`scope`,`idempotency_key`),
  KEY `idx_workflow_event_ledger_event_id` (`event_id`),
  KEY `idx_workflow_event_ledger_workflow` (`workflow_run_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流事件 durable 幂等账本';
