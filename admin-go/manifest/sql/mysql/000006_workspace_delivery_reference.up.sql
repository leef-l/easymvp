ALTER TABLE `mvp_task_workspace`
  ADD COLUMN `delivery_ref` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '交付引用(PR草稿文件/外部链接)' AFTER `patch_ref`,
  ADD COLUMN `delivery_title` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '交付标题' AFTER `delivery_ref`;
