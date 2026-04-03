import { requestClient } from '#/api/request';

import type { TaskExecuteParams, TaskItem, TaskListParams, TaskLogItem } from './types';

const PREFIX = '/ai/task';

export function getTaskList(params: TaskListParams) {
  return requestClient.get<{ list: TaskItem[]; total: number }>(`${PREFIX}/list`, {
    params,
  });
}

export function getTaskDetail(id: string) {
  return requestClient.get<TaskItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

export function executeTask(data: TaskExecuteParams) {
  return requestClient.post<{ taskID: string; status: string }>(`${PREFIX}/execute`, data);
}

export function getTaskLogs(taskID: string) {
  return requestClient.get<{ list: TaskLogItem[] }>(`${PREFIX}/logs`, {
    params: { taskID },
  });
}

export function cancelTask(taskID: string) {
  return requestClient.post(`${PREFIX}/cancel`, { taskID });
}
