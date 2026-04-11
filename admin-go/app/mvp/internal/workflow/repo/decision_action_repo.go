package repo

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// DecisionActionRepo 决策动作仓储。
type DecisionActionRepo struct{}

// NewDecisionActionRepo 创建决策动作仓储。
func NewDecisionActionRepo() *DecisionActionRepo { return &DecisionActionRepo{} }

func (r *DecisionActionRepo) table() string { return "mvp_decision_action" }

// Create 创建决策动作记录。
func (r *DecisionActionRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	normalizeDecisionActionJSONFields(data)
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询。
func (r *DecisionActionRepo) GetByID(ctx context.Context, id int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).WhereNull("deleted_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetProjectIDByID 查询决策动作所属项目。
func (r *DecisionActionRepo) GetProjectIDByID(ctx context.Context, id int64) (int64, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Fields("project_id").
		Where("id", id).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return 0, err
	}
	return record["project_id"].Int64(), nil
}

// UpdateStatus 更新决策动作状态。
func (r *DecisionActionRepo) UpdateStatus(ctx context.Context, id int64, status string, extra g.Map) error {
	data := g.Map{"action_status": status, "updated_at": gtime.Now()}
	for k, v := range extra {
		data[k] = v
	}
	normalizeDecisionActionJSONFields(data)
	_, err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).Update(data)
	return err
}

func normalizeDecisionActionJSONFields(data g.Map) {
	for _, key := range []string{"trigger_context", "matched_gate_ids", "recommendation", "result"} {
		value, ok := data[key]
		if !ok {
			continue
		}
		if value == nil {
			continue
		}
		switch v := value.(type) {
		case string:
			trimmed := strings.TrimSpace(v)
			if trimmed == "" {
				data[key] = nil
				continue
			}
			if json.Valid([]byte(trimmed)) {
				data[key] = trimmed
				continue
			}
			if encoded, err := json.Marshal(v); err == nil {
				data[key] = string(encoded)
			}
		case []byte:
			if json.Valid(v) {
				data[key] = string(v)
				continue
			}
			if encoded, err := json.Marshal(string(v)); err == nil {
				data[key] = string(encoded)
			}
		default:
			if encoded, err := json.Marshal(v); err == nil {
				data[key] = string(encoded)
			}
		}
	}
}

// ListByWorkflow 按工作流查询决策记录。
func (r *DecisionActionRepo) ListByWorkflow(ctx context.Context, workflowRunID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderDesc("created_at").
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

// ListByProject 按项目和动作类型查询。
func (r *DecisionActionRepo) ListByProject(ctx context.Context, projectID int64, actionType string) ([]g.Map, error) {
	q := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at")
	if actionType != "" {
		q = q.Where("decision_type", actionType)
	}
	records, err := q.OrderDesc("created_at").All()
	if err != nil {
		return nil, err
	}
	result := make([]g.Map, 0, len(records))
	for _, rec := range records {
		result = append(result, rec.Map())
	}
	return result, nil
}

// ListPending 查询待处理的决策。
func (r *DecisionActionRepo) ListPending(ctx context.Context, projectID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("action_status", g.Slice{"pending", "waiting_human"}).
		WhereNull("deleted_at").
		OrderDesc("created_at").
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

// CountPendingByProject 统计项目下待处理动作数。
func (r *DecisionActionRepo) CountPendingByProject(ctx context.Context, projectID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("action_status", g.Slice{"pending", "waiting_human"}).
		WhereNull("deleted_at").
		Count()
}

// ListByProjectFiltered 按项目查询决策记录，支持可选的状态和类型过滤。
func (r *DecisionActionRepo) ListByProjectFiltered(ctx context.Context, projectID int64, actionStatus, decisionType string) ([]g.Map, error) {
	q := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at")
	if actionStatus != "" {
		q = q.Where("action_status", actionStatus)
	}
	if decisionType != "" {
		q = q.Where("decision_type", decisionType)
	}
	records, err := q.OrderDesc("created_at").All()
	if err != nil {
		return nil, err
	}
	result := make([]g.Map, 0, len(records))
	for _, rec := range records {
		result = append(result, rec.Map())
	}
	return result, nil
}

// CountByType 按类型统计。
func (r *DecisionActionRepo) CountByType(ctx context.Context, workflowRunID int64, actionType string) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("decision_type", actionType).
		WhereNull("deleted_at").
		Count()
}

// ListByWorkflowAndTask 查询工作流下指定任务的决策动作。
func (r *DecisionActionRepo) ListByWorkflowAndTask(ctx context.Context, workflowRunID, taskID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("domain_task_id", taskID).
		WhereNull("deleted_at").
		OrderDesc("created_at").
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
