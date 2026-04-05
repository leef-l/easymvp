-- Accept Stage 核心表（第一批：可闭环底座）
-- 对应设计文档：docs/AcceptStage自动验收阶段设计文档.md

-- 1. 自动验收运行记录
CREATE TABLE IF NOT EXISTS `mvp_accept_run` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `workflow_run_id` bigint NOT NULL COMMENT '工作流运行ID',
  `stage_run_id` bigint NOT NULL COMMENT 'accept阶段stage_run_id',
  `project_id` bigint NOT NULL COMMENT '项目ID',
  `plan_version_id` bigint DEFAULT NULL COMMENT '关联方案版本ID',
  `accept_round` int NOT NULL DEFAULT '1' COMMENT '第几轮验收',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/running/completed/failed/canceled',
  `decision` varchar(20) DEFAULT NULL COMMENT 'passed/failed/manual_review',
  `score` decimal(5,2) DEFAULT NULL COMMENT '验收评分',
  `summary` text COMMENT '验收摘要',
  `rules_version` varchar(64) DEFAULT NULL COMMENT '规则版本号',
  `rules_snapshot_ref` longtext COMMENT '规则快照引用或JSON',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `finished_at` datetime DEFAULT NULL COMMENT '结束时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_run_id` (`workflow_run_id`),
  KEY `idx_stage_run_id` (`stage_run_id`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_workflow_round` (`workflow_run_id`,`accept_round`),
  KEY `idx_status` (`status`),
  KEY `idx_decision` (`decision`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动验收运行记录';

-- 2. 自动验收问题
CREATE TABLE IF NOT EXISTS `mvp_accept_issue` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `accept_run_id` bigint NOT NULL COMMENT '验收运行ID',
  `workflow_run_id` bigint NOT NULL COMMENT '工作流运行ID',
  `project_id` bigint NOT NULL COMMENT '项目ID',
  `domain_task_id` bigint DEFAULT NULL COMMENT '主关联任务ID',
  `issue_type` varchar(32) NOT NULL COMMENT 'artifact/process/quality/risk',
  `rule_code` varchar(64) DEFAULT NULL COMMENT '规则编码',
  `severity` varchar(16) NOT NULL COMMENT 'info/warn/error/blocker',
  `title` varchar(255) NOT NULL COMMENT '问题标题',
  `detail` text COMMENT '问题详情',
  `expected_value` text COMMENT '预期值',
  `actual_value` text COMMENT '实际值',
  `suggested_action` text COMMENT '建议动作',
  `resource_ref` text COMMENT '关联资源引用(JSON)',
  `status` varchar(20) NOT NULL DEFAULT 'open' COMMENT 'open/resolved/ignored',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_accept_run_id` (`accept_run_id`),
  KEY `idx_workflow_run_id` (`workflow_run_id`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_domain_task_id` (`domain_task_id`),
  KEY `idx_rule_code` (`rule_code`),
  KEY `idx_severity` (`severity`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动验收问题';

-- 3. 自动验收规则定义
CREATE TABLE IF NOT EXISTS `mvp_accept_rule` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `project_type` varchar(64) NOT NULL COMMENT '项目类型模板',
  `rule_code` varchar(64) NOT NULL COMMENT '规则编码',
  `rule_name` varchar(255) NOT NULL COMMENT '规则名称',
  `rule_type` varchar(32) NOT NULL COMMENT 'artifact/process/quality',
  `scope_type` varchar(32) NOT NULL COMMENT 'project/task/file/stage',
  `config_json` longtext NOT NULL COMMENT '规则配置',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `priority` int NOT NULL DEFAULT '100' COMMENT '优先级(越小越先执行)',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_type_rule_code` (`project_type`,`rule_code`),
  KEY `idx_rule_type` (`rule_type`),
  KEY `idx_scope_type` (`scope_type`),
  KEY `idx_enabled_priority` (`enabled`,`priority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动验收规则定义';

-- 4. 自动验收证据
CREATE TABLE IF NOT EXISTS `mvp_accept_evidence` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `accept_run_id` bigint NOT NULL COMMENT '验收运行ID',
  `evidence_type` varchar(32) NOT NULL COMMENT 'task_output/file/log/diff/stage_output/handoff/summary',
  `source_type` varchar(32) NOT NULL COMMENT 'domain_task/stage_run/file/handoff_record/workflow_run',
  `source_id` bigint DEFAULT NULL COMMENT '来源对象ID',
  `content_ref` longtext COMMENT '证据引用或JSON',
  `summary` text COMMENT '证据摘要',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_accept_run_id` (`accept_run_id`),
  KEY `idx_evidence_type` (`evidence_type`),
  KEY `idx_source_type_source_id` (`source_type`,`source_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动验收证据';
