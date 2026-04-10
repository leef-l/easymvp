package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/model/entity"
	"easymvp/utility/snowflake"
)

// PlanVersionRepo 计划版本仓储。
type PlanVersionRepo struct{}

func NewPlanVersionRepo() *PlanVersionRepo { return &PlanVersionRepo{} }

func (r *PlanVersionRepo) table() string { return "mvp_plan_version" }

// Create 创建计划版本。
func (r *PlanVersionRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询。
func (r *PlanVersionRepo) GetByID(ctx context.Context, id int64) (*entity.MvpPlanVersion, error) {
	var ent entity.MvpPlanVersion
	err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).Scan(&ent)
	return &ent, err
}

// NextVersionNo 获取项目下一个版本号。
func (r *PlanVersionRepo) NextVersionNo(ctx context.Context, projectID int64) (int, error) {
	val, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Max("version_no")
	if err != nil {
		return 0, err
	}
	return int(val) + 1, nil
}

// ListByProject 查询项目所有版本。
func (r *PlanVersionRepo) ListByProject(ctx context.Context, projectID int64) ([]entity.MvpPlanVersion, error) {
	var list []entity.MvpPlanVersion
	err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("version_no").
		Scan(&list)
	return list, err
}

// ListByProjectStatuses 查询项目下给定状态集合的计划版本。
func (r *PlanVersionRepo) ListByProjectStatuses(ctx context.Context, projectID int64, statuses []string, fields ...string) (gdb.Result, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at")
	if len(statuses) > 0 {
		model = model.WhereIn("status", statuses)
	}
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	return model.OrderDesc("version_no").All()
}

// GetLatestByProjectStatusAndReviewStatus 查询项目下最新命中的计划版本。
func (r *PlanVersionRepo) GetLatestByProjectStatusAndReviewStatus(ctx context.Context, projectID int64, status, reviewStatus string, fields ...string) (gdb.Record, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		Where("status", status).
		Where("review_status", reviewStatus).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.OrderDesc("version_no").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record, nil
}

// RestoreRejectedForManualApprove 恢复被驳回的方案版本并确认草稿蓝图。
func (r *PlanVersionRepo) RestoreRejectedForManualApprove(ctx context.Context, planVersionID int64) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Model(r.table()).Ctx(ctx).
			Where("id", planVersionID).
			Where("status", "draft").
			Where("review_status", "rejected").
			Data(g.Map{
				"status":        "active",
				"review_status": "approved",
				"approved_at":   gdb.Raw("NOW()"),
				"rejected_at":   nil,
			}).
			Update()
		if err != nil {
			return err
		}
		if rows, _ := result.RowsAffected(); rows == 0 {
			return gerror.New("方案版本已不是可人工放行的 rejected 状态")
		}
		_, err = tx.Model("mvp_task_blueprint").Ctx(ctx).
			Where("plan_version_id", planVersionID).
			Where("blueprint_status", "draft").
			Data(g.Map{"blueprint_status": "confirmed"}).
			Update()
		return err
	})
}
