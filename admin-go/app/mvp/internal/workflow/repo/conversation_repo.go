package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// ConversationRepo 对话仓储。
type ConversationRepo struct{}

func NewConversationRepo() *ConversationRepo { return &ConversationRepo{} }

func (r *ConversationRepo) table() string { return "mvp_conversation" }

// GetByID 按 ID 查询对话。
func (r *ConversationRepo) GetByID(ctx context.Context, conversationID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", conversationID).
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

// GetArchitectProjectConversation 查询项目级架构师对话。
func (r *ConversationRepo) GetArchitectProjectConversation(ctx context.Context, projectID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		Where("role_type", "architect").
		Where("task_id IS NULL OR task_id = 0").
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

// GetFirstByProject 查询项目最早一条对话。
func (r *ConversationRepo) GetFirstByProject(ctx context.Context, projectID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderAsc("created_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}
