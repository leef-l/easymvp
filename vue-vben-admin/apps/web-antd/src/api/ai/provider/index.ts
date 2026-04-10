import { requestClient } from '#/api/request';

import type {
  ProviderItem,
  ProviderListParams,
  ProviderCreateParams,
  ProviderUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/ai/provider';

/** 获取AI供应商表列表 */
export function getProviderList(params: ProviderListParams) {
  return requestClient.get<{ list: ProviderItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取AI供应商表详情 */
export function getProviderDetail(id: string) {
  return requestClient.get<ProviderItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建AI供应商表 */
export function createProvider(data: ProviderCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新AI供应商表 */
export function updateProvider(data: ProviderUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除AI供应商表 */
export function deleteProvider(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除AI供应商表 */
export function batchDeleteProvider(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出AI供应商表 */
export function exportProvider(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 导入AI供应商表 */
export function importProvider(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载AI供应商表导入模板 */
export function downloadImportTemplateProvider() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}

/** 批量编辑AI供应商表 */
export function batchUpdateProvider(data: { ids: string[]; status?: number | string }) {
  return requestClient.put(`${PREFIX}/batch-update`, data);
}
