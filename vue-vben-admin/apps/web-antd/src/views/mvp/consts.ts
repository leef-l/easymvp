// ==================== 项目状态 ====================

/** 项目状态枚举 */
export const ProjectStatus = {
  DESIGNING: 'designing',
  REVIEWING: 'reviewing',
  RUNNING: 'running',
  PAUSED: 'paused',
  COMPLETED: 'completed',
} as const;

/** 项目状态映射（用于列表/详情展示） */
export const projectStatusMap: Record<string, { label: string; color: string }> = {
  designing: { label: '设计中', color: 'blue' },
  reviewing: { label: '方案审核中', color: 'cyan' },
  running: { label: '执行中', color: 'green' },
  paused: { label: '已暂停', color: 'orange' },
  completed: { label: '已完成', color: 'default' },
};

/** 项目状态选项 */
export const projectStatusOptions = [
  { label: '设计中', value: 'designing' },
  { label: '方案审核中', value: 'reviewing' },
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

/** 项目分类静态选项（兜底，新建项目表单已改为从 API 动态加载） */
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

/** 编码类分类的 familyCode */
export const CODING_FAMILY = 'coding';

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

// ==================== 引擎版本 ====================

/** 引擎版本枚举 */
export const EngineVersion = {
  WORKFLOW_V2: 'workflow_v2',
} as const;

/** 引擎版本映射 */
export const engineVersionMap: Record<string, { label: string; color: string }> = {
  workflow_v2: { label: 'Workflow V2', color: 'green' },
};

/** 引擎版本选项 */
export const engineVersionOptions = [
  { label: 'Workflow V2', value: 'workflow_v2' },
];

// ==================== 工作流状态 ====================

/** 工作流运行状态枚举（阶段化语义） */
export const WorkflowRunStatus = {
  DESIGNING: 'designing',
  REVIEWING: 'reviewing',
  EXECUTING: 'executing',
  REWORKING: 'reworking',
  PAUSED: 'paused',
  COMPLETED: 'completed',
  FAILED: 'failed',
  CANCELED: 'canceled',
} as const;

/** 工作流状态映射 */
export const workflowRunStatusMap: Record<string, { label: string; color: string }> = {
  designing: { label: '设计中', color: 'cyan' },
  reviewing: { label: '审核中', color: 'purple' },
  executing: { label: '执行中', color: 'green' },
  reworking: { label: '返工中', color: 'orange' },
  paused: { label: '已暂停', color: 'default' },
  completed: { label: '已完成', color: 'blue' },
  failed: { label: '已失败', color: 'red' },
  canceled: { label: '已取消', color: 'volcano' },
};

// ==================== 工作流阶段 ====================

/** 阶段类型枚举 */
export const StageType = {
  DESIGN: 'design',
  REVIEW: 'review',
  EXECUTE: 'execute',
  REWORK: 'rework',
  COMPLETE: 'complete',
} as const;

/** 阶段类型映射 */
export const stageTypeMap: Record<string, { label: string; color: string }> = {
  design: { label: '设计', color: 'blue' },
  review: { label: '审核', color: 'cyan' },
  execute: { label: '执行', color: 'green' },
  accept: { label: '验收', color: 'purple' },
  rework: { label: '返工', color: 'orange' },
  complete: { label: '完成', color: 'default' },
};

/** 阶段运行状态映射 */
export const stageRunStatusMap: Record<string, { label: string; color: string }> = {
  pending: { label: '待启动', color: 'default' },
  running: { label: '进行中', color: 'processing' },
  completed: { label: '已完成', color: 'success' },
  failed: { label: '失败', color: 'error' },
};

// ==================== 执行方式 ====================

/** 执行方式选项 */
export const executionModeOptions = [
  { label: '🤖 Auto（自动选择）', value: 'auto' },
  { label: '💬 Chat 对话', value: 'chat' },
  { label: '✏️ Aider', value: 'aider' },
  { label: '🐙 OpenHands', value: 'openhands' },
  { label: '🔵 Claude Code', value: 'claude_code' },
  { label: '🟢 Codex CLI', value: 'codex_cli' },
  { label: '🔴 Gemini CLI', value: 'gemini_cli' },
];

/** 执行方式映射 */
export const executionModeMap: Record<string, { label: string; color: string }> = {
  auto: { label: 'Auto', color: 'gold' },
  chat: { label: 'Chat', color: 'blue' },
  aider: { label: 'Aider', color: 'green' },
  openhands: { label: 'OpenHands', color: 'purple' },
  claude_code: { label: 'Claude Code', color: 'cyan' },
  codex_cli: { label: 'Codex CLI', color: 'lime' },
  gemini_cli: { label: 'Gemini CLI', color: 'orange' },
};
