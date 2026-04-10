import { requestClient } from '#/api/request';

import type {
  ProjectCategoryItem,
  ProjectCategoryListParams,
  ProjectCategoryCreateParams,
  ProjectCategoryUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/mvp/project_category';

/** 获取项目分类配置表列表 */
export function getProjectCategoryList(params: ProjectCategoryListParams) {
  return requestClient.get<{ list: ProjectCategoryItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取项目分类配置表详情 */
export function getProjectCategoryDetail(id: string) {
  return requestClient.get<ProjectCategoryItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建项目分类配置表 */
export function createProjectCategory(data: ProjectCategoryCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新项目分类配置表 */
export function updateProjectCategory(data: ProjectCategoryUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除项目分类配置表 */
export function deleteProjectCategory(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除项目分类配置表 */
export function batchDeleteProjectCategory(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出项目分类配置表 */
export function exportProjectCategory(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 导入项目分类配置表 */
export function importProjectCategory(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载项目分类配置表导入模板 */
export function downloadImportTemplateProjectCategory() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}

/** 批量编辑项目分类配置表 */
export function batchUpdateProjectCategory(data: { ids: string[]; status?: number | string }) {
  return requestClient.put(`${PREFIX}/batch-update`, data);
}
