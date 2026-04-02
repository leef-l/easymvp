import { requestClient } from '#/api/request';

/** API 前缀 */
const PREFIX = '/mvp/chat';

/** 聊天消息类型 */
export interface ChatMessage {
  id: string;
  role: 'assistant' | 'user';
  content: string;
  status: 'completed' | 'failed' | 'streaming';
  modelName?: string;
  createdAt: string;
}

/** 发送消息请求参数 */
export interface SendMessageParams {
  conversationID: string;
  content: string;
}

/** 发送消息响应 */
export interface SendMessageResult {
  messageID: string;
  replyID: string;
}

/** 发送消息 */
export function sendMessage(data: SendMessageParams) {
  return requestClient.post<SendMessageResult>(`${PREFIX}/send`, data);
}

/** 获取对话历史 */
export function getChatHistory(conversationID: string) {
  return requestClient.get<{ list: ChatMessage[] }>(`${PREFIX}/history`, {
    params: { conversationID },
  });
}
