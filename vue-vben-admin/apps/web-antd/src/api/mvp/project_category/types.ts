/** 项目分类类型定义 */

/** 项目分类项 */
export interface ProjectCategoryItem {
  id: string;
  categoryCode?: string;
  displayName?: string;
  familyCode?: string;
  description?: string;
  status?: number;
  sort?: number;
  createdBy?: string;
  deptId?: string;
  createdAt?: string;
  updatedAt?: string;
}

/** 项目分类列表查询参数 */
export interface ProjectCategoryListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
  displayName?: string;
  categoryCode?: string;
  status?: number;
}

/** 项目分类创建参数 */
export interface ProjectCategoryCreateParams {
  categoryCode: string;
  displayName: string;
  familyCode?: string;
  description?: string;
  status?: number;
  sort?: number;
}

/** 项目分类更新参数 */
export interface ProjectCategoryUpdateParams {
  id: string;
  categoryCode?: string;
  displayName?: string;
  familyCode?: string;
  description?: string;
  status?: number;
  sort?: number;
}
