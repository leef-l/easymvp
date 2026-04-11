-- BrainKernel v1 持久化层：新增 5 张表支撑下一代执行器
-- 对应设计文档 docs/next-gen-executor/26-持久化与恢复.md 附录 B

CREATE TABLE `mvp_brain_plan` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `run_id` bigint unsigned NOT NULL COMMENT '所属 Run ID',
  `brain_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '产出该 plan 的 brain 标识',
  `version` bigint unsigned NOT NULL DEFAULT '1' COMMENT '当前 plan 版本号（每次 delta 提交 +1）',
  `current_state` json NOT NULL COMMENT '当前 plan 完整快照',
  `archived` tinyint NOT NULL DEFAULT '0' COMMENT '是否已归档: 0-活跃 1-已归档',
  `archived_at` datetime DEFAULT NULL COMMENT '归档时间',
  `archive_ref` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '归档后的冷存储引用',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人用户ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '创建人部门ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_brain_plan_run` (`run_id`),
  KEY `idx_brain_plan_created` (`created_at`),
  KEY `idx_brain_plan_brain_archived` (`brain_id`,`archived`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='BrainKernel 计划表（中央大脑产出的任务树快照）';

CREATE TABLE `mvp_brain_plan_delta` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `plan_id` bigint unsigned NOT NULL COMMENT '所属 plan ID',
  `version` bigint unsigned NOT NULL COMMENT '该 delta 对应的 plan 版本号',
  `op_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '操作类型: add_node/update_status/mark_done/...',
  `payload` json NOT NULL COMMENT 'delta 具体内容',
  `actor` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '发起变更的 brain 或 user 标识',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人用户ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '创建人部门ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_brain_plan_delta_plan_version` (`plan_id`,`version`),
  KEY `idx_brain_plan_delta_plan` (`plan_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='BrainKernel 计划增量日志（append-only，用于审计与重放）';

CREATE TABLE `mvp_run_checkpoint` (
  `run_id` bigint unsigned NOT NULL COMMENT 'Run ID（一个 run 只有一个最新 checkpoint）',
  `turn_index` int NOT NULL COMMENT 'Turn 序号（从 0 开始）',
  `brain_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '当前活跃 brain',
  `state` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Run 状态: running/paused/waiting_tool/...',
  `messages_ref` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'messages 的 CAS 引用（sha256/...）',
  `system_ref` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'system prompt 的 CAS 引用',
  `tools_ref` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'tools 定义的 CAS 引用',
  `cost_snapshot` json DEFAULT NULL COMMENT '截至当前 Turn 的累计成本（USD）',
  `token_snapshot` json DEFAULT NULL COMMENT '截至当前 Turn 的累计 token',
  `budget_remain` json DEFAULT NULL COMMENT '剩余 Run/Turn/Tool/LLMCall 预算',
  `trace_parent` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'W3C Trace Context parent',
  `turn_uuid` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '当前 Turn 的唯一标识（恢复时幂等判据）',
  `resume_attempts` int NOT NULL DEFAULT '0' COMMENT '恢复尝试次数',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人用户ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '创建人部门ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`run_id`),
  UNIQUE KEY `uk_run_checkpoint_turn_uuid` (`turn_uuid`),
  KEY `idx_run_checkpoint_updated` (`updated_at`),
  KEY `idx_run_checkpoint_state_updated` (`state`,`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='BrainKernel Run 检查点（Turn 边界覆盖式，用于崩溃恢复）';

CREATE TABLE `mvp_brain_usage` (
  `id` bigint unsigned NOT NULL COMMENT '雪花ID',
  `run_id` bigint unsigned NOT NULL COMMENT '所属 Run ID',
  `turn_index` int NOT NULL COMMENT '对应 Turn 序号',
  `provider` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'LLM 供应商: anthropic/openai/...',
  `model` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模型名',
  `input_tokens` bigint unsigned NOT NULL COMMENT '输入 token 数',
  `output_tokens` bigint unsigned NOT NULL COMMENT '输出 token 数',
  `cache_read` bigint unsigned NOT NULL DEFAULT '0' COMMENT 'Prompt Cache 命中读取的 token',
  `cache_creation` bigint unsigned NOT NULL DEFAULT '0' COMMENT 'Prompt Cache 写入创建的 token',
  `cost_usd` decimal(10,4) NOT NULL COMMENT '本次调用成本（USD）',
  `idempotency_key` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '幂等键（防止重复记账）',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人用户ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '创建人部门ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_brain_usage_idempotency` (`idempotency_key`),
  KEY `idx_brain_usage_run` (`run_id`),
  KEY `idx_brain_usage_provider_model` (`provider`,`model`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='BrainKernel LLM 调用计费明细（append-only）';

CREATE TABLE `mvp_artifact_meta` (
  `ref` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'CAS 引用键: sha256/<64 hex>',
  `mime_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'MIME 类型',
  `size_bytes` bigint unsigned NOT NULL COMMENT '字节数',
  `run_id` bigint unsigned DEFAULT NULL COMMENT '首次产出该 artifact 的 Run ID',
  `turn_index` int DEFAULT NULL COMMENT '首次产出时的 Turn 序号',
  `caption` text COLLATE utf8mb4_unicode_ci COMMENT 'artifact 描述（用于 LLM 引用）',
  `ref_count` bigint unsigned NOT NULL DEFAULT '1' COMMENT '引用计数（用于 CAS 垃圾回收）',
  `tags` json DEFAULT NULL COMMENT '自定义标签',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人用户ID',
  `dept_id` bigint unsigned DEFAULT NULL COMMENT '创建人部门ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`ref`),
  KEY `idx_artifact_meta_created` (`created_at`),
  KEY `idx_artifact_meta_refcount` (`ref_count`),
  KEY `idx_artifact_meta_run` (`run_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='BrainKernel Artifact 元数据（CAS 内容寻址）';
