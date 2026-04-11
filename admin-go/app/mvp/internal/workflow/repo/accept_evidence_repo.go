package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// AcceptEvidenceRepo 验收证据仓储。
type AcceptEvidenceRepo struct{}

func NewAcceptEvidenceRepo() *AcceptEvidenceRepo { return &AcceptEvidenceRepo{} }

func (r *AcceptEvidenceRepo) table() string { return "mvp_accept_evidence" }

// BatchCreate 批量创建验收证据。
func (r *AcceptEvidenceRepo) BatchCreate(ctx context.Context, items []g.Map) error {
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		items[i]["id"] = snowflake.Generate()
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(items)
	return err
}

// ListByAcceptRun 按验收运行查询证据列表（必须绑定 accept_run_id，不允许裸查）。
func (r *AcceptEvidenceRepo) ListByAcceptRun(ctx context.Context, acceptRunID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("accept_run_id", acceptRunID).
		WhereNull("deleted_at").
		OrderAsc("evidence_type").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListByTaskSources 查询和任务/工作空间关联的验收证据。
func (r *AcceptEvidenceRepo) ListByTaskSources(ctx context.Context, taskID, workspaceID int64) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).WhereNull("deleted_at")
	if workspaceID > 0 {
		model = model.Where(
			"(source_type = ? AND source_id = ?) OR (source_type = ? AND source_id = ?)",
			"domain_task", taskID, "workspace", workspaceID,
		)
	} else {
		model = model.Where("source_type", "domain_task").Where("source_id", taskID)
	}
	records, err := model.OrderDesc("created_at").All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}
