import { requestClient } from '#/api/request';

/** API 前缀 */
const PREFIX = '/mvp/workflow';

/** 创建项目（通过工作流，返回项目ID和对话ID） */
export function createProject(data: {
  name: string;
  description: string;
  workDir: string;
  architectModelID: string;
}) {
  return requestClient.post<{ projectID: string; conversationID: string }>(
    `${PREFIX}/create-project`,
    data,
  );
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

/** 获取角色预设列表 */
export function getRolePresets() {
  return requestClient.get<{ list: RolePresetItem[] }>(
    `${PREFIX}/role-presets`,
  );
}

/** 获取项目当前状态及任务统计 */
export function getProjectStatus(projectID: string) {
  return requestClient.get<{
    status: string;
    pauseReason?: string;
    activeBatch: number;
    totalTasks: number;
    statusCounts: Record<string, number>;
  }>(`${PREFIX}/project-status`, { params: { projectID } });
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
