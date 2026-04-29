package architect_chat

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/architect_chat/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) ListMessages(ctx context.Context, req *v1.ListMessagesReq) (res *v1.ListMessagesRes, err error) {
	conv, err := service.ArchitectChat().GetConversationByProject(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}
	messages := make([]v1.MessageItem, 0, len(conv.Messages))
	for _, m := range conv.Messages {
		messages = append(messages, v1.MessageItem{
			ID:          m.ID,
			SenderRole:  m.SenderRole,
			SenderName:  m.SenderName,
			Content:     m.Content,
			MessageKind: m.MessageKind,
			CreatedAt:   m.CreatedAt,
		})
	}
	return &v1.ListMessagesRes{Messages: messages}, nil
}
