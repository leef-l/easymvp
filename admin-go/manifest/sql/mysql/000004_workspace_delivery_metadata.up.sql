ALTER TABLE `mvp_task_workspace`
  ADD COLUMN `delivery_mode` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'patch' COMMENT '交付结果形态: patch/pr/manual' AFTER `cleanup_status`,
  ADD COLUMN `delivery_status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '交付状态: pending/ready/skipped/failed' AFTER `delivery_mode`,
  ADD COLUMN `sync_strategy` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'auto_apply' COMMENT '回写策略: auto_apply/manual' AFTER `delivery_status`,
  ADD COLUMN `sync_status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '回写状态: pending/applied/skipped/failed' AFTER `sync_strategy`,
  ADD COLUMN `risk_level` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'medium' COMMENT '风险等级: low/medium/high' AFTER `sync_status`,
  ADD COLUMN `patch_ref` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'patch产物路径' AFTER `risk_level`;
