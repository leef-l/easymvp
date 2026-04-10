import { requestClient } from '#/api/request';

import type {
  ConversationItem,
  ConversationListParams,
  ConversationCreateParams,
  ConversationUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/mvp/conversation';

/** 获取MVP对话表列表 */
export function getConversationList(params: ConversationListParams) {
  return requestClient.get<{ list: ConversationItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取MVP对话表详情 */
export function getConversationDetail(id: string) {
  return requestClient.get<ConversationItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建MVP对话表 */
export function createConversation(data: ConversationCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新MVP对话表 */
export function updateConversation(data: ConversationUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除MVP对话表 */
export function deleteConversation(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除MVP对话表 */
export function batchDeleteConversation(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出MVP对话表 */
export function exportConversation(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 导入MVP对话表 */
export function importConversation(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载MVP对话表导入模板 */
export function downloadImportTemplateConversation() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}

/** 批量编辑MVP对话表 */
export function batchUpdateConversation(data: { ids: string[]; status?: number | string }) {
  return requestClient.put(`${PREFIX}/batch-update`, data);
}
