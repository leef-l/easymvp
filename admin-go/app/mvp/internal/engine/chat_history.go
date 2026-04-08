package engine

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	mvpmodel "easymvp/app/mvp/internal/model"
	"easymvp/utility/provider"
)

// loadHistory 加载对话历史（排除当前正在 streaming 的消息）
func (e *ChatEngine) loadHistory(ctx context.Context, conversationID int64, excludeID int64) ([]provider.Message, error) {
	var records []gdb.Record
	err := g.DB().Model("mvp_message").Ctx(ctx).
		Where("conversation_id", conversationID).
		WhereNull("deleted_at").
		Where("status", "completed").
		Where("(message_type IS NULL OR message_type <> ?)", mvpmodel.MessageTypePoison).
		Where("id != ?", excludeID).
		Order("created_at ASC").
		Scan(&records)
	if err != nil {
		return nil, fmt.Errorf("加载对话历史失败: %w", err)
	}

	messages := make([]provider.Message, 0, len(records))
	for _, r := range records {
		role := provider.Role(r["role"].String())
		messages = append(messages, provider.Message{
			Role:    role,
			Content: r["content"].String(),
		})
	}
	return messages, nil
}
