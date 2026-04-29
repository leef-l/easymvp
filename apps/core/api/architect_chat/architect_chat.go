package architect_chat

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/architect_chat/v1"
)

type IArchitectChatV1 interface {
	SendMessage(ctx context.Context, req *v1.SendMessageReq) (res *v1.SendMessageRes, err error)
	ListMessages(ctx context.Context, req *v1.ListMessagesReq) (res *v1.ListMessagesRes, err error)
	GetConversation(ctx context.Context, req *v1.GetConversationReq) (res *v1.GetConversationRes, err error)
	ConfirmPlan(ctx context.Context, req *v1.ConfirmPlanReq) (res *v1.ConfirmPlanRes, err error)
}
