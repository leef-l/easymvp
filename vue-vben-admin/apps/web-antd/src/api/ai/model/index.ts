import { requestClient } from '#/api/request';

import type {
  ModelItem,
  ModelListParams,
  ModelCreateParams,
  ModelUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/ai/model';

/** 获取AI模型表列表 */
export function getModelList(params: ModelListParams) {
  return requestClient.get<{ list: ModelItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取AI模型表详情 */
export function getModelDetail(id: string) {
  return requestClient.get<ModelItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建AI模型表 */
export function createModel(data: ModelCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新AI模型表 */
export function updateModel(data: ModelUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除AI模型表 */
export function deleteModel(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除AI模型表 */
export function batchDeleteModel(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出AI模型表 */
export function exportModel(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 导入AI模型表 */
export function importModel(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载AI模型表导入模板 */
export function downloadImportTemplateModel() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}

/** 批量编辑AI模型表 */
export function batchUpdateModel(data: { ids: string[]; status?: number | string }) {
  return requestClient.put(`${PREFIX}/batch-update`, data);
}
