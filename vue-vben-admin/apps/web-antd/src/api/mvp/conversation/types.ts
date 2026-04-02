/** MVP对话表类型定义 */

/** MVP对话表项 */
export interface ConversationItem {
  id: string;
  projectID: string;
  projectName?: string;
  taskID?: string;
  taskName?: string;
  title?: string;
  roleType: string;
  status?: string;
  createdAt?: string;
  updatedAt?: string;
}

/** MVP对话表列表查询参数 */
export interface ConversationListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
  title?: string;
}

/** MVP对话表创建参数 */
export interface ConversationCreateParams {
  projectID: string;
  taskID?: string;
  title?: string;
  roleType: string;
  status?: string;
}

/** MVP对话表更新参数 */
export interface ConversationUpdateParams {
  id: string;
  projectID: string;
  taskID?: string;
  title?: string;
  roleType: string;
  status?: string;
}
