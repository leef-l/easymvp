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
  executionMode: string;
  systemPrompt: string;
  isDefault: boolean;
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

// ==================== Timeline / Rework / Stage History ====================

/** 时间线事件 */
export interface TimelineEvent {
  id: string;
  workflowRunID: string;
  stageRunID?: string;
  entityType: string;
  entityID?: string;
  eventType: string;
  label: string;
  payload?: string;
  createdAt: string;
}

/** 获取工作流事件时间线 */
export function getTimeline(projectID: string, limit = 50) {
  return requestClient.get<{ events: TimelineEvent[] }>(
    `${PREFIX}/timeline`,
    { params: { projectID, limit } },
  );
}

/** 返工阶段信息 */
export interface ReworkStageInfo {
  stageRunID: string;
  status: string;
  startedAt?: string;
}

/** 返工轮次信息 */
export interface ReworkRoundInfo {
  round: number;
  failedTaskID: string;
  failedTaskName: string;
  failedReason: string;
  analysisTaskID?: string;
  analysisResult?: string;
  handoffType: string;
  createdAt: string;
}

/** 获取返工阶段状态 */
export function getReworkStatus(projectID: string) {
  return requestClient.get<{
    hasRework: boolean;
    reworkRounds: number;
    currentStage?: ReworkStageInfo;
    history: ReworkRoundInfo[];
  }>(`${PREFIX}/rework-status`, { params: { projectID } });
}

/** 阶段历史项 */
export interface StageHistoryItem {
  id: string;
  stageType: string;
  stageNo: number;
  status: string;
  startedAt?: string;
  finishedAt?: string;
  error?: string;
}

/** 获取阶段历史 */
export function getStageHistory(projectID: string) {
  return requestClient.get<{ stages: StageHistoryItem[] }>(
    `${PREFIX}/stage-history`,
    { params: { projectID } },
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

// ==================== 完成总结 ====================

/** 完成总结 */
export interface CompletionSummaryResult {
  workflowRunID: string;
  projectID: string;
  totalTasks: number;
  completedTasks: number;
  failedTasks: number;
  escalatedTasks: number;
  skippedTasks: number;
  successRate: number;
  totalDuration: string;
  avgTaskDuration: string;
  stageDurations: Record<string, string>;
  reworkRounds: number;
  handoffCount: number;
  startedAt: string;
  finishedAt: string;
}

/** 获取项目完成总结 */
export function getCompletionSummary(projectID: string) {
  return requestClient.get<CompletionSummaryResult>(
    `${PREFIX}/completion-summary`,
    { params: { projectID } },
  );
}

// ==================== 执行控制台 ====================

/** 领域任务详情 */
export interface DomainTaskItem {
  id: string;
  name: string;
  description?: string;
  status: string;
  roleType: string;
  roleLevel: string;
  batchNo: number;
  sort: number;
  executionMode: string;
  affectedResources: string[];
  startedAt?: string;
  completedAt?: string;
  errorMessage?: string;
  result?: string;
  retryCount: number;
}

/** 资源锁详情 */
export interface ResourceLockItem {
  resource: string;
  taskID: string;
  taskName: string;
}

/** 执行状态结果 */
export interface ExecutionStatusResult {
  workflowRunID: string;
  stageRunID: string;
  stageStatus: string;
  activeBatch: number;
  totalTasks: number;
  completedTasks: number;
  runningTasks: number;
  failedTasks: number;
  pendingTasks: number;
  escalatedTasks: number;
  tasks: DomainTaskItem[];
  resourceLocks: ResourceLockItem[];
}

/** 获取执行阶段实时状态 */
export function getExecutionStatus(projectID: string) {
  return requestClient.get<ExecutionStatusResult>(
    `${PREFIX}/execution-status`,
    { params: { projectID } },
  );
}

/** 获取领域任务列表 */
export function getDomainTasks(projectID: string, status?: string, batchNo?: number) {
  return requestClient.get<{ tasks: DomainTaskItem[]; total: number }>(
    `${PREFIX}/domain-tasks`,
    { params: { projectID, ...(status ? { status } : {}), ...(batchNo ? { batchNo } : {}) } },
  );
}

/** 获取资源锁列表 */
export function getResourceLocks(projectID: string) {
  return requestClient.get<{ locks: ResourceLockItem[] }>(
    `${PREFIX}/resource-locks`,
    { params: { projectID } },
  );
}
