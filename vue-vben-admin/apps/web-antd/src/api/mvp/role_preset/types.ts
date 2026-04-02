/** 角色预设模板类型定义 */

/** 角色预设模板项 */
export interface RolePresetItem {
  id: string;
  roleType: string;
  roleLevel?: string;
  modelID: string;
  modelName?: string;
  systemPrompt?: string;
  status?: number;
  sort?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 角色预设模板列表查询参数 */
export interface RolePresetListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
  status?: number;
}

/** 角色预设模板创建参数 */
export interface RolePresetCreateParams {
  roleType: string;
  roleLevel?: string;
  modelID: string;
  systemPrompt?: string;
  status?: number;
  sort?: number;
}

/** 角色预设模板更新参数 */
export interface RolePresetUpdateParams {
  id: string;
  roleType: string;
  roleLevel?: string;
  modelID: string;
  systemPrompt?: string;
  status?: number;
  sort?: number;
}
