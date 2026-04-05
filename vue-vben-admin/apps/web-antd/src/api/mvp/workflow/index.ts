import { requestClient } from '#/api/request';

/** API 前缀 */
const PREFIX = '/mvp/workflow';

/** 分类列表项 */
export interface CategoryItem {
  categoryCode: string;
  displayName: string;
  familyCode: string;
  description: string;
}

/** 获取项目��类列表 */
export function getCategories() {
  return requestClient.get<{ list: CategoryItem[] }>(`${PREFIX}/categories`);
}

/** 创建项目（通过工作流，返回项目ID和对话ID） */
export function createProject(data: {
  name: string;
  projectCategory?: string;
  categoryCode?: string;
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

/** 获取角色预设列表（可按项目分类过滤，优先 categoryCode） */
export function getRolePresets(categoryCodeOrDisplayName?: string) {
  if (!categoryCodeOrDisplayName) {
    return requestClient.get<{ list: RolePresetItem[] }>(`${PREFIX}/role-presets`);
  }
  // 判断是 categoryCode（英文下划线）还是 displayName（中文）
  const isCategoryCode = /^[a-z_]+$/.test(categoryCodeOrDisplayName);
  const params = isCategoryCode
    ? { categoryCode: categoryCodeOrDisplayName }
    : { projectCategory: categoryCodeOrDisplayName };
  return requestClient.get<{ list: RolePresetItem[] }>(
    `${PREFIX}/role-presets`,
    { params },
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

// ==================== 验收控制台 ====================

/** 验收问题条目 */
export interface AcceptIssueItem {
  id: string;
  issueType: string;
  ruleCode: string;
  severity: string;
  title: string;
  detail: string;
  expectedValue: string;
  actualValue: string;
  suggestedAction: string;
  domainTaskID?: string;
  resourceRef?: string;
  status: string;
  createdAt: string;
}

/** 验收证据条目 */
export interface AcceptEvidenceItem {
  id: string;
  evidenceType: string;
  sourceType: string;
  sourceID?: string;
  contentRef?: string;
  summary: string;
  createdAt: string;
}

/** 验收状态总览结果 */
export interface AcceptStatusResult {
  acceptRunID: string;
  workflowRunID: string;
  acceptRound: number;
  status: string;
  decision: string;
  score: number;
  summary: string;
  rulesSnapshot?: string;
  startedAt?: string;
  finishedAt?: string;
  blockerCount: number;
  errorCount: number;
  warnCount: number;
  infoCount: number;
  evidenceCount: number;
}

/** 获取验收状态总览 */
export function getAcceptStatus(projectID: string) {
  return requestClient.get<AcceptStatusResult>(
    `${PREFIX}/accept-status`,
    { params: { projectID } },
  );
}

/** 获取验收问题列表 */
export function getAcceptIssues(projectID: string, severity?: string) {
  return requestClient.get<{ issues: AcceptIssueItem[] }>(
    `${PREFIX}/accept-issues`,
    { params: { projectID, ...(severity ? { severity } : {}) } },
  );
}

/** 获取验收证据列表 */
export function getAcceptEvidence(projectID: string) {
  return requestClient.get<{ evidence: AcceptEvidenceItem[] }>(
    `${PREFIX}/accept-evidence`,
    { params: { projectID } },
  );
}

/** 人工放行验收 */
export function acceptApprove(projectID: string, reason?: string) {
  return requestClient.post(`${PREFIX}/accept-approve`, { projectID, reason });
}

/** 驳回验收 */
export function acceptReject(projectID: string, reason: string) {
  return requestClient.post(`${PREFIX}/accept-reject`, { projectID, reason });
}

/** 重新验收 */
export function acceptRerun(projectID: string) {
  return requestClient.post(`${PREFIX}/accept-rerun`, { projectID });
}

/** 驳回并返工 */
export function acceptRework(projectID: string, reason: string) {
  return requestClient.post(`${PREFIX}/accept-rework`, { projectID, reason });
}

// ==================== 自治管理 ====================

/** @deprecated 旧自治决策条目，已迁移至 L3.5 自治中台 */
export interface AutonomyDecisionItem {
  id: string;
  decisionType: string;
  triggerSource: string;
  triggerContext?: string;
  recommendation: string;
  decisionMode: string;
  humanAction: string;
  executedAt?: string;
  result?: string;
  createdAt: string;
}

/** L3.5 自治中台：人工节点条目 */
export interface AutonomyCheckpointItem {
  id: string;
  workflowRunId: string;
  projectId: string;
  decisionActionId: string;
  checkpointType: string;
  title: string;
  description?: string;
  status: string;
  assignedTo?: string;
  handledBy?: string;
  handleAction?: string;
  handleReason?: string;
  handledAt?: string;
  expiresAt?: string;
  createdAt: string;
}

/** L3.5 自治中台：决策动作条目 */
export interface AutonomyActionItem {
  id: string;
  workflowRunId: string;
  projectId: string;
  decisionType: string;
  decisionLevel: string;
  triggerSource: string;
  triggerContext?: string;
  matchedRuleId?: string;
  matchedGateIds?: string;
  actionType: string;
  recommendation?: string;
  finalAction?: string;
  actionStatus: string;
  autoExecutable: number;
  humanRequired: number;
  executedAt?: string;
  result?: string;
  createdAt: string;
}

/** 项目报告条目 */
export interface ProjectReportItem {
  id: string;
  reportType: string;
  stageType?: string;
  title: string;
  content: string;
  metrics?: string;
  createdAt: string;
}

/** @deprecated 旧接口，已迁移至 getAutonomyCheckpoints */
export function getAutonomyDecisions(projectID: string, decisionType?: string) {
  return requestClient.get<{ decisions: AutonomyDecisionItem[] }>(
    `${PREFIX}/autonomy-decisions`,
    { params: { projectID, ...(decisionType ? { decisionType } : {}) } },
  );
}

/** @deprecated 旧接口，已迁移至 autonomyApprove */
export function approveDecision(projectID: string, decisionID: string) {
  return requestClient.post(`${PREFIX}/approve-decision`, { projectID, decisionID });
}

/** @deprecated 旧接口，已迁移至 autonomyReject */
export function rejectDecision(projectID: string, decisionID: string) {
  return requestClient.post(`${PREFIX}/reject-decision`, { projectID, decisionID });
}

// ==================== L3.5 自治中台 ====================

/** 查询项目待处理人工节点 + 关联决策动作 */
export function getAutonomyCheckpoints(projectID: string) {
  return requestClient.get<{ checkpoints: AutonomyCheckpointItem[]; actions: AutonomyActionItem[] }>(
    `${PREFIX}/autonomy-checkpoints`,
    { params: { projectID } },
  );
}

/** 审批通过决策动作 */
export function autonomyApprove(actionID: string) {
  return requestClient.post(`${PREFIX}/autonomy-approve`, { actionID });
}

/** 驳回决策动作 */
export function autonomyReject(actionID: string, reason: string) {
  return requestClient.post(`${PREFIX}/autonomy-reject`, { actionID, reason });
}

/** 查询全量决策记录（审计历史） */
export function getAutonomyActions(projectID: string, actionStatus?: string, decisionType?: string) {
  return requestClient.get<{ actions: AutonomyActionItem[] }>(
    `${PREFIX}/autonomy-actions`,
    { params: { projectID, ...(actionStatus ? { actionStatus } : {}), ...(decisionType ? { decisionType } : {}) } },
  );
}

/** 风险闸门规则条目 */
export interface RiskGateRuleItem {
  id: string;
  gateCode: string;
  gateName: string;
  gateType: string;
  projectFamily?: string;
  projectCategoryCode?: string;
  triggerExpression?: string;
  blockAction: string;
  fallbackAction?: string;
  enabled: number;
  priority: number;
  createdAt: string;
}

/** 策略规则条目 */
export interface PolicyRuleItem {
  id: string;
  ruleCode: string;
  ruleName: string;
  decisionType: string;
  decisionLevel: string;
  triggerSource: string;
  projectFamily?: string;
  projectCategoryCode?: string;
  configJson?: string;
  enabled: number;
  priority: number;
  createdAt: string;
}

/** 查询风险闸门规则 */
export function getAutonomyGateRules(projectID: string) {
  return requestClient.get<{ rules: RiskGateRuleItem[] }>(
    `${PREFIX}/autonomy-gate-rules`,
    { params: { projectID } },
  );
}

/** 查询策略规则 */
export function getAutonomyPolicyRules(projectID: string) {
  return requestClient.get<{ rules: PolicyRuleItem[] }>(
    `${PREFIX}/autonomy-policy-rules`,
    { params: { projectID } },
  );
}

// ==================== 飞书协作管理 ====================

export interface FeishuConfigItem {
  enabled: number;
  appId: string;
  appSecret: string;
  verificationToken: string;
  encryptKey: string;
  defaultNotifyUserIds: string;
  connectionMode: string; // webhook | websocket
  callbackPath: string;
  wsRunning?: boolean; // WebSocket 长连接是否在线
}

export interface FeishuBindingItem {
  id: string;
  userId: string;
  platform: string;
  platformUserId: string;
  platformName?: string;
  createdBy: string;
  deptId: string;
  createdAt?: string;
  updatedAt?: string;
}

export function getFeishuConfig() {
  return requestClient.get<{ config: FeishuConfigItem }>(`${PREFIX}/feishu-config`);
}

export function saveFeishuConfig(data: FeishuConfigItem) {
  return requestClient.post(`${PREFIX}/save-feishu-config`, data);
}

export function getFeishuBindings() {
  return requestClient.get<{ bindings: FeishuBindingItem[] }>(`${PREFIX}/feishu-bindings`);
}

export function bindFeishuUser(data: {
  userId: string;
  platformUserId: string;
  platformName?: string;
}) {
  return requestClient.post<{ id: string }>(`${PREFIX}/bind-feishu-user`, data);
}

export function unbindFeishuUser(bindingId: string) {
  return requestClient.post(`${PREFIX}/unbind-feishu-user`, { bindingId });
}

export function testFeishuMessage(data: { bindingId: string; content?: string }) {
  return requestClient.post(`${PREFIX}/test-feishu-message`, data);
}

/** 手动触发重规划 */
export function triggerReplan(projectID: string) {
  return requestClient.post(`${PREFIX}/trigger-replan`, { projectID });
}

/** 获取项目报告列表 */
export function getProjectReports(projectID: string, reportType?: string) {
  return requestClient.get<{ reports: ProjectReportItem[] }>(
    `${PREFIX}/project-reports`,
    { params: { projectID, ...(reportType ? { reportType } : {}) } },
  );
}

/** 手动生成报告 */
export function triggerReport(projectID: string, stageType?: string) {
  return requestClient.post(`${PREFIX}/trigger-report`, { projectID, ...(stageType ? { stageType } : {}) });
}

// ==================== 自治模式配置 ====================

/** 查询当前自治模式 */
export function getAutonomyMode() {
  return requestClient.get<{ mode: string }>(`${PREFIX}/autonomy-mode`);
}

/** 设置自治模式 */
export function setAutonomyMode(mode: 'auto' | 'suggest') {
  return requestClient.post(`${PREFIX}/set-autonomy-mode`, { mode });
}

// ==================== 项目列表批量状态 ====================

export interface ProjectRuntimeStat {
  projectID: string;
  currentStage: string;
  totalTasks: number;
  completedTasks: number;
  failedTasks: number;
  runningTasks: number;
}

/** 批量查询项目运行时统计 */
export function batchProjectStats(projectIDs: string[]) {
  return requestClient.post<{ stats: ProjectRuntimeStat[] }>(
    `${PREFIX}/batch-project-stats`,
    { projectIDs },
  );
}

// ==================== 态势感知（Phase A）====================

export interface ProgressMetrics {
  totalTasks: number;
  completedTasks: number;
  runningTasks: number;
  failedTasks: number;
  pendingTasks: number;
  completionRate: number;
  currentBatchNo: number;
  totalBatches: number;
  batchProgress: number;
}

export interface HealthMetrics {
  consecutiveFailures: number;
  recentFailureRate: number;
  avgTaskDuration: number;
  medianTaskDuration: number;
  retryCount: number;
  escalationCount: number;
  reworkRounds: number;
  acceptRounds: number;
  replanCount: number;
  staleTaskCount: number;
}

export interface ResourceMetrics {
  runningConcurrency: number;
  maxConcurrency: number;
  resourceUtilization: number;
  lockedResourceCount: number;
  conflictCount: number;
  tokensConsumed: number;
  estimatedTokensLeft: number;
}

export interface TrendMetrics {
  failureRateTrend: string; // rising / falling / stable
  durationTrend: string;
  throughputTrend: string;
}

export interface AnomalySignal {
  type: string;
  severity: string; // critical / warning / info
  message: string;
  confidence: number;
}

export interface SituationData {
  workflowRunId: string;
  projectId: string;
  activeStage: string;
  workflowStatus: string;
  workflowStartedAt?: string;
  snapshotAt: string;
  progress?: ProgressMetrics;
  health?: HealthMetrics;
  resource?: ResourceMetrics;
  trend?: TrendMetrics;
  anomalySignals?: AnomalySignal[];
}

export interface SituationSnapshot {
  id: string;
  projectId: string;
  workflowRunId: string;
  snapshotAt: string;
  progress?: ProgressMetrics;
  health?: HealthMetrics;
  resource?: ResourceMetrics;
  trend?: TrendMetrics;
}

/** 查询当前态势 */
export function getSituation(workflowRunID: string) {
  return requestClient.get<{ situation: SituationData }>(
    `${PREFIX}/situation`,
    { params: { workflowRunID } },
  );
}

/** 查询态势快照历史 */
export function getSituationHistory(params: {
  projectID: string;
  workflowRunID?: string;
  limit?: number;
}) {
  return requestClient.get<{ snapshots: SituationSnapshot[] }>(
    `${PREFIX}/situation-history`,
    { params },
  );
}

// ==================== 目标层管理（Phase A）====================

export interface ObjectiveData {
  projectId: string;
  deliveryGoal: string;
  qualityFloor: number;
  tokenBudget: number;
  timeBudgetHours: number;
  costBudgetCents: number;
  riskTolerance: string;   // low / medium / high
  maxAutoRetries: number;
  maxAutoReworks: number;
  maxAutoReplans: number;
  deadlineAt?: string;
  maxStallMinutes: number;
  autonomyLevel: string;   // supervised / assisted / full_auto
  maxSideEffectLevel: string; // none / reversible / irreversible
}

/** 查询项目目标约束 */
export function getObjective(projectID: string) {
  return requestClient.get<{ objective: ObjectiveData }>(
    `${PREFIX}/objective`,
    { params: { projectID } },
  );
}

/** 保存项目目标约束 */
export function saveObjective(data: { projectID: string } & Partial<ObjectiveData>) {
  return requestClient.post(`${PREFIX}/save-objective`, data);
}

// ==================== 元认知（Phase D / L7）====================

export interface ObservationRecord {
  id: string;
  decisionActionId: string;
  workflowRunId: string;
  projectId: string;
  decisionType: string;
  triggerSource: string;
  decisionLevel: string;
  actionType: string;
  outcome: string;
  effectScore: number;
  humanOverride: number;
  overrideReason: string;
  signalWeight: number;
  createdAt: string;
}

export interface ObservationStats {
  total: number;
  outcomeDistribution: Record<string, number>;
  levelDistribution: Record<string, number>;
  humanOverrideCount: number;
  humanOverrideRate: number;
}

export interface AssessmentData {
  id: string;
  projectId: string;
  periodStart: string;
  periodEnd: string;
  sampleCount: number;
  policyAccuracy: number;
  gateFalsePositive: number;
  gateFalseNegative: number;
  humanOverrideRate: number;
  matchAccuracy: number;
  costEfficiency: number;
  drifts: DriftItem[];
  summary: string;
}

export interface DriftItem {
  parameter: string;
  currentValue: number;
  optimalValue: number;
  confidence: number;
  evidence: string;
}

export interface TuneRecommendationItem {
  id: string;
  assessmentId: string;
  projectId: string;
  parameter: string;
  currentValue: string;
  suggestedValue: string;
  direction: string;
  reasoning: string;
  confidence: number;
  autoApplicable: boolean;
  riskLevel: string;
  status: string;
  appliedAt?: string;
  appliedBy?: string;
}

export interface LearningRecordItem {
  id: string;
  metricKey: string;
  projectId: string;
  emaValue: number;
  rawValue: number;
  sampleCount: number;
  lastUpdated: string;
  decayFactor: number;
}

/** 查询决策观测记录 */
export function getMetaObservations(projectID: string, limit?: number) {
  return requestClient.get<{ observations: ObservationRecord[] }>(
    `${PREFIX}/meta/observations`,
    { params: { projectID, ...(limit ? { limit } : {}) } },
  );
}

/** 查询观测统计 */
export function getMetaObservationStats(projectID: string) {
  return requestClient.get<{ stats: ObservationStats }>(
    `${PREFIX}/meta/observation-stats`,
    { params: { projectID } },
  );
}

/** 查询最新评估结果 */
export function getMetaAssessment(projectID: string) {
  return requestClient.get<{ assessment: AssessmentData }>(
    `${PREFIX}/meta/assessment`,
    { params: { projectID } },
  );
}

/** 查询评估历史 */
export function getMetaAssessmentHistory(projectID: string, limit?: number) {
  return requestClient.get<{ assessments: AssessmentData[] }>(
    `${PREFIX}/meta/assessment-history`,
    { params: { projectID, ...(limit ? { limit } : {}) } },
  );
}

/** 手动触发评估 */
export function runMetaAssessment(projectID: string, days?: number) {
  return requestClient.post<{ assessment: AssessmentData }>(
    `${PREFIX}/meta/run-assessment`,
    { projectID, ...(days ? { days } : {}) },
  );
}

/** 查询调参建议 */
export function getMetaRecommendations(projectID?: string, status?: string, limit?: number) {
  return requestClient.get<{ recommendations: TuneRecommendationItem[] }>(
    `${PREFIX}/meta/recommendations`,
    { params: { ...(projectID ? { projectID } : {}), ...(status ? { status } : {}), ...(limit ? { limit } : {}) } },
  );
}

/** 应用一条调参建议 */
export function applyMetaRecommendation(recommendationID: string) {
  return requestClient.post(`${PREFIX}/meta/apply-recommendation`, { recommendationID });
}

/** 驳回一条调参建议 */
export function rejectMetaRecommendation(recommendationID: string) {
  return requestClient.post(`${PREFIX}/meta/reject-recommendation`, { recommendationID });
}

/** 查询 EMA 学习记录 */
export function getMetaLearning(projectID: string) {
  return requestClient.get<{ records: LearningRecordItem[] }>(
    `${PREFIX}/meta/learning`,
    { params: { projectID } },
  );
}
