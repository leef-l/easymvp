import { requestClient } from '#/api/request';

import type {
  ProjectRoleItem,
  ProjectRoleListParams,
  ProjectRoleCreateParams,
  ProjectRoleUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/mvp/project_role';

/** 获取项目角色配置表列表 */
export function getProjectRoleList(params: ProjectRoleListParams) {
  return requestClient.get<{ list: ProjectRoleItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取项目角色配置表详情 */
export function getProjectRoleDetail(id: string) {
  return requestClient.get<ProjectRoleItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建项目角色配置表 */
export function createProjectRole(data: ProjectRoleCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新项目角色配置表 */
export function updateProjectRole(data: ProjectRoleUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除项目角色配置表 */
export function deleteProjectRole(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除项目角色配置表 */
export function batchDeleteProjectRole(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出项目角色配置表 */
export function exportProjectRole(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 导入项目角色配置表 */
export function importProjectRole(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载项目角色配置表导入模板 */
export function downloadImportTemplateProjectRole() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}

/** 批量编辑项目角色配置表 */
export function batchUpdateProjectRole(data: { ids: string[]; status?: number | string }) {
  return requestClient.put(`${PREFIX}/batch-update`, data);
}
