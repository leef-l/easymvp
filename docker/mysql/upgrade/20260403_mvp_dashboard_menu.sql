-- MVP 系统概览菜单（配置检测 Dashboard）
INSERT INTO `system_menu` (`id`, `parent_id`, `title`, `type`, `path`, `component`, `permission`, `icon`, `sort`, `is_show`, `is_cache`, `link_url`, `status`, `created_by`, `dept_id`, `created_at`, `updated_at`, `deleted_at`)
SELECT 315200100000000001, 315013060379021312, '系统概览', 2, '/mvp/dashboard', 'mvp/dashboard/index', NULL, 'lucide:layout-dashboard', -1, 1, 0, NULL, 1, 1, 0, NOW(), NOW(), NULL
WHERE NOT EXISTS (SELECT 1 FROM `system_menu` WHERE `id` = 315200100000000001);
