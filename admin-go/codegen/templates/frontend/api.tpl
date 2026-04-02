import { requestClient } from '#/api/request';

import type {
  {{.ModelName}}Item,
  {{.ModelName}}ListParams,
  {{.ModelName}}CreateParams,
  {{.ModelName}}UpdateParams,{{if .HasParentID}}
  {{.ModelName}}TreeParams,{{end}}
} from './types';

/** API 前缀 */
const PREFIX = '/{{.AppName}}/{{.ModuleName}}';

/** 获取{{.Comment}}列表 */
export function get{{.ModelName}}List(params: {{.ModelName}}ListParams) {
  return requestClient.get<{ list: {{.ModelName}}Item[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取{{.Comment}}详情 */
export function get{{.ModelName}}Detail(id: string) {
  return requestClient.get<{{.ModelName}}Item>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建{{.Comment}} */
export function create{{.ModelName}}(data: {{.ModelName}}CreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新{{.Comment}} */
export function update{{.ModelName}}(data: {{.ModelName}}UpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除{{.Comment}} */
export function delete{{.ModelName}}(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 批量删除{{.Comment}} */
export function batchDelete{{.ModelName}}(ids: string[]) {
  return requestClient.delete(`${PREFIX}/batch-delete`, { data: { ids } });
}

/** 导出{{.Comment}} */
export function export{{.ModelName}}(params?: Record<string, any>) {
  return requestClient.get(`${PREFIX}/export`, {
    params,
    responseType: 'blob',
  });
}
{{- if .HasParentID}}

/** 获取{{.Comment}}树形结构 */
export async function get{{.ModelName}}Tree(params?: {{.ModelName}}TreeParams) {
  const res = await requestClient.get<{ list: {{.ModelName}}Item[] }>(`${PREFIX}/tree`, { params });
  return res?.list ?? [];
}
{{- end}}
{{- if .HasImport}}

/** 导入{{.Comment}} */
export function import{{.ModelName}}(data: FormData) {
  return requestClient.post<{ success: number; fail: number }>(
    `${PREFIX}/import`,
    data,
  );
}

/** 下载{{.Comment}}导入模板 */
export function downloadImportTemplate{{.ModelName}}() {
  return requestClient.get(`${PREFIX}/import-template`, {
    responseType: 'blob',
  });
}
{{- end}}
{{- if .HasBatchEdit}}

/** 批量编辑{{.Comment}} */
export function batchUpdate{{.ModelName}}(data: { ids: string[]; status?: number }) {
  return requestClient.put(`${PREFIX}/batch-update`, data);
}
{{- end}}
