import { requestClient } from '#/api/request';

import type {
  PlanItem,
  PlanListParams,
  PlanCreateParams,
  PlanUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/ai/plan';

/** 获取AI套餐表列表 */
export function getPlanList(params: PlanListParams) {
  return requestClient.get<{ list: PlanItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取AI套餐表详情 */
export function getPlanDetail(id: string) {
  return requestClient.get<PlanItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建AI套餐表 */
export function createPlan(data: PlanCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新AI套餐表 */
export function updatePlan(data: PlanUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除AI套餐表 */
export function deletePlan(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除AI套餐表 */
export function batchDeletePlan(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出AI套餐表 */
export function exportPlan(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}

/** 导入AI套餐表 */
export function importPlan(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载AI套餐表导入模板 */
export function downloadImportTemplatePlan() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}

/** 批量编辑AI套餐表 */
export function batchUpdatePlan(data: { ids: string[]; status?: number }) {
  return requestClient.put(`${PREFIX}/batch-update`, data);
}
