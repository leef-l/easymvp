/** MVP配置表类型定义 */

/** MVP配置表项 */
export interface ConfigItem {
  id: string;
  configKey: string;
  configValue: string;
  configType?: string;
  category?: string;
  description?: string;
  createdAt?: string;
  updatedAt?: string;
}

/** MVP配置表列表查询参数 */
export interface ConfigListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
}

/** MVP配置表创建参数 */
export interface ConfigCreateParams {
  configKey?: string;
  configValue?: string;
  configType?: string;
  category?: string;
  description?: string;
}

/** MVP配置表更新参数 */
export interface ConfigUpdateParams {
  id: string;
  configKey?: string;
  configValue?: string;
  configType?: string;
  category?: string;
  description?: string;
}
