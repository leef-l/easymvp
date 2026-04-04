// ==================== 项目状态 ====================

/** 项目状态枚举 */
export const ProjectStatus = {
  DESIGNING: 'designing',
  RUNNING: 'running',
  PAUSED: 'paused',
  COMPLETED: 'completed',
} as const;

/** 项目状态映射（用于列表/详情展示） */
export const projectStatusMap: Record<string, { label: string; color: string }> = {
  designing: { label: '设计中', color: 'blue' },
  running: { label: '执行中', color: 'green' },
  paused: { label: '已暂停', color: 'orange' },
  completed: { label: '已完成', color: 'default' },
};

/** 项目状态选项 */
export const projectStatusOptions = [
  { label: '设计中', value: 'designing' },
  { label: '执行中', value: 'running' },
  { label: '已暂停', value: 'paused' },
  { label: '已完成', value: 'completed' },
];

// ==================== 任务状态 ====================

/** 任务状态枚举 */
export const TaskStatus = {
  DRAFT: 'draft',
  PENDING: 'pending',
  RUNNING: 'running',
  COMPLETED: 'completed',
  FAILED: 'failed',
  ESCALATED: 'escalated',
  AUDITING: 'auditing',
  BUG_FOUND: 'bug_found',
  BUG_DISPATCHED: 'bug_dispatched',
  SUBMIT_ERROR: 'submit_error',
} as const;

/** 任务状态映射 */
export const taskStatusMap: Record<string, { label: string; color: string }> = {
  draft: { label: '草稿', color: 'default' },
  pending: { label: '待执行', color: 'blue' },
  running: { label: '执行中', color: 'processing' },
  completed: { label: '已完成', color: 'success' },
  failed: { label: '失败', color: 'error' },
  escalated: { label: '已升级', color: 'warning' },
  auditing: { label: '审计中', color: 'purple' },
  bug_found: { label: '发现Bug', color: 'error' },
  bug_dispatched: { label: 'Bug已派发', color: 'orange' },
  submit_error: { label: '提交异常', color: 'error' },
};

/** 任务状态选项 */
export const taskStatusOptions = Object.entries(taskStatusMap).map(([value, { label }]) => ({
  label,
  value,
}));

// ==================== 项目分类 ====================

/** 项目分类选项（与数据库 mvp_role_preset 保持一致） */
export const projectCategoryOptions = [
  { label: '软件开发', value: '软件开发' },
  { label: '游戏开发', value: '游戏开发' },
  { label: '小说创作', value: '小说创作' },
  { label: '动漫创作', value: '动漫创作' },
  { label: '漫剧创作', value: '漫剧创作' },
  { label: '大电影创作', value: '大电影创作' },
  { label: '动画创作', value: '动画创作' },
  { label: '数据分析', value: '数据分析' },
  { label: '产品设计', value: '产品设计' },
];

/** 角色类型选项 */
export const roleTypeOptions = [
  { label: '架构师', value: 'architect' },
  { label: '实现者', value: 'implementer' },
  { label: '审核者', value: 'auditor' },
  { label: '协调者', value: 'coordinator' },
];

/** 角色类型映射（用于列表/详情展示） */
export const roleTypeMap: Record<string, { label: string; color: string }> = {
  architect: { label: '架构师', color: 'purple' },
  implementer: { label: '实现者', color: 'blue' },
  auditor: { label: '审核者', color: 'green' },
  coordinator: { label: '协调者', color: 'orange' },
};

/** 角色等级选项 */
export const roleLevelOptions = [
  { label: 'Lite - 轻量级', value: 'lite' },
  { label: 'Pro - 专业级', value: 'pro' },
  { label: 'Max - 旗舰级', value: 'max' },
];

/** 角色等级映射（用于列表/详情展示） */
export const roleLevelMap: Record<string, { label: string; color: string }> = {
  lite: { label: 'Lite', color: 'default' },
  pro: { label: 'Pro', color: 'blue' },
  max: { label: 'Max', color: 'gold' },
};

// ==================== 执行方式 ====================

/** 执行方式选项 */
export const executionModeOptions = [
  { label: '对话模式 (Chat)', value: 'chat' },
  { label: 'Aider 代码编辑', value: 'aider' },
  { label: 'OpenHands 沙箱', value: 'openhands' },
];

/** 执行方式映射 */
export const executionModeMap: Record<string, { label: string; color: string }> = {
  chat: { label: 'Chat', color: 'blue' },
  aider: { label: 'Aider', color: 'green' },
  openhands: { label: 'OpenHands', color: 'purple' },
};
