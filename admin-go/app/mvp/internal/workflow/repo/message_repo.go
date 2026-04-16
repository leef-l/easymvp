package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// MessageRepo 消息仓储。
type MessageRepo struct{}

func NewMessageRepo() *MessageRepo { return &MessageRepo{} }

func (r *MessageRepo) table() string { return "mvp_message" }

// GetByID 按 ID 查询消息。
func (r *MessageRepo) GetByID(ctx context.Context, messageID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", messageID).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByConversation 查询对话下的消息列表。
func (r *MessageRepo) ListByConversation(ctx context.Context, conversationID int64, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("conversation_id", conversationID).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.OrderAsc("created_at").All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListByConversationRoleStatus 查询对话下指定角色和状态的消息列表。
func (r *MessageRepo) ListByConversationRoleStatus(ctx context.Context, conversationID int64, role string, status string, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("conversation_id", conversationID).
		WhereNull("deleted_at")
	if role != "" {
		model = model.Where("role", role)
	}
	if status != "" {
		model = model.Where("status", status)
	}
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.OrderAsc("created_at").All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// GetLatestByConversationRoleStatus 查询对话下最近一条指定角色和状态的消息。
func (r *MessageRepo) GetLatestByConversationRoleStatus(ctx context.Context, conversationID int64, role string, status string, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("conversation_id", conversationID).
		WhereNull("deleted_at")
	if role != "" {
		model = model.Where("role", role)
	}
	if status != "" {
		model = model.Where("status", status)
	}
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.OrderDesc("created_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListHistoryByConversation 查询对话历史，附带模型名称。
func (r *MessageRepo) ListHistoryByConversation(ctx context.Context, conversationID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()+" m").Ctx(ctx).
		LeftJoin("ai_model am", "am.id = m.model_id").
		Fields("m.id, m.role, m.message_type, m.content, m.status, m.created_at, am.name AS model_name").
		Where("m.conversation_id", conversationID).
		Where("m.deleted_at IS NULL").
		OrderAsc("m.created_at").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// GetLatestAssistantContentByConversation 查询对话下最近一条 assistant 内容消息。
func (r *MessageRepo) GetLatestAssistantContentByConversation(ctx context.Context, conversationID int64) (string, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("conversation_id", conversationID).
		Where("role", "assistant").
		Where("content <> ''").
		WhereNull("deleted_at").
		OrderDesc("created_at").
		OrderDesc("id").
		Fields("content").
		One()
	if err != nil || record.IsEmpty() {
		return "", err
	}
	return strings.TrimSpace(record["content"].String()), nil
}
