package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// HandoffRecordRepo 任务交接记录仓储。
type HandoffRecordRepo struct{}

func NewHandoffRecordRepo() *HandoffRecordRepo { return &HandoffRecordRepo{} }

func (r *HandoffRecordRepo) table() string { return "mvp_handoff_record" }

// Create 创建交接记录。
func (r *HandoffRecordRepo) Create(ctx context.Context, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return err
}

// CreateInTx 在事务中创建交接记录。
func (r *HandoffRecordRepo) CreateInTx(ctx context.Context, tx gdb.TX, data g.Map) error {
	_, err := tx.Model(r.table()).Ctx(ctx).Insert(data)
	return err
}

// ListByWorkflow 查询工作流下交接记录。
func (r *HandoffRecordRepo) ListByWorkflow(ctx context.Context, workflowRunID int64, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		OrderAsc("created_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListByWorkflowAndTask 查询工作流下指定任务参与的交接记录。
func (r *HandoffRecordRepo) ListByWorkflowAndTask(ctx context.Context, workflowRunID, taskID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("(from_task_id = ? OR to_task_id = ?)", taskID, taskID).
		OrderAsc("created_at").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListByWorkflowAndType 查询工作流下指定交接类型的记录。
func (r *HandoffRecordRepo) ListByWorkflowAndType(ctx context.Context, workflowRunID int64, handoffType string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		OrderAsc("created_at")
	if handoffType != "" {
		model = model.Where("handoff_type", handoffType)
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// CountByWorkflow 统计工作流下交接记录数。
func (r *HandoffRecordRepo) CountByWorkflow(ctx context.Context, workflowRunID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Count()
}

// CountByWorkflowAndFromTask 统计工作流下指定来源任务的交接记录数。
func (r *HandoffRecordRepo) CountByWorkflowAndFromTask(ctx context.Context, workflowRunID, fromTaskID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("from_task_id", fromTaskID).
		Count()
}

// GetLatestPayloadByWorkflowFromTaskType 查询最近一条指定来源任务和交接类型的 payload。
func (r *HandoffRecordRepo) GetLatestPayloadByWorkflowFromTaskType(ctx context.Context, workflowRunID, fromTaskID int64, handoffType string) (string, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("from_task_id", fromTaskID).
		OrderDesc("created_at")
	if handoffType != "" {
		model = model.Where("handoff_type", handoffType)
	}
	record, err := model.Fields("payload").One()
	if err != nil || record.IsEmpty() {
		return "", err
	}
	return record["payload"].String(), nil
}
