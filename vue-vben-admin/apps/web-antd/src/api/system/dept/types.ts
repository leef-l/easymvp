/** 部门表类型定义 */

/** 部门表项 */
export interface DeptItem {
  id: string;
  parentID?: string;
  deptTitle?: string;
  title: string;
  username?: string;
  email?: string;
  sort?: number;
  status?: number;
  createdAt?: string;
  updatedAt?: string;
  children?: DeptItem[];
}

/** 部门表列表查询参数 */
export interface DeptListParams {
  pageNum: number;
  pageSize: number;
  status?: number;
}

/** 部门表创建参数 */
export interface DeptCreateParams {
  parentID?: string;
  title: string;
  username?: string;
  email?: string;
  sort?: number;
  status?: number;
}

/** 部门表更新参数 */
export interface DeptUpdateParams {
  id: string;
  parentID?: string;
  title: string;
  username?: string;
  email?: string;
  sort?: number;
  status?: number;
}
