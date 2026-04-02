/** MVP消息表类型定义 */

/** MVP消息表项 */
export interface MessageItem {
  id: string;
  conversationID: string;
  conversationTitle?: string;
  role: string;
  content: string;
  modelID?: string;
  tokenUsage?: string;
  status?: string;
  createdAt?: string;
  updatedAt?: string;
}

/** MVP消息表列表查询参数 */
export interface MessageListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
}

/** MVP消息表创建参数 */
export interface MessageCreateParams {
  conversationID: string;
  role: string;
  content: string;
  modelID?: string;
  tokenUsage?: string;
  status?: string;
}

/** MVP消息表更新参数 */
export interface MessageUpdateParams {
  id: string;
  conversationID: string;
  role: string;
  content: string;
  modelID?: string;
  tokenUsage?: string;
  status?: string;
}
