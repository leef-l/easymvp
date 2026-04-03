SET NAMES utf8mb4;

-- MVP 配置管理菜单
INSERT INTO `system_menu` (`id`, `parent_id`, `title`, `type`, `path`, `component`, `permission`, `icon`, `sort`, `is_show`, `is_cache`, `link_url`, `status`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`)
SELECT 315200100000000002, 315013060379021312, CONVERT(0xE9858DE7BDAEE7AEA1E79086 USING utf8mb4), 2, '/mvp/config', 'mvp/config/index', 'mvp:config:list', 'lucide:settings-2', 11, 1, 0, NULL, 1, 1, 0, NOW(), NOW(), NULL
WHERE NOT EXISTS (SELECT 1 FROM `system_menu` WHERE `id` = 315200100000000002);
