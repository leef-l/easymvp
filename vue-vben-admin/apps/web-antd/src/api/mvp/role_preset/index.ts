import { requestClient } from '#/api/request';

import type {
  RolePresetItem,
  RolePresetListParams,
  RolePresetCreateParams,
  RolePresetUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/mvp/role_preset';

/** 获取角色预设模板列表 */
export function getRolePresetList(params: RolePresetListParams) {
  return requestClient.get<{ list: RolePresetItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取角色预设模板详情 */
export function getRolePresetDetail(id: string) {
  return requestClient.get<RolePresetItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建角色预设模板 */
export function createRolePreset(data: RolePresetCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新角色预设模板 */
export function updateRolePreset(data: RolePresetUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除角色预设模板 */
export function deleteRolePreset(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除角色预设模板 */
export function batchDeleteRolePreset(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出角色预设模板 */
export function exportRolePreset(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 导入角色预设模板 */
export function importRolePreset(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载角色预设模板导入模板 */
export function downloadImportTemplateRolePreset() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}

/** 批量编辑角色预设模板 */
export function batchUpdateRolePreset(data: { ids: string[]; status?: number }) {
  return requestClient.put(`${PREFIX}/batch-update`, data);
}
