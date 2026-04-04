-- 项目分类差异化执行：添加 execution_mode 字段
-- 执行时间：2026-04-04

-- 1. 角色预设表加 execution_mode
ALTER TABLE mvp_role_preset
ADD COLUMN execution_mode VARCHAR(20) NOT NULL DEFAULT 'chat'
COMMENT '执行方式: chat=对话模式, aider=Aider代码编辑, openhands=OpenHands沙箱'
AFTER system_prompt;

-- 2. 项目角色表加 execution_mode
ALTER TABLE mvp_project_role
ADD COLUMN execution_mode VARCHAR(20) NOT NULL DEFAULT 'chat'
COMMENT '执行方式: chat=对话模式, aider=Aider代码编辑, openhands=OpenHands沙箱'
AFTER system_prompt;

-- 3. 更新角色预设：软件开发和游戏开发的 implementer 用 aider
UPDATE mvp_role_preset SET execution_mode = 'aider'
WHERE project_category IN ('软件开发', '游戏开发') AND role_type = 'implementer';

-- 4. 更新已有项目角色
UPDATE mvp_project_role SET execution_mode = 'aider'
WHERE project_category IN ('软件开发', '游戏开发') AND role_type = 'implementer';
