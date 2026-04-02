/** 项目角色配置表类型定义 */

/** 项目角色配置表项 */
export interface ProjectRoleItem {
  id: string;
  projectID: string;
  projectName?: string;
  roleType: string;
  roleLevel?: string;
  modelID: string;
  systemPrompt?: string;
  status?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 项目角色配置表列表查询参数 */
export interface ProjectRoleListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
  status?: number;
}

/** 项目角色配置表创建参数 */
export interface ProjectRoleCreateParams {
  projectID: string;
  roleType: string;
  roleLevel?: string;
  modelID: string;
  systemPrompt?: string;
  status?: number;
}

/** 项目角色配置表更新参数 */
export interface ProjectRoleUpdateParams {
  id: string;
  projectID: string;
  roleType: string;
  roleLevel?: string;
  modelID: string;
  systemPrompt?: string;
  status?: number;
}
