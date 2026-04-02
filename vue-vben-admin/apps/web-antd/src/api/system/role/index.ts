import { requestClient } from '#/api/request';

import type {
  RoleItem,
  RoleListParams,
  RoleCreateParams,
  RoleUpdateParams,
  RoleGrantMenuParams,
  RoleGrantDeptParams,
} from './types';

/** API 前缀 */
const PREFIX = '/system/role';

/** 获取角色表列表 */
export function getRoleList(params: RoleListParams) {
  return requestClient.get<{ list: RoleItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取角色表详情 */
export function getRoleDetail(id: string) {
  return requestClient.get<RoleItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建角色表 */
export function createRole(data: RoleCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新角色表 */
export function updateRole(data: RoleUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除角色表 */
export function deleteRole(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 获取角色表树形结构 */
export async function getRoleTree(params?: Record<string, any>) {
  const res = await requestClient.get<{ list: RoleItem[] }>(
    `${PREFIX}/tree`,
    { params },
  );
  return res?.list ?? [];
}

/** 授权菜单 */
export function grantRoleMenu(data: RoleGrantMenuParams) {
  return requestClient.post(`${PREFIX}/grant-menu`, data);
}

/** 获取角色已授权菜单ID列表 */
export async function getRoleMenuIds(roleId: string) {
  const res = await requestClient.get<{ menuIds: string[] }>(
    `${PREFIX}/menu-ids`,
    { params: { id: roleId } },
  );
  return res?.menuIds ?? [];
}

/** 授权部门（数据权限） */
export function grantRoleDept(data: RoleGrantDeptParams) {
  return requestClient.post(`${PREFIX}/grant-dept`, data);
}

/** 获取角色已授权部门ID列表 */
export async function getRoleDeptIds(roleId: string) {
  const res = await requestClient.get<{ deptIds: string[] }>(
    `${PREFIX}/dept-ids`,
    { params: { id: roleId } },
  );
  return res?.deptIds ?? [];
}
