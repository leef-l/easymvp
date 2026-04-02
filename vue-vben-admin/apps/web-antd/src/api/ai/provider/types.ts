/** AI供应商表类型定义 */

/** AI供应商表项 */
export interface ProviderItem {
  id: string;
  name: string;
  code: string;
  providerType: string;
  baseURL: string;
  icon?: string;
  status?: number;
  sort?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** AI供应商表列表查询参数 */
export interface ProviderListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
  status?: number;
  name?: string;
}

/** AI供应商表创建参数 */
export interface ProviderCreateParams {
  name: string;
  code: string;
  providerType: string;
  baseURL: string;
  icon?: string;
  status?: number;
  sort?: number;
}

/** AI供应商表更新参数 */
export interface ProviderUpdateParams {
  id: string;
  name: string;
  code: string;
  providerType: string;
  baseURL: string;
  icon?: string;
  status?: number;
  sort?: number;
}
