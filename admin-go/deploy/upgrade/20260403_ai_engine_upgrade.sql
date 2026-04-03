ALTER TABLE `system_role`
  ADD COLUMN IF NOT EXISTS `default_ai_engine` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '默认AI执行引擎编码' AFTER `data_scope`;

CREATE TABLE IF NOT EXISTS `system_role_ai_engine` (
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `engine_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行引擎编码',
  PRIMARY KEY (`role_id`,`engine_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色可用AI执行引擎';

CREATE TABLE IF NOT EXISTS `ai_engine` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '引擎编码',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '引擎名称',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '引擎描述',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI执行引擎表';

CREATE TABLE IF NOT EXISTS `ai_engine_config` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `engine_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '引擎编码',
  `base_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '服务地址',
  `api_key` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'API Key',
  `default_model_id` bigint unsigned DEFAULT NULL COMMENT '默认模型ID',
  `timeout_seconds` int NOT NULL DEFAULT '600' COMMENT '超时秒数',
  `max_steps` int NOT NULL DEFAULT '20' COMMENT '最大步骤数',
  `workspace_root` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '工作区根目录',
  `command_template` text COLLATE utf8mb4_unicode_ci COMMENT '命令模板',
  `callback_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '回调地址',
  `callback_secret` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '回调密钥',
  `extra_config` text COLLATE utf8mb4_unicode_ci COMMENT '额外配置JSON',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_engine_code` (`engine_code`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI执行引擎配置表';

CREATE TABLE IF NOT EXISTS `ai_task` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `title` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务标题',
  `engine_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行引擎编码',
  `project_id` bigint unsigned DEFAULT NULL COMMENT '项目ID',
  `role_id` bigint unsigned DEFAULT NULL COMMENT '角色ID',
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID',
  `repo_path` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '仓库路径',
  `worktree_path` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '工作目录',
  `branch_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分支名',
  `instruction` text COLLATE utf8mb4_unicode_ci COMMENT '执行指令',
  `request_payload` longtext COLLATE utf8mb4_unicode_ci COMMENT '请求快照',
  `engine_config_snapshot` longtext COLLATE utf8mb4_unicode_ci COMMENT '引擎配置快照',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '状态',
  `response_summary` longtext COLLATE utf8mb4_unicode_ci COMMENT '执行摘要',
  `error_message` longtext COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `finished_at` datetime DEFAULT NULL COMMENT '结束时间',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_engine_code` (`engine_code`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI执行任务表';

CREATE TABLE IF NOT EXISTS `ai_task_log` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `task_id` bigint unsigned NOT NULL COMMENT '任务ID',
  `seq` int NOT NULL DEFAULT '1' COMMENT '日志序号',
  `log_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '日志类型',
  `content` longtext COLLATE utf8mb4_unicode_ci COMMENT '日志内容',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_seq` (`task_id`,`seq`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI执行任务日志表';

INSERT INTO `system_menu` (`id`,`parent_id`,`title`,`type`,`path`,`component`,`permission`,`icon`,`sort`,`is_show`,`is_cache`,`link_url`,`status`,`created_by`,`dept_id`,`created_at`,`updated_at`,`deleted_at`)
SELECT 315020000000000001,315012657751003136,'执行引擎配置',2,'/ai/engine','ai/engine/index','ai:engine:list','',1,1,0,NULL,1,0,0,'2026-04-03 11:02:29','2026-04-03 11:02:29',NULL
WHERE NOT EXISTS (SELECT 1 FROM `system_menu` WHERE `id` = 315020000000000001);

INSERT INTO `system_menu` (`id`,`parent_id`,`title`,`type`,`path`,`component`,`permission`,`icon`,`sort`,`is_show`,`is_cache`,`link_url`,`status`,`created_by`,`dept_id`,`created_at`,`updated_at`,`deleted_at`)
SELECT 315020000000000011,315012657751003136,'执行任务',2,'/ai/task','ai/task/index','ai:task:list','',2,1,0,NULL,1,0,0,'2026-04-03 11:02:30','2026-04-03 11:02:30',NULL
WHERE NOT EXISTS (SELECT 1 FROM `system_menu` WHERE `id` = 315020000000000011);
