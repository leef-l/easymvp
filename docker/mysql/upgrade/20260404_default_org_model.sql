-- M5.5: 默认组织模型收敛
-- 在完整角色预设库之上新增 is_default 层，标记新建项目的默认团队模板。
-- 不删除任何已有预设，扩展角色保留用于高级配置和特定任务类型。

-- 1. 添加 is_default 字段，区分默认模板与扩展模板
ALTER TABLE `mvp_role_preset` ADD COLUMN `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为默认模板（1=默认，0=扩展）' AFTER `execution_mode`;

-- 2. 软件开发：标记核心 4 角色为默认
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `project_category` = '软件开发' AND `deleted_at` IS NULL AND `id` IN (
  316100000000000001, -- architect max
  316100000000000002, -- implementer pro
  316100000000000003, -- auditor pro
  316100000000000004  -- coordinator lite
);

-- 3. 非软件开发分类：按角色类型标记默认（每类型取 1 个主力等级）
-- architect: 每分类只有 1 个，全标
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `deleted_at` IS NULL AND `project_category` != '软件开发' AND `role_type` = 'architect';

-- implementer: 只标 pro（主力等级），max/lite 保留为扩展
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `deleted_at` IS NULL AND `project_category` != '软件开发' AND `role_type` = 'implementer' AND `role_level` = 'pro';

-- auditor + coordinator: 每分类各 1 个，全标
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `deleted_at` IS NULL AND `project_category` != '软件开发' AND `role_type` IN ('auditor', 'coordinator');
