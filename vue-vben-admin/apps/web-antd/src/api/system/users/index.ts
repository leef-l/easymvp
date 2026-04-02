import { requestClient } from '#/api/request';

import type {
  UsersItem,
  UsersListParams,
  UsersCreateParams,
  UsersUpdateParams,
  UsersResetPasswordParams,
} from './types';

/** API 前缀 */
const PREFIX = '/system/users';

/** 获取用户表列表 */
export function getUsersList(params: UsersListParams) {
  return requestClient.get<{ list: UsersItem[]; total: number }>(
    `${PREFIX}/list`,
    { params },
  );
}

/** 获取用户表详情 */
export function getUsersDetail(id: string) {
  return requestClient.get<UsersItem>(`${PREFIX}/detail`, {
    params: { id },
  });
}

/** 创建用户表 */
export function createUsers(data: UsersCreateParams) {
  return requestClient.post(`${PREFIX}/create`, data);
}

/** 更新用户表 */
export function updateUsers(data: UsersUpdateParams) {
  return requestClient.put(`${PREFIX}/update`, data);
}

/** 删除用户表 */
export function deleteUsers(id: string) {
  return requestClient.delete(`${PREFIX}/delete`, { data: { id } });
}

/** 重置用户密码 */
export function resetUsersPassword(data: UsersResetPasswordParams) {
  return requestClient.put(`${PREFIX}/reset-password`, data);
}
