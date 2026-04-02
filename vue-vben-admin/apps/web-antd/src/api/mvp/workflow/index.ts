import { requestClient } from '#/api/request';

/** API 前缀 */
const PREFIX = '/mvp/workflow';

/** 创建项目（通过工作流，返回项目ID和对话ID） */
export function createProject(data: {
  name: string;
  description: string;
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

/** 获取项目当前状态及任务统计 */
export function getProjectStatus(projectID: string) {
  return requestClient.get<{
    status: string;
    pauseReason?: string;
    totalTasks: number;
    statusCounts: Record<string, number>;
  }>(`${PREFIX}/project-status`, { params: { projectID } });
}
