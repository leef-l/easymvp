-- 回滚：删除本次新增的预设
DELETE FROM `mvp_role_preset` WHERE `id` >= 1000001 AND `id` <= 1000135;
