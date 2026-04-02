import type { UserInfo } from '@vben/types';

import { requestClient } from '#/api/request';

/**
 * 获取用户信息
 */
export async function getUserInfoApi() {
  const res = await requestClient.get<{
    userId: string;
    username: string;
    nickname: string;
    email: string;
    avatar: string;
    deptId: string;
    status: number;
    roles: string[];
    perms: string[];
  }>('/system/auth/info');

  // 转换为 Vben UserInfo 格式
  const userInfo: UserInfo = {
    userId: res.userId,
    username: res.username,
    realName: res.nickname || res.username,
    avatar: res.avatar || '',
    roles: res.roles || [],
    desc: '',
    homePath: '/dashboard',
    token: '',
  };
  return userInfo;
}
