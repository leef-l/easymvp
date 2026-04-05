-- L3.5 自治中台第三批：飞书协作平台底座
-- 新建绑定表 + 插入配置项

-- 1. mvp_user_collab_binding: 用户协作平台绑定
CREATE TABLE IF NOT EXISTS `mvp_user_collab_binding` (
  `id`                bigint unsigned NOT NULL COMMENT '雪花ID',
  `user_id`           bigint unsigned NOT NULL COMMENT '关联 system_users.id',
  `platform`          varchar(32)     NOT NULL DEFAULT 'feishu' COMMENT '平台: feishu/dingtalk/wecom',
  `platform_user_id`  varchar(128)    NOT NULL COMMENT '平台用户标识(飞书 open_id)',
  `platform_name`     varchar(128)    DEFAULT NULL COMMENT '平台显示名',
  `created_by`        bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`           bigint unsigned NOT NULL DEFAULT 0,
  `created_at`        datetime        DEFAULT NULL,
  `updated_at`        datetime        DEFAULT NULL,
  `deleted_at`        datetime        DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_platform` (`user_id`, `platform`),
  KEY `idx_platform_user` (`platform`, `platform_user_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户协作平台绑定';

-- 2. 飞书协作配置项
INSERT INTO `mvp_config` (`config_key`, `config_value`, `config_type`, `category`, `description`, `created_at`, `updated_at`)
VALUES
  ('workflow.collab.feishu_enabled', '0', 'int', 'collab', '飞书通知总开关(0关/1开)', NOW(), NOW()),
  ('workflow.collab.feishu_app_id', '', 'string', 'collab', '飞书应用 App ID', NOW(), NOW()),
  ('workflow.collab.feishu_app_secret', '', 'string', 'collab', '飞书应用 App Secret', NOW(), NOW()),
  ('workflow.collab.feishu_encrypt_key', '', 'string', 'collab', '飞书事件回调加密 Key(签名验证)', NOW(), NOW()),
  ('workflow.collab.feishu_default_notify_user_ids', '', 'string', 'collab', '降级通知的系统用户ID列表(逗号分隔)', NOW(), NOW())
ON DUPLICATE KEY UPDATE `updated_at` = NOW();
