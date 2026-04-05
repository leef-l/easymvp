-- =====================================================================
-- 自治中台种子数据：策略规则 + 闸门规则 + 灰度开关
-- =====================================================================

-- ==================== 灰度开关 ====================
INSERT INTO `mvp_config` (`config_key`, `config_value`, `description`, `config_type`, `created_at`, `updated_at`)
VALUES
  ('workflow.autonomy.enabled', '0', '自治中台总开关（0=关闭, 1=开启）', 'autonomy', NOW(), NOW()),
  ('workflow.autonomy.audit_only', '1', '仅审计模式（1=只写记录不执行, 0=正式接管）', 'autonomy', NOW(), NOW()),
  ('workflow.autonomy.policy_engine_enabled', '1', '策略引擎开关', 'autonomy', NOW(), NOW()),
  ('workflow.autonomy.risk_gate_enabled', '1', '风险闸门开关', 'autonomy', NOW(), NOW())
ON DUPLICATE KEY UPDATE `updated_at` = NOW();

-- ==================== A 级策略规则（自动执行）====================
INSERT INTO `mvp_policy_rule` (`id`, `rule_code`, `rule_name`, `decision_type`, `decision_level`, `trigger_source`, `config_json`, `enabled`, `priority`, `created_at`, `updated_at`)
VALUES
  (900000000000001, 'POL_TASK_FAIL_RETRY', '任务失败自动重试', 'task_recovery', 'A', 'task.failed',
   '{"action_type":"retry_task","max_retry":3}', 1, 10, NOW(), NOW()),
  (900000000000002, 'POL_REWORK_DONE_ACCEPT', '返工完成回验收', 'rework_flow', 'A', 'rework.completed',
   '{"action_type":"rerun_accept"}', 1, 10, NOW(), NOW())
ON DUPLICATE KEY UPDATE `updated_at` = NOW();

-- ==================== B 级策略规则（建议型，需人工确认）====================
INSERT INTO `mvp_policy_rule` (`id`, `rule_code`, `rule_name`, `decision_type`, `decision_level`, `trigger_source`, `config_json`, `enabled`, `priority`, `created_at`, `updated_at`)
VALUES
  (900000000000003, 'POL_RETRY_EXHAUST_REWORK', '重试耗尽建议返工', 'task_recovery', 'B', 'task.retry_exhausted',
   '{"action_type":"trigger_rework"}', 1, 20, NOW(), NOW()),
  (900000000000004, 'POL_ACCEPT_FAIL_REWORK', '验收失败建议返工', 'accept_flow', 'B', 'accept.failed',
   '{"action_type":"trigger_rework"}', 1, 20, NOW(), NOW()),
  (900000000000005, 'POL_ACCEPT_PASS_COMPLETE', '验收通过建议完成', 'accept_flow', 'B', 'accept.passed',
   '{"action_type":"approve_complete"}', 1, 20, NOW(), NOW())
ON DUPLICATE KEY UPDATE `updated_at` = NOW();

-- ==================== C 级策略规则（必须人工决定）====================
INSERT INTO `mvp_policy_rule` (`id`, `rule_code`, `rule_name`, `decision_type`, `decision_level`, `trigger_source`, `config_json`, `enabled`, `priority`, `created_at`, `updated_at`)
VALUES
  (900000000000006, 'POL_CIRCUIT_BREAK_PAUSE', '熔断暂停', 'circuit_break', 'C', 'workflow.circuit_break',
   '{"action_type":"pause_workflow"}', 1, 30, NOW(), NOW()),
  (900000000000007, 'POL_ACCEPT_MANUAL', '验收人工审核', 'accept_flow', 'C', 'accept.manual_review',
   '{"action_type":"notify_human"}', 1, 30, NOW(), NOW()),
  (900000000000008, 'POL_REPLAN_SUGGEST', '重规划建议', 'replan', 'C', 'replan.suggested',
   '{"action_type":"replan_workflow"}', 1, 30, NOW(), NOW())
ON DUPLICATE KEY UPDATE `updated_at` = NOW();

-- ==================== 闸门规则 ====================
INSERT INTO `mvp_risk_gate_rule` (`id`, `gate_code`, `gate_name`, `gate_type`, `trigger_expression`, `block_action`, `fallback_action`, `enabled`, `priority`, `created_at`, `updated_at`)
VALUES
  (900000000000101, 'GATE_RUNTIME_FAIL_LIMIT', '连续失败上限闸门', 'runtime',
   '{"trigger_source":"task.failed","min_failure_count":10}',
   'pause_workflow', 'notify_human', 1, 10, NOW(), NOW()),
  (900000000000102, 'GATE_QUALITY_REWORK_LIMIT', '返工轮次上限闸门', 'quality',
   '{"trigger_source":"rework.completed","max_rework_rounds":5}',
   'pause_workflow', 'notify_human', 1, 20, NOW(), NOW()),
  (900000000000103, 'GATE_COST_ACCEPT_LIMIT', '验收轮次上限闸门', 'cost',
   '{"trigger_source":"accept.failed","max_accept_rounds":5}',
   'pause_workflow', 'notify_human', 1, 30, NOW(), NOW())
ON DUPLICATE KEY UPDATE `updated_at` = NOW();
