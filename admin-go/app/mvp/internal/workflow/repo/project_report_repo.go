package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// ProjectReportRepo 项目汇报仓储。
type ProjectReportRepo struct{}

func NewProjectReportRepo() *ProjectReportRepo { return &ProjectReportRepo{} }

func (r *ProjectReportRepo) table() string { return "mvp_project_report" }

// Create 创建报告。
func (r *ProjectReportRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	data["created_at"] = gtime.Now()
	data["updated_at"] = gtime.Now()
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询。
func (r *ProjectReportRepo) GetByID(ctx context.Context, id int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).WhereNull("deleted_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByProject 按项目查询报告列表。
func (r *ProjectReportRepo) ListByProject(ctx context.Context, projectID int64, reportType string) ([]g.Map, error) {
	m := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("created_at")
	if reportType != "" {
		m = m.Where("report_type", reportType)
	}
	records, err := m.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// GetLatestByType 获取最新的指定类型报告。
func (r *ProjectReportRepo) GetLatestByType(ctx context.Context, workflowRunID int64, reportType string) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("report_type", reportType).
		WhereNull("deleted_at").
		OrderDesc("created_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}
