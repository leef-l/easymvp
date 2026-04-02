import { requestClient } from '#/api/request';

import type {
  ProjectItem,
  ProjectListParams,
  ProjectCreateParams,
  ProjectUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/mvp/project';

/** 获取MVP项目表列表 */
export function getProjectList(params: ProjectListParams) {
  return requestClient.get<{ list: ProjectItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取MVP项目表详情 */
export function getProjectDetail(id: string) {
  return requestClient.get<ProjectItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建MVP项目表 */
export function createProject(data: ProjectCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新MVP项目表 */
export function updateProject(data: ProjectUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除MVP项目表 */
export function deleteProject(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除MVP项目表 */
export function batchDeleteProject(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出MVP项目表 */
export function exportProject(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 导入MVP项目表 */
export function importProject(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载MVP项目表导入模板 */
export function downloadImportTemplateProject() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}

/** 批量编辑MVP项目表 */
export function batchUpdateProject(data: { ids: string[]; status?: number }) {
  return requestClient.put(`${PREFIX}/batch-update`, data);
}
