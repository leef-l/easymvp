import { requestClient } from '#/api/request';

import type {
  ConfigItem,
  ConfigListParams,
  ConfigCreateParams,
  ConfigUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/mvp/config';

/** 获取MVP配置表列表 */
export function getConfigList(params: ConfigListParams) {
  return requestClient.get<{ list: ConfigItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取MVP配置表详情 */
export function getConfigDetail(id: string) {
  return requestClient.get<ConfigItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建MVP配置表 */
export function createConfig(data: ConfigCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新MVP配置表 */
export function updateConfig(data: ConfigUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除MVP配置表 */
export function deleteConfig(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除MVP配置表 */
export function batchDeleteConfig(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出MVP配置表 */
export function exportConfig(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 导入MVP配置表 */
export function importConfig(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载MVP配置表导入模板 */
export function downloadImportTemplateConfig() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}
