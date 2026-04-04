import { requestClient } from '#/api/request';

/** API 前缀 */
const PREFIX = '/mvp/workflow';

/** 创建项目（通过工作流，返回项目ID和对话ID） */
export function createProject(data: {
  name: string;
  projectCategory: string;
  description: string;
  workDir: string;
  architectModelID: string;
  engineVersion?: string;
}) {
  return requestClient.post<{
    projectID: string;
    conversationID: string;
    workflowRunID: string;
  }>(`${PREFIX}/create-project`, data);
}

/** 确认实施方案（由设计阶段进入执行阶段） */
export function confirmPlan(projectID: string) {
  return requestClient.post(`${PREFIX}/confirm-plan`, { projectID });
}

/** 手动解析架构师回复中的任务清单（托底机制） */
export function parseTasks(projectID: string, dryRun = false) {
  return requestClient.post<{ hasTasks: boolean; taskCount: number }>(
    `${PREFIX}/parse-tasks`,
    { projectID, dryRun },
  );
}

/** 暂停项目 */
export function pauseProject(data: {
  projectID: string;
  pauseReason: string;
}) {
  return requestClient.post(`${PREFIX}/pause`, data);
}

/** 恢复项目（从暂停状态继续执行） */
export function resumeProject(projectID: string) {
  return requestClient.post(`${PREFIX}/resume`, { projectID });
}

/** 重试失败的任务 */
export function retryTask(data: { projectID: string; taskID: string }) {
  return requestClient.post(`${PREFIX}/retry-task`, data);
}

/** 跳过失败任务（防止批次永久阻塞） */
export function skipTask(data: {
  projectID: string;
  taskID: string;
  reason: string;
}) {
  return requestClient.post(`${PREFIX}/skip-task`, data);
}

/** 角色预设项 */
export interface RolePresetItem {
  roleType: string;
  roleLevel: string;
  modelID: string;
  modelName: string;
  systemPrompt: string;
}

/** 获取角色预设列表（可按项目分类过滤） */
export function getRolePresets(projectCategory?: string) {
  return requestClient.get<{ list: RolePresetItem[] }>(
    `${PREFIX}/role-presets`,
    { params: projectCategory ? { projectCategory } : {} },
  );
}

/** 项目状态响应 */
export interface ProjectStatusResult {
  status: string;
  pauseReason?: string;
  activeBatch: number;
  totalTasks: number;
  statusCounts: Record<string, number>;
  lastActiveAt?: string;
  isActuallyWorking: boolean;
  activeRunningTasks: number;
  stalledTaskCount: number;
  // V2 聚合字段
  engineVersion?: string;
  workflowStatus?: string;
  currentStage?: string;
  progressPercent?: number;
}

/** 获取项目当前状态及任务统计 */
export function getProjectStatus(projectID: string) {
  return requestClient.get<ProjectStatusResult>(
    `${PREFIX}/project-status`,
    { params: { projectID } },
  );
}

/** 系统检测项 */
export interface SystemCheckItem {
  key: string;
  name: string;
  status: 'error' | 'ok' | 'warning';
  message: string;
  link: string;
}

/** 获取系统配置检测结果 */
export function getSystemCheck() {
  return requestClient.get<{ items: SystemCheckItem[]; allPass: boolean }>(
    `${PREFIX}/system-check`,
  );
}

// ==================== 审核相关 ====================

/** 审核阶段子任务 */
export interface ReviewStageTask {
  id: string;
  taskType: string;
  roleType: string;
  status: string;
  startedAt?: string;
  completedAt?: string;
  errorMessage?: string;
}

/** 审核问题项 */
export interface ReviewIssueItem {
  id: string;
  severity: string;
  issueCode: string;
  sourceRole: string;
  taskName: string;
  message: string;
  suggestion?: string;
  status: string;
  createdAt: string;
}

/** 获取审核状态 */
export function getReviewStatus(projectID: string) {
  return requestClient.get<{
    planVersionID: string;
    reviewStatus: string;
    stageRunID: string;
    stageStatus: string;
    stageTasks: ReviewStageTask[];
    errorCount: number;
    warningCount: number;
    blueprintCount: number;
  }>(`${PREFIX}/review-status`, { params: { projectID } });
}

/** 获取审核问题列表 */
export function getReviewIssues(projectID: string) {
  return requestClient.get<{ issues: ReviewIssueItem[] }>(
    `${PREFIX}/review-issues`,
    { params: { projectID } },
  );
}

/** 手动审批通过 */
export function manualApprove(projectID: string) {
  return requestClient.post(`${PREFIX}/manual-approve`, { projectID });
}

/** 手动驳回 */
export function manualReject(projectID: string, reason: string) {
  return requestClient.post(`${PREFIX}/manual-reject`, { projectID, reason });
}
