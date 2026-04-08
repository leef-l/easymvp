package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/workflow/autonomy"
	"easymvp/app/mvp/internal/workflow/repo"
)

// Objective 查询项目目标约束。
func (c *cWorkflow) Objective(ctx context.Context, req *v1.WorkflowObjectiveReq) (res *v1.WorkflowObjectiveRes, err error) {
	projectID := int64(req.ProjectID)
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	obj, err := autonomy.NewObjectiveService().Load(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowObjectiveRes{Objective: gconv.Map(obj)}, nil
}

// SaveObjective 保存项目目标约束。
func (c *cWorkflow) SaveObjective(ctx context.Context, req *v1.WorkflowSaveObjectiveReq) (res *v1.WorkflowSaveObjectiveRes, err error) {
	projectID := int64(req.ProjectID)
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	payload, mErr := json.Marshal(g.Map{
		"deliveryGoal":       req.DeliveryGoal,
		"qualityFloor":       req.QualityFloor,
		"tokenBudget":        req.TokenBudget,
		"timeBudgetHours":    req.TimeBudgetHours,
		"costBudgetCents":    req.CostBudgetCents,
		"riskTolerance":      req.RiskTolerance,
		"maxAutoRetries":     req.MaxAutoRetries,
		"maxAutoReworks":     req.MaxAutoReworks,
		"maxAutoReplans":     req.MaxAutoReplans,
		"deadlineAt":         req.DeadlineAt,
		"maxStallMinutes":    req.MaxStallMinutes,
		"autonomyLevel":      req.AutonomyLevel,
		"maxSideEffectLevel": req.MaxSideEffectLevel,
	})
	if mErr != nil {
		return nil, fmt.Errorf("序列化目标配置失败: %w", mErr)
	}
	_, err = g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).
		WhereNull("deleted_at").
		Update(g.Map{
			"objective_json": string(payload),
		})
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowSaveObjectiveRes{}, nil
}

// Situation 查询当前工作流态势。
func (c *cWorkflow) Situation(ctx context.Context, req *v1.WorkflowSituationReq) (res *v1.WorkflowSituationRes, err error) {
	workflowRunID := int64(req.WorkflowRunID)
	projectID, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).
		WhereNull("deleted_at").
		Value("project_id")
	if err != nil {
		return nil, err
	}
	if err = checkProjectOwnership(ctx, projectID.Int64()); err != nil {
		return nil, err
	}
	sensor := autonomy.NewSensor(repo.NewSituationSnapshotRepo())
	sit, err := sensor.Perceive(ctx, workflowRunID)
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowSituationRes{Situation: gconv.Map(sit)}, nil
}

// SituationHistory 查询态势快照历史。
func (c *cWorkflow) SituationHistory(ctx context.Context, req *v1.WorkflowSituationHistoryReq) (res *v1.WorkflowSituationHistoryRes, err error) {
	projectID := int64(req.ProjectID)
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	snapshots, err := repo.NewSituationSnapshotRepo().ListByProjectID(ctx, projectID, limit)
	if err != nil {
		return nil, err
	}
	if req.WorkflowRunID > 0 {
		filtered := make([]g.Map, 0, len(snapshots))
		for _, item := range snapshots {
			if int64(mapJsonInt64(item, "workflow_run_id")) == int64(req.WorkflowRunID) {
				filtered = append(filtered, item)
			}
		}
		snapshots = filtered
	}
	return &v1.WorkflowSituationHistoryRes{Snapshots: snapshots}, nil
}
