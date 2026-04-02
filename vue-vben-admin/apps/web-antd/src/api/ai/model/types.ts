/** AI模型表类型定义 */

/** AI模型表项 */
export interface ModelItem {
  id: string;
  planID: string;
  planName?: string;
  providerID: string;
  providerName?: string;
  name: string;
  modelCode: string;
  capability?: string;
  maxTokens?: number;
  contextWindow?: number;
  supportsStream?: number;
  rolePrompt?: string;
  status?: number;
  sort?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** AI模型表列表查询参数 */
export interface ModelListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
  supportsStream?: number;
  status?: number;
  name?: string;
}

/** AI模型表创建参数 */
export interface ModelCreateParams {
  planID: string;
  providerID: string;
  name: string;
  modelCode: string;
  capability?: string;
  maxTokens?: number;
  contextWindow?: number;
  supportsStream?: number;
  rolePrompt?: string;
  status?: number;
  sort?: number;
}

/** AI模型表更新参数 */
export interface ModelUpdateParams {
  id: string;
  planID: string;
  providerID: string;
  name: string;
  modelCode: string;
  capability?: string;
  maxTokens?: number;
  contextWindow?: number;
  supportsStream?: number;
  rolePrompt?: string;
  status?: number;
  sort?: number;
}
