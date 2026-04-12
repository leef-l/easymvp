-- 回滚：删除本次新增的 V2 预设 + 恢复 Legacy 预设为默认
DELETE FROM `mvp_role_preset` WHERE `id` >= 2000001 AND `id` <= 2000117;

-- 恢复 Legacy 中文预设为默认
UPDATE `mvp_role_preset`
SET `is_default` = 1, `updated_at` = NOW()
WHERE `project_category` IN ('软件开发','游戏开发','小说创作','动漫创作','漫剧创作','大电影创作','动画创作','数据分析','产品设计')
  AND `deleted_at` IS NULL;
