-- M5.5: 默认组织模型收敛
-- 在完整角色预设库之上新增 is_default 层，标记新建项目的默认团队模板。
-- 不删除任何已有预设，扩展角色保留用于高级配置和特定任务类型。

-- 1. 添加 is_default 字段
ALTER TABLE `mvp_role_preset` ADD COLUMN `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为默认模板（1=默认，0=扩展）' AFTER `execution_mode`;

-- 2. 所有分类统一规则：每个分类按角色类型标记默认等级
-- architect: 取 max（每分类最多 1 个 max）
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `deleted_at` IS NULL AND `role_type` = 'architect' AND `role_level` = 'max';

-- implementer: 取 pro
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `deleted_at` IS NULL AND `role_type` = 'implementer' AND `role_level` = 'pro';

-- auditor: 取 pro
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `deleted_at` IS NULL AND `role_type` = 'auditor' AND `role_level` = 'pro';

-- coordinator: 取 lite
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `deleted_at` IS NULL AND `role_type` = 'coordinator' AND `role_level` = 'lite';
