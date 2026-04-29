package v1

import "github.com/gogf/gf/v2/frame/g"

type SendMessageReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/architect-chat/messages" tags:"ArchitectChat" method:"post" summary:"Send message to architect"`
	ProjectID string `json:"id" in:"path" v:"required"`
	Content   string `json:"content" v:"required"`
}

type SendMessageRes struct {
	CommandID      string `json:"command_id"`
	MessageID      string `json:"message_id"`
	ArchitectReply string `json:"architect_reply"`
	Accepted       bool   `json:"accepted"`
}

type ListMessagesReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/architect-chat/messages" tags:"ArchitectChat" method:"get" summary:"List architect chat messages"`
	ProjectID string `json:"id" in:"path" v:"required"`
}

type ListMessagesRes struct {
	Messages []MessageItem `json:"messages"`
}

type GetConversationReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/architect-chat" tags:"ArchitectChat" method:"get" summary:"Get architect conversation"`
	ProjectID string `json:"id" in:"path" v:"required"`
}

type GetConversationRes struct {
	ConversationID string        `json:"conversation_id"`
	Status         string        `json:"status"`
	PlanDraftID    string        `json:"plan_draft_id"`
	Messages       []MessageItem `json:"messages"`
}

type ConfirmPlanReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/architect-chat/confirm-plan" tags:"ArchitectChat" method:"post" summary:"Confirm architect plan"`
	ProjectID string `json:"id" in:"path" v:"required"`
}

type ConfirmPlanRes struct {
	CommandID      string `json:"command_id"`
	Accepted       bool   `json:"accepted"`
	CompiledPlanID string `json:"compiled_plan_id"`
	Reason         string `json:"reason"`
	ReviewID       string `json:"review_id"`
}

type MessageItem struct {
	ID          string `json:"id"`
	SenderRole  string `json:"sender_role"`
	SenderName  string `json:"sender_name"`
	Content     string `json:"content"`
	MessageKind string `json:"message_kind"`
	CreatedAt   string `json:"created_at"`
}
