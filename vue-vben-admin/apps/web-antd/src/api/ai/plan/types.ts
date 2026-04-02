/** AI套餐表类型定义 */

/** AI套餐表项 */
export interface PlanItem {
  id: string;
  providerID: string;
  providerName?: string;
  name: string;
  code: string;
  apiKey?: string;
  apiSecret?: string;
  status?: number;
  sort?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** AI套餐表列表查询参数 */
export interface PlanListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
  status?: number;
  name?: string;
}

/** AI套餐表创建参数 */
export interface PlanCreateParams {
  providerID: string;
  name: string;
  code: string;
  apiKey?: string;
  apiSecret?: string;
  status?: number;
  sort?: number;
}

/** AI套餐表更新参数 */
export interface PlanUpdateParams {
  id: string;
  providerID: string;
  name: string;
  code: string;
  apiKey?: string;
  apiSecret?: string;
  status?: number;
  sort?: number;
}
