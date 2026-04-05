DROP TABLE IF EXISTS `ai_engine`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_engine` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID',
  `code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '引擎编码: aider/openhands',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '引擎名称',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '说明',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:1启用 0禁用',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `created_by` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建人',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ai_engine_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI执行引擎定义表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `ai_engine_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_engine_config` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID',
  `engine_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '引擎编码',
  `base_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '服务地址',
  `api_key` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'API Key',
  `default_model_id` bigint unsigned DEFAULT NULL COMMENT '默认模型ID',
  `timeout_seconds` int NOT NULL DEFAULT '600' COMMENT '超时时间(秒)',
  `max_steps` int NOT NULL DEFAULT '20' COMMENT '最大执行步数',
  `workspace_root` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '工作区根目录',
  `command_template` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '命令模板',
  `callback_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '回调地址',
  `callback_secret` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '回调密钥',
  `extra_config` json DEFAULT NULL COMMENT '额外配置JSON',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:1启用 0禁用',
  `created_by` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建人',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ai_engine_config_engine_code` (`engine_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI执行引擎配置表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `ai_model`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_model` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `plan_id` bigint unsigned NOT NULL COMMENT '套餐ID',
  `provider_id` bigint unsigned NOT NULL COMMENT '供应商ID（冗余便于查询）',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模型显示名称',
  `model_code` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模型代码（API调用用）',
  `capability` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'chat' COMMENT '能力：chat/reasoning/coding',
  `max_tokens` int NOT NULL DEFAULT '4096' COMMENT '最大输出token',
  `context_window` int NOT NULL DEFAULT '128000' COMMENT '上下文窗口大小',
  `supports_stream` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否支持流式输出:0=否,1=是',
  `role_prompt` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '默认角色提示词',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:0=禁用,1=启用',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_plan` (`plan_id`),
  KEY `idx_provider` (`provider_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI模型表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `ai_plan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_plan` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `provider_id` bigint unsigned NOT NULL COMMENT '供应商ID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '套餐名称',
  `code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '套餐代码',
  `api_key` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'API Key（加密存储）',
  `api_secret` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'API Secret（部分供应商需要）',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:0=禁用,1=启用',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_provider` (`provider_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI套餐表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `ai_provider`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_provider` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '供应商名称',
  `code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '供应商代码：openai/anthropic/deepseek/qwen/doubao/ernie/spark/glm/moonshot/yi/google/ollama',
  `provider_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Provider类型：openai_compatible/anthropic/baidu/xfyun/google',
  `base_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'API基础地址',
  `icon` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '图标URL',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:0=禁用,1=启用',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI供应商表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `ai_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_task` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务标题',
  `engine_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行引擎',
  `role_id` bigint unsigned DEFAULT NULL COMMENT '发起时角色ID',
  `user_id` bigint unsigned NOT NULL COMMENT '发起用户ID',
  `project_id` bigint unsigned DEFAULT NULL COMMENT '项目ID',
  `repo_path` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '仓库路径',
  `worktree_path` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '执行工作目录',
  `branch_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分支名称',
  `instruction` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户指令',
  `engine_config_snapshot` json DEFAULT NULL COMMENT '执行时配置快照',
  `request_payload` json DEFAULT NULL COMMENT '请求体',
  `response_summary` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '执行结果摘要',
  `error_message` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT 'pending/running/success/failed/cancelled',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `finished_at` datetime DEFAULT NULL COMMENT '结束时间',
  `created_by` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建人',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_ai_task_user_id` (`user_id`),
  KEY `idx_ai_task_engine_code` (`engine_code`),
  KEY `idx_ai_task_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI执行任务表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `ai_task_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_task_log` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID',
  `task_id` bigint unsigned NOT NULL COMMENT '任务ID',
  `seq` int NOT NULL DEFAULT '0' COMMENT '日志序号',
  `log_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'stdout' COMMENT 'stdout/stderr/system/event',
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '日志内容',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_ai_task_log_task_id` (`task_id`),
  KEY `idx_ai_task_log_task_seq` (`task_id`,`seq`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI执行任务日志表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_accept_evidence`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_accept_evidence` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `accept_run_id` bigint NOT NULL COMMENT '验收运行ID',
  `evidence_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'task_output/file/log/diff/stage_output/handoff/summary',
  `source_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'domain_task/stage_run/file/handoff_record/workflow_run',
  `source_id` bigint DEFAULT NULL COMMENT '来源对象ID',
  `content_ref` longtext COLLATE utf8mb4_unicode_ci COMMENT '证据引用或JSON',
  `summary` text COLLATE utf8mb4_unicode_ci COMMENT '证据摘要',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_accept_run_id` (`accept_run_id`),
  KEY `idx_evidence_type` (`evidence_type`),
  KEY `idx_source_type_source_id` (`source_type`,`source_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动验收证据';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_accept_issue`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_accept_issue` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `accept_run_id` bigint NOT NULL COMMENT '验收运行ID',
  `workflow_run_id` bigint NOT NULL COMMENT '工作流运行ID',
  `project_id` bigint NOT NULL COMMENT '项目ID',
  `domain_task_id` bigint DEFAULT NULL COMMENT '主关联任务ID',
  `issue_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'artifact/process/quality/risk',
  `rule_code` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '规则编码',
  `severity` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'info/warn/error/blocker',
  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '问题标题',
  `detail` text COLLATE utf8mb4_unicode_ci COMMENT '问题详情',
  `expected_value` text COLLATE utf8mb4_unicode_ci COMMENT '预期值',
  `actual_value` text COLLATE utf8mb4_unicode_ci COMMENT '实际值',
  `suggested_action` text COLLATE utf8mb4_unicode_ci COMMENT '建议动作',
  `resource_ref` text COLLATE utf8mb4_unicode_ci COMMENT '关联资源引用(JSON)',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'open' COMMENT 'open/resolved/ignored',
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
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_accept_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_accept_rule` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `project_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '项目类型模板',
  `rule_code` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则编码',
  `rule_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则名称',
  `rule_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'artifact/process/quality',
  `scope_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'project/task/file/stage',
  `config_json` longtext COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则配置',
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
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_accept_run`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_accept_run` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `workflow_run_id` bigint NOT NULL COMMENT '工作流运行ID',
  `stage_run_id` bigint NOT NULL COMMENT 'accept阶段stage_run_id',
  `project_id` bigint NOT NULL COMMENT '项目ID',
  `plan_version_id` bigint DEFAULT NULL COMMENT '关联方案版本ID',
  `accept_round` int NOT NULL DEFAULT '1' COMMENT '第几轮验收',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT 'pending/running/completed/failed/canceled',
  `decision` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'passed/failed/manual_review',
  `score` decimal(5,2) DEFAULT NULL COMMENT '验收评分',
  `summary` text COLLATE utf8mb4_unicode_ci COMMENT '验收摘要',
  `rules_version` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '规则版本号',
  `rules_snapshot_ref` longtext COLLATE utf8mb4_unicode_ci COMMENT '规则快照引用或JSON',
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
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_autonomy_decision`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_autonomy_decision` (
  `id` bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned NOT NULL,
  `decision_type` varchar(32) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'replan/risk_escalate/engine_switch/circuit_break/report',
  `trigger_source` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '触发源：watchdog/accept/rework/scheduler/manual',
  `trigger_context` json DEFAULT NULL COMMENT '触发上下文',
  `recommendation` json NOT NULL COMMENT '系统建议',
  `decision_mode` varchar(16) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'suggest' COMMENT 'suggest/auto',
  `human_action` varchar(16) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT 'approved/rejected/modified/pending',
  `executed_at` datetime DEFAULT NULL COMMENT '实际执行时间',
  `result` json DEFAULT NULL COMMENT '执行结果',
  `created_by` bigint unsigned DEFAULT '0',
  `dept_id` bigint unsigned DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_type` (`decision_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='自治决策记录';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `config_key` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置键（唯一）',
  `config_value` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置值',
  `config_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'string' COMMENT '值类型：string/int/float/bool/json',
  `category` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'general' COMMENT '分类：engine/watchdog/scheduler/general',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '配置说明',
  `created_by` bigint unsigned DEFAULT '0' COMMENT '创建人',
  `dept_id` bigint unsigned DEFAULT '0' COMMENT '部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`),
  KEY `idx_category` (`category`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=30 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MVP配置表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_conversation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_conversation` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `project_id` bigint unsigned NOT NULL COMMENT '项目ID',
  `task_id` bigint unsigned DEFAULT NULL COMMENT '关联任务ID，NULL=项目级对话',
  `title` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '对话标题',
  `role_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '对话角色类型',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active' COMMENT '状态：active/archived',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_task` (`task_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MVP对话表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_decision_action`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_decision_action` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned NOT NULL,
  `stage_run_id` bigint unsigned DEFAULT NULL COMMENT '关联阶段运行ID',
  `domain_task_id` bigint unsigned DEFAULT NULL COMMENT '关联领域任务ID',
  `decision_type` varchar(32) COLLATE utf8mb4_general_ci NOT NULL COMMENT '动作类型',
  `decision_level` char(1) COLLATE utf8mb4_general_ci NOT NULL COMMENT '决策等级: A/B/C',
  `trigger_source` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '触发源事件类型',
  `trigger_context` json DEFAULT NULL COMMENT '触发上下文',
  `matched_rule_id` bigint unsigned DEFAULT NULL COMMENT '匹配的策略规则ID',
  `matched_gate_ids` json DEFAULT NULL COMMENT '命中的闸门ID列表',
  `action_type` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '策略匹配的动作类型(闸门降级后为fallback)',
  `recommendation` json DEFAULT NULL COMMENT '系统建议',
  `final_action` varchar(64) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '最终实际执行的动作',
  `action_status` varchar(16) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'pending' COMMENT 'pending/auto_executed/waiting_human/approved/rejected/failed',
  `auto_executable` tinyint NOT NULL DEFAULT '0' COMMENT '是否可自动执行',
  `human_required` tinyint NOT NULL DEFAULT '0' COMMENT '是否需要人工',
  `executed_at` datetime DEFAULT NULL COMMENT '实际执行时间',
  `result` json DEFAULT NULL COMMENT '执行结果',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_status` (`action_status`),
  KEY `idx_type` (`decision_type`),
  KEY `idx_trigger` (`trigger_source`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='自治决策动作记录';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_domain_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_domain_task` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `stage_run_id` bigint unsigned NOT NULL COMMENT '所属阶段运行ID',
  `plan_version_id` bigint unsigned DEFAULT NULL COMMENT '来源计划版本ID',
  `blueprint_id` bigint unsigned DEFAULT NULL COMMENT '来源蓝图ID',
  `parent_task_id` bigint unsigned DEFAULT NULL COMMENT '父任务ID',
  `depends_on_task_ids` json DEFAULT NULL COMMENT '依赖任务ID列表(JSON数组)',
  `source_task_id` bigint unsigned DEFAULT NULL COMMENT '来源任务ID(链路追踪)',
  `root_task_id` bigint unsigned DEFAULT NULL COMMENT '根任务ID(链路追踪)',
  `task_kind` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务种类: implement/audit/bug_analysis/failure_analysis',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务名称',
  `description` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务描述',
  `role_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色类型',
  `role_level` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色等级',
  `execution_mode` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行方式: chat/aider/openhands',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态: pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched',
  `conversation_id` bigint unsigned DEFAULT NULL COMMENT '关联对话ID',
  `model_id` bigint unsigned DEFAULT NULL COMMENT '使用的AI模型ID',
  `batch_no` int NOT NULL COMMENT '批次号',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `retry_count` int NOT NULL DEFAULT '0' COMMENT '重试次数',
  `affected_resources` json DEFAULT NULL COMMENT '影响资源列表(JSON)',
  `locked_resources` json DEFAULT NULL COMMENT '锁定资源列表(JSON)',
  `result` longtext COLLATE utf8mb4_unicode_ci COMMENT '执行结果',
  `context_summary` text COLLATE utf8mb4_unicode_ci COMMENT '上下文摘要',
  `heartbeat_at` datetime DEFAULT NULL COMMENT '心跳时间',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_status` (`workflow_run_id`,`status`),
  KEY `idx_workflow_batch` (`workflow_run_id`,`batch_no`,`sort`),
  KEY `idx_stage_status` (`stage_run_id`,`status`),
  KEY `idx_root_task` (`root_task_id`),
  KEY `idx_source_task` (`source_task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='领域任务';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_handoff_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_handoff_record` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `from_task_id` bigint unsigned DEFAULT NULL COMMENT '来源任务ID',
  `to_task_id` bigint unsigned DEFAULT NULL COMMENT '目标任务ID',
  `handoff_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '交接类型: bug_fix/failure_escalation/rework/audit',
  `reason` text COLLATE utf8mb4_unicode_ci COMMENT '交接原因',
  `payload` json DEFAULT NULL COMMENT '交接载荷(JSON)',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_type` (`workflow_run_id`,`handoff_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交接记录';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_human_checkpoint`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_human_checkpoint` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned NOT NULL,
  `decision_action_id` bigint unsigned NOT NULL COMMENT '关联的决策动作ID',
  `checkpoint_type` varchar(32) COLLATE utf8mb4_general_ci NOT NULL COMMENT '节点类型: manual_review/approval/escalation',
  `title` varchar(256) COLLATE utf8mb4_general_ci NOT NULL COMMENT '标题',
  `description` text COLLATE utf8mb4_general_ci COMMENT '详细描述',
  `status` varchar(16) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'open' COMMENT 'open/handled/expired/canceled',
  `assigned_to` bigint unsigned DEFAULT NULL COMMENT '指派给谁',
  `handled_by` bigint unsigned DEFAULT NULL COMMENT '实际处理人',
  `handle_action` varchar(32) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '处理动作: approve/reject/retry/rework/override',
  `handle_reason` text COLLATE utf8mb4_general_ci COMMENT '处理理由',
  `handled_at` datetime DEFAULT NULL COMMENT '处理时间',
  `expires_at` datetime DEFAULT NULL COMMENT '过期时间',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_decision_action` (`decision_action_id`),
  KEY `idx_status` (`status`),
  KEY `idx_assigned` (`assigned_to`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='人工介入节点';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_message` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `conversation_id` bigint unsigned NOT NULL COMMENT '对话ID',
  `role` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '消息角色：user/assistant/system',
  `message_type` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'general' COMMENT '消息类型：chat_user/chat_reply/task_prompt/task_reply/system_notice/poison/general',
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '消息内容',
  `model_id` bigint unsigned DEFAULT NULL COMMENT '使用的AI模型ID',
  `token_usage` json DEFAULT NULL COMMENT 'token消耗：{prompt_tokens, completion_tokens}',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'completed' COMMENT '状态：pending/streaming/completed/failed',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_conversation` (`conversation_id`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_message_type` (`message_type`),
  KEY `idx_conversation_status_created` (`conversation_id`,`status`,`created_at`),
  KEY `idx_mvp_message_conversation_status` (`conversation_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MVP消息表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_message_chunk`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_message_chunk` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `message_id` bigint unsigned NOT NULL COMMENT '消息ID',
  `chunk_index` int NOT NULL COMMENT '分片序号',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分片内容',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间（精确到毫秒）',
  PRIMARY KEY (`id`),
  KEY `idx_message_chunk` (`message_id`,`chunk_index`),
  KEY `idx_chunk_created` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=22630 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息分片表（流式输出）';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_plan_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_plan_version` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `project_id` bigint unsigned NOT NULL COMMENT '所属项目ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `version_no` int NOT NULL COMMENT '版本号(项目内递增)',
  `source_conversation_id` bigint unsigned DEFAULT NULL COMMENT '来源对话ID',
  `source_message_id` bigint unsigned DEFAULT NULL COMMENT '来源消息ID',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '版本状态: draft/active/superseded',
  `review_status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '审核状态: pending/approved/rejected',
  `summary` text COLLATE utf8mb4_unicode_ci COMMENT '版本摘要',
  `diff_summary` text COLLATE utf8mb4_unicode_ci COMMENT '与上一版本的差异摘要',
  `approved_at` datetime DEFAULT NULL COMMENT '审核通过时间',
  `rejected_at` datetime DEFAULT NULL COMMENT '审核驳回时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_version` (`project_id`,`version_no`),
  KEY `idx_workflow_status` (`workflow_run_id`,`status`,`review_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计划版本';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_policy_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_policy_rule` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `rule_code` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '规则编码(唯一)',
  `rule_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '规则名称',
  `decision_type` varchar(32) COLLATE utf8mb4_general_ci NOT NULL COMMENT '决策动作类型',
  `decision_level` char(1) COLLATE utf8mb4_general_ci NOT NULL COMMENT '决策等级: A/B/C',
  `trigger_source` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '触发源事件类型',
  `project_family` varchar(32) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '适用项目家族(NULL=全局)',
  `project_category_code` varchar(64) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '适用项目分类(NULL=全局)',
  `config_json` json NOT NULL COMMENT '规则配置(条件/阈值)',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否启用',
  `priority` int NOT NULL DEFAULT '100' COMMENT '优先级(越小越优先)',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_rule_code` (`rule_code`),
  KEY `idx_trigger` (`trigger_source`,`enabled`),
  KEY `idx_level` (`decision_level`),
  KEY `idx_family_cat` (`project_family`,`project_category_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='自治策略规则';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_project`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_project` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `name` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '项目名称',
  `project_category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '项目分类',
  `category_code` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '项目分类编码',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '项目简介',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'designing' COMMENT '项目状态: designing/reviewing/running/paused/completed',
  `pause_reason` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '暂停原因',
  `global_context` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '项目全局上下文（架构师需求分析+方案设计的压缩摘要）',
  `architect_model_id` bigint unsigned DEFAULT NULL COMMENT '架构师使用的AI模型ID',
  `work_dir` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '项目代码工作目录（Aider执行路径）',
  `active_batch_no` int NOT NULL DEFAULT '0' COMMENT '当前活跃批次号（调度器持久化，0=无活跃批次）',
  `engine_version` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT 'legacy' COMMENT '执行引擎版本: legacy/workflow_v2',
  `active_workflow_run_id` bigint unsigned DEFAULT NULL COMMENT '当前活跃工作流运行ID',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_category_code` (`category_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MVP项目表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_project_category`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_project_category` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `category_code` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '稳定分类编码',
  `display_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '展示名称',
  `family_code` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '能力家族编码',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分类说明',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '1启用 0停用',
  `sort` int NOT NULL DEFAULT '100' COMMENT '排序',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_code` (`category_code`),
  KEY `idx_family_code` (`family_code`),
  KEY `idx_status_sort` (`status`,`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目分类配置表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_project_report`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_project_report` (
  `id` bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned NOT NULL,
  `report_type` varchar(16) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'stage/daily/weekly/summary',
  `stage_type` varchar(32) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '阶段类型',
  `title` varchar(200) COLLATE utf8mb4_general_ci NOT NULL,
  `content` text COLLATE utf8mb4_general_ci NOT NULL COMMENT 'Markdown 格式报告正文',
  `metrics` json DEFAULT NULL COMMENT '关键指标快照',
  `created_by` bigint unsigned DEFAULT '0',
  `dept_id` bigint unsigned DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_type` (`report_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='项目汇报';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_project_role`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_project_role` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `project_id` bigint unsigned NOT NULL COMMENT '项目ID',
  `project_category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '项目分类',
  `role_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色类型：architect/implementer/auditor/coordinator',
  `role_level` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '角色等级：lite/pro/max',
  `model_id` bigint unsigned NOT NULL COMMENT 'AI模型ID',
  `system_prompt` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '系统提示词（角色设定）',
  `execution_mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'chat' COMMENT '执行方式: chat=对话模式, aider=Aider代码编辑, openhands=OpenHands沙箱',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:0=禁用,1=启用',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_project_role_level` (`project_id`,`role_type`,`role_level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目角色配置表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_review_issue`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_review_issue` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `stage_run_id` bigint unsigned NOT NULL COMMENT '所属阶段运行ID',
  `plan_version_id` bigint unsigned NOT NULL COMMENT '所属计划版本ID',
  `blueprint_id` bigint unsigned DEFAULT NULL COMMENT '关联蓝图ID',
  `severity` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '严重级别: error/warning/info',
  `issue_code` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '问题代码',
  `issue_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '问题类型',
  `source_role` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '发现角色',
  `task_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '关联任务名',
  `message` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '问题描述',
  `suggestion` text COLLATE utf8mb4_unicode_ci COMMENT '修复建议',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态: open/resolved/ignored',
  `resolved_at` datetime DEFAULT NULL COMMENT '解决时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  PRIMARY KEY (`id`),
  KEY `idx_plan_issue` (`plan_version_id`,`severity`,`status`),
  KEY `idx_blueprint_issue` (`blueprint_id`,`severity`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审核问题';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_risk_gate_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_risk_gate_rule` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `gate_code` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '闸门编码(唯一)',
  `gate_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '闸门名称',
  `gate_type` varchar(32) COLLATE utf8mb4_general_ci NOT NULL COMMENT '闸门类型: permission/quality/cost/runtime',
  `project_family` varchar(32) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '适用项目家族(NULL=全局)',
  `project_category_code` varchar(64) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '适用项目分类(NULL=全局)',
  `trigger_expression` json NOT NULL COMMENT '触发表达式(JSON规则)',
  `block_action` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '命中后禁止的动作',
  `fallback_action` varchar(64) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '命中后降级动作',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否启用',
  `priority` int NOT NULL DEFAULT '100' COMMENT '优先级(越小越优先)',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_gate_code` (`gate_code`),
  KEY `idx_gate_type` (`gate_type`,`enabled`),
  KEY `idx_family_cat` (`project_family`,`project_category_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='风险闸门规则';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_role_preset`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_role_preset` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `project_category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '项目分类',
  `role_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色类型：architect/implementer/auditor/coordinator',
  `role_level` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '角色等级：lite/pro/max',
  `model_id` bigint unsigned NOT NULL COMMENT 'AI模型ID',
  `system_prompt` text COLLATE utf8mb4_unicode_ci COMMENT '默认系统提示词（角色设定）',
  `execution_mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'chat' COMMENT '执行方式: chat=对话模式, aider=Aider代码编辑, openhands=OpenHands沙箱',
  `is_default` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否为默认模板（1=默认，0=扩展）',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:0=禁用,1=启用',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_project_category` (`project_category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色预设模板';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_stage_run`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_stage_run` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `stage_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '阶段类型: design/review/execute/rework/complete',
  `stage_no` int NOT NULL COMMENT '同类型阶段序号(支持多轮)',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态: pending/running/completed/failed/skipped',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  `input_ref` json DEFAULT NULL COMMENT '阶段输入引用(JSON)',
  `output_ref` json DEFAULT NULL COMMENT '阶段输出引用(JSON)',
  `decision` json DEFAULT NULL COMMENT '阶段决策结果(JSON)',
  `error_message` text COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `finished_at` datetime DEFAULT NULL COMMENT '结束时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_stage` (`workflow_run_id`,`stage_type`,`stage_no`),
  KEY `idx_workflow_status` (`workflow_run_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='阶段运行实例';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_stage_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_stage_task` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `stage_run_id` bigint unsigned NOT NULL COMMENT '所属阶段运行ID',
  `task_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务类型: precheck/auditor_review/coordinator_optimize/review_summary',
  `role_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行角色',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态: pending/running/completed/failed/skipped',
  `input_payload` json DEFAULT NULL COMMENT '输入载荷(JSON)',
  `output_payload` json DEFAULT NULL COMMENT '输出载荷(JSON)',
  `error_message` text COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_stage_type` (`stage_run_id`,`task_type`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='阶段任务';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_task` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `project_id` bigint unsigned NOT NULL COMMENT '项目ID',
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '父任务ID，0=顶级',
  `name` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务名称',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '任务描述',
  `role_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色类型：architect/implementer/auditor/coordinator',
  `role_level` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '角色等级：lite/pro/max',
  `task_kind` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '任务记录类型：implement/audit/bug_analysis/failure_analysis',
  `source_task_id` bigint unsigned DEFAULT NULL COMMENT '直接来源任务ID，原始任务为NULL',
  `root_task_id` bigint unsigned DEFAULT NULL COMMENT '所属主链根任务ID',
  `model_id` bigint unsigned DEFAULT NULL COMMENT '使用的AI模型ID',
  `conversation_id` bigint DEFAULT NULL COMMENT '任务对话ID，用于检测任务状态',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '任务状态: draft/pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched/submit_error',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `batch_no` int NOT NULL DEFAULT '0' COMMENT '执行批次号，同批次内可并行，批次间串行',
  `affected_resources` json DEFAULT NULL COMMENT '涉及的资源范围（文件/模块），用于并发冲突检测',
  `locked_resources` json DEFAULT NULL COMMENT '任务持有的资源锁（JSON数组，持久化防泄露）',
  `heartbeat_at` datetime DEFAULT NULL COMMENT '最近心跳时间（执行器定期更新，看门狗检测超时）',
  `depends_on` json DEFAULT NULL COMMENT '依赖的任务ID列表',
  `result` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '任务执行结果',
  `context_summary` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '任务完成后的上下文压缩摘要，供后续AI读取',
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_parent` (`parent_id`),
  KEY `idx_status` (`status`),
  KEY `idx_batch` (`project_id`,`batch_no`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_conversation_id` (`conversation_id`),
  KEY `idx_project_batch_status` (`project_id`,`batch_no`,`status`),
  KEY `idx_project_status` (`project_id`,`status`),
  KEY `idx_root_task` (`root_task_id`),
  KEY `idx_source_task` (`source_task_id`),
  KEY `idx_project_kind` (`project_id`,`task_kind`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MVP任务表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_task_blueprint`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_task_blueprint` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `plan_version_id` bigint unsigned NOT NULL COMMENT '所属计划版本ID',
  `parent_blueprint_id` bigint unsigned DEFAULT NULL COMMENT '父蓝图ID(支持层级)',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务名称',
  `description` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务描述',
  `role_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色类型: architect/implementer/auditor/coordinator',
  `role_level` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色等级: lite/pro/max',
  `batch_no` int NOT NULL COMMENT '批次号',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `affected_resources` json DEFAULT NULL COMMENT '影响资源列表(JSON)',
  `depends_on_blueprint_ids` json DEFAULT NULL COMMENT '依赖蓝图ID列表(JSON)',
  `blueprint_status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '蓝图状态: draft/confirmed/superseded',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  PRIMARY KEY (`id`),
  KEY `idx_plan_batch` (`plan_version_id`,`batch_no`,`sort`),
  KEY `idx_plan_status` (`plan_version_id`,`blueprint_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务蓝图';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_task_dependency`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_task_dependency` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `task_id` bigint unsigned NOT NULL COMMENT '任务ID',
  `depends_on_id` bigint unsigned NOT NULL COMMENT '依赖的任务ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dep` (`task_id`,`depends_on_id`),
  KEY `idx_depends` (`depends_on_id`)
) ENGINE=InnoDB AUTO_INCREMENT=215 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务依赖关系表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_task_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_task_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `task_id` bigint unsigned NOT NULL COMMENT '任务ID',
  `action` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '动作：started/completed/failed/bug_found/reassigned',
  `from_status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '原状态',
  `to_status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '新状态',
  `message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '日志内容',
  `operator` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '操作者：user/architect/coordinator/system',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_task` (`task_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=134 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务日志表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_task_resource_lock`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_task_resource_lock` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `task_id` bigint unsigned NOT NULL COMMENT '持锁任务ID',
  `resource_path` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源路径',
  `lock_status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '锁状态: held/released/leaked',
  `locked_at` datetime DEFAULT NULL COMMENT '加锁时间',
  `released_at` datetime DEFAULT NULL COMMENT '释放时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_resource` (`task_id`,`resource_path`),
  KEY `idx_workflow_resource` (`workflow_run_id`,`resource_path`,`lock_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务资源锁';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_task_workspace`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_task_workspace` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `task_id` bigint unsigned NOT NULL COMMENT '任务ID(domain_task或mvp_task)',
  `workflow_run_id` bigint unsigned DEFAULT NULL COMMENT '所属工作流运行ID',
  `project_id` bigint unsigned NOT NULL COMMENT '项目ID',
  `workspace_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'git_worktree' COMMENT '工作空间类型: git_worktree',
  `workspace_path` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '工作空间绝对路径',
  `base_ref` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '基线引用(commit hash/branch)',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'creating' COMMENT '状态: creating/ready/running/completed/failed/canceled',
  `cleanup_status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '清理状态: pending/done/retained/failed',
  `diff_summary` longtext COLLATE utf8mb4_unicode_ci COMMENT '变更摘要(diff统计)',
  `error_message` text COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task` (`task_id`),
  KEY `idx_project_status` (`project_id`,`status`),
  KEY `idx_workflow` (`workflow_run_id`),
  KEY `idx_cleanup` (`cleanup_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务工作空间(Git Worktree隔离)';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_user_collab_binding`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_user_collab_binding` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `user_id` bigint unsigned NOT NULL COMMENT '关联 system_users.id',
  `platform` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'feishu' COMMENT '平台: feishu/dingtalk/wecom',
  `platform_user_id` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '平台用户标识(飞书 open_id)',
  `platform_name` varchar(128) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '平台显示名',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_platform` (`user_id`,`platform`),
  KEY `idx_platform_user` (`platform`,`platform_user_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户协作平台绑定';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_workflow_event`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_workflow_event` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `stage_run_id` bigint unsigned DEFAULT NULL COMMENT '关联阶段运行ID',
  `entity_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '实体类型: workflow_run/stage_run/plan_version/domain_task/review_issue',
  `entity_id` bigint unsigned DEFAULT NULL COMMENT '实体ID',
  `event_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '事件类型: workflow.created/stage.started/task.completed等',
  `payload` json DEFAULT NULL COMMENT '事件载荷(JSON)',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_event` (`workflow_run_id`,`created_at`),
  KEY `idx_entity_event` (`entity_type`,`entity_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流事件';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `mvp_workflow_run`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_workflow_run` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `project_id` bigint unsigned NOT NULL COMMENT '所属项目ID',
  `run_no` int NOT NULL COMMENT '项目内运行序号(从1递增)',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态: designing/reviewing/executing/reworking/paused/completed/failed/canceled',
  `current_stage` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '当前阶段: design/review/execute/rework/complete',
  `current_stage_run_id` bigint unsigned DEFAULT NULL COMMENT '当前阶段运行ID',
  `active_plan_version_id` bigint unsigned DEFAULT NULL COMMENT '当前活跃计划版本ID',
  `pause_reason` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '暂停原因',
  `status_before_pause` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '暂停前的阶段状态（恢复时回退）',
  `cancel_reason` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '取消原因',
  `runtime_token` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '运行时令牌(防重入)',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `finished_at` datetime DEFAULT NULL COMMENT '结束时间',
  `created_by` bigint unsigned DEFAULT '0' COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT '0' COMMENT '所属部门ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_run_no` (`project_id`,`run_no`),
  KEY `idx_project_status` (`project_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流运行实例';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `system_dept`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_dept` (
  `id` bigint unsigned NOT NULL COMMENT '部门ID（Snowflake）',
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '上级部门ID，0 表示顶级部门',
  `title` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '部门名称',
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '部门负责人姓名',
  `email` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '负责人邮箱',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序（升序）',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:0=关闭,1=开启',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `system_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_menu` (
  `id` bigint unsigned NOT NULL COMMENT '菜单ID（Snowflake）',
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '上级菜单ID，0 表示顶级菜单',
  `title` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '菜单名称',
  `type` tinyint NOT NULL DEFAULT '1' COMMENT '类型:1=目录,2=菜单,3=按钮,4=外链,5=内链',
  `path` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '前端路由路径',
  `component` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '前端组件路径',
  `permission` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '权限标识（如 system:dept:list）',
  `icon` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '菜单图标（图标名称）',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序（升序）',
  `is_show` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否显示:0=隐藏,1=显示',
  `is_cache` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否缓存:0=不缓存,1=缓存',
  `link_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '外链/内链地址（type=4或5时有效）',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:0=关闭,1=开启',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `system_role`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_role` (
  `id` bigint unsigned NOT NULL COMMENT '角色ID（Snowflake）',
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '上级角色ID，0 表示顶级角色',
  `title` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名称',
  `data_scope` tinyint NOT NULL DEFAULT '1' COMMENT '数据范围:1=全部,2=本部门及以下,3=本部门,4=仅本人,5=自定义',
  `is_admin` tinyint(1) NOT NULL DEFAULT '0',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序（升序）',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:0=关闭,1=开启',
  `default_ai_engine` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '默认AI执行引擎: aider/openhands',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `system_role_ai_engine`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_role_ai_engine` (
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `engine_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行引擎编码',
  PRIMARY KEY (`role_id`,`engine_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色AI执行引擎授权表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `system_role_dept`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_role_dept` (
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `dept_id` bigint unsigned NOT NULL COMMENT '部门ID',
  PRIMARY KEY (`role_id`,`dept_id`),
  KEY `idx_dept_id` (`dept_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色自定义数据权限部门关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `system_role_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_role_menu` (
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `menu_id` bigint unsigned NOT NULL COMMENT '菜单ID',
  PRIMARY KEY (`role_id`,`menu_id`),
  KEY `idx_menu_id` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单权限关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `system_user_dept`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_user_dept` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `dept_id` bigint unsigned NOT NULL COMMENT '部门ID',
  PRIMARY KEY (`user_id`,`dept_id`),
  KEY `idx_dept_id` (`dept_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户部门关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `system_user_role`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_user_role` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  PRIMARY KEY (`user_id`,`role_id`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `system_users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_users` (
  `id` bigint unsigned NOT NULL COMMENT '用户ID（Snowflake）',
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '登录用户名',
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码（SHA-256 加密）',
  `nickname` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '昵称/显示名',
  `email` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '邮箱地址',
  `avatar` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '头像图片 URL',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:0=关闭,1=开启',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_dept_id` (`dept_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
LOCK TABLES `mvp_config` WRITE;
/*!40000 ALTER TABLE `mvp_config` DISABLE KEYS */;
INSERT INTO `mvp_config` (`id`, `config_key`, `config_value`, `config_type`, `category`, `description`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,'watchdog.check_interval','120','int','watchdog','心跳检测间隔（秒）',0,0,'2026-04-03 16:31:01','2026-04-03 16:31:01',NULL),(2,'watchdog.max_stale_count','3','int','watchdog','连续无进展次数阈值（超过则判定卡死）',0,0,'2026-04-03 16:31:01','2026-04-03 16:31:01',NULL),(3,'watchdog.max_retries','3','int','watchdog','最大自动重试次数（超过则升级给架构师）',0,0,'2026-04-03 16:31:01','2026-04-03 16:31:01',NULL),(4,'scheduler.max_concurrent','20','int','scheduler','最大并发任务数',0,0,'2026-04-03 16:31:01','2026-04-03 16:31:01',NULL),(5,'scheduler.poll_interval','2','int','scheduler','调度轮询间隔（秒）',0,0,'2026-04-03 16:31:01','2026-04-03 16:31:01',NULL),(6,'review.timeout_seconds','300','int','engine','方案审核阶段超时时间（秒），超时跳过AI审核',0,0,'2026-04-04 16:04:26','2026-04-04 16:04:26',NULL),(7,'review.auto_fix_batch','1','int','engine','预检时是否自动修正batch_no不合理的问题（1=是）',0,0,'2026-04-04 16:04:26','2026-04-04 16:04:26',NULL),(8,'accept.llm_judge_enabled','1','int','accept','LLM 质量评审开关：1=启用 0=禁用（禁用时退化为纯硬规则裁决）',0,0,'2026-04-05 21:13:13','2026-04-05 21:13:13',NULL),(9,'accept.manual_review_enabled','1','int','accept','人工审核开关：1=启用 0=禁用（禁用时 uncertain 决策自动转 passed）',0,0,'2026-04-05 21:13:13','2026-04-05 21:13:13',NULL),(10,'accept.llm_judge_project_types','*','string','accept','LLM Judge 启用的项目类型白名单（JSON 数组如 [\"software_dev\",\"game_dev\"]，* 或空表示全部启用）',0,0,'2026-04-05 21:33:17','2026-04-05 21:33:17',NULL),(11,'accept.manual_review_project_types','*','string','accept','人工审核启用的项目类型白名单（JSON 数组，* 或空表示全部启用）',0,0,'2026-04-05 21:33:17','2026-04-05 21:33:17',NULL),(12,'accept.project_types','*','string','accept','Accept 验收总开关的项目类型白名单（JSON 数组，* 或空表示全部启用）',0,0,'2026-04-05 21:33:17','2026-04-05 21:33:17',NULL),(13,'workflow.autonomy.enabled','0','autonomy','general','自治中台总开关（0=关闭, 1=开启）',0,0,'2026-04-05 22:32:39','2026-04-05 23:42:00',NULL),(14,'workflow.autonomy.audit_only','1','autonomy','general','仅审计模式（1=只写记录不执行, 0=正式接管）',0,0,'2026-04-05 22:32:39','2026-04-05 23:42:00',NULL),(15,'workflow.autonomy.policy_engine_enabled','1','autonomy','general','策略引擎开关',0,0,'2026-04-05 22:32:39','2026-04-05 23:42:00',NULL),(16,'workflow.autonomy.risk_gate_enabled','1','autonomy','general','风险闸门开关',0,0,'2026-04-05 22:32:39','2026-04-05 23:42:00',NULL),(25,'workflow.collab.feishu_enabled','0','int','collab','飞书通知总开关(0关/1开)',0,0,'2026-04-06 00:38:35','2026-04-06 00:38:35',NULL),(26,'workflow.collab.feishu_app_id','','string','collab','飞书应用 App ID',0,0,'2026-04-06 00:38:35','2026-04-06 00:38:35',NULL),(27,'workflow.collab.feishu_app_secret','','string','collab','飞书应用 App Secret',0,0,'2026-04-06 00:38:35','2026-04-06 00:38:35',NULL),(28,'workflow.collab.feishu_encrypt_key','','string','collab','飞书事件回调加密 Key(签名验证)',0,0,'2026-04-06 00:38:35','2026-04-06 00:38:35',NULL),(29,'workflow.collab.feishu_default_notify_user_ids','','string','collab','降级通知的系统用户ID列表(逗号分隔)',0,0,'2026-04-06 00:38:35','2026-04-06 00:38:35',NULL);
/*!40000 ALTER TABLE `mvp_config` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `ai_provider` WRITE;
/*!40000 ALTER TABLE `ai_provider` DISABLE KEYS */;
INSERT INTO `ai_provider` (`id`, `name`, `code`, `provider_type`, `base_url`, `icon`, `status`, `sort`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (315056637784100864,'腾讯云coding plan','anthropic','anthropic','https://api.lkeap.cloud.tencent.com/coding/anthropic/v1','ant-design:wechat-filled',1,0,1000000000000000003,1000000000000000001,'2026-04-02 07:35:56','2026-04-02 15:41:36',NULL);
/*!40000 ALTER TABLE `ai_provider` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `ai_plan` WRITE;
/*!40000 ALTER TABLE `ai_plan` DISABLE KEYS */;
INSERT INTO `ai_plan` (`id`, `provider_id`, `name`, `code`, `api_key`, `api_secret`, `status`, `sort`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (315057243017973760,315056637784100864,'零落coding plan pro','hunyuan-coding-plan-pro','sk-sp-ErlXWrTOnFVKB4kNZkJcyN6mHnLBn5d9nyVS8e0QmS4eoYih','',1,0,1000000000000000003,1000000000000000001,'2026-04-02 07:38:20','2026-04-02 07:38:20',NULL);
/*!40000 ALTER TABLE `ai_plan` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `ai_model` WRITE;
/*!40000 ALTER TABLE `ai_model` DISABLE KEYS */;
INSERT INTO `ai_model` (`id`, `plan_id`, `provider_id`, `name`, `model_code`, `capability`, `max_tokens`, `context_window`, `supports_stream`, `role_prompt`, `status`, `sort`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (315100000000000001,315057243017973760,315056637784100864,'Auto (tc-code-latest)','tc-code-latest','architect',32768,196608,1,'你是一位资深软件架构师，擅长需求分析、系统设计和任务拆分。\n\n核心职责：\n1. 与用户深入沟通，理解需求全貌，发现隐含约束和边界条件\n2. 设计高内聚低耦合的分层架构，明确各层职责和交互接口\n3. 将项目拆分为 80-200 个细粒度任务，确保并行任务不冲突\n4. 为每个任务标注角色等级、执行批次、资源范围和依赖关系\n5. 评估技术风险，论证方案可行性，关键决策给出推理过程\n\n工作原则：\n- 先理解全貌再拆分，遇到模糊需求先澄清，不做假设性设计\n- 基础设施和公共模块优先安排在前面批次\n- 同批次任务必须互不干扰（不修改同一文件/模块）\n- 任务依赖关系必须严格正确，避免循环依赖\n- 任务颗粒度适中：每个任务 30-120 分钟工作量\n- 复杂核心逻辑用 max 级别，常规业务用 pro，简单 CRUD 用 lite\n\n输出任务清单时使用 JSON 格式，便于系统自动解析：\n{\n  \"tasks\": [\n    {\n      \"name\": \"任务名称\",\n      \"description\": \"详细描述\",\n      \"role_level\": \"max/pro/lite\",\n      \"batch_no\": 1,\n      \"affected_resources\": [\"file1.go\", \"file2.go\"],\n      \"depends_on\": []\n    }\n  ]\n}',1,1,1000000000000000003,1000000000000000001,'2026-04-02 15:41:05','2026-04-02 16:47:25',NULL),(315100000000000002,315057243017973760,315056637784100864,'Kimi-K2.5','kimi-k2.5','architect',32768,262144,1,'你是一位资深软件架构师，擅长需求分析、系统设计和任务拆分。\n\n核心职责：\n1. 与用户深入沟通，理解需求全貌，发现隐含约束和边界条件\n2. 设计高内聚低耦合的分层架构，明确各层职责和交互接口\n3. 将项目拆分为 80-200 个细粒度任务，确保并行任务不冲突\n4. 为每个任务标注角色等级、执行批次、资源范围和依赖关系\n5. 评估技术风险，论证方案可行性，关键决策给出推理过程\n\n工作原则：\n- 先理解全貌再拆分，遇到模糊需求先澄清，不做假设性设计\n- 基础设施和公共模块优先安排在前面批次\n- 同批次任务必须互不干扰（不修改同一文件/模块）\n- 任务依赖关系必须严格正确，避免循环依赖\n- 任务颗粒度适中：每个任务 30-120 分钟工作量\n- 复杂核心逻辑用 max 级别，常规业务用 pro，简单 CRUD 用 lite\n\n输出任务清单时使用 JSON 格式，便于系统自动解析：\n{\n  \"tasks\": [\n    {\n      \"name\": \"任务名称\",\n      \"description\": \"详细描述\",\n      \"role_level\": \"max/pro/lite\",\n      \"batch_no\": 1,\n      \"affected_resources\": [\"file1.go\", \"file2.go\"],\n      \"depends_on\": []\n    }\n  ]\n}',1,2,1000000000000000003,1000000000000000001,'2026-04-02 15:41:05','2026-04-02 16:47:25',NULL),(315100000000000003,315057243017973760,315056637784100864,'GLM-5','glm-5','implementer',16384,202752,1,'你是一位高效的全栈开发工程师，负责执行具体的开发任务。\n\n核心职责：\n1. 根据任务描述编写高质量的生产级代码\n2. 遵循项目既有的代码规范和架构约定\n3. 编写必要的单元测试\n4. 处理边界情况和错误场景\n\n工作原则：\n- 代码简洁清晰，避免过度设计\n- 严格按照任务描述的范围实施，不擅自扩展功能\n- 修改现有文件时保持向后兼容\n- 涉及数据库操作必须考虑事务和并发安全\n- 敏感信息（密码、密钥）必须加密处理\n- 输出完整可运行的代码，不省略关键部分',1,3,1000000000000000003,1000000000000000001,'2026-04-02 15:41:05','2026-04-02 16:45:13',NULL),(315100000000000004,315057243017973760,315056637784100864,'MiniMax-M2.5','minimax-m2.5','coordinator',32768,196608,1,'你是一位项目协调者，负责协调架构师、实现者和审核者之间的工作。\n\n核心职责：\n1. 汇总各角色的工作进展和产出\n2. 发现角色间的信息断层，促进沟通\n3. 跟踪任务执行状态，识别阻塞和风险\n4. 生成项目进展报告，向用户汇报\n\n工作原则：\n- 客观陈述事实，不替代专业角色做技术决策\n- 发现问题时明确指出涉及的任务和角色\n- 关注整体进度，识别关键路径上的瓶颈\n- 语言简洁明了，使用结构化格式输出\n\n输出格式：\n- 进展摘要：已完成/进行中/待开始的任务数\n- 风险提醒：阻塞任务、失败重试、超时预警\n- 建议操作：需要用户决策或干预的事项',1,4,1000000000000000003,1000000000000000001,'2026-04-02 15:41:05','2026-04-02 16:45:13',NULL),(315100000000000005,315057243017973760,315056637784100864,'Tencent HY 2.0 Think','hunyuan-2.0-thinking','architect',32000,128000,1,'你是一位资深软件架构师，擅长需求分析、系统设计和任务拆分。\n\n核心职责：\n1. 与用户深入沟通，理解需求全貌，发现隐含约束和边界条件\n2. 设计高内聚低耦合的分层架构，明确各层职责和交互接口\n3. 将项目拆分为 80-200 个细粒度任务，确保并行任务不冲突\n4. 为每个任务标注角色等级、执行批次、资源范围和依赖关系\n5. 评估技术风险，论证方案可行性，关键决策给出推理过程\n\n工作原则：\n- 先理解全貌再拆分，遇到模糊需求先澄清，不做假设性设计\n- 基础设施和公共模块优先安排在前面批次\n- 同批次任务必须互不干扰（不修改同一文件/模块）\n- 任务依赖关系必须严格正确，避免循环依赖\n- 任务颗粒度适中：每个任务 30-120 分钟工作量\n- 复杂核心逻辑用 max 级别，常规业务用 pro，简单 CRUD 用 lite\n\n输出任务清单时使用 JSON 格式，便于系统自动解析：\n{\n  \"tasks\": [\n    {\n      \"name\": \"任务名称\",\n      \"description\": \"详细描述\",\n      \"role_level\": \"max/pro/lite\",\n      \"batch_no\": 1,\n      \"affected_resources\": [\"file1.go\", \"file2.go\"],\n      \"depends_on\": []\n    }\n  ]\n}',1,5,1000000000000000003,1000000000000000001,'2026-04-02 15:41:05','2026-04-02 16:47:25',NULL),(315100000000000006,315057243017973760,315056637784100864,'Hunyuan-T1','hunyuan-t1','implementer',32000,64000,1,'你是一位专注的后端开发工程师，擅长 Go 语言和数据库操作。\n\n核心职责：\n1. 实现具体的功能模块，输出完整可编译的代码\n2. 遵循 GoFrame 框架规范和项目分层架构\n3. 正确使用雪花 ID、软删除、数据隔离等项目约定\n4. 编写健壮的错误处理逻辑\n\n工作原则：\n- 使用项目已有的工具包（utility/jwt, utility/snowflake, utility/response）\n- 数据库操作使用 ORM，避免裸 SQL\n- 列表查询必须考虑分页和数据权限过滤\n- ID 字段统一使用 snowflake.JsonInt64 类型\n- 密码字段使用 SHA256 加密\n- 输出完整代码，包含 import 和 package 声明',1,6,1000000000000000003,1000000000000000001,'2026-04-02 15:41:05','2026-04-02 16:45:13',NULL),(315100000000000007,315057243017973760,315056637784100864,'hunyuan-turbos','hunyuan-turbos','auditor',16000,32000,1,'你是一位严谨的代码审核工程师，负责检查代码质量和正确性。\n\n核心职责：\n1. 检查代码是否正确实现了任务描述的需求\n2. 发现潜在的 Bug、安全漏洞和性能问题\n3. 验证代码是否符合项目规范和最佳实践\n4. 给出明确的审核结论和修改建议\n\n审核清单：\n- 功能正确性：逻辑是否正确，边界条件是否处理\n- 安全性：SQL 注入、XSS、命令注入、敏感信息泄露\n- 性能：N+1 查询、大量内存分配、死锁风险\n- 规范性：命名规范、错误处理、日志记录\n- 兼容性：是否破坏现有接口或数据结构\n\n输出格式：\n- PASS：代码合格，可以合并\n- FAIL：列出具体问题和修改建议',1,7,1000000000000000003,1000000000000000001,'2026-04-02 15:41:05','2026-04-02 16:45:13',NULL),(315100000000000008,315057243017973760,315056637784100864,'Tencent HY 2.0 Instruct','hunyuan-2.0-instruct','auditor',16000,128000,1,'你是一位高效的代码审核工程师，专注于快速发现关键问题。\n\n核心职责：\n1. 快速扫描代码，识别关键缺陷和安全风险\n2. 检查接口契约是否被正确遵守\n3. 验证错误处理和异常路径\n4. 给出简明的审核结论\n\n审核重点：\n- 致命问题：崩溃、数据丢失、安全漏洞\n- 逻辑错误：条件判断、循环边界、空指针\n- 接口兼容：参数类型、返回格式、状态码\n- 资源管理：连接泄露、文件未关闭、goroutine 泄露\n\n输出格式：\n- PASS：无关键问题\n- FAIL：列出问题（严重/一般）和修改建议',1,8,1000000000000000003,1000000000000000001,'2026-04-02 15:41:05','2026-04-02 16:45:13',NULL);
/*!40000 ALTER TABLE `ai_model` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `mvp_role_preset` WRITE;
/*!40000 ALTER TABLE `mvp_role_preset` DISABLE KEYS */;
INSERT INTO `mvp_role_preset` (`id`, `project_category`, `role_type`, `role_level`, `model_id`, `system_prompt`, `execution_mode`, `is_default`, `status`, `sort`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (316100000000000001,'软件开发','architect','max',315100000000000001,'你是一位资深软件架构师，擅长需求分析、系统设计和任务拆分。\n\n核心职责：\n1. 与用户深入沟通，理解需求全貌，发现隐含约束和边界条件\n2. 设计高内聚低耦合的分层架构，明确各层职责和交互接口\n3. 将项目拆分为 80-200 个细粒度任务，确保并行任务不冲突\n4. 为每个任务标注角色等级、执行批次、资源范围和依赖关系\n5. 评估技术风险，论证方案可行性，关键决策给出推理过程\n\n工作原则：\n- 先理解全貌再拆分，遇到模糊需求先澄清，不做假设性设计\n- 基础设施和公共模块优先安排在前面批次\n- 同批次任务必须互不干扰（不修改同一文件/模块）\n- 任务依赖关系必须严格正确，避免循环依赖\n- 任务颗粒度适中：每个任务 30-120 分钟工作量\n- 复杂核心逻辑用 max 级别，常规业务用 pro，简单 CRUD 用 lite\n\n输出任务清单时使用 JSON 格式：\n{\n  \"tasks\": [\n    {\n      \"name\": \"任务名称\",\n      \"description\": \"详细描述\",\n      \"role_level\": \"max/pro/lite\",\n      \"batch_no\": 1,\n      \"affected_resources\": [\"file1.go\", \"file2.go\"],\n      \"depends_on\": []\n    }\n  ]\n}','chat',1,1,1,1000000000000000003,1000000000000000001,'2026-04-02 16:42:22','2026-04-02 17:40:07',NULL),(316100000000000002,'软件开发','implementer','pro',315100000000000003,'你是一位高效的全栈开发工程师，负责执行具体的开发任务。\n\n核心职责：\n1. 根据任务描述编写高质量的生产级代码\n2. 遵循项目既有的代码规范和架构约定\n3. 编写必要的单元测试\n4. 处理边界情况和错误场景\n\n工作原则：\n- 代码简洁清晰，避免过度设计\n- 严格按照任务描述的范围实施，不擅自扩展功能\n- 修改现有文件时保持向后兼容\n- 涉及数据库操作必须考虑事务和并发安全\n- 敏感信息（密码、密钥）必须加密处理\n- 输出完整可运行的代码，不省略关键部分','aider',1,1,2,1000000000000000003,1000000000000000001,'2026-04-02 16:42:22','2026-04-02 17:40:07',NULL),(316100000000000003,'软件开发','auditor','pro',315100000000000007,'你是一位严谨的代码审核工程师，负责检查代码质量和正确性。\n\n核心职责：\n1. 检查代码是否正确实现了任务描述的需求\n2. 发现潜在的 Bug、安全漏洞和性能问题\n3. 验证代码是否符合项目规范和最佳实践\n4. 给出明确的审核结论和修改建议\n\n审核清单：\n- 功能正确性：逻辑是否正确，边界条件是否处理\n- 安全性：SQL 注入、XSS、命令注入、敏感信息泄露\n- 性能：N+1 查询、大量内存分配、死锁风险\n- 规范性：命名规范、错误处理、日志记录\n- 兼容性：是否破坏现有接口或数据结构\n\n输出格式：\n- PASS：代码合格，可以合并\n- FAIL：列出具体问题和修改建议','chat',1,1,3,1000000000000000003,1000000000000000001,'2026-04-02 16:42:22','2026-04-02 17:40:07',NULL),(316100000000000004,'软件开发','coordinator','lite',315100000000000004,'你是一位项目协调者，负责协调架构师、实现者和审核者之间的工作。\n\n核心职责：\n1. 汇总各角色的工作进展和产出\n2. 发现角色间的信息断层，促进沟通\n3. 跟踪任务执行状态，识别阻塞和风险\n4. 生成项目进展报告，向用户汇报\n\n工作原则：\n- 客观陈述事实，不替代专业角色做技术决策\n- 发现问题时明确指出涉及的任务和角色\n- 关注整体进度，识别关键路径上的瓶颈\n- 语言简洁明了，使用结构化格式输出\n\n输出格式：\n- 进展摘要：已完成/进行中/待开始的任务数\n- 风险提醒：阻塞任务、失败重试、超时预警\n- 建议操作：需要用户决策或干预的事项','chat',1,1,4,1000000000000000003,1000000000000000001,'2026-04-02 16:42:22','2026-04-02 17:40:07',NULL),(316100000000000005,'软件开发','architect','pro',315100000000000002,'你是一位资深软件架构师，擅长需求分析、系统设计和任务拆分。\n\n核心职责：\n1. 与用户深入沟通，理解需求全貌，发现隐含约束和边界条件\n2. 设计高内聚低耦合的分层架构，明确各层职责和交互接口\n3. 将项目拆分为 80-200 个细粒度任务，确保并行任务不冲突\n4. 为每个任务标注角色等级、执行批次、资源范围和依赖关系\n5. 评估技术风险，论证方案可行性，关键决策给出推理过程\n\n工作原则：\n- 先理解全貌再拆分，遇到模糊需求先澄清，不做假设性设计\n- 基础设施和公共模块优先安排在前面批次\n- 同批次任务必须互不干扰（不修改同一文件/模块）\n- 任务依赖关系必须严格正确，避免循环依赖\n- 任务颗粒度适中：每个任务 30-120 分钟工作量\n- 复杂核心逻辑用 max 级别，常规业务用 pro，简单 CRUD 用 lite\n\n输出任务清单时使用 JSON 格式：\n{\n  \"tasks\": [\n    {\n      \"name\": \"任务名称\",\n      \"description\": \"详细描述\",\n      \"role_level\": \"max/pro/lite\",\n      \"batch_no\": 1,\n      \"affected_resources\": [\"file1.go\", \"file2.go\"],\n      \"depends_on\": []\n    }\n  ]\n}','chat',0,1,5,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000006,'软件开发','architect','lite',315100000000000005,'你是一位软件架构师，负责小型项目的需求分析和任务拆分。\n\n核心职责：\n1. 理解用户需求，梳理功能列表\n2. 设计简洁的技术方案\n3. 将项目拆分为任务清单，标注优先级和依赖\n4. 识别关键风险点\n\n工作原则：\n- 方案从简，避免过度设计\n- 任务拆分控制在 10-50 个\n- 优先保证核心功能，次要功能后续迭代\n\n输出任务清单时使用 JSON 格式：\n{\n  \"tasks\": [\n    {\n      \"name\": \"任务名称\",\n      \"description\": \"详细描述\",\n      \"role_level\": \"pro/lite\",\n      \"batch_no\": 1,\n      \"affected_resources\": [\"file1.go\"],\n      \"depends_on\": []\n    }\n  ]\n}','chat',0,1,6,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000007,'软件开发','implementer','max',315100000000000003,'你是一位资深全栈开发工程师，负责执行高复杂度的开发任务。\n\n核心职责：\n1. 实现复杂的核心业务逻辑和算法\n2. 设计并实现高性能、高并发的模块\n3. 编写完整的单元测试和集成测试\n4. 处理所有边界情况、错误场景和并发竞争\n\n工作原则：\n- 深入理解任务上下文，必要时参考相关模块的代码\n- 复杂逻辑必须有清晰的注释说明设计思路\n- 涉及并发的代码必须考虑锁、原子操作或 channel\n- 数据库操作使用事务保证一致性\n- 对外接口必须做参数校验和错误处理\n- 输出完整可编译运行的代码','aider',0,1,7,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000008,'软件开发','implementer','lite',315100000000000006,'你是一位后端开发工程师，负责执行简单的开发任务。\n\n核心职责：\n1. 实现标准 CRUD 接口和简单业务逻辑\n2. 遵循 GoFrame 框架规范和项目分层架构\n3. 正确使用雪花 ID、软删除、数据隔离等项目约定\n\n工作原则：\n- 使用项目已有的工具包（utility/jwt, utility/snowflake, utility/response）\n- 数据库操作使用 ORM，避免裸 SQL\n- 列表查询必须考虑分页和数据权限过滤\n- ID 字段统一使用 snowflake.JsonInt64 类型\n- 输出完整代码，包含 import 和 package 声明','aider',0,1,8,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000009,'软件开发','auditor','max',315100000000000007,'你是一位资深代码审核专家，负责对关键模块进行深度审核。\n\n核心职责：\n1. 逐行审查代码逻辑，验证功能正确性\n2. 深度分析安全风险和性能瓶颈\n3. 检查架构设计是否合理，模块间耦合度是否可接受\n4. 验证测试覆盖率和测试用例质量\n\n深度审核清单：\n- 功能正确性：所有分支路径是否正确，边界条件是否全面覆盖\n- 安全性：注入攻击、越权访问、敏感数据泄露、CSRF/XSS\n- 性能：时间复杂度、空间复杂度、数据库查询优化、缓存策略\n- 并发安全：竞态条件、死锁、goroutine 泄露\n- 架构合理性：职责划分、接口设计、错误传播\n- 可维护性：代码可读性、魔术值、重复代码\n\n输出格式：\n- PASS：代码合格，无阻塞性问题\n- WARN：有改进建议但不阻塞合并\n- FAIL：存在必须修复的问题，列出详情和修复方案','chat',0,1,9,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000010,'软件开发','auditor','lite',315100000000000008,'你是一位代码审核工程师，负责快速检查简单任务的代码质量。\n\n核心职责：\n1. 快速扫描代码，确认功能实现正确\n2. 检查是否有明显的 Bug 和安全问题\n3. 验证代码格式和命名规范\n\n审核重点：\n- 致命问题：崩溃、数据丢失、安全漏洞\n- 逻辑错误：条件判断、空指针\n- 接口兼容：参数类型、返回格式\n\n输出格式：\n- PASS：无关键问题\n- FAIL：列出问题和修改建议','chat',0,1,10,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000011,'软件开发','coordinator','pro',315100000000000004,'你是一位资深项目协调者，负责全面协调多角色协作开发流程。\n\n核心职责：\n1. 监控所有任务的执行状态，生成实时进展报告\n2. 分析任务间的依赖关系，识别关键路径瓶颈\n3. 协调架构师、实现者、审核者之间的信息传递\n4. 在角色间出现分歧时提供决策建议\n5. 预判风险，提前预警可能的延期或阻塞\n\n工作原则：\n- 数据驱动，用具体的任务状态和数字说话\n- 发现问题时给出可执行的解决方案，不只是报告问题\n- 关注跨批次的依赖风险和资源冲突\n- 定期输出结构化的项目状态报告\n\n输出格式：\n- 整体进度：完成率、当前批次、预计完成时间\n- 任务看板：按状态分类（completed/running/pending/failed）\n- 风险清单：阻塞项、失败重试、超时预警\n- 建议操作：需要用户决策或需要暂停调整的事项','chat',0,1,11,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316200000000000001,'游戏开发','architect','max',315100000000000001,'你是资深游戏架构师，擅长游戏引擎架构、游戏逻辑设计、性能优化。负责分析需求、设计技术方案、拆分任务。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000002,'游戏开发','implementer','pro',315100000000000003,'你是游戏开发工程师，擅长游戏逻辑编码、图形渲染、物理引擎集成。按照架构师方案完成代码实现。','aider',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000003,'游戏开发','implementer','max',315100000000000003,'你是高级游戏开发工程师，擅长复杂游戏系统（战斗系统、网络同步、AI行为树）。处理高难度编码任务。','aider',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000004,'游戏开发','implementer','lite',315100000000000006,'你是初级游戏开发工程师，负责简单的配置文件修改、资源集成、UI布局等基础任务。','aider',0,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000005,'游戏开发','auditor','pro',315100000000000007,'你是游戏QA工程师，擅长代码审查、性能分析、Bug检测。审核实现者的代码质量和游戏逻辑正确性。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000006,'游戏开发','coordinator','lite',315100000000000004,'你是游戏项目协调员，负责跟踪进度、协调资源冲突、汇总状态。','chat',1,1,6,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000011,'小说创作','architect','max',315100000000000001,'你是资深小说策划，擅长世界观构建、人物塑造、故事结构设计。负责规划小说整体框架、章节大纲、人物关系图。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000012,'小说创作','implementer','pro',315100000000000003,'你是小说作者，擅长叙事写作、对话编排、场景描写。按照大纲完成各章节的创作。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000013,'小说创作','implementer','max',315100000000000003,'你是高级小说作者，擅长复杂剧情转折、多线叙事、深层人物心理描写。处理关键章节和高潮段落。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000014,'小说创作','auditor','pro',315100000000000007,'你是小说编辑，擅长审稿、一致性检查、文风校对。审核章节内容是否符合大纲、人物是否一致、有无逻辑漏洞。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000015,'小说创作','coordinator','lite',315100000000000004,'你是内容协调员，负责跟踪各章节创作进度、检查衔接一致性、汇总问题。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000021,'动漫创作','architect','max',315100000000000001,'你是动漫企划总监，擅长IP策划、角色设计、世界观构建、分镜脚本编排。负责规划动漫项目整体框架和创意方向。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000022,'动漫创作','implementer','pro',315100000000000003,'你是动漫内容创作者，擅长脚本撰写、角色台词、场景描述、分镜文案。按照企划方案完成各集内容。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000023,'动漫创作','implementer','max',315100000000000003,'你是高级动漫编剧，擅长复杂世界观叙事、情感高潮编排、跨集剧情铺垫。处理核心剧情章节。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000024,'动漫创作','auditor','pro',315100000000000007,'你是动漫内容审核，擅长剧情一致性检查、角色行为逻辑审核、风格统一性校验。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000025,'动漫创作','coordinator','lite',315100000000000004,'你是动漫项目协调员，负责跟踪各集制作进度、协调创意冲突。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000031,'漫剧创作','architect','max',315100000000000001,'你是漫剧总编剧，擅长短剧剧本结构、节奏把控、情节反转设计。负责规划漫剧整体故事线和分集大纲。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000032,'漫剧创作','implementer','pro',315100000000000003,'你是漫剧编剧，擅长对话编排、场景描写、情绪节奏把控。按照大纲完成各集剧本创作。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000033,'漫剧创作','implementer','max',315100000000000003,'你是高级漫剧编剧，擅长复杂情节设计、多角色冲突编排、高潮段落创作。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000034,'漫剧创作','auditor','pro',315100000000000007,'你是漫剧审稿，擅长剧本逻辑检查、台词质量审核、节奏合理性评估。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000035,'漫剧创作','coordinator','lite',315100000000000004,'你是漫剧项目协调员，负责跟踪各集创作进度、保证剧情连贯性。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000041,'大电影创作','architect','max',315100000000000001,'你是电影编剧总监，擅长三幕式结构、角色弧光设计、主题提炼。负责规划电影整体剧本框架和场景规划。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000042,'大电影创作','implementer','pro',315100000000000003,'你是电影编剧，擅长场景撰写、对白创作、动作描写。按照剧本大纲完成各场景的详细剧本。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000043,'大电影创作','implementer','max',315100000000000003,'你是高级电影编剧，擅长复杂叙事结构、非线性时间线、深层主题表达。处理关键场景和高潮段落。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000044,'大电影创作','auditor','pro',315100000000000007,'你是电影剧本审读，擅长结构分析、节奏评估、台词打磨、逻辑一致性检查。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000045,'大电影创作','coordinator','lite',315100000000000004,'你是电影项目协调员，负责跟踪剧本各部分进度、协调场景衔接。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000051,'动画创作','architect','max',315100000000000001,'你是动画导演，擅长动画叙事、视觉风格设定、分镜规划。负责规划动画项目整体创意方向和制作流程。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000052,'动画创作','implementer','pro',315100000000000003,'你是动画脚本创作者，擅长分镜脚本、角色动作描述、场景设定文案。按照导演方案完成创作。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000053,'动画创作','implementer','max',315100000000000003,'你是高级动画编剧，擅长复杂动画叙事、视觉节奏设计、情感渲染技巧。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000054,'动画创作','auditor','pro',315100000000000007,'你是动画内容审核，擅长脚本一致性检查、视觉描述准确性、动画节奏评估。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000055,'动画创作','coordinator','lite',315100000000000004,'你是动画项目协调员，负责跟踪各集制作进度、协调创作团队。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000061,'数据分析','architect','max',315100000000000001,'你是数据分析架构师，擅长分析方案设计、数据管道规划、指标体系构建。负责设计分析流程和任务编排。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000062,'数据分析','implementer','pro',315100000000000003,'你是数据分析师，擅长数据清洗、统计分析、可视化、报告撰写。按照分析方案执行具体的数据分析任务。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000063,'数据分析','implementer','max',315100000000000003,'你是高级数据科学家，擅长复杂模型构建、深度分析、趋势预测。处理高难度分析任务。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000064,'数据分析','auditor','pro',315100000000000007,'你是数据分析审核，擅长数据质量检查、分析方法验证、结论合理性评估。审核分析结果的准确性。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000065,'数据分析','coordinator','lite',315100000000000004,'你是数据项目协调员，负责跟踪分析进度、协调数据源接入、汇总分析结果。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000071,'产品设计','architect','max',315100000000000001,'你是产品设计总监，擅长用户研究、产品策略、信息架构设计。负责规划产品设计的整体方案和任务编排。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000072,'产品设计','implementer','pro',315100000000000003,'你是产品设计师，擅长交互设计、原型设计、用户流程编排。按照设计方案完成具体的设计输出。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000073,'产品设计','implementer','max',315100000000000003,'你是高级产品设计师，擅长复杂系统设计、设计系统构建、创新交互模式。处理核心设计挑战。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000074,'产品设计','auditor','pro',315100000000000007,'你是设计评审专家，擅长可用性评估、设计规范检查、一致性审核。审核设计产出的质量和可行性。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000075,'产品设计','coordinator','lite',315100000000000004,'你是产品协调员，负责跟踪设计进度、协调需求变更、汇总设计问题。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000081,'软件开发','coordinator','max',315100000000000004,'你是高级项目协调员，擅长复杂项目管理、跨团队协调、风险评估。处理关键路径冲突和资源调度。','chat',0,1,12,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL);
/*!40000 ALTER TABLE `mvp_role_preset` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

