import { apiGet, apiPost, getCoreBaseUrl } from "@/shared/lib/api";

export type MessageItem = {
  id: string;
  sender_role: string;
  sender_name: string;
  content: string;
  message_kind: string;
  created_at: string;
};

export type Conversation = {
  id: string;
  project_id: string;
  conversation_kind: string;
  status: string;
  plan_draft_id: string;
  messages: MessageItem[];
};

export type StreamEvent = {
  type: string;
  content?: string;
  tool_name?: string;
  detail?: string;
  done?: boolean;
  error?: string;
  command_id?: string;
  message_id?: string;
};

export async function getConversation(projectId: string): Promise<Conversation> {
  return apiGet<Conversation>(`/api/v3/projects/${encodeURIComponent(projectId)}/architect-chat`);
}

export async function listMessages(projectId: string): Promise<MessageItem[]> {
  const res = await apiGet<{ messages: MessageItem[] }>(
    `/api/v3/projects/${encodeURIComponent(projectId)}/architect-chat/messages`
  );
  return res.messages ?? [];
}

export async function sendMessage(projectId: string, content: string): Promise<{ message_id: string; reply_id: string }> {
  return apiPost<{ message_id: string; reply_id: string }>(
    `/api/v3/projects/${encodeURIComponent(projectId)}/architect-chat/messages`,
    { project_id: projectId, content }
  );
}

export async function* sendMessageStream(projectId: string, content: string): AsyncGenerator<StreamEvent, void, unknown> {
  const response = await fetch(
    `${getCoreBaseUrl()}/api/v3/projects/${encodeURIComponent(projectId)}/architect-chat/messages/stream`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "text/event-stream",
      },
      body: JSON.stringify({ project_id: projectId, content }),
    }
  );

  if (!response.ok || !response.body) {
    const text = await response.text();
    throw new Error(text || `HTTP ${response.status}`);
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });

    const lines = buffer.split("\n");
    buffer = lines.pop() ?? "";

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line.startsWith("data: ")) continue;
      const data = line.slice(6);
      if (data === "[DONE]") return;
      try {
        const ev: StreamEvent = JSON.parse(data);
        yield ev;
        if (ev.done) return;
      } catch {
        // ignore malformed JSON lines
      }
    }
  }
}

export async function confirmPlan(projectId: string): Promise<{ accepted: boolean; compiled_plan_id?: string; reason?: string; review_id?: string }> {
  return apiPost<{ accepted: boolean; compiled_plan_id?: string; reason?: string; review_id?: string }>(
    `/api/v3/projects/${encodeURIComponent(projectId)}/architect-chat/confirm-plan`,
    { project_id: projectId }
  );
}
