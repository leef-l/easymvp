package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/workflow/autonomy"
	"easymvp/app/mvp/internal/workflow/contract"
	"easymvp/app/mvp/internal/workflow/repo"
)

const (
	defaultSituationHistoryPageSize = 50
	maxSituationHistoryScanPages    = 10
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
	objective := gconv.Map(obj)
	projectContract, contractErr := contract.Load(ctx, projectID)
	if contractErr != nil {
		g.Log().Warningf(ctx, "[WorkflowObjective] 加载项目级硬约束失败: project=%d err=%v", projectID, contractErr)
	} else if projectContract != nil && !projectContract.IsEmpty() {
		objective["technicalContract"] = gconv.Map(projectContract)
	}
	return &v1.WorkflowObjectiveRes{Objective: objective}, nil
}

// SaveObjective 保存项目目标约束。
func (c *cWorkflow) SaveObjective(ctx context.Context, req *v1.WorkflowSaveObjectiveReq) (res *v1.WorkflowSaveObjectiveRes, err error) {
	projectID := int64(req.ProjectID)
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	project, getErr := repo.NewProjectRepo().GetByID(ctx, projectID, "objective_json")
	if getErr != nil {
		return nil, getErr
	}
	payload, mErr := contract.MergeObjectiveFields(mapString(project, "objective_json"), g.Map{
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
		return nil, fmt.Errorf("合并目标配置失败: %w", mErr)
	}
	err = repo.NewProjectRepo().UpdateFields(ctx, projectID, g.Map{
		"objective_json": payload,
	})
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowSaveObjectiveRes{}, nil
}

// Situation 查询当前工作流态势。
func (c *cWorkflow) Situation(ctx context.Context, req *v1.WorkflowSituationReq) (res *v1.WorkflowSituationRes, err error) {
	workflowRunID := int64(req.WorkflowRunID)
	workflowRun, err := repo.NewWorkflowRunRepo().GetByIDMap(ctx, workflowRunID, "project_id")
	if err != nil {
		return nil, err
	}
	projectID := g.NewVar(workflowRun["project_id"]).Int64()
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	sensor := autonomy.NewSensor(repo.NewSituationSnapshotRepo())
	taskID := int64(req.TaskID)
	var sit *autonomy.Situation
	if taskID > 0 {
		sit, err = sensor.PerceiveForTask(ctx, workflowRunID, taskID)
	} else {
		sit, err = sensor.Perceive(ctx, workflowRunID)
	}
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
	snapshots, err := loadSituationHistorySnapshots(
		func(offset, pageLimit int) ([]g.Map, error) {
			return repo.NewSituationSnapshotRepo().ListByProjectIDWindow(ctx, projectID, int64(req.WorkflowRunID), offset, pageLimit)
		},
		int64(req.WorkflowRunID),
		int64(req.TaskID),
		limit,
	)
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowSituationHistoryRes{Snapshots: snapshots}, nil
}

func buildSituationSnapshotItem(record g.Map) (g.Map, bool) {
	raw := mapString(record, "snapshot_data")
	if raw == "" {
		return nil, false
	}

	var sit autonomy.Situation
	if err := json.Unmarshal([]byte(raw), &sit); err != nil {
		return nil, false
	}

	item := gconv.Map(sit)
	if item == nil {
		item = g.Map{}
	}
	item["id"] = g.NewVar(record["id"]).Int64()
	if gconv.Int64(item["workflowRunId"]) == 0 {
		item["workflowRunId"] = g.NewVar(record["workflow_run_id"]).Int64()
	}
	if gconv.Int64(item["projectId"]) == 0 {
		item["projectId"] = g.NewVar(record["project_id"]).Int64()
	}
	if item["snapshotAt"] == nil {
		item["snapshotAt"] = normalizeDBUTCGTime(g.NewVar(record["created_at"]).GTime())
	}
	if item["snapshotAt"] == nil {
		item["snapshotAt"] = normalizeDBUTCGTime(gtime.Now())
	}
	return item, true
}

func matchSituationSnapshotFilters(item g.Map, workflowRunID, taskID int64) bool {
	if workflowRunID > 0 && int64(mapJsonInt64(item, "workflowRunId")) != workflowRunID {
		return false
	}
	health := gconv.Map(item["health"])
	focusedTaskID := gconv.Int64(health["focusedTaskId"])
	if taskID > 0 {
		return focusedTaskID == taskID
	}
	return focusedTaskID == 0
}

func loadSituationHistorySnapshots(
	loader func(offset, limit int) ([]g.Map, error),
	workflowRunID, taskID int64,
	limit int,
) ([]g.Map, error) {
	if limit <= 0 {
		limit = 20
	}

	pageSize := limit
	if pageSize < defaultSituationHistoryPageSize {
		pageSize = defaultSituationHistoryPageSize
	}

	filtered := make([]g.Map, 0, limit)
	for page := 0; page < maxSituationHistoryScanPages && len(filtered) < limit; page++ {
		records, err := loader(page*pageSize, pageSize)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			break
		}

		for _, item := range records {
			snapshotItem, ok := buildSituationSnapshotItem(item)
			if !ok {
				continue
			}
			if !matchSituationSnapshotFilters(snapshotItem, workflowRunID, taskID) {
				continue
			}
			filtered = append(filtered, snapshotItem)
			if len(filtered) >= limit {
				break
			}
		}

		if len(records) < pageSize {
			break
		}
	}
	return filtered, nil
}
