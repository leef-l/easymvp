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
	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

// AutonomyDecisions 自治决策列表
func (c *cWorkflow) AutonomyDecisions(ctx context.Context, req *v1.WorkflowAutonomyDecisionsReq) (res *v1.WorkflowAutonomyDecisionsRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	decisionRepo := repo.NewAutonomyDecisionRepo()
	records, err := decisionRepo.ListByProject(ctx, projectID, req.DecisionType)
	if err != nil {
		return nil, err
	}

	var items []v1.AutonomyDecisionItem
	for _, r := range records {
		items = append(items, v1.AutonomyDecisionItem{
			ID:             snowflake.JsonInt64(g.NewVar(r["id"]).Int64()),
			DecisionType:   g.NewVar(r["decision_type"]).String(),
			TriggerSource:  g.NewVar(r["trigger_source"]).String(),
			TriggerContext: g.NewVar(r["trigger_context"]).String(),
			Recommendation: g.NewVar(r["recommendation"]).String(),
			DecisionMode:   g.NewVar(r["decision_mode"]).String(),
			HumanAction:    g.NewVar(r["human_action"]).String(),
			ExecutedAt:     g.NewVar(r["executed_at"]).GTime(),
			Result:         g.NewVar(r["result"]).String(),
			CreatedAt:      g.NewVar(r["created_at"]).GTime(),
		})
	}
	if items == nil {
		items = []v1.AutonomyDecisionItem{}
	}
	return &v1.WorkflowAutonomyDecisionsRes{Decisions: items}, nil
}

// ApproveDecision 批准自治决策
func (c *cWorkflow) ApproveDecision(ctx context.Context, req *v1.WorkflowApproveDecisionReq) (res *v1.WorkflowApproveDecisionRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	decisionRepo := repo.NewAutonomyDecisionRepo()
	if err := decisionRepo.UpdateHumanAction(ctx, int64(req.DecisionID), "approved"); err != nil {
		return nil, err
	}
	return &v1.WorkflowApproveDecisionRes{}, nil
}

// RejectDecision 拒绝自治决策
func (c *cWorkflow) RejectDecision(ctx context.Context, req *v1.WorkflowRejectDecisionReq) (res *v1.WorkflowRejectDecisionRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	decisionRepo := repo.NewAutonomyDecisionRepo()
	if err := decisionRepo.UpdateHumanAction(ctx, int64(req.DecisionID), "rejected"); err != nil {
		return nil, err
	}
	return &v1.WorkflowRejectDecisionRes{}, nil
}

// TriggerReplan 手动触发重规划
func (c *cWorkflow) TriggerReplan(ctx context.Context, req *v1.WorkflowTriggerReplanReq) (res *v1.WorkflowTriggerReplanRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	workflowRunRepo := repo.NewWorkflowRunRepo()
	wfRun, err := workflowRunRepo.GetLatestByProjectStatuses(ctx, projectID, []string{"executing", "reworking", "accepting", "paused"}, "id")
	if err != nil || len(wfRun) == 0 {
		return nil, fmt.Errorf("无活跃的工作流运行")
	}

	projRole, roleErr := repo.GetProjectRole(ctx, projectID, "architect")
	if roleErr != nil {
		return nil, fmt.Errorf("查询架构师角色失败: %w", roleErr)
	}
	if projRole.IsEmpty() || projRole["model_id"].Int64() == 0 {
		return nil, fmt.Errorf("项目未配置架构师(architect)角色或模型，无法执行重规划。请先在项目角色中配置架构师。")
	}

	modelRow, modelErr := repo.NewAIModelRepo().GetWithPlanByID(ctx, projRole["model_id"].Int64(), "m.model_code", "p.api_key")
	if modelErr != nil {
		return nil, fmt.Errorf("查询架构师模型失败: %w", modelErr)
	}
	if len(modelRow) == 0 {
		return nil, fmt.Errorf("架构师角色关联的 AI 模型(ID=%d)不存在或已删除", projRole["model_id"].Int64())
	}
	if g.NewVar(modelRow["api_key"]).String() == "" {
		return nil, fmt.Errorf("架构师角色关联的 AI 模型(%s)没有配置 API Key，无法调用", g.NewVar(modelRow["model_code"]).String())
	}

	decisionRepo := repo.NewAutonomyDecisionRepo()
	replanner := autonomy.NewReplanner(decisionRepo)

	workflowRunID := g.NewVar(wfRun["id"]).Int64()
	failedTasks, ftErr := repo.NewDomainTaskRepo().ListByWorkflowAndStatuses(ctx, workflowRunID, []string{"failed", "escalated"}, "id", "name", "result", "retry_count")
	if ftErr != nil {
		g.Log().Warningf(ctx, "[TriggerReplan] 查询失败任务列表失败: %v", ftErr)
	}

	var failed []autonomy.FailedTaskInfo
	for _, t := range failedTasks {
		failed = append(failed, autonomy.FailedTaskInfo{
			TaskID:       g.NewVar(t["id"]).Int64(),
			TaskName:     g.NewVar(t["name"]).String(),
			ErrorMessage: g.NewVar(t["result"]).String(),
			RetryCount:   g.NewVar(t["retry_count"]).Int(),
		})
	}

	input := &autonomy.ReplanInput{
		WorkflowRunID: workflowRunID,
		ProjectID:     projectID,
		TriggerSource: "manual",
		FailedTasks:   failed,
	}

	wfRunID := workflowRunID

	go func() {
		bgCtx := context.Background()
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(bgCtx, "[TriggerReplan] panic: %v", r)
			}
		}()

		result, err := replanner.Evaluate(bgCtx, input)

		eventType := "replan.completed"
		payloadMap := map[string]interface{}{
			"trigger": "manual",
		}
		if err != nil {
			eventType = "replan.failed"
			payloadMap["error"] = err.Error()
			g.Log().Errorf(bgCtx, "[TriggerReplan] 重规划失败: projectID=%d err=%v", projectID, err)
		} else if result != nil {
			payloadMap["action"] = result.Action
			payloadMap["reasoning"] = result.Reasoning
			if result.Action == autonomy.ReplanAbort {
				eventType = "replan.aborted"
				g.Log().Warningf(bgCtx, "[TriggerReplan] 重规划中止: projectID=%d reason=%s", projectID, result.Reasoning)
			} else {
				g.Log().Infof(bgCtx, "[TriggerReplan] 重规划完成: projectID=%d action=%s", projectID, result.Action)
			}
		}
		if insErr := event.PersistRecord(bgCtx, event.Event{
			WorkflowRunID: wfRunID,
			EntityType:    "workflow",
			EventType:     eventType,
			Payload:       payloadMap,
		}); insErr != nil {
			g.Log().Warningf(bgCtx, "[TriggerReplan] 记录重规划事件失败: wfRun=%d err=%v", wfRunID, insErr)
		}
	}()

	return &v1.WorkflowTriggerReplanRes{}, nil
}

// ProjectReports 项目报告列表
func (c *cWorkflow) ProjectReports(ctx context.Context, req *v1.WorkflowProjectReportsReq) (res *v1.WorkflowProjectReportsRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	reportRepo := repo.NewProjectReportRepo()
	records, err := reportRepo.ListByProject(ctx, projectID, req.ReportType)
	if err != nil {
		return nil, err
	}

	var items []v1.ProjectReportItem
	for _, r := range records {
		items = append(items, v1.ProjectReportItem{
			ID:         snowflake.JsonInt64(g.NewVar(r["id"]).Int64()),
			ReportType: g.NewVar(r["report_type"]).String(),
			StageType:  g.NewVar(r["stage_type"]).String(),
			Title:      g.NewVar(r["title"]).String(),
			Content:    g.NewVar(r["content"]).String(),
			Metrics:    g.NewVar(r["metrics"]).String(),
			CreatedAt:  g.NewVar(r["created_at"]).GTime(),
		})
	}
	if items == nil {
		items = []v1.ProjectReportItem{}
	}
	return &v1.WorkflowProjectReportsRes{Reports: items}, nil
}

// TriggerReport 手动生成报告
func (c *cWorkflow) TriggerReport(ctx context.Context, req *v1.WorkflowTriggerReportReq) (res *v1.WorkflowTriggerReportRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := repo.NewWorkflowRunRepo().GetLatestByProject(ctx, projectID)
	if err != nil || wfRun.IsEmpty() {
		return nil, fmt.Errorf("无工作流运行记录")
	}

	reportRepo := repo.NewProjectReportRepo()
	reporter := autonomy.NewReporter(reportRepo)

	stageType := req.StageType
	if stageType == "" {
		stageType = "complete"
	}

	if err := reporter.GenerateStageReport(ctx, wfRun["id"].Int64(), stageType); err != nil {
		return nil, err
	}
	return &v1.WorkflowTriggerReportRes{}, nil
}

// AutonomyMode 查询当前自治模式
func (c *cWorkflow) AutonomyMode(ctx context.Context, req *v1.WorkflowAutonomyModeReq) (res *v1.WorkflowAutonomyModeRes, err error) {
	return &v1.WorkflowAutonomyModeRes{Mode: autonomy.GetAutonomyMode(ctx)}, nil
}

// SetAutonomyMode 设置自治模式（写入 mvp_config）
func (c *cWorkflow) SetAutonomyMode(ctx context.Context, req *v1.WorkflowSetAutonomyModeReq) (res *v1.WorkflowSetAutonomyModeRes, err error) {
	if err := repo.NewConfigRepo().UpsertByKey(ctx, "autonomy.mode", g.Map{
		"config_value": req.Mode,
		"config_type":  "string",
		"category":     "autonomy",
		"description":  "自治模式：suggest=建议型 auto=全自动",
	}); err != nil {
		return nil, fmt.Errorf("写入自治模式配置失败: %w", err)
	}
	return &v1.WorkflowSetAutonomyModeRes{}, nil
}

// AutonomyCheckpoints 查询项目待处理的人工节点和决策动作。
func (c *cWorkflow) AutonomyCheckpoints(ctx context.Context, req *v1.WorkflowAutonomyCheckpointsReq) (res *v1.WorkflowAutonomyCheckpointsRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	rawCheckpoints, cpErr := dc.ListOpenCheckpoints(ctx, projectID)
	if cpErr != nil {
		return nil, cpErr
	}
	rawActions, acErr := dc.ListPendingActions(ctx, projectID)
	if acErr != nil {
		return nil, acErr
	}

	checkpoints := make([]v1.CheckpointDTO, 0, len(rawCheckpoints))
	for _, m := range rawCheckpoints {
		checkpoints = append(checkpoints, mapToCheckpointDTO(m))
	}
	actions := make([]v1.DecisionActionDTO, 0, len(rawActions))
	for _, m := range rawActions {
		actions = append(actions, mapToDecisionActionDTO(m))
	}

	return &v1.WorkflowAutonomyCheckpointsRes{
		Checkpoints: checkpoints,
		Actions:     actions,
	}, nil
}

func mapToCheckpointDTO(m g.Map) v1.CheckpointDTO {
	return v1.CheckpointDTO{
		ID:               mapJsonInt64(m, "id"),
		WorkflowRunID:    mapJsonInt64(m, "workflow_run_id"),
		ProjectID:        mapJsonInt64(m, "project_id"),
		DecisionActionID: mapJsonInt64(m, "decision_action_id"),
		CheckpointType:   mapString(m, "checkpoint_type"),
		Title:            mapString(m, "title"),
		Description:      mapString(m, "description"),
		Status:           mapString(m, "status"),
		AssignedTo:       mapJsonInt64(m, "assigned_to"),
		HandledBy:        mapJsonInt64(m, "handled_by"),
		HandleAction:     mapString(m, "handle_action"),
		HandleReason:     mapString(m, "handle_reason"),
		HandledAt:        mapGTime(m, "handled_at"),
		ExpiresAt:        mapGTime(m, "expires_at"),
		CreatedAt:        mapGTime(m, "created_at"),
	}
}

func mapToDecisionActionDTO(m g.Map) v1.DecisionActionDTO {
	return v1.DecisionActionDTO{
		ID:             mapJsonInt64(m, "id"),
		WorkflowRunID:  mapJsonInt64(m, "workflow_run_id"),
		ProjectID:      mapJsonInt64(m, "project_id"),
		StageRunID:     mapJsonInt64(m, "stage_run_id"),
		DomainTaskID:   mapJsonInt64(m, "domain_task_id"),
		DecisionType:   mapString(m, "decision_type"),
		DecisionLevel:  mapString(m, "decision_level"),
		TriggerSource:  mapString(m, "trigger_source"),
		TriggerContext: mapJSONString(m, "trigger_context"),
		MatchedRuleID:  mapJsonInt64(m, "matched_rule_id"),
		MatchedGateIDs: mapJSONString(m, "matched_gate_ids"),
		ActionType:     mapString(m, "action_type"),
		Recommendation: mapJSONString(m, "recommendation"),
		FinalAction:    mapString(m, "final_action"),
		ActionStatus:   mapString(m, "action_status"),
		AutoExecutable: mapInt(m, "auto_executable"),
		HumanRequired:  mapInt(m, "human_required"),
		ExecutedAt:     mapGTime(m, "executed_at"),
		Result:         mapJSONString(m, "result"),
		CreatedAt:      mapGTime(m, "created_at"),
	}
}

// AutonomyApprove 审批通过决策动作。
func (c *cWorkflow) AutonomyApprove(ctx context.Context, req *v1.WorkflowAutonomyApproveReq) (res *v1.WorkflowAutonomyApproveRes, err error) {
	actionID := int64(req.ActionID)
	projectID, lookupErr := autonomyActionProjectID(ctx, actionID)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	if err := dc.ApproveAction(ctx, actionID); err != nil {
		return nil, err
	}
	return &v1.WorkflowAutonomyApproveRes{}, nil
}

// AutonomyReject 驳回决策动作。
func (c *cWorkflow) AutonomyReject(ctx context.Context, req *v1.WorkflowAutonomyRejectReq) (res *v1.WorkflowAutonomyRejectRes, err error) {
	actionID := int64(req.ActionID)
	projectID, lookupErr := autonomyActionProjectID(ctx, actionID)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	if err := dc.RejectAction(ctx, actionID, req.Reason); err != nil {
		return nil, err
	}
	return &v1.WorkflowAutonomyRejectRes{}, nil
}

func autonomyActionProjectID(ctx context.Context, actionID int64) (int64, error) {
	projectID, err := repo.NewDecisionActionRepo().GetProjectIDByID(ctx, actionID)
	if err != nil {
		return 0, fmt.Errorf("查询决策记录失败: %w", err)
	}
	if projectID == 0 {
		return 0, fmt.Errorf("决策记录不存在: %d", actionID)
	}
	return projectID, nil
}

// AutonomyActions 查询项目全量决策记录。
func (c *cWorkflow) AutonomyActions(ctx context.Context, req *v1.WorkflowAutonomyActionsReq) (res *v1.WorkflowAutonomyActionsRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	rawActions, queryErr := dc.ListAllActions(ctx, projectID, req.ActionStatus, req.DecisionType)
	if queryErr != nil {
		return nil, queryErr
	}

	actions := make([]v1.DecisionActionDTO, 0, len(rawActions))
	for _, m := range rawActions {
		actions = append(actions, mapToDecisionActionDTO(m))
	}
	return &v1.WorkflowAutonomyActionsRes{Actions: actions}, nil
}

// AutonomyGateRules 查询项目适用的风险闸门规则。
func (c *cWorkflow) AutonomyGateRules(ctx context.Context, req *v1.WorkflowAutonomyGateRulesReq) (res *v1.WorkflowAutonomyGateRulesRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	rawRules, queryErr := dc.ListGateRules(ctx, projectID)
	if queryErr != nil {
		return nil, queryErr
	}

	rules := make([]v1.RiskGateRuleDTO, 0, len(rawRules))
	for _, m := range rawRules {
		rules = append(rules, mapToRiskGateRuleDTO(m))
	}
	return &v1.WorkflowAutonomyGateRulesRes{Rules: rules}, nil
}

// AutonomyPolicyRules 查询项目适用的策略规则。
func (c *cWorkflow) AutonomyPolicyRules(ctx context.Context, req *v1.WorkflowAutonomyPolicyRulesReq) (res *v1.WorkflowAutonomyPolicyRulesRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	rawRules, queryErr := dc.ListPolicyRules(ctx, projectID)
	if queryErr != nil {
		return nil, queryErr
	}

	rules := make([]v1.PolicyRuleDTO, 0, len(rawRules))
	for _, m := range rawRules {
		rules = append(rules, mapToPolicyRuleDTO(m))
	}
	return &v1.WorkflowAutonomyPolicyRulesRes{Rules: rules}, nil
}

func mapToRiskGateRuleDTO(m g.Map) v1.RiskGateRuleDTO {
	return v1.RiskGateRuleDTO{
		ID:                  mapJsonInt64(m, "id"),
		GateCode:            mapString(m, "gate_code"),
		GateName:            mapString(m, "gate_name"),
		GateType:            mapString(m, "gate_type"),
		ProjectFamily:       mapString(m, "project_family"),
		ProjectCategoryCode: mapString(m, "project_category_code"),
		TriggerExpression:   mapJSONString(m, "trigger_expression"),
		BlockAction:         mapString(m, "block_action"),
		FallbackAction:      mapString(m, "fallback_action"),
		Enabled:             mapInt(m, "enabled"),
		Priority:            mapInt(m, "priority"),
		CreatedAt:           mapGTime(m, "created_at"),
	}
}

func mapToPolicyRuleDTO(m g.Map) v1.PolicyRuleDTO {
	return v1.PolicyRuleDTO{
		ID:                  mapJsonInt64(m, "id"),
		RuleCode:            mapString(m, "rule_code"),
		RuleName:            mapString(m, "rule_name"),
		DecisionType:        mapString(m, "decision_type"),
		DecisionLevel:       mapString(m, "decision_level"),
		TriggerSource:       mapString(m, "trigger_source"),
		ProjectFamily:       mapString(m, "project_family"),
		ProjectCategoryCode: mapString(m, "project_category_code"),
		ConfigJSON:          mapJSONString(m, "config_json"),
		Enabled:             mapInt(m, "enabled"),
		Priority:            mapInt(m, "priority"),
		CreatedAt:           mapGTime(m, "created_at"),
	}
}

func mapString(m g.Map, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func mapInt(m g.Map, key string) int {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return gconv.Int(v)
	}
}

func mapJsonInt64(m g.Map, key string) snowflake.JsonInt64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return snowflake.JsonInt64(n)
	case float64:
		return snowflake.JsonInt64(int64(n))
	case json.Number:
		i, _ := n.Int64()
		return snowflake.JsonInt64(i)
	default:
		return snowflake.JsonInt64(gconv.Int64(v))
	}
}

func mapGTime(m g.Map, key string) *gtime.Time {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case *gtime.Time:
		return t
	default:
		s := fmt.Sprintf("%v", v)
		if s == "" || s == "<nil>" {
			return nil
		}
		return gtime.New(s)
	}
}

func mapJSONString(m g.Map, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}
