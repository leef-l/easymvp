package repo

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/middleware"
	"easymvp/utility/snowflake"
)

// BindingRepo 用户协作平台绑定仓储。
type BindingRepo struct{}

// NewBindingRepo 创建绑定仓储。
func NewBindingRepo() *BindingRepo { return &BindingRepo{} }

func (r *BindingRepo) table() string { return "mvp_user_collab_binding" }

func (r *BindingRepo) model(ctx context.Context) *gdb.Model {
	return g.DB().Model(r.table()).Ctx(ctx)
}

// GetByUserID 按系统用户ID和平台查询绑定。
func (r *BindingRepo) GetByUserID(ctx context.Context, userID int64, platform string) (g.Map, error) {
	record, err := r.model(ctx).
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
	record, err := r.model(ctx).
		Where("platform", platform).
		Where("platform_user_id", platformUserID).
		WhereNull("deleted_at").
		OrderDesc("updated_at").
		OrderDesc("created_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetByID 获取单条绑定记录。
func (r *BindingRepo) GetByID(ctx context.Context, id int64) (g.Map, error) {
	record, err := r.model(ctx).
		Where("id", id).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetByIDScoped 获取当前权限范围内的绑定记录。
func (r *BindingRepo) GetByIDScoped(ctx context.Context, id int64) (g.Map, error) {
	record, err := middleware.ApplyDataScope(ctx, r.model(ctx), "created_by", "dept_id").
		Where("id", id).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByUserIDs 批量查询用户绑定。
func (r *BindingRepo) ListByUserIDs(ctx context.Context, userIDs []int64, platform string) ([]g.Map, error) {
	records, err := r.model(ctx).
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

// List 获取当前权限范围内的平台绑定列表。
func (r *BindingRepo) List(ctx context.Context, platform string) ([]g.Map, error) {
	model := middleware.ApplyDataScope(ctx, r.model(ctx), "created_by", "dept_id").
		WhereNull("deleted_at")
	if platform != "" {
		model = model.Where("platform", platform)
	}
	records, err := model.OrderDesc("updated_at").OrderDesc("created_at").All()
	if err != nil {
		return nil, err
	}
	result := make([]g.Map, 0, len(records))
	for _, rec := range records {
		result = append(result, rec.Map())
	}
	return result, nil
}

// CountByPlatform 统计平台绑定数量。
func (r *BindingRepo) CountByPlatform(ctx context.Context, platform string) (int, error) {
	model := r.model(ctx).WhereNull("deleted_at")
	if platform != "" {
		model = model.Where("platform", platform)
	}
	return model.Count()
}

// Rebind 重绑用户平台标识，确保同平台下 user_id/platform_user_id 只有一条有效记录。
func (r *BindingRepo) Rebind(ctx context.Context, data g.Map) (int64, error) {
	id := int64(snowflake.Generate())
	now := gtime.Now()
	data["id"] = id
	data["created_at"] = now
	data["updated_at"] = now

	userID := data["user_id"]
	platform := data["platform"]
	platformUserID := data["platform_user_id"]

	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err := tx.Model(r.table()).
			Where("platform", platform).
			WhereNull("deleted_at").
			Where("(user_id = ? OR platform_user_id = ?)", userID, platformUserID).
			Update(g.Map{
				"deleted_at": now,
				"updated_at": now,
			})
		if err != nil {
			return err
		}
		_, err = tx.Model(r.table()).Insert(data)
		return err
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Unbind 解绑（软删除）。
func (r *BindingRepo) Unbind(ctx context.Context, userID int64, platform string) error {
	_, err := r.model(ctx).
		Where("user_id", userID).
		Where("platform", platform).
		WhereNull("deleted_at").
		Update(g.Map{"deleted_at": gtime.Now(), "updated_at": gtime.Now()})
	return err
}

// UnbindByID 按记录 ID 解绑（软删除）。
func (r *BindingRepo) UnbindByID(ctx context.Context, id int64) error {
	_, err := r.model(ctx).
		Where("id", id).
		WhereNull("deleted_at").
		Update(g.Map{"deleted_at": gtime.Now(), "updated_at": gtime.Now()})
	return err
}
