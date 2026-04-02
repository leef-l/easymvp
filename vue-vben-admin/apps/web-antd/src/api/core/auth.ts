import { baseRequestClient, requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    password?: string;
    username?: string;
  }

  /** 登录接口返回值 */
  export interface LoginResult {
    token: string;
    userId: string;
    username: string;
    nickname: string;
    avatar: string;
  }
}

/**
 * 登录
 */
export async function loginApi(data: AuthApi.LoginParams) {
  return requestClient.post<AuthApi.LoginResult>(
    '/system/auth/login',
    data,
    {
      // 登录接口不需要 token
    },
  );
}

/**
 * 刷新accessToken（暂不支持，直接返回空）
 */
export async function refreshTokenApi() {
  return baseRequestClient.post<{ data: string; status: number }>(
    '/system/auth/refresh',
    { withCredentials: true },
  );
}

/**
 * 退出登录
 */
export async function logoutApi() {
  // 后端暂无 logout 接口，前端清除 token 即可
  return Promise.resolve();
}

/**
 * 获取用户权限码
 */
export async function getAccessCodesApi() {
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
  return res?.perms ?? [];
}
