-- Accept Stage 种子数据：4类模板验收规则
-- 对应设计文档：docs/AcceptStage自动验收阶段设计文档.md 第21节

-- ============================================
-- 1. software_dev 模板规则
-- ============================================

INSERT INTO `mvp_accept_rule` (`id`, `project_type`, `rule_code`, `rule_name`, `rule_type`, `scope_type`, `config_json`, `enabled`, `priority`, `created_at`, `updated_at`) VALUES
(1, 'software_dev', 'software.no_failed_tasks', '不得存在失败任务', 'process', 'project',
 '{"forbid_status":["failed","escalated"]}',
 1, 10, NOW(), NOW()),

(2, 'software_dev', 'software.required_file_exists', '关键文件必须存在', 'artifact', 'file',
 '{"required_files":["README.md"]}',
 1, 20, NOW(), NOW()),

(3, 'software_dev', 'software.output_not_empty', '关键输出不得为空', 'artifact', 'task',
 '{"task_kinds":["implement","refactor","fix"],"require_non_empty_result":true}',
 1, 30, NOW(), NOW());

-- ============================================
-- 2. document 模板规则
-- ============================================

INSERT INTO `mvp_accept_rule` (`id`, `project_type`, `rule_code`, `rule_name`, `rule_type`, `scope_type`, `config_json`, `enabled`, `priority`, `created_at`, `updated_at`) VALUES
(4, 'document', 'document.required_output_exists', '文档产物必须存在', 'artifact', 'file',
 '{"required_extensions":[".md",".docx"]}',
 1, 10, NOW(), NOW()),

(5, 'document', 'document.summary_present', '必须生成总结', 'process', 'stage',
 '{"required_stage_outputs":["execute"]}',
 1, 20, NOW(), NOW());

-- ============================================
-- 3. creative 模板规则
-- ============================================

INSERT INTO `mvp_accept_rule` (`id`, `project_type`, `rule_code`, `rule_name`, `rule_type`, `scope_type`, `config_json`, `enabled`, `priority`, `created_at`, `updated_at`) VALUES
(6, 'creative', 'creative.required_sections_present', '核心章节必须齐全', 'artifact', 'project',
 '{"required_sections":["outline","main_content","ending"]}',
 1, 10, NOW(), NOW());

-- ============================================
-- 4. analysis 模板规则
-- ============================================

INSERT INTO `mvp_accept_rule` (`id`, `project_type`, `rule_code`, `rule_name`, `rule_type`, `scope_type`, `config_json`, `enabled`, `priority`, `created_at`, `updated_at`) VALUES
(7, 'analysis', 'analysis.must_have_conclusion', '分析报告必须包含结论', 'quality', 'project',
 '{"required_keywords":["结论","建议"]}',
 1, 10, NOW(), NOW());
