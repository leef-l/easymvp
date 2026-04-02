/** 角色表类型定义 */

/** 角色表项 */
export interface RoleItem {
  id: string;
  parentID?: string;
  roleTitle?: string;
  title: string;
  dataScope?: number;
  sort?: number;
  status?: number;
  isAdmin?: number;
  createdAt?: string;
  updatedAt?: string;
  children?: RoleItem[];
}

/** 角色表列表查询参数 */
export interface RoleListParams {
  pageNum: number;
  pageSize: number;
  dataScope?: number;
  status?: number;
}

/** 角色表创建参数 */
export interface RoleCreateParams {
  parentID?: string;
  title: string;
  dataScope?: number;
  sort?: number;
  status?: number;
  isAdmin?: number;
}

/** 角色表更新参数 */
export interface RoleUpdateParams {
  id: string;
  parentID?: string;
  title: string;
  dataScope?: number;
  sort?: number;
  status?: number;
  isAdmin?: number;
}

/** 角色授权菜单参数 */
export interface RoleGrantMenuParams {
  id: string;
  menuIds: string[];
}

/** 角色授权部门参数 */
export interface RoleGrantDeptParams {
  id: string;
  dataScope: number;
  deptIds: string[];
}
