SET NAMES utf8mb4;

-- MVP 配置表（引擎参数可配置化）
CREATE TABLE IF NOT EXISTS `mvp_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `config_key` varchar(100) NOT NULL COMMENT '配置键（唯一）',
  `config_value` text NOT NULL COMMENT '配置值',
  `config_type` varchar(20) NOT NULL DEFAULT 'string' COMMENT '值类型：string/int/float/bool/json',
  `category` varchar(50) NOT NULL DEFAULT 'general' COMMENT '分类：engine/watchdog/scheduler/general',
  `description` varchar(255) DEFAULT NULL COMMENT '配置说明',
  `created_by` bigint unsigned DEFAULT 0 COMMENT '创建人',
  `dept_id` bigint unsigned DEFAULT 0 COMMENT '部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`),
  KEY `idx_category` (`category`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MVP配置表';

INSERT IGNORE INTO `mvp_config` (`config_key`, `config_value`, `config_type`, `category`, `description`, `created_by`, `dept_id`, `created_at`, `updated_at`) VALUES
('runtime.task_timeout_seconds', '600', 'int', 'runtime', CONVERT(0x4D565020E9A1B9E79BAEE4BBBBE58AA1E689A7E8A18CE8B685E697B6EFBC88E7A792EFBC89 USING utf8mb4), 0, 0, NOW(), NOW()),
('runtime.max_steps', '2', 'int', 'runtime', CONVERT(0xE5A49AE6ADA5E4BBA3E79086E69C80E5A4A7E8BFADE4BBA3E6ACA1E695B0EFBC88E5BD93E5898D20416964657220E4BB85E68EA7E588B6E698AFE590A6E58581E8AEB8E7B2BEE7AE80E9878DE8AF95EFBC89 USING utf8mb4), 0, 0, NOW(), NOW()),
('watchdog.check_interval', '120', 'int', 'watchdog', CONVERT(0xE5BF83E8B7B3E6A380E6B58BE997B4E99A94EFBC88E7A792EFBC89 USING utf8mb4), 0, 0, NOW(), NOW()),
('watchdog.max_stale_count', '3', 'int', 'watchdog', CONVERT(0xE8BF9EE7BBADE697A0E8BF9BE5B195E6ACA1E695B0E99888E580BCEFBC88E8B685E8BF87E58899E588A4E5AE9AE58DA1E6ADBBEFBC89 USING utf8mb4), 0, 0, NOW(), NOW()),
('watchdog.max_retries', '3', 'int', 'watchdog', CONVERT(0xE69C80E5A4A7E887AAE58AA8E9878DE8AF95E6ACA1E695B0EFBC88E8B685E8BF87E58899E58D87E7BAA7E7BB99E69EB6E69E84E5B888EFBC89 USING utf8mb4), 0, 0, NOW(), NOW()),
('scheduler.max_concurrent', '20', 'int', 'scheduler', CONVERT(0xE69C80E5A4A7E5B9B6E58F91E4BBBBE58AA1E695B0 USING utf8mb4), 0, 0, NOW(), NOW()),
('scheduler.poll_interval', '2', 'int', 'scheduler', CONVERT(0xE8B083E5BAA6E8BDAEE8AFA2E997B4E99A94EFBC88E7A792EFBC89 USING utf8mb4), 0, 0, NOW(), NOW());
