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

-- 初始配置数据
INSERT IGNORE INTO `mvp_config` (`config_key`, `config_value`, `config_type`, `category`, `description`, `created_at`, `updated_at`) VALUES
('watchdog.check_interval', '120', 'int', 'watchdog', '心跳检测间隔（秒）', NOW(), NOW()),
('watchdog.max_stale_count', '3', 'int', 'watchdog', '连续无进展次数阈值（超过则判定卡死）', NOW(), NOW()),
('watchdog.max_retries', '3', 'int', 'watchdog', '最大自动重试次数（超过则升级给架构师）', NOW(), NOW()),
('scheduler.max_concurrent', '20', 'int', 'scheduler', '最大并发任务数', NOW(), NOW()),
('scheduler.poll_interval', '2', 'int', 'scheduler', '调度轮询间隔（秒）', NOW(), NOW());
