-- M5.5: 默认组织模型收敛
-- 目标：每个项目分类保留 4 条默认角色预设：architect(max) + implementer(pro) + auditor(pro) + coordinator(lite)

-- 1. 添加 is_default 字段，区分默认模板与扩展模板
ALTER TABLE `mvp_role_preset` ADD COLUMN `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为默认模板（1=默认，0=扩展）' AFTER `execution_mode`;

-- 2. 软件开发：标记 sort 1-4 为默认（architect max / implementer pro / auditor pro / coordinator lite）
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `project_category` = '软件开发' AND `deleted_at` IS NULL AND `id` IN (
  316100000000000001, -- architect max
  316100000000000002, -- implementer pro
  316100000000000003, -- auditor pro
  316100000000000004  -- coordinator lite
);

-- 3. 非软件开发分类：按角色类型精确标记默认
-- architect: 每分类 1 个，全标
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `deleted_at` IS NULL AND `project_category` != '软件开发' AND `role_type` = 'architect';

-- implementer: 只标 pro 等级
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `deleted_at` IS NULL AND `project_category` != '软件开发' AND `role_type` = 'implementer' AND `role_level` = 'pro';

-- auditor + coordinator: 每分类各 1 个，全标
UPDATE `mvp_role_preset` SET `is_default` = 1
WHERE `deleted_at` IS NULL AND `project_category` != '软件开发' AND `role_type` IN ('auditor', 'coordinator');

-- 4. 软删除非默认的多余 implementer 变体（非软件开发分类的 implementer max/lite）
UPDATE `mvp_role_preset` SET `deleted_at` = NOW()
WHERE `deleted_at` IS NULL AND `project_category` != '软件开发'
  AND `role_type` = 'implementer' AND `role_level` != 'pro';
