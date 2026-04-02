import { requestClient } from '#/api/request';

import type {
  MenuItem,
  MenuListParams,
  MenuCreateParams,
  MenuUpdateParams,
} from './types';

/** API 前缀 */
const PREFIX = '/system/menu';

/** 获取菜单表列表 */
export function getMenuList(params: MenuListParams) {
  return requestClient.get<{ list: MenuItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取菜单表详情 */
export function getMenuDetail(id: string) {
  return requestClient.get<MenuItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建菜单表 */
export function createMenu(data: MenuCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新菜单表 */
export function updateMenu(data: MenuUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除菜单表 */
export function deleteMenu(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 获取菜单表树形结构 */
export async function getMenuTree(params?: Record<string, any>) {
  const res = await requestClient.get<{ list: MenuItem[] }>(
    `${PREFIX}/tree`,
    { params },
  );
  return res?.list ?? [];
}
