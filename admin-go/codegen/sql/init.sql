
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
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MVP配置表';
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
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_version` (`project_id`,`version_no`),
  KEY `idx_workflow_status` (`workflow_run_id`,`status`,`review_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计划版本';
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
  PRIMARY KEY (`id`),
  KEY `idx_plan_issue` (`plan_version_id`,`severity`,`status`),
  KEY `idx_blueprint_issue` (`blueprint_id`,`severity`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审核问题';
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

LOCK TABLES `mvp_role_preset` WRITE;
/*!40000 ALTER TABLE `mvp_role_preset` DISABLE KEYS */;
INSERT INTO `mvp_role_preset` (`id`, `project_category`, `role_type`, `role_level`, `model_id`, `system_prompt`, `execution_mode`, `is_default`, `status`, `sort`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (316100000000000001,'软件开发','architect','max',315100000000000001,'你是一位资深软件架构师，擅长需求分析、系统设计和任务拆分。\n\n核心职责：\n1. 与用户深入沟通，理解需求全貌，发现隐含约束和边界条件\n2. 设计高内聚低耦合的分层架构，明确各层职责和交互接口\n3. 将项目拆分为 80-200 个细粒度任务，确保并行任务不冲突\n4. 为每个任务标注角色等级、执行批次、资源范围和依赖关系\n5. 评估技术风险，论证方案可行性，关键决策给出推理过程\n\n工作原则：\n- 先理解全貌再拆分，遇到模糊需求先澄清，不做假设性设计\n- 基础设施和公共模块优先安排在前面批次\n- 同批次任务必须互不干扰（不修改同一文件/模块）\n- 任务依赖关系必须严格正确，避免循环依赖\n- 任务颗粒度适中：每个任务 30-120 分钟工作量\n- 复杂核心逻辑用 max 级别，常规业务用 pro，简单 CRUD 用 lite\n\n输出任务清单时使用 JSON 格式：\n{\n  \"tasks\": [\n    {\n      \"name\": \"任务名称\",\n      \"description\": \"详细描述\",\n      \"role_level\": \"max/pro/lite\",\n      \"batch_no\": 1,\n      \"affected_resources\": [\"file1.go\", \"file2.go\"],\n      \"depends_on\": []\n    }\n  ]\n}','chat',1,1,1,1000000000000000003,1000000000000000001,'2026-04-02 16:42:22','2026-04-02 17:40:07',NULL),(316100000000000002,'软件开发','implementer','pro',315100000000000003,'你是一位高效的全栈开发工程师，负责执行具体的开发任务。\n\n核心职责：\n1. 根据任务描述编写高质量的生产级代码\n2. 遵循项目既有的代码规范和架构约定\n3. 编写必要的单元测试\n4. 处理边界情况和错误场景\n\n工作原则：\n- 代码简洁清晰，避免过度设计\n- 严格按照任务描述的范围实施，不擅自扩展功能\n- 修改现有文件时保持向后兼容\n- 涉及数据库操作必须考虑事务和并发安全\n- 敏感信息（密码、密钥）必须加密处理\n- 输出完整可运行的代码，不省略关键部分','aider',1,1,2,1000000000000000003,1000000000000000001,'2026-04-02 16:42:22','2026-04-02 17:40:07',NULL),(316100000000000003,'软件开发','auditor','pro',315100000000000007,'你是一位严谨的代码审核工程师，负责检查代码质量和正确性。\n\n核心职责：\n1. 检查代码是否正确实现了任务描述的需求\n2. 发现潜在的 Bug、安全漏洞和性能问题\n3. 验证代码是否符合项目规范和最佳实践\n4. 给出明确的审核结论和修改建议\n\n审核清单：\n- 功能正确性：逻辑是否正确，边界条件是否处理\n- 安全性：SQL 注入、XSS、命令注入、敏感信息泄露\n- 性能：N+1 查询、大量内存分配、死锁风险\n- 规范性：命名规范、错误处理、日志记录\n- 兼容性：是否破坏现有接口或数据结构\n\n输出格式：\n- PASS：代码合格，可以合并\n- FAIL：列出具体问题和修改建议','chat',1,1,3,1000000000000000003,1000000000000000001,'2026-04-02 16:42:22','2026-04-02 17:40:07',NULL),(316100000000000004,'软件开发','coordinator','lite',315100000000000004,'你是一位项目协调者，负责协调架构师、实现者和审核者之间的工作。\n\n核心职责：\n1. 汇总各角色的工作进展和产出\n2. 发现角色间的信息断层，促进沟通\n3. 跟踪任务执行状态，识别阻塞和风险\n4. 生成项目进展报告，向用户汇报\n\n工作原则：\n- 客观陈述事实，不替代专业角色做技术决策\n- 发现问题时明确指出涉及的任务和角色\n- 关注整体进度，识别关键路径上的瓶颈\n- 语言简洁明了，使用结构化格式输出\n\n输出格式：\n- 进展摘要：已完成/进行中/待开始的任务数\n- 风险提醒：阻塞任务、失败重试、超时预警\n- 建议操作：需要用户决策或干预的事项','chat',1,1,4,1000000000000000003,1000000000000000001,'2026-04-02 16:42:22','2026-04-02 17:40:07',NULL),(316100000000000005,'软件开发','architect','pro',315100000000000002,'你是一位资深软件架构师，擅长需求分析、系统设计和任务拆分。\n\n核心职责：\n1. 与用户深入沟通，理解需求全貌，发现隐含约束和边界条件\n2. 设计高内聚低耦合的分层架构，明确各层职责和交互接口\n3. 将项目拆分为 80-200 个细粒度任务，确保并行任务不冲突\n4. 为每个任务标注角色等级、执行批次、资源范围和依赖关系\n5. 评估技术风险，论证方案可行性，关键决策给出推理过程\n\n工作原则：\n- 先理解全貌再拆分，遇到模糊需求先澄清，不做假设性设计\n- 基础设施和公共模块优先安排在前面批次\n- 同批次任务必须互不干扰（不修改同一文件/模块）\n- 任务依赖关系必须严格正确，避免循环依赖\n- 任务颗粒度适中：每个任务 30-120 分钟工作量\n- 复杂核心逻辑用 max 级别，常规业务用 pro，简单 CRUD 用 lite\n\n输出任务清单时使用 JSON 格式：\n{\n  \"tasks\": [\n    {\n      \"name\": \"任务名称\",\n      \"description\": \"详细描述\",\n      \"role_level\": \"max/pro/lite\",\n      \"batch_no\": 1,\n      \"affected_resources\": [\"file1.go\", \"file2.go\"],\n      \"depends_on\": []\n    }\n  ]\n}','chat',0,1,5,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000006,'软件开发','architect','lite',315100000000000005,'你是一位软件架构师，负责小型项目的需求分析和任务拆分。\n\n核心职责：\n1. 理解用户需求，梳理功能列表\n2. 设计简洁的技术方案\n3. 将项目拆分为任务清单，标注优先级和依赖\n4. 识别关键风险点\n\n工作原则：\n- 方案从简，避免过度设计\n- 任务拆分控制在 10-50 个\n- 优先保证核心功能，次要功能后续迭代\n\n输出任务清单时使用 JSON 格式：\n{\n  \"tasks\": [\n    {\n      \"name\": \"任务名称\",\n      \"description\": \"详细描述\",\n      \"role_level\": \"pro/lite\",\n      \"batch_no\": 1,\n      \"affected_resources\": [\"file1.go\"],\n      \"depends_on\": []\n    }\n  ]\n}','chat',0,1,6,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000007,'软件开发','implementer','max',315100000000000003,'你是一位资深全栈开发工程师，负责执行高复杂度的开发任务。\n\n核心职责：\n1. 实现复杂的核心业务逻辑和算法\n2. 设计并实现高性能、高并发的模块\n3. 编写完整的单元测试和集成测试\n4. 处理所有边界情况、错误场景和并发竞争\n\n工作原则：\n- 深入理解任务上下文，必要时参考相关模块的代码\n- 复杂逻辑必须有清晰的注释说明设计思路\n- 涉及并发的代码必须考虑锁、原子操作或 channel\n- 数据库操作使用事务保证一致性\n- 对外接口必须做参数校验和错误处理\n- 输出完整可编译运行的代码','aider',0,1,7,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000008,'软件开发','implementer','lite',315100000000000006,'你是一位后端开发工程师，负责执行简单的开发任务。\n\n核心职责：\n1. 实现标准 CRUD 接口和简单业务逻辑\n2. 遵循 GoFrame 框架规范和项目分层架构\n3. 正确使用雪花 ID、软删除、数据隔离等项目约定\n\n工作原则：\n- 使用项目已有的工具包（utility/jwt, utility/snowflake, utility/response）\n- 数据库操作使用 ORM，避免裸 SQL\n- 列表查询必须考虑分页和数据权限过滤\n- ID 字段统一使用 snowflake.JsonInt64 类型\n- 输出完整代码，包含 import 和 package 声明','aider',0,1,8,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000009,'软件开发','auditor','max',315100000000000007,'你是一位资深代码审核专家，负责对关键模块进行深度审核。\n\n核心职责：\n1. 逐行审查代码逻辑，验证功能正确性\n2. 深度分析安全风险和性能瓶颈\n3. 检查架构设计是否合理，模块间耦合度是否可接受\n4. 验证测试覆盖率和测试用例质量\n\n深度审核清单：\n- 功能正确性：所有分支路径是否正确，边界条件是否全面覆盖\n- 安全性：注入攻击、越权访问、敏感数据泄露、CSRF/XSS\n- 性能：时间复杂度、空间复杂度、数据库查询优化、缓存策略\n- 并发安全：竞态条件、死锁、goroutine 泄露\n- 架构合理性：职责划分、接口设计、错误传播\n- 可维护性：代码可读性、魔术值、重复代码\n\n输出格式：\n- PASS：代码合格，无阻塞性问题\n- WARN：有改进建议但不阻塞合并\n- FAIL：存在必须修复的问题，列出详情和修复方案','chat',0,1,9,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000010,'软件开发','auditor','lite',315100000000000008,'你是一位代码审核工程师，负责快速检查简单任务的代码质量。\n\n核心职责：\n1. 快速扫描代码，确认功能实现正确\n2. 检查是否有明显的 Bug 和安全问题\n3. 验证代码格式和命名规范\n\n审核重点：\n- 致命问题：崩溃、数据丢失、安全漏洞\n- 逻辑错误：条件判断、空指针\n- 接口兼容：参数类型、返回格式\n\n输出格式：\n- PASS：无关键问题\n- FAIL：列出问题和修改建议','chat',0,1,10,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316100000000000011,'软件开发','coordinator','pro',315100000000000004,'你是一位资深项目协调者，负责全面协调多角色协作开发流程。\n\n核心职责：\n1. 监控所有任务的执行状态，生成实时进展报告\n2. 分析任务间的依赖关系，识别关键路径瓶颈\n3. 协调架构师、实现者、审核者之间的信息传递\n4. 在角色间出现分歧时提供决策建议\n5. 预判风险，提前预警可能的延期或阻塞\n\n工作原则：\n- 数据驱动，用具体的任务状态和数字说话\n- 发现问题时给出可执行的解决方案，不只是报告问题\n- 关注跨批次的依赖风险和资源冲突\n- 定期输出结构化的项目状态报告\n\n输出格式：\n- 整体进度：完成率、当前批次、预计完成时间\n- 任务看板：按状态分类（completed/running/pending/failed）\n- 风险清单：阻塞项、失败重试、超时预警\n- 建议操作：需要用户决策或需要暂停调整的事项','chat',0,1,11,1000000000000000003,1000000000000000001,'2026-04-02 17:40:07','2026-04-02 17:40:07',NULL),(316200000000000001,'游戏开发','architect','max',315100000000000001,'你是资深游戏架构师，擅长游戏引擎架构、游戏逻辑设计、性能优化。负责分析需求、设计技术方案、拆分任务。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000002,'游戏开发','implementer','pro',315100000000000003,'你是游戏开发工程师，擅长游戏逻辑编码、图形渲染、物理引擎集成。按照架构师方案完成代码实现。','aider',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000003,'游戏开发','implementer','max',315100000000000003,'你是高级游戏开发工程师，擅长复杂游戏系统（战斗系统、网络同步、AI行为树）。处理高难度编码任务。','aider',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000004,'游戏开发','implementer','lite',315100000000000006,'你是初级游戏开发工程师，负责简单的配置文件修改、资源集成、UI布局等基础任务。','aider',0,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000005,'游戏开发','auditor','pro',315100000000000007,'你是游戏QA工程师，擅长代码审查、性能分析、Bug检测。审核实现者的代码质量和游戏逻辑正确性。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000006,'游戏开发','coordinator','lite',315100000000000004,'你是游戏项目协调员，负责跟踪进度、协调资源冲突、汇总状态。','chat',1,1,6,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000011,'小说创作','architect','max',315100000000000001,'你是资深小说策划，擅长世界观构建、人物塑造、故事结构设计。负责规划小说整体框架、章节大纲、人物关系图。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000012,'小说创作','implementer','pro',315100000000000003,'你是小说作者，擅长叙事写作、对话编排、场景描写。按照大纲完成各章节的创作。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000013,'小说创作','implementer','max',315100000000000003,'你是高级小说作者，擅长复杂剧情转折、多线叙事、深层人物心理描写。处理关键章节和高潮段落。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000014,'小说创作','auditor','pro',315100000000000007,'你是小说编辑，擅长审稿、一致性检查、文风校对。审核章节内容是否符合大纲、人物是否一致、有无逻辑漏洞。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000015,'小说创作','coordinator','lite',315100000000000004,'你是内容协调员，负责跟踪各章节创作进度、检查衔接一致性、汇总问题。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000021,'动漫创作','architect','max',315100000000000001,'你是动漫企划总监，擅长IP策划、角色设计、世界观构建、分镜脚本编排。负责规划动漫项目整体框架和创意方向。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000022,'动漫创作','implementer','pro',315100000000000003,'你是动漫内容创作者，擅长脚本撰写、角色台词、场景描述、分镜文案。按照企划方案完成各集内容。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000023,'动漫创作','implementer','max',315100000000000003,'你是高级动漫编剧，擅长复杂世界观叙事、情感高潮编排、跨集剧情铺垫。处理核心剧情章节。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000024,'动漫创作','auditor','pro',315100000000000007,'你是动漫内容审核，擅长剧情一致性检查、角色行为逻辑审核、风格统一性校验。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000025,'动漫创作','coordinator','lite',315100000000000004,'你是动漫项目协调员，负责跟踪各集制作进度、协调创意冲突。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000031,'漫剧创作','architect','max',315100000000000001,'你是漫剧总编剧，擅长短剧剧本结构、节奏把控、情节反转设计。负责规划漫剧整体故事线和分集大纲。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000032,'漫剧创作','implementer','pro',315100000000000003,'你是漫剧编剧，擅长对话编排、场景描写、情绪节奏把控。按照大纲完成各集剧本创作。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000033,'漫剧创作','implementer','max',315100000000000003,'你是高级漫剧编剧，擅长复杂情节设计、多角色冲突编排、高潮段落创作。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000034,'漫剧创作','auditor','pro',315100000000000007,'你是漫剧审稿，擅长剧本逻辑检查、台词质量审核、节奏合理性评估。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000035,'漫剧创作','coordinator','lite',315100000000000004,'你是漫剧项目协调员，负责跟踪各集创作进度、保证剧情连贯性。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000041,'大电影创作','architect','max',315100000000000001,'你是电影编剧总监，擅长三幕式结构、角色弧光设计、主题提炼。负责规划电影整体剧本框架和场景规划。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000042,'大电影创作','implementer','pro',315100000000000003,'你是电影编剧，擅长场景撰写、对白创作、动作描写。按照剧本大纲完成各场景的详细剧本。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000043,'大电影创作','implementer','max',315100000000000003,'你是高级电影编剧，擅长复杂叙事结构、非线性时间线、深层主题表达。处理关键场景和高潮段落。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000044,'大电影创作','auditor','pro',315100000000000007,'你是电影剧本审读，擅长结构分析、节奏评估、台词打磨、逻辑一致性检查。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000045,'大电影创作','coordinator','lite',315100000000000004,'你是电影项目协调员，负责跟踪剧本各部分进度、协调场景衔接。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000051,'动画创作','architect','max',315100000000000001,'你是动画导演，擅长动画叙事、视觉风格设定、分镜规划。负责规划动画项目整体创意方向和制作流程。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000052,'动画创作','implementer','pro',315100000000000003,'你是动画脚本创作者，擅长分镜脚本、角色动作描述、场景设定文案。按照导演方案完成创作。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000053,'动画创作','implementer','max',315100000000000003,'你是高级动画编剧，擅长复杂动画叙事、视觉节奏设计、情感渲染技巧。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000054,'动画创作','auditor','pro',315100000000000007,'你是动画内容审核，擅长脚本一致性检查、视觉描述准确性、动画节奏评估。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000055,'动画创作','coordinator','lite',315100000000000004,'你是动画项目协调员，负责跟踪各集制作进度、协调创作团队。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000061,'数据分析','architect','max',315100000000000001,'你是数据分析架构师，擅长分析方案设计、数据管道规划、指标体系构建。负责设计分析流程和任务编排。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000062,'数据分析','implementer','pro',315100000000000003,'你是数据分析师，擅长数据清洗、统计分析、可视化、报告撰写。按照分析方案执行具体的数据分析任务。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000063,'数据分析','implementer','max',315100000000000003,'你是高级数据科学家，擅长复杂模型构建、深度分析、趋势预测。处理高难度分析任务。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000064,'数据分析','auditor','pro',315100000000000007,'你是数据分析审核，擅长数据质量检查、分析方法验证、结论合理性评估。审核分析结果的准确性。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000065,'数据分析','coordinator','lite',315100000000000004,'你是数据项目协调员，负责跟踪分析进度、协调数据源接入、汇总分析结果。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000071,'产品设计','architect','max',315100000000000001,'你是产品设计总监，擅长用户研究、产品策略、信息架构设计。负责规划产品设计的整体方案和任务编排。','chat',1,1,1,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000072,'产品设计','implementer','pro',315100000000000003,'你是产品设计师，擅长交互设计、原型设计、用户流程编排。按照设计方案完成具体的设计输出。','chat',1,1,2,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000073,'产品设计','implementer','max',315100000000000003,'你是高级产品设计师，擅长复杂系统设计、设计系统构建、创新交互模式。处理核心设计挑战。','chat',0,1,3,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000074,'产品设计','auditor','pro',315100000000000007,'你是设计评审专家，擅长可用性评估、设计规范检查、一致性审核。审核设计产出的质量和可行性。','chat',1,1,4,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000075,'产品设计','coordinator','lite',315100000000000004,'你是产品协调员，负责跟踪设计进度、协调需求变更、汇总设计问题。','chat',1,1,5,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL),(316200000000000081,'软件开发','coordinator','max',315100000000000004,'你是高级项目协调员，擅长复杂项目管理、跨团队协调、风险评估。处理关键路径冲突和资源调度。','chat',0,1,12,NULL,NULL,'2026-04-04 15:32:49','2026-04-04 15:32:49',NULL);
/*!40000 ALTER TABLE `mvp_role_preset` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `mvp_config` WRITE;
/*!40000 ALTER TABLE `mvp_config` DISABLE KEYS */;
INSERT INTO `mvp_config` (`id`, `config_key`, `config_value`, `config_type`, `category`, `description`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,'watchdog.check_interval','120','int','watchdog','心跳检测间隔（秒）',0,0,'2026-04-03 16:31:01','2026-04-03 16:31:01',NULL),(2,'watchdog.max_stale_count','3','int','watchdog','连续无进展次数阈值（超过则判定卡死）',0,0,'2026-04-03 16:31:01','2026-04-03 16:31:01',NULL),(3,'watchdog.max_retries','3','int','watchdog','最大自动重试次数（超过则升级给架构师）',0,0,'2026-04-03 16:31:01','2026-04-03 16:31:01',NULL),(4,'scheduler.max_concurrent','20','int','scheduler','最大并发任务数',0,0,'2026-04-03 16:31:01','2026-04-03 16:31:01',NULL),(5,'scheduler.poll_interval','2','int','scheduler','调度轮询间隔（秒）',0,0,'2026-04-03 16:31:01','2026-04-03 16:31:01',NULL),(6,'review.timeout_seconds','300','int','engine','方案审核阶段超时时间（秒），超时跳过AI审核',0,0,'2026-04-04 16:04:26','2026-04-04 16:04:26',NULL),(7,'review.auto_fix_batch','1','int','engine','预检时是否自动修正batch_no不合理的问题（1=是）',0,0,'2026-04-04 16:04:26','2026-04-04 16:04:26',NULL);
/*!40000 ALTER TABLE `mvp_config` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `mvp_accept_rule` WRITE;
/*!40000 ALTER TABLE `mvp_accept_rule` DISABLE KEYS */;
INSERT INTO `mvp_accept_rule` (`id`, `project_type`, `rule_code`, `rule_name`, `rule_type`, `scope_type`, `config_json`, `enabled`, `priority`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,'software_dev','software.no_failed_tasks','不得存在失败任务','process','project','{\"forbid_status\":[\"failed\",\"escalated\"]}',1,10,'2026-04-05 16:57:08','2026-04-05 16:57:08',NULL),(2,'software_dev','software.required_file_exists','关键文件必须存在','artifact','file','{\"required_files\":[\"README.md\"]}',1,20,'2026-04-05 16:57:08','2026-04-05 16:57:08',NULL),(3,'software_dev','software.output_not_empty','关键输出不得为空','artifact','task','{\"task_kinds\":[\"implement\",\"refactor\",\"fix\"],\"require_non_empty_result\":true}',1,30,'2026-04-05 16:57:08','2026-04-05 16:57:08',NULL),(4,'document','document.required_output_exists','文档产物必须存在','artifact','file','{\"required_extensions\":[\".md\",\".docx\"]}',1,10,'2026-04-05 16:57:08','2026-04-05 16:57:08',NULL),(5,'document','document.summary_present','必须生成总结','process','stage','{\"required_stage_outputs\":[\"execute\"]}',1,20,'2026-04-05 16:57:08','2026-04-05 16:57:08',NULL);
/*!40000 ALTER TABLE `mvp_accept_rule` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `mvp_project_category` WRITE;
/*!40000 ALTER TABLE `mvp_project_category` DISABLE KEYS */;
INSERT INTO `mvp_project_category` (`id`, `category_code`, `display_name`, `family_code`, `description`, `status`, `sort`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (317100000000000001,'software_dev','软件开发','coding','标准软件开发项目',1,10,0,0,'2026-04-05 20:45:46','2026-04-05 20:45:46',NULL),(317100000000000002,'game_dev','游戏开发','coding','游戏开发项目',1,20,0,0,'2026-04-05 20:45:46','2026-04-05 20:45:46',NULL),(317100000000000003,'novel_writing','小说创作','creative','小说创作项目',1,30,0,0,'2026-04-05 20:45:46','2026-04-05 20:45:46',NULL),(317100000000000004,'animation_writing','动漫创作','creative','动漫创作项目',1,40,0,0,'2026-04-05 20:45:46','2026-04-05 20:45:46',NULL),(317100000000000005,'comic_drama_writing','漫剧创作','creative','漫剧创作项目',1,50,0,0,'2026-04-05 20:45:46','2026-04-05 20:45:46',NULL),(317100000000000006,'movie_writing','大电影创作','creative','大电影创作项目',1,60,0,0,'2026-04-05 20:45:46','2026-04-05 20:45:46',NULL),(317100000000000007,'animation_project','动画创作','creative','动画创作项目',1,70,0,0,'2026-04-05 20:45:46','2026-04-05 20:45:46',NULL),(317100000000000008,'data_analysis','数据分析','analysis','数据分析项目',1,80,0,0,'2026-04-05 20:45:46','2026-04-05 20:45:46',NULL),(317100000000000009,'product_design','产品设计','analysis','产品设计项目',1,90,0,0,'2026-04-05 20:45:46','2026-04-05 20:45:46',NULL);
/*!40000 ALTER TABLE `mvp_project_category` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `system_role` WRITE;
/*!40000 ALTER TABLE `system_role` DISABLE KEYS */;
INSERT INTO `system_role` (`id`, `parent_id`, `title`, `data_scope`, `is_admin`, `sort`, `status`, `default_ai_engine`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (1000000000000000002,0,'超级管理员',1,1,0,1,'openhands',0,1000000000000000001,'2026-03-30 21:20:22','2026-04-01 11:55:43',NULL);
/*!40000 ALTER TABLE `system_role` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `system_menu` WRITE;
/*!40000 ALTER TABLE `system_menu` DISABLE KEYS */;
INSERT INTO `system_menu` (`id`, `parent_id`, `title`, `type`, `path`, `component`, `permission`, `icon`, `sort`, `is_show`, `is_cache`, `link_url`, `status`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (314253730209861632,0,'上传管理',1,'/upload',NULL,'','CloudUploadOutlined',50,1,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730235027456,314253730209861632,'上传配置',2,'/upload/config','upload/config/index','upload:config:list','',0,1,0,NULL,0,0,0,'2026-03-31 10:25:27','2026-04-02 18:06:51',NULL),(314253730251804672,314253730235027456,'上传配置新增',3,NULL,NULL,'upload:config:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730268581888,314253730235027456,'上传配置修改',3,NULL,NULL,'upload:config:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730285359104,314253730235027456,'上传配置删除',3,NULL,NULL,'upload:config:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730344079360,314253730209861632,'文件目录',2,'/upload/dir','upload/dir/index','upload:dir:list','',0,1,0,NULL,0,0,0,'2026-03-31 10:25:27','2026-04-02 18:06:51',NULL),(314253730365050880,314253730344079360,'文件目录新增',3,NULL,NULL,'upload:dir:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730386022400,314253730344079360,'文件目录修改',3,NULL,NULL,'upload:dir:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730415382528,314253730344079360,'文件目录删除',3,NULL,NULL,'upload:dir:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730461519872,314253730209861632,'文件目录规则',2,'/upload/dir-rule','upload/dir_rule/index','upload:dir_rule:list','',0,1,0,NULL,0,0,0,'2026-03-31 10:25:27','2026-04-02 18:06:51',NULL),(314253730478297088,314253730461519872,'文件目录规则新增',3,NULL,NULL,'upload:dir_rule:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730490880000,314253730461519872,'文件目录规则修改',3,NULL,NULL,'upload:dir_rule:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730503462912,314253730461519872,'文件目录规则删除',3,NULL,NULL,'upload:dir_rule:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730566377472,314253730209861632,'文件记录',2,'/upload/file','upload/file/index','upload:file:list','',0,1,0,NULL,0,0,0,'2026-03-31 10:25:27','2026-04-02 18:06:51',NULL),(314253730583154688,314253730566377472,'文件记录新增',3,NULL,NULL,'upload:file:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730595737600,314253730566377472,'文件记录修改',3,NULL,NULL,'upload:file:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253730620903424,314253730566377472,'文件记录删除',3,NULL,NULL,'upload:file:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:27','2026-03-31 10:25:27',NULL),(314253751944744960,0,'陪玩管理',1,'/play',NULL,'','game-icons:joystick',50,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253751965716480,314253751944744960,'活动',2,'/play/activity','play/activity/index','play:activity:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253751986688000,314253751965716480,'活动新增',3,NULL,NULL,'play:activity:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752007659520,314253751965716480,'活动修改',3,NULL,NULL,'play:activity:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752024436736,314253751965716480,'活动删除',3,NULL,NULL,'play:activity:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752108322816,314253751944744960,'活动参与记录',2,'/play/activity-join','play/activity_join/index','play:activity_join:list','',0,0,0,'',1,0,0,'2026-03-31 10:25:33','2026-03-31 06:55:37','2026-04-02 17:02:19'),(314253752120905728,314253752108322816,'活动参与记录新增',3,NULL,NULL,'play:activity_join:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752141877248,314253752108322816,'活动参与记录修改',3,NULL,NULL,'play:activity_join:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752154460160,314253752108322816,'活动参与记录删除',3,NULL,NULL,'play:activity_join:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752196403200,314253751944744960,'活动奖励',2,'/play/activity-reward','play/activity_reward/index','play:activity_reward:list','',0,0,0,'',1,0,0,'2026-03-31 10:25:33','2026-03-31 06:55:52','2026-04-02 17:02:19'),(314253752208986112,314253752196403200,'活动奖励新增',3,NULL,NULL,'play:activity_reward:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752225763328,314253752196403200,'活动奖励修改',3,NULL,NULL,'play:activity_reward:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752238346240,314253752196403200,'活动奖励删除',3,NULL,NULL,'play:activity_reward:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752284483584,314253751944744960,'活动步骤',2,'/play/activity-step','play/activity_step/index','play:activity_step:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253752297066496,314253752284483584,'活动步骤新增',3,NULL,NULL,'play:activity_step:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752318038016,314253752284483584,'活动步骤修改',3,NULL,NULL,'play:activity_step:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752339009536,314253752284483584,'活动步骤删除',3,NULL,NULL,'play:activity_step:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752435478528,314253751944744960,'余额流水',2,'/play/balance-log','play/balance_log/index','play:balance_log:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253752460644352,314253752435478528,'余额流水新增',3,NULL,NULL,'play:balance_log:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752477421568,314253752435478528,'余额流水修改',3,NULL,NULL,'play:balance_log:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752490004480,314253752435478528,'余额流水删除',3,NULL,NULL,'play:balance_log:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752536141824,314253751944744960,'商品分类',2,'/play/category','play/category/index','play:category:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253752557113344,314253752536141824,'商品分类新增',3,NULL,NULL,'play:category:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752569696256,314253752536141824,'商品分类修改',3,NULL,NULL,'play:category:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752586473472,314253752536141824,'商品分类删除',3,NULL,NULL,'play:category:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752661970944,314253751944744960,'陪玩师',2,'/play/coach','play/coach/index','play:coach:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253752682942464,314253752661970944,'陪玩师新增',3,NULL,NULL,'play:coach:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752699719680,314253752661970944,'陪玩师修改',3,NULL,NULL,'play:coach:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752716496896,314253752661970944,'陪玩师删除',3,NULL,NULL,'play:coach:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752783605760,314253751944744960,'陪玩师申请',2,'/play/coach-apply','play/coach_apply/index','play:coach_apply:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253752800382976,314253752783605760,'陪玩师申请新增',3,NULL,NULL,'play:coach_apply:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752821354496,314253752783605760,'陪玩师申请修改',3,NULL,NULL,'play:coach_apply:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752842326016,314253752783605760,'陪玩师申请删除',3,NULL,NULL,'play:coach_apply:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752884269056,314253751944744960,'陪玩师等级',2,'/play/coach-level','play/coach_level/index','play:coach_level:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253752901046272,314253752884269056,'陪玩师等级新增',3,NULL,NULL,'play:coach_level:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752922017792,314253752884269056,'陪玩师等级修改',3,NULL,NULL,'play:coach_level:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752951377920,314253752884269056,'陪玩师等级删除',3,NULL,NULL,'play:coach_level:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253752989126656,314253751944744960,'优惠券模板',2,'/play/coupon','play/coupon/index','play:coupon:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253753001709568,314253752989126656,'优惠券模板新增',3,NULL,NULL,'play:coupon:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753018486784,314253752989126656,'优惠券模板修改',3,NULL,NULL,'play:coupon:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753035264000,314253752989126656,'优惠券模板删除',3,NULL,NULL,'play:coupon:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753123344384,314253751944744960,'会员优惠券',2,'/play/coupon-member','play/coupon_member/index','play:coupon_member:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253753140121600,314253753123344384,'会员优惠券新增',3,NULL,NULL,'play:coupon_member:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753152704512,314253753123344384,'会员优惠券修改',3,NULL,NULL,'play:coupon_member:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753173676032,314253753123344384,'会员优惠券删除',3,NULL,NULL,'play:coupon_member:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753244979200,314253751944744960,'商品',2,'/play/goods','play/goods/index','play:goods:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253753261756416,314253753244979200,'商品新增',3,NULL,NULL,'play:goods:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753282727936,314253753244979200,'商品修改',3,NULL,NULL,'play:goods:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753299505152,314253753244979200,'商品删除',3,NULL,NULL,'play:goods:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753349836800,314253751944744960,'会员',2,'/play/member','play/member/index','play:member:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253753362419712,314253753349836800,'会员新增',3,NULL,NULL,'play:member:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753383391232,314253753349836800,'会员修改',3,NULL,NULL,'play:member:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753400168448,314253753349836800,'会员删除',3,NULL,NULL,'play:member:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753437917184,314253751944744960,'会员等级',2,'/play/member-level','play/member_level/index','play:member_level:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253753463083008,314253753437917184,'会员等级新增',3,NULL,NULL,'play:member_level:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753479860224,314253753437917184,'会员等级修改',3,NULL,NULL,'play:member_level:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753496637440,314253753437917184,'会员等级删除',3,NULL,NULL,'play:member_level:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753609883648,314253751944744960,'第三方登录绑定',2,'/play/oauth','play/oauth/index','play:oauth:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253753626660864,314253753609883648,'第三方登录绑定新增',3,NULL,NULL,'play:oauth:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753651826688,314253753609883648,'第三方登录绑定修改',3,NULL,NULL,'play:oauth:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753668603904,314253753609883648,'第三方登录绑定删除',3,NULL,NULL,'play:oauth:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753798627328,314253751944744960,'订单',2,'/play/order','play/order/index','play:order:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253753815404544,314253753798627328,'订单新增',3,NULL,NULL,'play:order:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753832181760,314253753798627328,'订单修改',3,NULL,NULL,'play:order:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753848958976,314253753798627328,'订单删除',3,NULL,NULL,'play:order:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753945427968,314253751944744960,'支付记录',2,'/play/payment','play/payment/index','play:payment:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253753966399488,314253753945427968,'支付记录新增',3,NULL,NULL,'play:payment:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753983176704,314253753945427968,'支付记录修改',3,NULL,NULL,'play:payment:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253753999953920,314253753945427968,'支付记录删除',3,NULL,NULL,'play:payment:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754096422912,314253751944744960,'利润分成流水',2,'/play/profit-log','play/profit_log/index','play:profit_log:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253754109005824,314253754096422912,'利润分成流水新增',3,NULL,NULL,'play:profit_log:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754125783040,314253754096422912,'利润分成流水修改',3,NULL,NULL,'play:profit_log:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754146754560,314253754096422912,'利润分成流水删除',3,NULL,NULL,'play:profit_log:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754222252032,314253751944744960,'充值订单',2,'/play/recharge-order','play/recharge_order/index','play:recharge_order:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253754243223552,314253754222252032,'充值订单新增',3,NULL,NULL,'play:recharge_order:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754255806464,314253754222252032,'充值订单修改',3,NULL,NULL,'play:recharge_order:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754272583680,314253754222252032,'充值订单删除',3,NULL,NULL,'play:recharge_order:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754318721024,314253751944744960,'充值方案',2,'/play/recharge-plan','play/recharge_plan/index','play:recharge_plan:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253754331303936,314253754318721024,'充值方案新增',3,NULL,NULL,'play:recharge_plan:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754343886848,314253754318721024,'充值方案修改',3,NULL,NULL,'play:recharge_plan:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754356469760,314253754318721024,'充值方案删除',3,NULL,NULL,'play:recharge_plan:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754465521664,314253751944744960,'评价',2,'/play/review','play/review/index','play:review:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253754478104576,314253754465521664,'评价新增',3,NULL,NULL,'play:review:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754494881792,314253754465521664,'评价修改',3,NULL,NULL,'play:review:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754507464704,314253754465521664,'评价删除',3,NULL,NULL,'play:review:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754557796352,314253751944744960,'店铺',2,'/play/shop','play/shop/index','play:shop:list','',0,1,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:19'),(314253754570379264,314253754557796352,'店铺新增',3,NULL,NULL,'play:shop:create','',1,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754591350784,314253754557796352,'店铺修改',3,NULL,NULL,'play:shop:update','',2,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314253754603933696,314253754557796352,'店铺删除',3,NULL,NULL,'play:shop:delete','',3,0,0,NULL,1,0,0,'2026-03-31 10:25:33','2026-03-31 10:25:33','2026-04-02 17:02:30'),(314616957221474304,314253751944744960,'活动步骤提交记录',2,'/play/activity-step-log','play/activity_step_log/index','play:activity_step_log:list','',0,1,0,NULL,1,0,0,'2026-04-01 10:28:47','2026-04-01 10:28:47','2026-04-02 17:02:19'),(314616957238251520,314616957221474304,'活动步骤提交记录新增',3,NULL,NULL,'play:activity_step_log:create','',1,0,0,NULL,1,0,0,'2026-04-01 10:28:47','2026-04-01 10:28:47','2026-04-02 17:02:30'),(314616957271805952,314616957221474304,'活动步骤提交记录修改',3,NULL,NULL,'play:activity_step_log:update','',2,0,0,NULL,1,0,0,'2026-04-01 10:28:47','2026-04-01 10:28:47','2026-04-02 17:02:30'),(314616957284388864,314616957221474304,'活动步骤提交记录删除',3,NULL,NULL,'play:activity_step_log:delete','',3,0,0,NULL,1,0,0,'2026-04-01 10:28:47','2026-04-01 10:28:47','2026-04-02 17:02:30'),(314738628100100096,314253751944744960,'首页Banner轮播',2,'/play/banner','play/banner/index','play:banner:list','',0,1,0,NULL,1,0,0,'2026-04-01 18:32:16','2026-04-01 18:45:26','2026-04-02 17:02:19'),(314738628116877312,314738628100100096,'首页Banner轮播新增',3,NULL,NULL,'play:banner:create','',1,0,0,NULL,1,0,0,'2026-04-01 18:32:16','2026-04-01 18:45:26','2026-04-02 17:02:30'),(314738628125265920,314738628100100096,'首页Banner轮播修改',3,NULL,NULL,'play:banner:update','',2,0,0,NULL,1,0,0,'2026-04-01 18:32:16','2026-04-01 18:45:26','2026-04-02 17:02:30'),(314738628133654528,314738628100100096,'首页Banner轮播删除',3,NULL,NULL,'play:banner:delete','',3,0,0,NULL,1,0,0,'2026-04-01 18:32:16','2026-04-01 18:45:26','2026-04-02 17:02:30'),(314738628142043136,314738628100100096,'首页Banner轮播查看',3,NULL,NULL,'play:banner:detail','',4,0,0,NULL,1,0,0,'2026-04-01 18:32:16','2026-04-01 18:45:26','2026-04-02 17:02:30'),(314738628150431744,314738628100100096,'首页Banner轮播导出',3,NULL,NULL,'play:banner:export','',5,0,0,NULL,1,0,0,'2026-04-01 18:32:16','2026-04-01 18:45:26','2026-04-02 17:02:30'),(314897283948744704,314253751944744960,'会员消息',2,'/play/message','play/message/index','play:message:list','',0,1,0,NULL,1,0,0,'2026-04-02 05:02:43','2026-04-02 05:02:43','2026-04-02 17:02:19'),(314897283969716224,314897283948744704,'会员消息新增',3,NULL,NULL,'play:message:create','',1,0,0,NULL,1,0,0,'2026-04-02 05:02:43','2026-04-02 05:02:43','2026-04-02 17:02:30'),(314897283982299136,314897283948744704,'会员消息修改',3,NULL,NULL,'play:message:update','',2,0,0,NULL,1,0,0,'2026-04-02 05:02:43','2026-04-02 05:02:43','2026-04-02 17:02:30'),(314897283994882048,314897283948744704,'会员消息删除',3,NULL,NULL,'play:message:delete','',3,0,0,NULL,1,0,0,'2026-04-02 05:02:43','2026-04-02 05:02:43','2026-04-02 17:02:30'),(314897284003270656,314897283948744704,'会员消息批量删除',3,NULL,NULL,'play:message:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 05:02:43','2026-04-02 05:02:43','2026-04-02 17:02:30'),(314897284007464960,314897283948744704,'会员消息查看',3,NULL,NULL,'play:message:detail','',5,0,0,NULL,1,0,0,'2026-04-02 05:02:43','2026-04-02 05:02:43','2026-04-02 17:02:30'),(314897284020047872,314897283948744704,'会员消息导出',3,NULL,NULL,'play:message:export','',6,0,0,NULL,1,0,0,'2026-04-02 05:02:43','2026-04-02 05:02:43','2026-04-02 17:02:30'),(314897400193880064,314253751944744960,'陪玩师提现记录',2,'/play/withdraw','play/withdraw/index','play:withdraw:list','',0,1,0,NULL,1,0,0,'2026-04-02 05:03:10','2026-04-02 05:03:10','2026-04-02 17:02:19'),(314897400206462976,314897400193880064,'陪玩师提现记录新增',3,NULL,NULL,'play:withdraw:create','',1,0,0,NULL,1,0,0,'2026-04-02 05:03:10','2026-04-02 05:03:10','2026-04-02 17:02:30'),(314897400219045888,314897400193880064,'陪玩师提现记录修改',3,NULL,NULL,'play:withdraw:update','',2,0,0,NULL,1,0,0,'2026-04-02 05:03:10','2026-04-02 05:03:10','2026-04-02 17:02:30'),(314897400227434496,314897400193880064,'陪玩师提现记录删除',3,NULL,NULL,'play:withdraw:delete','',3,0,0,NULL,1,0,0,'2026-04-02 05:03:10','2026-04-02 05:03:10','2026-04-02 17:02:30'),(314897400235823104,314897400193880064,'陪玩师提现记录批量删除',3,NULL,NULL,'play:withdraw:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 05:03:10','2026-04-02 05:03:10','2026-04-02 17:02:30'),(314897400244211712,314897400193880064,'陪玩师提现记录查看',3,NULL,NULL,'play:withdraw:detail','',5,0,0,NULL,1,0,0,'2026-04-02 05:03:10','2026-04-02 05:03:10','2026-04-02 17:02:30'),(314897400256794624,314897400193880064,'陪玩师提现记录导出',3,NULL,NULL,'play:withdraw:export','',6,0,0,NULL,1,0,0,'2026-04-02 05:03:10','2026-04-02 05:03:10','2026-04-02 17:02:30'),(315012657751003136,0,'AI模型管理',1,'/ai',NULL,'','RobotOutlined',20,1,0,NULL,1,0,0,'2026-04-02 12:41:10','2026-04-02 12:41:10',NULL),(315012657767780352,315012657751003136,'AI供应商',2,'/ai/provider','ai/provider/index','ai:provider:list','',0,1,0,NULL,1,0,0,'2026-04-02 12:41:10','2026-04-02 12:41:10',NULL),(315012657780363264,315012657767780352,'AI供应商新增',3,NULL,NULL,'ai:provider:create','',1,0,0,NULL,1,0,0,'2026-04-02 12:41:10','2026-04-02 12:41:10',NULL),(315012657788751872,315012657767780352,'AI供应商修改',3,NULL,NULL,'ai:provider:update','',2,0,0,NULL,1,0,0,'2026-04-02 12:41:10','2026-04-02 12:41:10',NULL),(315012657805529088,315012657767780352,'AI供应商删除',3,NULL,NULL,'ai:provider:delete','',3,0,0,NULL,1,0,0,'2026-04-02 12:41:10','2026-04-02 12:41:10',NULL),(315012657809723392,315012657767780352,'AI供应商批量删除',3,NULL,NULL,'ai:provider:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 12:41:10','2026-04-02 12:41:10',NULL),(315012657822306304,315012657767780352,'AI供应商查看',3,NULL,NULL,'ai:provider:detail','',5,0,0,NULL,1,0,0,'2026-04-02 12:41:10','2026-04-02 12:41:10',NULL),(315012657830694912,315012657767780352,'AI供应商导出',3,NULL,NULL,'ai:provider:export','',6,0,0,NULL,1,0,0,'2026-04-02 12:41:10','2026-04-02 12:41:10',NULL),(315012657839083520,315012657767780352,'AI供应商导入',3,NULL,NULL,'ai:provider:import','',7,0,0,NULL,1,0,0,'2026-04-02 12:41:10','2026-04-02 12:41:10',NULL),(315012657847472128,315012657767780352,'AI供应商批量编辑',3,NULL,NULL,'ai:provider:batch-update','',8,0,0,NULL,1,0,0,'2026-04-02 12:41:10','2026-04-02 12:41:10',NULL),(315012674029096960,315012657751003136,'AI套餐',2,'/ai/plan','ai/plan/index','ai:plan:list','',0,1,0,NULL,1,0,0,'2026-04-02 12:41:14','2026-04-02 12:41:14',NULL),(315012674041679872,315012674029096960,'AI套餐新增',3,NULL,NULL,'ai:plan:create','',1,0,0,NULL,1,0,0,'2026-04-02 12:41:14','2026-04-02 12:41:14',NULL),(315012674050068480,315012674029096960,'AI套餐修改',3,NULL,NULL,'ai:plan:update','',2,0,0,NULL,1,0,0,'2026-04-02 12:41:14','2026-04-02 12:41:14',NULL),(315012674062651392,315012674029096960,'AI套餐删除',3,NULL,NULL,'ai:plan:delete','',3,0,0,NULL,1,0,0,'2026-04-02 12:41:14','2026-04-02 12:41:14',NULL),(315012674071040000,315012674029096960,'AI套餐批量删除',3,NULL,NULL,'ai:plan:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 12:41:14','2026-04-02 12:41:14',NULL),(315012674092011520,315012674029096960,'AI套餐查看',3,NULL,NULL,'ai:plan:detail','',5,0,0,NULL,1,0,0,'2026-04-02 12:41:14','2026-04-02 12:41:14',NULL),(315012674100400128,315012674029096960,'AI套餐导出',3,NULL,NULL,'ai:plan:export','',6,0,0,NULL,1,0,0,'2026-04-02 12:41:14','2026-04-02 12:41:14',NULL),(315012674117177344,315012674029096960,'AI套餐导入',3,NULL,NULL,'ai:plan:import','',7,0,0,NULL,1,0,0,'2026-04-02 12:41:14','2026-04-02 12:41:14',NULL),(315012674129760256,315012674029096960,'AI套餐批量编辑',3,NULL,NULL,'ai:plan:batch-update','',8,0,0,NULL,1,0,0,'2026-04-02 12:41:14','2026-04-02 12:41:14',NULL),(315012677640392704,315012657751003136,'AI模型',2,'/ai/model','ai/model/index','ai:model:list','',0,1,0,NULL,1,0,0,'2026-04-02 12:41:15','2026-04-02 12:41:15',NULL),(315012677661364224,315012677640392704,'AI模型新增',3,NULL,NULL,'ai:model:create','',1,0,0,NULL,1,0,0,'2026-04-02 12:41:15','2026-04-02 12:41:15',NULL),(315012677669752832,315012677640392704,'AI模型修改',3,NULL,NULL,'ai:model:update','',2,0,0,NULL,1,0,0,'2026-04-02 12:41:15','2026-04-02 12:41:15',NULL),(315012677678141440,315012677640392704,'AI模型删除',3,NULL,NULL,'ai:model:delete','',3,0,0,NULL,1,0,0,'2026-04-02 12:41:15','2026-04-02 12:41:15',NULL),(315012677686530048,315012677640392704,'AI模型批量删除',3,NULL,NULL,'ai:model:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 12:41:15','2026-04-02 12:41:15',NULL),(315012677694918656,315012677640392704,'AI模型查看',3,NULL,NULL,'ai:model:detail','',5,0,0,NULL,1,0,0,'2026-04-02 12:41:15','2026-04-02 12:41:15',NULL),(315012677703307264,315012677640392704,'AI模型导出',3,NULL,NULL,'ai:model:export','',6,0,0,NULL,1,0,0,'2026-04-02 12:41:15','2026-04-02 12:41:15',NULL),(315012677711695872,315012677640392704,'AI模型导入',3,NULL,NULL,'ai:model:import','',7,0,0,NULL,1,0,0,'2026-04-02 12:41:15','2026-04-02 12:41:15',NULL),(315012677715890176,315012677640392704,'AI模型批量编辑',3,NULL,NULL,'ai:model:batch-update','',8,0,0,NULL,1,0,0,'2026-04-02 12:41:15','2026-04-02 12:41:15',NULL),(315013060379021312,0,'MVP项目管理',1,'/mvp',NULL,'','ProjectOutlined',30,1,0,NULL,1,0,0,'2026-04-02 12:42:46','2026-04-02 12:42:46',NULL),(315013060395798528,315013060379021312,'MVP项目',2,'/mvp/project','mvp/project/index','mvp:project:list','',0,1,0,NULL,1,0,0,'2026-04-02 12:42:46','2026-04-02 12:42:46',NULL),(315013060408381440,315013060395798528,'MVP项目新增',3,NULL,NULL,'mvp:project:create','',1,0,0,NULL,1,0,0,'2026-04-02 12:42:46','2026-04-02 12:42:46',NULL),(315013060416770048,315013060395798528,'MVP项目修改',3,NULL,NULL,'mvp:project:update','',2,0,0,NULL,1,0,0,'2026-04-02 12:42:46','2026-04-02 12:42:46',NULL),(315013060425158656,315013060395798528,'MVP项目删除',3,NULL,NULL,'mvp:project:delete','',3,0,0,NULL,1,0,0,'2026-04-02 12:42:46','2026-04-02 12:42:46',NULL),(315013060437741568,315013060395798528,'MVP项目批量删除',3,NULL,NULL,'mvp:project:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 12:42:46','2026-04-02 12:42:46',NULL),(315013060446130176,315013060395798528,'MVP项目查看',3,NULL,NULL,'mvp:project:detail','',5,0,0,NULL,1,0,0,'2026-04-02 12:42:46','2026-04-02 12:42:46',NULL),(315013060454518784,315013060395798528,'MVP项目导出',3,NULL,NULL,'mvp:project:export','',6,0,0,NULL,1,0,0,'2026-04-02 12:42:46','2026-04-02 12:42:46',NULL),(315013060458713088,315013060395798528,'MVP项目导入',3,NULL,NULL,'mvp:project:import','',7,0,0,NULL,1,0,0,'2026-04-02 12:42:46','2026-04-02 12:42:46',NULL),(315013060467101696,315013060395798528,'MVP项目批量编辑',3,NULL,NULL,'mvp:project:batch-update','',8,0,0,NULL,1,0,0,'2026-04-02 12:42:46','2026-04-02 12:42:46',NULL),(315013075654676480,315013060379021312,'MVP任务',2,'/mvp/task','mvp/task/index','mvp:task:list','',0,1,0,NULL,1,0,0,'2026-04-02 12:42:49','2026-04-02 12:42:49',NULL),(315013075671453696,315013075654676480,'MVP任务新增',3,NULL,NULL,'mvp:task:create','',1,0,0,NULL,1,0,0,'2026-04-02 12:42:49','2026-04-02 12:42:49',NULL),(315013075684036608,315013075654676480,'MVP任务修改',3,NULL,NULL,'mvp:task:update','',2,0,0,NULL,1,0,0,'2026-04-02 12:42:49','2026-04-02 12:42:49',NULL),(315013075696619520,315013075654676480,'MVP任务删除',3,NULL,NULL,'mvp:task:delete','',3,0,0,NULL,1,0,0,'2026-04-02 12:42:49','2026-04-02 12:42:49',NULL),(315013075705008128,315013075654676480,'MVP任务批量删除',3,NULL,NULL,'mvp:task:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 12:42:49','2026-04-02 12:42:49',NULL),(315013075730173952,315013075654676480,'MVP任务查看',3,NULL,NULL,'mvp:task:detail','',5,0,0,NULL,1,0,0,'2026-04-02 12:42:49','2026-04-02 12:42:49',NULL),(315013075742756864,315013075654676480,'MVP任务导出',3,NULL,NULL,'mvp:task:export','',6,0,0,NULL,1,0,0,'2026-04-02 12:42:49','2026-04-02 12:42:49',NULL),(315013075751145472,315013075654676480,'MVP任务导入',3,NULL,NULL,'mvp:task:import','',7,0,0,NULL,1,0,0,'2026-04-02 12:42:49','2026-04-02 12:42:49',NULL),(315013075763728384,315013075654676480,'MVP任务批量编辑',3,NULL,NULL,'mvp:task:batch-update','',8,0,0,NULL,1,0,0,'2026-04-02 12:42:49','2026-04-02 12:42:49',NULL),(315013078984953856,315013060379021312,'MVP对话',2,'/mvp/conversation','mvp/conversation/index','mvp:conversation:list','',0,1,0,NULL,1,0,0,'2026-04-02 12:42:50','2026-04-02 12:42:50',NULL),(315013078997536768,315013078984953856,'MVP对话新增',3,NULL,NULL,'mvp:conversation:create','',1,0,0,NULL,1,0,0,'2026-04-02 12:42:50','2026-04-02 12:42:50',NULL),(315013079005925376,315013078984953856,'MVP对话修改',3,NULL,NULL,'mvp:conversation:update','',2,0,0,NULL,1,0,0,'2026-04-02 12:42:50','2026-04-02 12:42:50',NULL),(315013079010119680,315013078984953856,'MVP对话删除',3,NULL,NULL,'mvp:conversation:delete','',3,0,0,NULL,1,0,0,'2026-04-02 12:42:50','2026-04-02 12:42:50',NULL),(315013079022702592,315013078984953856,'MVP对话批量删除',3,NULL,NULL,'mvp:conversation:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 12:42:50','2026-04-02 12:42:50',NULL),(315013079031091200,315013078984953856,'MVP对话查看',3,NULL,NULL,'mvp:conversation:detail','',5,0,0,NULL,1,0,0,'2026-04-02 12:42:50','2026-04-02 12:42:50',NULL),(315013079043674112,315013078984953856,'MVP对话导出',3,NULL,NULL,'mvp:conversation:export','',6,0,0,NULL,1,0,0,'2026-04-02 12:42:50','2026-04-02 12:42:50',NULL),(315013079052062720,315013078984953856,'MVP对话导入',3,NULL,NULL,'mvp:conversation:import','',7,0,0,NULL,1,0,0,'2026-04-02 12:42:50','2026-04-02 12:42:50',NULL),(315013079064645632,315013078984953856,'MVP对话批量编辑',3,NULL,NULL,'mvp:conversation:batch-update','',8,0,0,NULL,1,0,0,'2026-04-02 12:42:50','2026-04-02 12:42:50',NULL),(315013082873073664,315013060379021312,'MVP消息',2,'/mvp/message','mvp/message/index','mvp:message:list','',0,1,0,NULL,1,0,0,'2026-04-02 12:42:51','2026-04-02 12:42:51',NULL),(315013082881462272,315013082873073664,'MVP消息新增',3,NULL,NULL,'mvp:message:create','',1,0,0,NULL,1,0,0,'2026-04-02 12:42:51','2026-04-02 12:42:51',NULL),(315013082898239488,315013082873073664,'MVP消息修改',3,NULL,NULL,'mvp:message:update','',2,0,0,NULL,1,0,0,'2026-04-02 12:42:51','2026-04-02 12:42:51',NULL),(315013082906628096,315013082873073664,'MVP消息删除',3,NULL,NULL,'mvp:message:delete','',3,0,0,NULL,1,0,0,'2026-04-02 12:42:51','2026-04-02 12:42:51',NULL),(315013082910822400,315013082873073664,'MVP消息批量删除',3,NULL,NULL,'mvp:message:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 12:42:51','2026-04-02 12:42:51',NULL),(315013082919211008,315013082873073664,'MVP消息查看',3,NULL,NULL,'mvp:message:detail','',5,0,0,NULL,1,0,0,'2026-04-02 12:42:51','2026-04-02 12:42:51',NULL),(315013082927599616,315013082873073664,'MVP消息导出',3,NULL,NULL,'mvp:message:export','',6,0,0,NULL,1,0,0,'2026-04-02 12:42:51','2026-04-02 12:42:51',NULL),(315013082935988224,315013082873073664,'MVP消息导入',3,NULL,NULL,'mvp:message:import','',7,0,0,NULL,1,0,0,'2026-04-02 12:42:51','2026-04-02 12:42:51',NULL),(315013082940182528,315013082873073664,'MVP消息批量编辑',3,NULL,NULL,'mvp:message:batch-update','',8,0,0,NULL,1,0,0,'2026-04-02 12:42:51','2026-04-02 12:42:51',NULL),(315013085918138368,315013060379021312,'项目角色配置',2,'/mvp/project-role','mvp/project_role/index','mvp:project_role:list','',0,1,0,NULL,1,0,0,'2026-04-02 12:42:52','2026-04-02 12:42:52',NULL),(315013085926526976,315013085918138368,'项目角色配置新增',3,NULL,NULL,'mvp:project_role:create','',1,0,0,NULL,1,0,0,'2026-04-02 12:42:52','2026-04-02 12:42:52',NULL),(315013085934915584,315013085918138368,'项目角色配置修改',3,NULL,NULL,'mvp:project_role:update','',2,0,0,NULL,1,0,0,'2026-04-02 12:42:52','2026-04-02 12:42:52',NULL),(315013085947498496,315013085918138368,'项目角色配置删除',3,NULL,NULL,'mvp:project_role:delete','',3,0,0,NULL,1,0,0,'2026-04-02 12:42:52','2026-04-02 12:42:52',NULL),(315013085955887104,315013085918138368,'项目角色配置批量删除',3,NULL,NULL,'mvp:project_role:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 12:42:52','2026-04-02 12:42:52',NULL),(315013085964275712,315013085918138368,'项目角色配置查看',3,NULL,NULL,'mvp:project_role:detail','',5,0,0,NULL,1,0,0,'2026-04-02 12:42:52','2026-04-02 12:42:52',NULL),(315013085972664320,315013085918138368,'项目角色配置导出',3,NULL,NULL,'mvp:project_role:export','',6,0,0,NULL,1,0,0,'2026-04-02 12:42:52','2026-04-02 12:42:52',NULL),(315013085989441536,315013085918138368,'项目角色配置导入',3,NULL,NULL,'mvp:project_role:import','',7,0,0,NULL,1,0,0,'2026-04-02 12:42:52','2026-04-02 12:42:52',NULL),(315013085997830144,315013085918138368,'项目角色配置批量编辑',3,NULL,NULL,'mvp:project_role:batch-update','',8,0,0,NULL,1,0,0,'2026-04-02 12:42:52','2026-04-02 12:42:52',NULL),(315013088828985344,315013060379021312,'任务日志',2,'/mvp/task-log','mvp/task_log/index','mvp:task_log:list','',0,1,0,NULL,1,0,0,'2026-04-02 12:42:53','2026-04-02 12:42:53',NULL),(315013088837373952,315013088828985344,'任务日志新增',3,NULL,NULL,'mvp:task_log:create','',1,0,0,NULL,1,0,0,'2026-04-02 12:42:53','2026-04-02 12:42:53',NULL),(315013088849956864,315013088828985344,'任务日志修改',3,NULL,NULL,'mvp:task_log:update','',2,0,0,NULL,1,0,0,'2026-04-02 12:42:53','2026-04-02 12:42:53',NULL),(315013088862539776,315013088828985344,'任务日志删除',3,NULL,NULL,'mvp:task_log:delete','',3,0,0,NULL,1,0,0,'2026-04-02 12:42:53','2026-04-02 12:42:53',NULL),(315013088870928384,315013088828985344,'任务日志批量删除',3,NULL,NULL,'mvp:task_log:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 12:42:53','2026-04-02 12:42:53',NULL),(315013088879316992,315013088828985344,'任务日志查看',3,NULL,NULL,'mvp:task_log:detail','',5,0,0,NULL,1,0,0,'2026-04-02 12:42:53','2026-04-02 12:42:53',NULL),(315013088887705600,315013088828985344,'任务日志导出',3,NULL,NULL,'mvp:task_log:export','',6,0,0,NULL,1,0,0,'2026-04-02 12:42:53','2026-04-02 12:42:53',NULL),(315013088896094208,315013088828985344,'任务日志导入',3,NULL,NULL,'mvp:task_log:import','',7,0,0,NULL,1,0,0,'2026-04-02 12:42:53','2026-04-02 12:42:53',NULL),(315013088904482816,315013088828985344,'任务日志批量编辑',3,NULL,NULL,'mvp:task_log:batch-update','',8,0,0,NULL,1,0,0,'2026-04-02 12:42:53','2026-04-02 12:42:53',NULL),(315013100000000001,315013060379021312,'AI对话',2,'/mvp/chat','mvp/chat/index',NULL,'lucide:message-square',10,0,0,NULL,1,NULL,NULL,'2026-04-02 13:29:19','2026-04-02 13:29:19',NULL),(315020000000000001,315012657751003136,'执行引擎配置',2,'/ai/engine','ai/engine/index','ai:engine:list','',1,1,0,NULL,1,0,0,'2026-04-03 11:02:29','2026-04-03 11:02:29',NULL),(315020000000000002,315020000000000001,'执行引擎查看',3,NULL,NULL,'ai:engine:detail','',1,0,0,NULL,1,0,0,'2026-04-03 11:02:29','2026-04-03 11:02:29',NULL),(315020000000000003,315020000000000001,'执行引擎修改',3,NULL,NULL,'ai:engine:update','',2,0,0,NULL,1,0,0,'2026-04-03 11:02:30','2026-04-03 11:02:30',NULL),(315020000000000004,315020000000000001,'执行引擎测试',3,NULL,NULL,'ai:engine:test-connection','',3,0,0,NULL,1,0,0,'2026-04-03 11:02:30','2026-04-03 11:02:30',NULL),(315020000000000011,315012657751003136,'执行任务',2,'/ai/task','ai/task/index','ai:task:list','',2,1,0,NULL,1,0,0,'2026-04-03 11:02:30','2026-04-03 11:02:30',NULL),(315020000000000012,315020000000000011,'执行任务创建',3,NULL,NULL,'ai:task:execute','',1,0,0,NULL,1,0,0,'2026-04-03 11:02:30','2026-04-03 11:02:30',NULL),(315020000000000013,315020000000000011,'执行任务查看',3,NULL,NULL,'ai:task:detail','',2,0,0,NULL,1,0,0,'2026-04-03 11:02:31','2026-04-03 11:02:31',NULL),(315020000000000014,315020000000000011,'执行任务日志',3,NULL,NULL,'ai:task:logs','',3,0,0,NULL,1,0,0,'2026-04-03 11:02:31','2026-04-03 11:02:31',NULL),(315020000000000015,315020000000000011,'执行任务取消',3,NULL,NULL,'ai:task:cancel','',4,0,0,NULL,1,0,0,'2026-04-03 11:02:31','2026-04-03 11:02:31',NULL),(315072919958982656,315013060379021312,'角色预设模板',2,'/mvp/role-preset','mvp/role_preset/index','mvp:role_preset:list','',0,1,0,NULL,1,0,0,'2026-04-02 16:40:37','2026-04-02 16:40:37',NULL),(315072919979954176,315072919958982656,'角色预设模板新增',3,NULL,NULL,'mvp:role_preset:create','',1,0,0,NULL,1,0,0,'2026-04-02 16:40:37','2026-04-02 16:40:37',NULL),(315072919992537088,315072919958982656,'角色预设模板修改',3,NULL,NULL,'mvp:role_preset:update','',2,0,0,NULL,1,0,0,'2026-04-02 16:40:37','2026-04-02 16:40:37',NULL),(315072920000925696,315072919958982656,'角色预设模板删除',3,NULL,NULL,'mvp:role_preset:delete','',3,0,0,NULL,1,0,0,'2026-04-02 16:40:37','2026-04-02 16:40:37',NULL),(315072920009314304,315072919958982656,'角色预设模板批量删除',3,NULL,NULL,'mvp:role_preset:batch-delete','',4,0,0,NULL,1,0,0,'2026-04-02 16:40:37','2026-04-02 16:40:37',NULL),(315072920021897216,315072919958982656,'角色预设模板查看',3,NULL,NULL,'mvp:role_preset:detail','',5,0,0,NULL,1,0,0,'2026-04-02 16:40:37','2026-04-02 16:40:37',NULL),(315072920034480128,315072919958982656,'角色预设模板导出',3,NULL,NULL,'mvp:role_preset:export','',6,0,0,NULL,1,0,0,'2026-04-02 16:40:37','2026-04-02 16:40:37',NULL),(315072920042868736,315072919958982656,'角色预设模板导入',3,NULL,NULL,'mvp:role_preset:import','',7,0,0,NULL,1,0,0,'2026-04-02 16:40:37','2026-04-02 16:40:37',NULL),(315072920051257344,315072919958982656,'角色预设模板批量编辑',3,NULL,NULL,'mvp:role_preset:batch-update','',8,0,0,NULL,1,0,0,'2026-04-02 16:40:37','2026-04-02 16:40:37',NULL),(315200100000000001,315013060379021312,'系统概览',2,'/mvp/dashboard','mvp/dashboard/index',NULL,'lucide:layout-dashboard',-1,1,0,NULL,1,1,0,'2026-04-03 17:08:37','2026-04-03 17:08:37',NULL),(315200100000000002,315013060379021312,'配置管理',2,'/mvp/config','mvp/config/index','mvp:config:list','lucide:settings-2',11,1,0,NULL,1,1,0,'2026-04-05 17:28:29','2026-04-05 17:28:29',NULL),(315300100000000001,315013060379021312,'工作流仪表板',1,'/mvp/workflow/dashboard','mvp/workflow/dashboard',NULL,NULL,20,0,0,NULL,1,NULL,NULL,'2026-04-05 18:26:18','2026-04-05 18:26:18',NULL),(315300100000000002,315013060379021312,'审核工作台',1,'/mvp/workflow/review','mvp/workflow/review',NULL,NULL,21,0,0,NULL,1,NULL,NULL,'2026-04-05 18:26:18','2026-04-05 18:26:18',NULL),(315300100000000003,315013060379021312,'执行控制台',1,'/mvp/workflow/execution','mvp/workflow/execution',NULL,NULL,22,0,0,NULL,1,NULL,NULL,'2026-04-05 18:26:18','2026-04-05 18:26:18',NULL),(315300100000000004,315013060379021312,'返工记录',1,'/mvp/workflow/rework','mvp/workflow/rework',NULL,NULL,23,0,0,NULL,1,NULL,NULL,'2026-04-05 18:26:18','2026-04-05 18:26:18',NULL),(315300100000000005,315013060379021312,'事件时间线',1,'/mvp/workflow/timeline','mvp/workflow/timeline',NULL,NULL,24,0,0,NULL,1,NULL,NULL,'2026-04-05 18:26:18','2026-04-05 18:26:18',NULL),(315300100000000006,315013060379021312,'验收控制台',1,'/mvp/workflow/accept','mvp/workflow/accept',NULL,NULL,25,0,0,NULL,1,NULL,NULL,'2026-04-05 18:26:18','2026-04-05 18:26:18',NULL),(315300100000000007,315013060379021312,'自治控制台',1,'/mvp/workflow/autonomy','mvp/workflow/autonomy',NULL,NULL,26,0,0,NULL,1,NULL,NULL,'2026-04-05 18:55:38','2026-04-05 18:55:38',NULL),(1000000000000000010,0,'系统管理',1,'/system',NULL,'','SettingOutlined',100,1,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000011,1000000000000000010,'部门管理',2,'/system/dept','system/dept/index','system:dept:list','ApartmentOutlined',1,1,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000012,1000000000000000010,'角色管理',2,'/system/role','system/role/index','system:role:list','TeamOutlined',2,1,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000013,1000000000000000010,'菜单管理',2,'/system/menu','system/menu/index','system:menu:list','MenuOutlined',3,1,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000014,1000000000000000010,'用户管理',2,'/system/users','system/users/index','system:user:list','UserOutlined',4,1,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000021,1000000000000000011,'部门新增',3,NULL,NULL,'system:dept:create','',1,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000022,1000000000000000011,'部门修改',3,NULL,NULL,'system:dept:update','',2,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000023,1000000000000000011,'部门删除',3,NULL,NULL,'system:dept:delete','',3,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000031,1000000000000000012,'角色新增',3,NULL,NULL,'system:role:create','',1,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000032,1000000000000000012,'角色修改',3,NULL,NULL,'system:role:update','',2,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000033,1000000000000000012,'角色删除',3,NULL,NULL,'system:role:delete','',3,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000034,1000000000000000012,'资源授权',3,NULL,NULL,'system:role:grant:menu','',4,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000035,1000000000000000012,'数据授权',3,NULL,NULL,'system:role:grant:dept','',5,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000041,1000000000000000013,'菜单新增',3,NULL,NULL,'system:menu:create','',1,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000042,1000000000000000013,'菜单修改',3,NULL,NULL,'system:menu:update','',2,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000043,1000000000000000013,'菜单删除',3,NULL,NULL,'system:menu:delete','',3,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000051,1000000000000000014,'用户新增',3,NULL,NULL,'system:user:create','',1,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000052,1000000000000000014,'用户修改',3,NULL,NULL,'system:user:update','',2,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000053,1000000000000000014,'用户删除',3,NULL,NULL,'system:user:delete','',3,0,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000060,0,'仪表盘',1,'/dashboard',NULL,'','DashboardOutlined',0,1,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000061,1000000000000000060,'分析页',2,'/analytics','dashboard/analytics/index','','AreaChartOutlined',1,1,1,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL),(1000000000000000062,1000000000000000060,'工作台',2,'/workspace','dashboard/workspace/index','','DesktopOutlined',2,1,0,NULL,1,0,1000000000000000001,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL);
/*!40000 ALTER TABLE `system_menu` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `system_dept` WRITE;
/*!40000 ALTER TABLE `system_dept` DISABLE KEYS */;
INSERT INTO `system_dept` (`id`, `parent_id`, `title`, `username`, `email`, `sort`, `status`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (1000000000000000001,0,'总公司','admin','admin@example.com',0,1,0,0,'2026-03-30 21:20:22','2026-03-30 21:20:22',NULL);
/*!40000 ALTER TABLE `system_dept` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `system_role_menu` WRITE;
/*!40000 ALTER TABLE `system_role_menu` DISABLE KEYS */;
INSERT INTO `system_role_menu` (`role_id`, `menu_id`) VALUES (1000000000000000002,314253730209861632),(1000000000000000002,314253730235027456),(1000000000000000002,314253730251804672),(1000000000000000002,314253730268581888),(1000000000000000002,314253730285359104),(1000000000000000002,314253730344079360),(1000000000000000002,314253730365050880),(1000000000000000002,314253730386022400),(1000000000000000002,314253730415382528),(1000000000000000002,314253730461519872),(1000000000000000002,314253730478297088),(1000000000000000002,314253730490880000),(1000000000000000002,314253730503462912),(1000000000000000002,314253730566377472),(1000000000000000002,314253730583154688),(1000000000000000002,314253730595737600),(1000000000000000002,314253730620903424),(1000000000000000002,314253751944744960),(1000000000000000002,314253751965716480),(1000000000000000002,314253751986688000),(1000000000000000002,314253752007659520),(1000000000000000002,314253752024436736),(1000000000000000002,314253752108322816),(1000000000000000002,314253752120905728),(1000000000000000002,314253752141877248),(1000000000000000002,314253752154460160),(1000000000000000002,314253752196403200),(1000000000000000002,314253752208986112),(1000000000000000002,314253752225763328),(1000000000000000002,314253752238346240),(1000000000000000002,314253752284483584),(1000000000000000002,314253752297066496),(1000000000000000002,314253752318038016),(1000000000000000002,314253752339009536),(1000000000000000002,314253752435478528),(1000000000000000002,314253752460644352),(1000000000000000002,314253752477421568),(1000000000000000002,314253752490004480),(1000000000000000002,314253752536141824),(1000000000000000002,314253752557113344),(1000000000000000002,314253752569696256),(1000000000000000002,314253752586473472),(1000000000000000002,314253752661970944),(1000000000000000002,314253752682942464),(1000000000000000002,314253752699719680),(1000000000000000002,314253752716496896),(1000000000000000002,314253752783605760),(1000000000000000002,314253752800382976),(1000000000000000002,314253752821354496),(1000000000000000002,314253752842326016),(1000000000000000002,314253752884269056),(1000000000000000002,314253752901046272),(1000000000000000002,314253752922017792),(1000000000000000002,314253752951377920),(1000000000000000002,314253752989126656),(1000000000000000002,314253753001709568),(1000000000000000002,314253753018486784),(1000000000000000002,314253753035264000),(1000000000000000002,314253753123344384),(1000000000000000002,314253753140121600),(1000000000000000002,314253753152704512),(1000000000000000002,314253753173676032),(1000000000000000002,314253753244979200),(1000000000000000002,314253753261756416),(1000000000000000002,314253753282727936),(1000000000000000002,314253753299505152),(1000000000000000002,314253753349836800),(1000000000000000002,314253753362419712),(1000000000000000002,314253753383391232),(1000000000000000002,314253753400168448),(1000000000000000002,314253753437917184),(1000000000000000002,314253753463083008),(1000000000000000002,314253753479860224),(1000000000000000002,314253753496637440),(1000000000000000002,314253753609883648),(1000000000000000002,314253753626660864),(1000000000000000002,314253753651826688),(1000000000000000002,314253753668603904),(1000000000000000002,314253753798627328),(1000000000000000002,314253753815404544),(1000000000000000002,314253753832181760),(1000000000000000002,314253753848958976),(1000000000000000002,314253753945427968),(1000000000000000002,314253753966399488),(1000000000000000002,314253753983176704),(1000000000000000002,314253753999953920),(1000000000000000002,314253754096422912),(1000000000000000002,314253754109005824),(1000000000000000002,314253754125783040),(1000000000000000002,314253754146754560),(1000000000000000002,314253754222252032),(1000000000000000002,314253754243223552),(1000000000000000002,314253754255806464),(1000000000000000002,314253754272583680),(1000000000000000002,314253754318721024),(1000000000000000002,314253754331303936),(1000000000000000002,314253754343886848),(1000000000000000002,314253754356469760),(1000000000000000002,314253754465521664),(1000000000000000002,314253754478104576),(1000000000000000002,314253754494881792),(1000000000000000002,314253754507464704),(1000000000000000002,314253754557796352),(1000000000000000002,314253754570379264),(1000000000000000002,314253754591350784),(1000000000000000002,314253754603933696),(1000000000000000002,315300100000000001),(1000000000000000002,315300100000000002),(1000000000000000002,315300100000000003),(1000000000000000002,315300100000000004),(1000000000000000002,315300100000000005),(1000000000000000002,315300100000000006),(1000000000000000002,315300100000000007),(1000000000000000002,1000000000000000010),(1000000000000000002,1000000000000000011),(1000000000000000002,1000000000000000012),(1000000000000000002,1000000000000000013),(1000000000000000002,1000000000000000014),(1000000000000000002,1000000000000000021),(1000000000000000002,1000000000000000022),(1000000000000000002,1000000000000000023),(1000000000000000002,1000000000000000031),(1000000000000000002,1000000000000000032),(1000000000000000002,1000000000000000033),(1000000000000000002,1000000000000000034),(1000000000000000002,1000000000000000035),(1000000000000000002,1000000000000000041),(1000000000000000002,1000000000000000042),(1000000000000000002,1000000000000000043),(1000000000000000002,1000000000000000051),(1000000000000000002,1000000000000000052),(1000000000000000002,1000000000000000053),(1000000000000000002,1000000000000000060),(1000000000000000002,1000000000000000061),(1000000000000000002,1000000000000000062);
/*!40000 ALTER TABLE `system_role_menu` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `system_role_dept` WRITE;
/*!40000 ALTER TABLE `system_role_dept` DISABLE KEYS */;
/*!40000 ALTER TABLE `system_role_dept` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `system_user_role` WRITE;
/*!40000 ALTER TABLE `system_user_role` DISABLE KEYS */;
INSERT INTO `system_user_role` (`user_id`, `role_id`) VALUES (1000000000000000003,1000000000000000002);
/*!40000 ALTER TABLE `system_user_role` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `system_user_dept` WRITE;
/*!40000 ALTER TABLE `system_user_dept` DISABLE KEYS */;
INSERT INTO `system_user_dept` (`user_id`, `dept_id`) VALUES (1000000000000000003,1000000000000000001);
/*!40000 ALTER TABLE `system_user_dept` ENABLE KEYS */;
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

LOCK TABLES `ai_engine` WRITE;
/*!40000 ALTER TABLE `ai_engine` DISABLE KEYS */;
INSERT INTO `ai_engine` (`id`, `code`, `name`, `description`, `status`, `sort`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (1000000000000001001,'aider','Aider','本地命令行代码执行引擎',1,10,1,0,'2026-04-03 10:54:32','2026-04-03 10:54:32',NULL),(1000000000000001002,'openhands','OpenHands','远程Agent执行引擎',1,20,1,0,'2026-04-03 10:54:32','2026-04-03 10:54:32',NULL);
/*!40000 ALTER TABLE `ai_engine` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `ai_engine_config` WRITE;
/*!40000 ALTER TABLE `ai_engine_config` DISABLE KEYS */;
INSERT INTO `ai_engine_config` (`id`, `engine_code`, `base_url`, `api_key`, `default_model_id`, `timeout_seconds`, `max_steps`, `workspace_root`, `command_template`, `callback_url`, `callback_secret`, `extra_config`, `status`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`) VALUES (315200000000000001,'aider',NULL,NULL,315100000000000001,900,20,'C:\\project\\easymvp',NULL,NULL,NULL,'{}',1,1,0,'2026-04-03 12:00:19','2026-04-03 12:00:19',NULL),(315200000000000002,'openhands',NULL,NULL,315100000000000001,1800,20,'C:\\project\\easymvp','docker run --rm -v \"${env:AI_TASK_WORKTREE_PATH}:/workspace\" -w /workspace -e LLM_API_KEY=\"${env:AI_MODEL_API_KEY}\" -e LLM_MODEL=\"anthropic/${env:AI_MODEL_CODE}\" -e LLM_BASE_URL=\"$(${env:AI_MODEL_BASE_URL} -replace \'/v1$\',\'\')\" -e OPENHANDS_SUPPRESS_BANNER=\"1\" -e SANDBOX_VOLUMES=\"/workspace:/workspace:rw\" -e PIP_INDEX_URL=\"https://mirrors.aliyun.com/pypi/simple/\" -e UV_DEFAULT_INDEX=\"https://mirrors.aliyun.com/pypi/simple/\" -e UV_NO_PROGRESS=\"true\" docker.1panel.live/library/python:3.12-slim bash -lc \"python -m pip install -q uv && uv tool run --python 3.12 openhands --headless --json --override-with-envs --always-approve --exit-without-confirmation -t \'$env:AI_TASK_INSTRUCTION\'\"',NULL,NULL,'{\"runner\": \"docker-cli\"}',1,1,0,'2026-04-03 12:00:19','2026-04-03 13:53:58',NULL);
/*!40000 ALTER TABLE `ai_engine_config` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

