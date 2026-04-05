package acceptance

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/repo"
)

// EvidenceCollector 证据收集器。
type EvidenceCollector struct {
	evidenceRepo *repo.AcceptEvidenceRepo
}

// NewEvidenceCollector 创建证据收集器。
func NewEvidenceCollector(evidenceRepo *repo.AcceptEvidenceRepo) *EvidenceCollector {
	return &EvidenceCollector{evidenceRepo: evidenceRepo}
}

// Collect 收集工作流运行的所有证据并持久化。
func (c *EvidenceCollector) Collect(ctx context.Context, in *AcceptContext) ([]EvidenceItem, error) {
	var items []EvidenceItem

	// 1. 收集 domain_task 结果
	taskItems, err := c.collectTaskOutputs(ctx, in.WorkflowRunID)
	if err != nil {
		g.Log().Warningf(ctx, "[EvidenceCollector] 收集任务输出失败: %v", err)
	} else {
		items = append(items, taskItems...)
	}

	// 2. 收集 stage_run 输出
	stageItems, err := c.collectStageOutputs(ctx, in.WorkflowRunID)
	if err != nil {
		g.Log().Warningf(ctx, "[EvidenceCollector] 收集阶段输出失败: %v", err)
	} else {
		items = append(items, stageItems...)
	}

	// 3. 收集 handoff_record（返工记录）
	handoffItems, err := c.collectHandoffs(ctx, in.WorkflowRunID)
	if err != nil {
		g.Log().Warningf(ctx, "[EvidenceCollector] 收集交接���录失败: %v", err)
	} else {
		items = append(items, handoffItems...)
	}

	// 4. 持久化证据
	if len(items) > 0 {
		now := gtime.Now()
		var dbItems []g.Map
		for _, item := range items {
			dbItems = append(dbItems, g.Map{
				"accept_run_id": in.AcceptRunID,
				"evidence_type": item.EvidenceType,
				"source_type":   item.SourceType,
				"source_id":     item.SourceID,
				"content_ref":   item.ContentRef,
				"summary":       item.Summary,
				"created_at":    now,
				"updated_at":    now,
			})
		}
		if err := c.evidenceRepo.BatchCreate(ctx, dbItems); err != nil {
			return items, fmt.Errorf("��久化证据失败: %w", err)
		}
	}

	g.Log().Infof(ctx, "[EvidenceCollector] 收集到 %d 条证据, acceptRunID=%d", len(items), in.AcceptRunID)
	return items, nil
}

// collectTaskOutputs 收集领域任务的执行结果。
func (c *EvidenceCollector) collectTaskOutputs(ctx context.Context, workflowRunID int64) ([]EvidenceItem, error) {
	tasks, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("id, name, status, result, task_kind").
		All()
	if err != nil {
		return nil, err
	}

	var items []EvidenceItem
	for _, t := range tasks {
		summary := fmt.Sprintf("[%s] %s: status=%s", t["task_kind"].String(), t["name"].String(), t["status"].String())
		items = append(items, EvidenceItem{
			EvidenceType: "task_output",
			SourceType:   "domain_task",
			SourceID:     t["id"].Int64(),
			ContentRef:   t["result"].String(),
			Summary:      summary,
		})
	}
	return items, nil
}

// collectStageOutputs 收集阶段运行记录。
func (c *EvidenceCollector) collectStageOutputs(ctx context.Context, workflowRunID int64) ([]EvidenceItem, error) {
	stages, err := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("id, stage_type, status, error_message").
		OrderAsc("stage_no").
		All()
	if err != nil {
		return nil, err
	}

	var items []EvidenceItem
	for _, s := range stages {
		summary := fmt.Sprintf("stage=%s status=%s", s["stage_type"].String(), s["status"].String())
		items = append(items, EvidenceItem{
			EvidenceType: "stage_output",
			SourceType:   "stage_run",
			SourceID:     s["id"].Int64(),
			Summary:      summary,
		})
	}
	return items, nil
}

// collectHandoffs 收集返工交接记录。
func (c *EvidenceCollector) collectHandoffs(ctx context.Context, workflowRunID int64) ([]EvidenceItem, error) {
	records, err := g.DB().Model("mvp_handoff_record").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Fields("id, from_task_id, to_task_id, handoff_type, reason").
		All()
	if err != nil {
		return nil, err
	}

	var items []EvidenceItem
	for _, r := range records {
		summary := fmt.Sprintf("handoff: %s from=%d to=%d", r["handoff_type"].String(), r["from_task_id"].Int64(), r["to_task_id"].Int64())
		items = append(items, EvidenceItem{
			EvidenceType: "handoff",
			SourceType:   "handoff_record",
			SourceID:     r["id"].Int64(),
			ContentRef:   r["reason"].String(),
			Summary:      summary,
		})
	}
	return items, nil
}
