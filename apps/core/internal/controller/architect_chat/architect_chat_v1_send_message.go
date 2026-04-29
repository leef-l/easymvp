package architect_chat

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/architect_chat/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) SendMessage(ctx context.Context, req *v1.SendMessageReq) (res *v1.SendMessageRes, err error) {
	result, err := service.ArchitectChat().SendMessage(ctx, service.SendMessageCommand{
		ProjectID: req.ProjectID,
		Content:   req.Content,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SendMessageRes{
		CommandID:      result.CommandID,
		MessageID:      result.MessageID,
		ArchitectReply: result.ArchitectReply,
		Accepted:       true,
	}, nil
}
