package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// MessageChunkRepo 消息分块仓储。
type MessageChunkRepo struct{}

func NewMessageChunkRepo() *MessageChunkRepo { return &MessageChunkRepo{} }

func (r *MessageChunkRepo) table() string { return "mvp_message_chunk" }

// ListByMessage 查询消息分块。
func (r *MessageChunkRepo) ListByMessage(ctx context.Context, messageID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("message_id", messageID).
		Order("chunk_index ASC").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}
