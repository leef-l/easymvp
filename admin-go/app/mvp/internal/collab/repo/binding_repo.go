package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BindingRepo 用户协作平台绑定仓储。
type BindingRepo struct{}

// NewBindingRepo 创建绑定仓储。
func NewBindingRepo() *BindingRepo { return &BindingRepo{} }

func (r *BindingRepo) table() string { return "mvp_user_collab_binding" }

// GetByUserID 按系统用户ID和平台查询绑定。
func (r *BindingRepo) GetByUserID(ctx context.Context, userID int64, platform string) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("user_id", userID).
		Where("platform", platform).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetByPlatformUserID 按平台用户标识反查系统用户。
func (r *BindingRepo) GetByPlatformUserID(ctx context.Context, platform, platformUserID string) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("platform", platform).
		Where("platform_user_id", platformUserID).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByUserIDs 批量查询用户绑定。
func (r *BindingRepo) ListByUserIDs(ctx context.Context, userIDs []int64, platform string) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("user_id", userIDs).
		Where("platform", platform).
		WhereNull("deleted_at").
		All()
	if err != nil {
		return nil, err
	}
	result := make([]g.Map, 0, len(records))
	for _, rec := range records {
		result = append(result, rec.Map())
	}
	return result, nil
}

// Bind 创建绑定。
func (r *BindingRepo) Bind(ctx context.Context, data g.Map) (int64, error) {
	data["created_at"] = gtime.Now()
	data["updated_at"] = gtime.Now()
	result, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

// Unbind 解绑（软删除）。
func (r *BindingRepo) Unbind(ctx context.Context, userID int64, platform string) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("user_id", userID).
		Where("platform", platform).
		WhereNull("deleted_at").
		Update(g.Map{"deleted_at": gtime.Now()})
	return err
}
