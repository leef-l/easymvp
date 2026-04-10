import { requestClient } from '#/api/request';

import type {
  MessageItem,
  MessageListParams,
  MessageCreateParams,
  MessageUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/mvp/message';

/** 获取MVP消息表列表 */
export function getMessageList(params: MessageListParams) {
  return requestClient.get<{ list: MessageItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取MVP消息表详情 */
export function getMessageDetail(id: string) {
  return requestClient.get<MessageItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建MVP消息表 */
export function createMessage(data: MessageCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新MVP消息表 */
export function updateMessage(data: MessageUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除MVP消息表 */
export function deleteMessage(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除MVP消息表 */
export function batchDeleteMessage(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出MVP消息表 */
export function exportMessage(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 导入MVP消息表 */
export function importMessage(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载MVP消息表导入模板 */
export function downloadImportTemplateMessage() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}

/** 批量编辑MVP消息表 */
export function batchUpdateMessage(data: { ids: string[]; status?: number | string }) {
  return requestClient.put(`${PREFIX}/batch-update`, data);
}
