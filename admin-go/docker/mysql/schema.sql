/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_provider` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '供应商名称',
  `code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '供应商代码：openai/anthropic/deepseek/qwen/doubao/ernie/spark/glm/moonshot/yi/google/ollama',
  `provider_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '供应商主类型/默认路由类型',
  `supported_protocols` json DEFAULT NULL COMMENT '支持的协议类型(JSON)：anthropic/openai_compatible/google 等',
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_accept_evidence` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `accept_run_id` bigint NOT NULL COMMENT '验收运行ID',
  `evidence_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'task_output/file/log/diff/stage_output/handoff/summary',
  `source_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'domain_task/stage_run/file/handoff_record/workflow_run',
  `source_id` bigint DEFAULT NULL COMMENT '来源对象ID',
  `content_ref` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '证据引用或JSON',
  `summary` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '证据摘要',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_accept_run_id` (`accept_run_id`),
  KEY `idx_evidence_type` (`evidence_type`),
  KEY `idx_source_type_source_id` (`source_type`,`source_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动验收证据';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_accept_issue` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `accept_run_id` bigint NOT NULL COMMENT '验收运行ID',
  `workflow_run_id` bigint NOT NULL COMMENT '工作流运行ID',
  `project_id` bigint NOT NULL COMMENT '项目ID',
  `domain_task_id` bigint DEFAULT NULL COMMENT '主关联任务ID',
  `issue_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'artifact/process/quality/risk',
  `rule_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '规则编码',
  `severity` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'info/warn/error/blocker',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '问题标题',
  `detail` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '问题详情',
  `expected_value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '预期值',
  `actual_value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '实际值',
  `suggested_action` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '建议动作',
  `resource_ref` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '关联资源引用(JSON)',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'open' COMMENT 'open/resolved/ignored',
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_accept_rule` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `project_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '项目类型模板',
  `rule_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则编码',
  `rule_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则名称',
  `rule_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'artifact/process/quality',
  `scope_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'project/task/file/stage',
  `config_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则配置',
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_accept_run` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `workflow_run_id` bigint NOT NULL COMMENT '工作流运行ID',
  `stage_run_id` bigint NOT NULL COMMENT 'accept阶段stage_run_id',
  `project_id` bigint NOT NULL COMMENT '项目ID',
  `plan_version_id` bigint DEFAULT NULL COMMENT '关联方案版本ID',
  `accept_round` int NOT NULL DEFAULT '1' COMMENT '第几轮验收',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT 'pending/running/completed/failed/canceled',
  `decision` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'passed/failed/manual_review',
  `score` decimal(5,2) DEFAULT NULL COMMENT '验收评分',
  `summary` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '验收摘要',
  `rules_version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '规则版本号',
  `rules_snapshot_ref` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '规则快照引用或JSON',
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_action_outcome` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `action_id` bigint unsigned NOT NULL COMMENT '关联 mvp_decision_action.id',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '关联 workflow_run',
  `project_id` bigint unsigned NOT NULL COMMENT '关联 project',
  `strategy_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '策略名称',
  `action_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '动作类型',
  `decision_level` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '决策级别(A/B/C)',
  `sit_before` json DEFAULT NULL COMMENT '执行前态势摘要(JSON)',
  `sit_after` json DEFAULT NULL COMMENT '执行后态势摘要(JSON)',
  `effective` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'unknown' COMMENT '效果: positive/negative/neutral/unknown',
  `effect_score` decimal(5,3) NOT NULL DEFAULT '0.000' COMMENT '效果得分(-1~1)',
  `eval_delay_ms` bigint NOT NULL DEFAULT '0' COMMENT '评估延迟(毫秒)',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_action` (`action_id`),
  KEY `idx_workflow_run` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_strategy` (`strategy_name`),
  KEY `idx_effective` (`effective`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='策略效果跟踪';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_assessment_result` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `project_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '关联 project(0=全局)',
  `period_start` datetime NOT NULL COMMENT '评估周期开始',
  `period_end` datetime NOT NULL COMMENT '评估周期结束',
  `sample_count` int NOT NULL DEFAULT '0' COMMENT '样本数量',
  `policy_accuracy` decimal(5,3) NOT NULL DEFAULT '0.000' COMMENT '策略准确率(0~1)',
  `gate_false_positive` decimal(5,3) NOT NULL DEFAULT '0.000' COMMENT '闸门误报率(0~1)',
  `gate_false_negative` decimal(5,3) NOT NULL DEFAULT '0.000' COMMENT '闸门漏报率(0~1)',
  `human_override_rate` decimal(5,3) NOT NULL DEFAULT '0.000' COMMENT '人工干预率(0~1)',
  `match_accuracy` decimal(5,3) NOT NULL DEFAULT '0.000' COMMENT '匹配准确率(0~1)',
  `cost_efficiency` decimal(5,3) NOT NULL DEFAULT '0.000' COMMENT '成本效率(0~1)',
  `drifts` json DEFAULT NULL COMMENT '参数偏差列表(JSON)',
  `summary` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT '评估摘要',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_period` (`period_start`,`period_end`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='系统评估结果';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_autonomy_decision` (
  `id` bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned NOT NULL,
  `decision_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'replan/risk_escalate/engine_switch/circuit_break/report',
  `trigger_source` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '触发源：watchdog/accept/rework/scheduler/manual',
  `trigger_context` json DEFAULT NULL COMMENT '触发上下文',
  `recommendation` json NOT NULL COMMENT '系统建议',
  `decision_mode` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'suggest' COMMENT 'suggest/auto',
  `human_action` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT 'approved/rejected/modified/pending',
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
  KEY `idx_type` (`decision_type`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='自治决策记录';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `config_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置键（唯一）',
  `config_value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置值',
  `config_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'string' COMMENT '值类型：string/int/float/bool/json',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'general' COMMENT '分类：engine/watchdog/scheduler/general',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '配置说明',
  `created_by` bigint unsigned DEFAULT '0' COMMENT '创建人',
  `dept_id` bigint unsigned DEFAULT '0' COMMENT '部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`),
  KEY `idx_category` (`category`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=315200000000000101 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MVP配置表';
/*!40101 SET character_set_client = @saved_cs_client */;
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
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_datascope` (`dept_id`,`created_by`),
  KEY `idx_project_role` (`project_id`,`role_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MVP对话表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_decision_action` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned NOT NULL,
  `stage_run_id` bigint unsigned DEFAULT NULL COMMENT '关联阶段运行ID',
  `domain_task_id` bigint unsigned DEFAULT NULL COMMENT '关联领域任务ID',
  `decision_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '动作类型',
  `decision_level` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '决策等级: A/B/C',
  `trigger_source` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '触发源事件类型',
  `trigger_context` json DEFAULT NULL COMMENT '触发上下文',
  `matched_rule_id` bigint unsigned DEFAULT NULL COMMENT '匹配的策略规则ID',
  `matched_gate_ids` json DEFAULT NULL COMMENT '命中的闸门ID列表',
  `action_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '策略匹配的动作类型(闸门降级后为fallback)',
  `recommendation` json DEFAULT NULL COMMENT '系统建议',
  `final_action` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '最终实际执行的动作',
  `action_status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'pending' COMMENT 'pending/auto_executed/waiting_human/approved/rejected/failed',
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
  `task_kind` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务种类: implement/audit/bug_analysis/failure_analysis',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务名称',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务描述',
  `role_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色类型',
  `role_level` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色等级',
  `execution_mode` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行方式: chat/aider/openhands',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态: pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched',
  `conversation_id` bigint unsigned DEFAULT NULL COMMENT '关联对话ID',
  `model_id` bigint unsigned DEFAULT NULL COMMENT '使用的AI模型ID',
  `batch_no` int NOT NULL COMMENT '批次号',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `retry_count` int NOT NULL DEFAULT '0' COMMENT '重试次数',
  `affected_resources` json DEFAULT NULL COMMENT '影响资源列表(JSON)',
  `locked_resources` json DEFAULT NULL COMMENT '锁定资源列表(JSON)',
  `result` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '执行结果',
  `context_summary` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '上下文摘要',
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
  KEY `idx_source_task` (`source_task_id`),
  KEY `idx_datascope` (`dept_id`,`created_by`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='领域任务';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_handoff_record` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `from_task_id` bigint unsigned DEFAULT NULL COMMENT '来源任务ID',
  `to_task_id` bigint unsigned DEFAULT NULL COMMENT '目标任务ID',
  `handoff_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '交接类型: bug_fix/failure_escalation/rework/audit',
  `reason` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '交接原因',
  `payload` json DEFAULT NULL COMMENT '交接载荷(JSON)',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_type` (`workflow_run_id`,`handoff_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交接记录';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_human_checkpoint` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned NOT NULL,
  `decision_action_id` bigint unsigned NOT NULL COMMENT '关联的决策动作ID',
  `checkpoint_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '节点类型: manual_review/approval/escalation',
  `title` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '标题',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT '详细描述',
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'open' COMMENT 'open/handled/expired/canceled',
  `assigned_to` bigint unsigned DEFAULT NULL COMMENT '指派给谁',
  `handled_by` bigint unsigned DEFAULT NULL COMMENT '实际处理人',
  `handle_action` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '处理动作: approve/reject/retry/rework/override',
  `handle_reason` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT '处理理由',
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_learning_record` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `metric_key` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '指标名称(如 strategy.cost_guard.accuracy)',
  `project_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '关联 project(0=全局)',
  `ema_value` decimal(10,6) NOT NULL DEFAULT '0.000000' COMMENT 'EMA 当前值',
  `raw_value` decimal(10,6) NOT NULL DEFAULT '0.000000' COMMENT '最新原始值',
  `sample_count` int NOT NULL DEFAULT '0' COMMENT '累计样本数',
  `last_updated` datetime NOT NULL COMMENT '最后更新时间',
  `decay_factor` decimal(5,3) NOT NULL DEFAULT '0.900' COMMENT 'EMA 衰减因子',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_metric_project` (`metric_key`,`project_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_last_updated` (`last_updated`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='EMA学习记录';
/*!40101 SET character_set_client = @saved_cs_client */;
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
) ENGINE=InnoDB AUTO_INCREMENT=262469 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息分片表（流式输出）';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_observation_record` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `decision_action_id` bigint unsigned NOT NULL COMMENT '关联 mvp_decision_action.id',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '关联 workflow_run',
  `project_id` bigint unsigned NOT NULL COMMENT '关联 project',
  `decision_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '决策类型: policy_match/strategy:xxx/objective_guard',
  `trigger_source` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '触发源',
  `decision_level` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '决策级别 A/B/C',
  `action_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '动作类型',
  `input_snapshot` json DEFAULT NULL COMMENT '决策输入快照(JSON)',
  `output_snapshot` json DEFAULT NULL COMMENT '决策输出快照(JSON)',
  `meta_snapshot` json DEFAULT NULL COMMENT 'DecisionMeta 快照(JSON)',
  `outcome` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'pending' COMMENT '结果: success/failure/neutral/pending',
  `effect_score` decimal(5,3) NOT NULL DEFAULT '0.000' COMMENT '效果得分(-1~1)',
  `human_override` tinyint NOT NULL DEFAULT '0' COMMENT '人工是否干预(0否/1是)',
  `override_reason` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '人工干预原因',
  `signal_weight` decimal(3,1) NOT NULL DEFAULT '0.0' COMMENT '学习信号权重(0-1)',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_decision_action` (`decision_action_id`),
  KEY `idx_workflow_run` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_decision_type` (`decision_type`),
  KEY `idx_outcome` (`outcome`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='决策观测记录';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_plan_version` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `project_id` bigint unsigned NOT NULL COMMENT '所属项目ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `version_no` int NOT NULL COMMENT '版本号(项目内递增)',
  `source_conversation_id` bigint unsigned DEFAULT NULL COMMENT '来源对话ID',
  `source_message_id` bigint unsigned DEFAULT NULL COMMENT '来源消息ID',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '版本状态: draft/active/superseded',
  `review_status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '审核状态: pending/approved/rejected',
  `summary` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '版本摘要',
  `diff_summary` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '与上一版本的差异摘要',
  `approved_at` datetime DEFAULT NULL COMMENT '审核通过时间',
  `rejected_at` datetime DEFAULT NULL COMMENT '审核驳回时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_version` (`project_id`,`version_no`),
  KEY `idx_workflow_status` (`workflow_run_id`,`status`,`review_status`),
  KEY `idx_datascope` (`dept_id`,`created_by`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计划版本';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_policy_rule` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `rule_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '规则编码(唯一)',
  `rule_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '规则名称',
  `decision_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '决策动作类型',
  `decision_level` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '决策等级: A/B/C',
  `trigger_source` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '触发源事件类型',
  `project_family` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '适用项目家族(NULL=全局)',
  `project_category_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '适用项目分类(NULL=全局)',
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_project` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `name` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '项目名称',
  `project_category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '项目分类',
  `category_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '项目分类编码',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '项目简介',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'designing' COMMENT '项目状态: designing/reviewing/running/paused/completed',
  `pause_reason` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '暂停原因',
  `global_context` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '项目全局上下文（架构师需求分析+方案设计的压缩摘要）',
  `objective_json` json DEFAULT NULL COMMENT '项目目标约束(JSON): budget/deadline/risk_tolerance/autonomy_level',
  `architect_model_id` bigint unsigned DEFAULT NULL COMMENT '架构师使用的AI模型ID',
  `work_dir` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '项目代码工作目录（Aider执行路径）',
  `active_batch_no` int NOT NULL DEFAULT '0' COMMENT '当前活跃批次号（调度器持久化，0=无活跃批次）',
  `engine_version` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'legacy' COMMENT '执行引擎版本: legacy/workflow_v2',
  `active_workflow_run_id` bigint unsigned DEFAULT NULL COMMENT '当前活跃工作流运行ID',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_category_code` (`category_code`),
  KEY `idx_datascope` (`dept_id`,`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MVP项目表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_project_category` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `category_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '稳定分类编码',
  `display_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '展示名称',
  `family_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '能力家族编码',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分类说明',
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_project_report` (
  `id` bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned NOT NULL,
  `report_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'stage/daily/weekly/summary',
  `stage_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '阶段类型',
  `title` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'Markdown 格式报告正文',
  `metrics` json DEFAULT NULL COMMENT '关键指标快照',
  `created_by` bigint unsigned DEFAULT '0',
  `dept_id` bigint unsigned DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_type` (`report_type`),
  KEY `idx_datascope` (`dept_id`,`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='项目汇报';
/*!40101 SET character_set_client = @saved_cs_client */;
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
  `execution_mode` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'chat' COMMENT '执行方式: auto=自动选择, chat=对话模式, aider=Aider, openhands=OpenHands, claude_code=Claude Code, codex_cli=Codex CLI, gemini_cli=Gemini CLI',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:0=禁用,1=启用',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '所属部门ID',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_project_role_level` (`project_id`,`role_type`,`role_level`),
  KEY `idx_datascope` (`dept_id`,`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目角色配置表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_review_issue` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `stage_run_id` bigint unsigned NOT NULL COMMENT '所属阶段运行ID',
  `plan_version_id` bigint unsigned NOT NULL COMMENT '所属计划版本ID',
  `blueprint_id` bigint unsigned DEFAULT NULL COMMENT '关联蓝图ID',
  `severity` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '严重级别: error/warning/info',
  `issue_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '问题代码',
  `issue_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '问题类型',
  `source_role` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '发现角色',
  `task_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '关联任务名',
  `message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '问题描述',
  `suggestion` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '修复建议',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态: open/resolved/ignored',
  `resolved_at` datetime DEFAULT NULL COMMENT '解决时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  PRIMARY KEY (`id`),
  KEY `idx_plan_issue` (`plan_version_id`,`severity`,`status`),
  KEY `idx_blueprint_issue` (`blueprint_id`,`severity`,`status`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_workflow_run` (`workflow_run_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审核问题';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_risk_gate_rule` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `gate_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '闸门编码(唯一)',
  `gate_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '闸门名称',
  `gate_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '闸门类型: permission/quality/cost/runtime',
  `project_family` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '适用项目家族(NULL=全局)',
  `project_category_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '适用项目分类(NULL=全局)',
  `trigger_expression` json NOT NULL COMMENT '触发表达式(JSON规则)',
  `block_action` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '命中后禁止的动作',
  `fallback_action` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '命中后降级动作',
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_role_preset` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `project_category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '项目分类',
  `role_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色类型：architect/implementer/auditor/coordinator',
  `role_level` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '角色等级：lite/pro/max',
  `model_id` bigint unsigned NOT NULL COMMENT 'AI模型ID',
  `system_prompt` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '默认系统提示词（角色设定）',
  `execution_mode` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'chat' COMMENT '执行方式: auto=自动选择, chat=对话模式, aider=Aider, openhands=OpenHands, claude_code=Claude Code, codex_cli=Codex CLI, gemini_cli=Gemini CLI',
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_situation_snapshot` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '关联 workflow_run',
  `project_id` bigint unsigned NOT NULL COMMENT '关联 project',
  `snapshot_data` json NOT NULL COMMENT '态势数据(JSON): progress/health/resource/trend',
  `anomaly_signals` json DEFAULT NULL COMMENT '异常信号列表(JSON)',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow_run` (`workflow_run_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_workflow_deleted_id` (`workflow_run_id`,`deleted_at`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='态势快照';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_stage_run` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `stage_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '阶段类型: design/review/execute/rework/complete',
  `stage_no` int NOT NULL COMMENT '同类型阶段序号(支持多轮)',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态: pending/running/completed/failed/skipped',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  `input_ref` json DEFAULT NULL COMMENT '阶段输入引用(JSON)',
  `output_ref` json DEFAULT NULL COMMENT '阶段输出引用(JSON)',
  `decision` json DEFAULT NULL COMMENT '阶段决策结果(JSON)',
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `finished_at` datetime DEFAULT NULL COMMENT '结束时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_stage` (`workflow_run_id`,`stage_type`,`stage_no`),
  KEY `idx_workflow_status` (`workflow_run_id`,`status`),
  KEY `idx_datascope` (`dept_id`,`created_by`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='阶段运行实例';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_stage_task` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `stage_run_id` bigint unsigned NOT NULL COMMENT '所属阶段运行ID',
  `task_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务类型: precheck/auditor_review/coordinator_optimize/review_summary',
  `role_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行角色',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态: pending/running/completed/failed/skipped',
  `input_payload` json DEFAULT NULL COMMENT '输入载荷(JSON)',
  `output_payload` json DEFAULT NULL COMMENT '输出载荷(JSON)',
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_stage_type` (`stage_run_id`,`task_type`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='阶段任务';
/*!40101 SET character_set_client = @saved_cs_client */;
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
  `task_kind` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '任务记录类型：implement/audit/bug_analysis/failure_analysis',
  `source_task_id` bigint unsigned DEFAULT NULL COMMENT '直接来源任务ID，原始任务为NULL',
  `root_task_id` bigint unsigned DEFAULT NULL COMMENT '所属主链根任务ID',
  `model_id` bigint unsigned DEFAULT NULL COMMENT '使用的AI模型ID',
  `conversation_id` bigint DEFAULT NULL COMMENT '任务对话ID，用于检测任务状态',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '任务状态: draft/pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched/submit_error',
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_task_blueprint` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `plan_version_id` bigint unsigned NOT NULL COMMENT '所属计划版本ID',
  `parent_blueprint_id` bigint unsigned DEFAULT NULL COMMENT '父蓝图ID(支持层级)',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务名称',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务描述',
  `role_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色类型: architect/implementer/auditor/coordinator',
  `role_level` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色等级: lite/pro/max',
  `batch_no` int NOT NULL COMMENT '批次号',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `affected_resources` json DEFAULT NULL COMMENT '影响资源列表(JSON)',
  `depends_on_blueprint_ids` json DEFAULT NULL COMMENT '依赖蓝图ID列表(JSON)',
  `blueprint_status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '蓝图状态: draft/confirmed/superseded',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `dept_id` bigint NOT NULL DEFAULT '0' COMMENT '部门ID',
  PRIMARY KEY (`id`),
  KEY `idx_plan_batch` (`plan_version_id`,`batch_no`,`sort`),
  KEY `idx_plan_status` (`plan_version_id`,`blueprint_status`),
  KEY `idx_datascope` (`dept_id`,`created_by`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务蓝图';
/*!40101 SET character_set_client = @saved_cs_client */;
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_task_resource_lock` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `task_id` bigint unsigned NOT NULL COMMENT '持锁任务ID',
  `resource_path` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源路径',
  `lock_status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '锁状态: held/released/leaked',
  `locked_at` datetime DEFAULT NULL COMMENT '加锁时间',
  `released_at` datetime DEFAULT NULL COMMENT '释放时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_resource` (`task_id`,`resource_path`),
  KEY `idx_workflow_resource` (`workflow_run_id`,`resource_path`,`lock_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务资源锁';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_task_workspace` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `task_id` bigint unsigned NOT NULL COMMENT '任务ID(domain_task或mvp_task)',
  `workflow_run_id` bigint unsigned DEFAULT NULL COMMENT '所属工作流运行ID',
  `project_id` bigint unsigned NOT NULL COMMENT '项目ID',
  `workspace_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'git_worktree' COMMENT '工作空间类型: git_worktree',
  `workspace_path` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '工作空间绝对路径',
  `base_ref` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '基线引用(commit hash/branch)',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'creating' COMMENT '状态: creating/ready/running/completed/failed/canceled',
  `cleanup_status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '清理状态: pending/done/retained/failed',
  `diff_summary` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '变更摘要(diff统计)',
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_tune_recommendation` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `assessment_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '关联评估结果',
  `project_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '关联 project(0=全局)',
  `parameter` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '参数名',
  `current_value` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '当前值',
  `suggested_value` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '建议值',
  `direction` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'conservative/aggressive',
  `reasoning` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT '建议理由',
  `confidence` decimal(5,3) NOT NULL DEFAULT '0.000' COMMENT '置信度(0~1)',
  `auto_applicable` tinyint NOT NULL DEFAULT '0' COMMENT '是否可自动应用(0否/1是)',
  `risk_level` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'low' COMMENT 'low/medium/high',
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'pending' COMMENT 'pending/applied/rejected/expired',
  `applied_at` datetime DEFAULT NULL COMMENT '应用时间',
  `applied_by` bigint unsigned NOT NULL DEFAULT '0' COMMENT '应用人',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_assessment` (`assessment_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_parameter` (`parameter`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='调参建议';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_user_collab_binding` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `user_id` bigint unsigned NOT NULL COMMENT '关联 system_users.id',
  `platform` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'feishu' COMMENT '平台: feishu/dingtalk/wecom',
  `platform_user_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '平台用户标识(飞书 open_id)',
  `platform_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '平台显示名',
  `created_by` bigint unsigned NOT NULL DEFAULT '0',
  `dept_id` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_platform_active` (`user_id`,`platform`,`deleted_at`),
  KEY `idx_platform_user` (`platform`,`platform_user_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户协作平台绑定';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_workflow_event` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `workflow_run_id` bigint unsigned NOT NULL COMMENT '所属工作流运行ID',
  `stage_run_id` bigint unsigned DEFAULT NULL COMMENT '关联阶段运行ID',
  `entity_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '实体类型: workflow_run/stage_run/plan_version/domain_task/review_issue',
  `entity_id` bigint unsigned DEFAULT NULL COMMENT '实体ID',
  `event_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '事件类型: workflow.created/stage.started/task.completed等',
  `payload` json DEFAULT NULL COMMENT '事件载荷(JSON)',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_event` (`workflow_run_id`,`created_at`),
  KEY `idx_entity_event` (`entity_type`,`entity_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流事件';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mvp_workflow_run` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `project_id` bigint unsigned NOT NULL COMMENT '所属项目ID',
  `run_no` int NOT NULL COMMENT '项目内运行序号(从1递增)',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态: designing/reviewing/executing/reworking/paused/completed/failed/canceled',
  `tokens_consumed` bigint NOT NULL DEFAULT '0' COMMENT '已消耗Token总量',
  `replan_count` int NOT NULL DEFAULT '0' COMMENT '重规划次数',
  `current_stage` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '当前阶段: design/review/execute/rework/complete',
  `current_stage_run_id` bigint unsigned DEFAULT NULL COMMENT '当前阶段运行ID',
  `active_plan_version_id` bigint unsigned DEFAULT NULL COMMENT '当前活跃计划版本ID',
  `pause_reason` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '暂停原因',
  `status_before_pause` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '暂停前的阶段状态（恢复时回退）',
  `cancel_reason` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '取消原因',
  `runtime_token` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '运行时令牌(防重入)',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `finished_at` datetime DEFAULT NULL COMMENT '结束时间',
  `created_by` bigint unsigned DEFAULT '0' COMMENT '创建人ID',
  `dept_id` bigint unsigned DEFAULT '0' COMMENT '所属部门ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_run_no` (`project_id`,`run_no`),
  KEY `idx_project_status` (`project_id`,`status`),
  KEY `idx_project_deleted_id` (`project_id`,`deleted_at`,`id`),
  KEY `idx_datascope` (`dept_id`,`created_by`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流运行实例';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_delete_queue` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `entity` varchar(64) NOT NULL COMMENT '实体类型，如 mvp_task',
  `ids` json NOT NULL COMMENT '待删除的主键列表',
  `status` varchar(16) NOT NULL DEFAULT 'pending' COMMENT 'pending/processing/done/failed',
  `retry_count` int NOT NULL DEFAULT '0',
  `error_msg` text,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='异步删除队列';
/*!40101 SET character_set_client = @saved_cs_client */;
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
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_role_ai_engine` (
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `engine_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行引擎编码',
  PRIMARY KEY (`role_id`,`engine_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色AI执行引擎授权表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_role_dept` (
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `dept_id` bigint unsigned NOT NULL COMMENT '部门ID',
  PRIMARY KEY (`role_id`,`dept_id`),
  KEY `idx_dept_id` (`dept_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色自定义数据权限部门关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_role_menu` (
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `menu_id` bigint unsigned NOT NULL COMMENT '菜单ID',
  PRIMARY KEY (`role_id`,`menu_id`),
  KEY `idx_menu_id` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单权限关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_user_dept` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `dept_id` bigint unsigned NOT NULL COMMENT '部门ID',
  PRIMARY KEY (`user_id`,`dept_id`),
  KEY `idx_dept_id` (`dept_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户部门关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_user_role` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  PRIMARY KEY (`user_id`,`role_id`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
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
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
