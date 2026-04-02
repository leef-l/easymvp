/** 菜单表类型定义 */

/** 菜单表项 */
export interface MenuItem {
  id: string;
  parentID?: string;
  title: string;
  type?: number;
  path?: string;
  component?: string;
  permission?: string;
  icon?: string;
  sort?: number;
  isShow?: number;
  isCache?: number;
  linkURL?: string;
  status?: number;
  createdAt?: string;
  updatedAt?: string;
  children?: MenuItem[];
}

/** 菜单表列表查询参数 */
export interface MenuListParams {
  pageNum: number;
  pageSize: number;
  type?: number;
  isShow?: number;
  isCache?: number;
  status?: number;
}

/** 菜单表创建参数 */
export interface MenuCreateParams {
  parentID?: string;
  title: string;
  type?: number;
  path?: string;
  component?: string;
  permission?: string;
  icon?: string;
  sort?: number;
  isShow?: number;
  isCache?: number;
  linkURL?: string;
  status?: number;
}

/** 菜单表更新参数 */
export interface MenuUpdateParams {
  id: string;
  parentID?: string;
  title: string;
  type?: number;
  path?: string;
  component?: string;
  permission?: string;
  icon?: string;
  sort?: number;
  isShow?: number;
  isCache?: number;
  linkURL?: string;
  status?: number;
}
