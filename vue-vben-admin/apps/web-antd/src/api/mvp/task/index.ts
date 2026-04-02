import { requestClient } from '#/api/request';

import type {
  TaskItem,
  TaskListParams,
  TaskCreateParams,
  TaskUpdateParams,
  TaskTreeParams,
} from './types';

/** API 前缀 */
const PREFIX = '/mvp/task';

/** 获取MVP任务表列表 */
export function getTaskList(params: TaskListParams) {
  return requestClient.get<{ list: TaskItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取MVP任务表详情 */
export function getTaskDetail(id: string) {
  return requestClient.get<TaskItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建MVP任务表 */
export function createTask(data: TaskCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新MVP任务表 */
export function updateTask(data: TaskUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除MVP任务表 */
export function deleteTask(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除MVP任务表 */
export function batchDeleteTask(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出MVP任务表 */
export function exportTask(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 获取MVP任务表树形结构 */
export async function getTaskTree(params?: TaskTreeParams) {
  const res = await requestClient.get<{ list: TaskItem[] }>(`${PREFIX}/tree`, { params });
  return res?.list ?? [];
}

/** 批量编辑MVP任务表 */
export function batchUpdateTask(data: { ids: string[]; status?: number }) {
  return requestClient.put(`${PREFIX}/batch-update`, data);
}
