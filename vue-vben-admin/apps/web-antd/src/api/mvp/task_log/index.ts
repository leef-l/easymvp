import { requestClient } from '#/api/request';

import type {
  TaskLogItem,
  TaskLogListParams,
  TaskLogCreateParams,
  TaskLogUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/mvp/task_log';

/** 获取任务日志表列表 */
export function getTaskLogList(params: TaskLogListParams) {
  return requestClient.get<{ list: TaskLogItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取任务日志表详情 */
export function getTaskLogDetail(id: string) {
  return requestClient.get<TaskLogItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建任务日志表 */
export function createTaskLog(data: TaskLogCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新任务日志表 */
export function updateTaskLog(data: TaskLogUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除任务日志表 */
export function deleteTaskLog(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除任务日志表 */
export function batchDeleteTaskLog(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出任务日志表 */
export function exportTaskLog(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 导入任务日志表 */
export function importTaskLog(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载任务日志表导入模板 */
export function downloadImportTemplateTaskLog() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}
