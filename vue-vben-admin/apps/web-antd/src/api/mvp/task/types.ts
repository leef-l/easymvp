/** MVP任务表类型定义 */

/** MVP任务表项 */
export interface TaskItem {
  id: string;
  projectID: string;
  projectName?: string;
  parentID?: string;
  taskName?: string;
  name: string;
  description?: string;
  roleType: string;
  roleLevel?: string;
  modelID?: string;
  conversationID?: string;
  status?: string;
  sort?: number;
  batchNo?: number;
  affectedResources?: string;
  dependsOn?: string;
  result?: string;
  contextSummary?: string;
  errorMessage?: string;
  startedAt?: string;
  completedAt?: string;
  createdAt?: string;
  updatedAt?: string;
  children?: TaskItem[];
}

/** MVP任务表列表查询参数 */
export interface TaskListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
  name?: string;
  projectID?: string;
  status?: string;
  batchNo?: number;
  roleType?: string;
}
/** MVP任务表树形查询参数 */
export interface TaskTreeParams {
  projectID?: string;
  name?: string;
  status?: string;
  batchNo?: number;
  roleType?: string;
  startTime?: string;
  endTime?: string;
}

/** MVP任务表创建参数 */
export interface TaskCreateParams {
  projectID: string;
  parentID?: string;
  name: string;
  description?: string;
  roleType: string;
  roleLevel?: string;
  modelID?: string;
  status?: string;
  sort?: number;
  batchNo?: number;
  affectedResources?: string;
  dependsOn?: string;
  result?: string;
  contextSummary?: string;
  errorMessage?: string;
  startedAt?: string;
  completedAt?: string;
}

/** MVP任务表更新参数 */
export interface TaskUpdateParams {
  id: string;
  projectID: string;
  parentID?: string;
  name: string;
  description?: string;
  roleType: string;
  roleLevel?: string;
  modelID?: string;
  status?: string;
  sort?: number;
  batchNo?: number;
  affectedResources?: string;
  dependsOn?: string;
  result?: string;
  contextSummary?: string;
  errorMessage?: string;
  startedAt?: string;
  completedAt?: string;
}
