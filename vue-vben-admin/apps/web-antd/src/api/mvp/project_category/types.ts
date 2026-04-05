/** 项目分类配置表类型定义 */

/** 项目分类配置表项 */
export interface ProjectCategoryItem {
  id: string;
  categoryCode: string;
  displayName: string;
  familyCode: string;
  description?: string;
  status?: number;
  sort?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 项目分类配置表列表查询参数 */
export interface ProjectCategoryListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
  displayName?: string;
}

/** 项目分类配置表创建参数 */
export interface ProjectCategoryCreateParams {
  categoryCode: string;
  displayName: string;
  familyCode: string;
  description?: string;
  status?: number;
  sort?: number;
}

/** 项目分类配置表更新参数 */
export interface ProjectCategoryUpdateParams {
  id: string;
  categoryCode: string;
  displayName: string;
  familyCode: string;
  description?: string;
  status?: number;
  sort?: number;
}
