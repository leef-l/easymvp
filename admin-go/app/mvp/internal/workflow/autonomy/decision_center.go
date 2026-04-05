package autonomy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/app/mvp/internal/workflow/repo"
)

// DecisionCenter 自治决策中台：统一决策入口，编排 PolicyEngine → RiskGate → 审计记录 → 动作执行/人工等待。
type DecisionCenter struct {
	policyEngine     *PolicyEngine
	riskGate         *RiskGate
	actionDispatcher *ActionDispatcher
	actionRepo       *repo.DecisionActionRepo
	checkpointRepo   *repo.HumanCheckpointRepo
	eventPublisher   *event.Publisher
}

// NewDecisionCenter 创建决策中台。
func NewDecisionCenter(
	pe *PolicyEngine,
	rg *RiskGate,
	ad *ActionDispatcher,
	actionRepo *repo.DecisionActionRepo,
	checkpointRepo *repo.HumanCheckpointRepo,
	eventPublisher *event.Publisher,
) *DecisionCenter {
	return &DecisionCenter{
		policyEngine:     pe,
		riskGate:         rg,
		actionDispatcher: ad,
		actionRepo:       actionRepo,
		checkpointRepo:   checkpointRepo,
		eventPublisher:   eventPublisher,
	}
}

// IsEnabled 灰度检查：workflow.autonomy.enabled 是否开启。
func (dc *DecisionCenter) IsEnabled(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.autonomy.enabled", "workflow.autonomy.enabled", 0) == 1
}

// isAuditOnly 是否仅审计模式（记录但不实际执行）。
func (dc *DecisionCenter) isAuditOnly(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.autonomy.audit_only", "workflow.autonomy.auditOnly", 1) == 1
}

// isPolicyEngineEnabled 策略引擎分层灰度。
func (dc *DecisionCenter) isPolicyEngineEnabled(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.autonomy.policy_engine_enabled", "workflow.autonomy.policyEngineEnabled", 1) == 1
}

// isRiskGateEnabled 风险闸门分层灰度。
func (dc *DecisionCenter) isRiskGateEnabled(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.autonomy.risk_gate_enabled", "workflow.autonomy.riskGateEnabled", 1) == 1
}

// Decide 统一决策入口。
//
// 流程：策略匹配 → 闸门检查 → 写审计记录 → 分流（自动执行 / 等待人工）。
//
// Handled 语义：
//   - audit_only=true → Handled=false（仅写审计记录，调用方应继续执行原逻辑）
//   - 真正执行或创建人工节点 → Handled=true（中台已接管，调用方不再执行原逻辑）
func (dc *DecisionCenter) Decide(ctx context.Context, req *DecisionRequest) *DecisionResponse {
	resp := &DecisionResponse{}

	// 1. 获取项目作用域信息（family, category_code）+ 数据权限归属
	family, categoryCode, createdBy, deptID := dc.resolveProjectScope(ctx, req.ProjectID)

	// 2. 策略匹配（受 policy_engine_enabled 灰度控制）
	var match *PolicyMatch
	if dc.isPolicyEngineEnabled(ctx) {
		match = dc.policyEngine.Match(ctx, req.TriggerSource, family, categoryCode)
	}
	if match == nil {
		g.Log().Infof(ctx, "[DecisionCenter] 无匹配策略，跳过: trigger=%s project=%d",
			req.TriggerSource, req.ProjectID)
		return resp // Handled=false → 调用方继续原逻辑
	}

	resp.DecisionLevel = match.DecisionLevel
	resp.ActionType = match.ActionType
	resp.AutoExecutable = match.AutoExecutable

	// 3. 闸门检查（受 risk_gate_enabled 灰度控制）
	var gateResult *GateCheckResult
	if dc.isRiskGateEnabled(ctx) {
		gateResult = dc.riskGate.Check(ctx, req, family, categoryCode)
	}
	// 记录原动作（闸门命中时写入 recommendation 供审计追溯）
	originalActionType := resp.ActionType

	if gateResult != nil && gateResult.Blocked {
		g.Log().Infof(ctx, "[DecisionCenter] 闸门阻断: trigger=%s gates=%d",
			req.TriggerSource, len(gateResult.BlockedGates))
		// 闸门阻断 → 强制降级为 C 级（必须人工）+ 动作切换为 fallback_action
		resp.DecisionLevel = consts.DecisionLevelC
		resp.AutoExecutable = false
		resp.HumanRequired = true

		// 取第一个闸门的 fallback_action 作为降级动作
		if len(gateResult.BlockedGates) > 0 && gateResult.BlockedGates[0].FallbackAction != "" {
			resp.ActionType = gateResult.BlockedGates[0].FallbackAction
		}

		dc.emitEvent(ctx, event.EventAutonomyGateBlocked, event.EntityDecisionAction,
			req.WorkflowRunID, 0, g.Map{
				"trigger_source":      req.TriggerSource,
				"blocked_gates":       gateResult.BlockedGates,
				"original_action":     originalActionType,
				"fallback_action":     resp.ActionType,
			})
	}

	// 4. 写审计记录 (mvp_decision_action)
	triggerJSON, _ := json.Marshal(req.TriggerContext)
	// 闸门命中时把原动作写入 recommendation，gate ID 列表写入 matched_gate_ids
	var recommendationJSON string
	var matchedGateIDsJSON string
	if gateResult != nil && gateResult.Blocked {
		recMap := map[string]interface{}{
			"original_action": originalActionType,
			"blocked_by":      gateResult.BlockedGates,
		}
		recBytes, _ := json.Marshal(recMap)
		recommendationJSON = string(recBytes)

		gateIDs := make([]int64, 0, len(gateResult.BlockedGates))
		for _, bg := range gateResult.BlockedGates {
			gateIDs = append(gateIDs, bg.GateID)
		}
		gateIDBytes, _ := json.Marshal(gateIDs)
		matchedGateIDsJSON = string(gateIDBytes)
	}

	actionData := g.Map{
		"workflow_run_id":  req.WorkflowRunID,
		"project_id":       req.ProjectID,
		"stage_run_id":     req.StageRunID,
		"domain_task_id":   req.DomainTaskID,
		"decision_type":    match.Rule.DecisionType,
		"decision_level":   resp.DecisionLevel,
		"trigger_source":   req.TriggerSource,
		"trigger_context":  string(triggerJSON),
		"matched_rule_id":  match.Rule.ID,
		"action_type":      resp.ActionType,
		"auto_executable":  resp.AutoExecutable,
		"human_required":   resp.HumanRequired,
		"action_status":    consts.ActionStatusPending,
		"created_by":       createdBy,
		"dept_id":          deptID,
		"created_at":       gtime.Now(),
		"updated_at":       gtime.Now(),
	}
	if recommendationJSON != "" {
		actionData["recommendation"] = recommendationJSON
	}
	if matchedGateIDsJSON != "" {
		actionData["matched_gate_ids"] = matchedGateIDsJSON
	}

	actionID, err := dc.actionRepo.Create(ctx, actionData)
	if err != nil {
		g.Log().Errorf(ctx, "[DecisionCenter] 创建决策记录失败: err=%v", err)
		resp.Error = err
		return resp
	}
	resp.ActionID = actionID

	dc.emitEvent(ctx, event.EventAutonomyDecisionCreated, event.EntityDecisionAction,
		req.WorkflowRunID, actionID, g.Map{
			"decision_level":  resp.DecisionLevel,
			"action_type":     resp.ActionType,
			"auto_executable": resp.AutoExecutable,
		})

	// 5. 审计模式：只写记录，不执行，不接管
	if dc.isAuditOnly(ctx) {
		g.Log().Infof(ctx, "[DecisionCenter] audit_only 模式，仅记录: actionID=%d type=%s",
			actionID, resp.ActionType)
		// Handled=false → 调用方继续执行原逻辑
		return resp
	}

	// 6. 分流执行
	if resp.AutoExecutable && !resp.HumanRequired {
		// A 级：自动执行
		execErr := dc.actionDispatcher.Execute(ctx, actionID, resp.ActionType, req)
		resp.Executed = execErr == nil
		resp.Handled = execErr == nil // 执行成功才算接管
		if execErr != nil {
			resp.Error = execErr
			dc.emitEvent(ctx, event.EventAutonomyActionFailed, event.EntityDecisionAction,
				req.WorkflowRunID, actionID, g.Map{"error": execErr.Error()})
			// Handled=false → 执行失败时调用方应降级到原逻辑
		} else {
			// 自动执行成功 → 写 auto_executed + final_action
			_ = dc.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusAutoExecuted, g.Map{
				"final_action": resp.ActionType,
				"executed_at":  gtime.Now(),
				"result":       "ok",
			})
			dc.emitEvent(ctx, event.EventAutonomyActionExecuted, event.EntityDecisionAction,
				req.WorkflowRunID, actionID, g.Map{"action_type": resp.ActionType})
		}
	} else {
		// B/C 级：等待人工
		resp.HumanRequired = true
		resp.Handled = true // 创建人工节点 = 已接管
		_ = dc.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusWaitingHuman, nil)

		// 创建人工介入节点
		checkpointType := consts.CheckpointApproval
		if resp.DecisionLevel == consts.DecisionLevelC {
			checkpointType = consts.CheckpointEscalation
		}
		cpID, cpErr := dc.checkpointRepo.Create(ctx, g.Map{
			"decision_action_id": actionID,
			"project_id":         req.ProjectID,
			"workflow_run_id":    req.WorkflowRunID,
			"checkpoint_type":    checkpointType,
			"status":             consts.CheckpointStatusOpen,
			"title":              fmt.Sprintf("[%s] %s", resp.DecisionLevel, resp.ActionType),
			"description":        string(triggerJSON),
			"created_by":         createdBy,
			"dept_id":            deptID,
			"created_at":         gtime.Now(),
			"updated_at":         gtime.Now(),
		})
		if cpErr != nil {
			g.Log().Warningf(ctx, "[DecisionCenter] 创建人工节点失败: err=%v", cpErr)
		} else {
			dc.emitEvent(ctx, event.EventAutonomyCheckpointOpened, event.EntityHumanCheckpoint,
				req.WorkflowRunID, cpID, g.Map{
					"action_id":       actionID,
					"checkpoint_type": checkpointType,
				})
		}
	}

	return resp
}

// ApproveAction 人工批准动作，执行降级后的动作（非原动作）。
func (dc *DecisionCenter) ApproveAction(ctx context.Context, actionID int64) error {
	action, err := dc.actionRepo.GetByID(ctx, actionID)
	if err != nil || action == nil {
		return fmt.Errorf("决策记录不存在: %d", actionID)
	}

	status := mapString(action, "action_status")
	if status != consts.ActionStatusWaitingHuman {
		return fmt.Errorf("决策状态不允许审批: %s", status)
	}

	_ = dc.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusApproved, g.Map{
		"executed_at": gtime.Now(),
	})

	// 更新人工节点
	wfRunID := mapInt64(action, "workflow_run_id")
	cp, _ := dc.checkpointRepo.GetByActionID(ctx, actionID)
	if cp != nil {
		cpID := mapInt64(cp, "id")
		_ = dc.checkpointRepo.UpdateHandle(ctx, cpID, g.Map{
			"status":        consts.CheckpointStatusHandled,
			"handle_action": consts.HandleActionApprove,
			"handled_at":    gtime.Now(),
		})
		dc.emitEvent(ctx, event.EventAutonomyCheckpointHandled, event.EntityHumanCheckpoint,
			wfRunID, cpID, g.Map{"action": consts.HandleActionApprove})
	}

	// 执行 action_type（即闸门降级后的 fallback_action，不是原始动作）
	actionType := mapString(action, "action_type")
	req := dc.rebuildRequest(action)
	execErr := dc.actionDispatcher.Execute(ctx, actionID, actionType, req)
	if execErr != nil {
		dc.emitEvent(ctx, event.EventAutonomyActionFailed, event.EntityDecisionAction,
			wfRunID, actionID, g.Map{"error": execErr.Error()})
		return execErr
	}

	// 执行成功 → 保持 approved 状态不变，只写 final_action + result
	_ = dc.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusApproved, g.Map{
		"final_action": actionType,
		"result":       "ok",
	})

	dc.emitEvent(ctx, event.EventAutonomyActionExecuted, event.EntityDecisionAction,
		wfRunID, actionID, g.Map{"action_type": actionType})
	return nil
}

// RejectAction 人工驳回动作。
func (dc *DecisionCenter) RejectAction(ctx context.Context, actionID int64, reason string) error {
	action, err := dc.actionRepo.GetByID(ctx, actionID)
	if err != nil || action == nil {
		return fmt.Errorf("决策记录不存在: %d", actionID)
	}

	status := mapString(action, "action_status")
	if status != consts.ActionStatusWaitingHuman {
		return fmt.Errorf("决策状态不允许驳回: %s", status)
	}

	_ = dc.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusRejected, g.Map{
		"result": reason,
	})

	wfRunID := mapInt64(action, "workflow_run_id")
	cp, _ := dc.checkpointRepo.GetByActionID(ctx, actionID)
	if cp != nil {
		cpID := mapInt64(cp, "id")
		_ = dc.checkpointRepo.UpdateHandle(ctx, cpID, g.Map{
			"status":        consts.CheckpointStatusHandled,
			"handle_action": consts.HandleActionReject,
			"handle_reason":  reason,
			"handled_at":    gtime.Now(),
		})
		dc.emitEvent(ctx, event.EventAutonomyCheckpointHandled, event.EntityHumanCheckpoint,
			wfRunID, cpID, g.Map{"action": consts.HandleActionReject, "reason": reason})
	}

	return nil
}

// ListOpenCheckpoints 查询项目下未处理的人工节点。
func (dc *DecisionCenter) ListOpenCheckpoints(ctx context.Context, projectID int64) ([]g.Map, error) {
	return dc.checkpointRepo.ListOpen(ctx, projectID)
}

// ListPendingActions 查询项目下待处理的决策动作。
func (dc *DecisionCenter) ListPendingActions(ctx context.Context, projectID int64) ([]g.Map, error) {
	return dc.actionRepo.ListPending(ctx, projectID)
}

// emitEvent 发布事件的便捷方法。
func (dc *DecisionCenter) emitEvent(ctx context.Context, eventType, entityType string, wfRunID, entityID int64, payload interface{}) {
	if dc.eventPublisher == nil {
		return
	}
	evt := event.Event{
		WorkflowRunID: wfRunID,
		EntityType:    entityType,
		EventType:     eventType,
		Payload:       payload,
	}
	if entityID > 0 {
		evt.EntityID = &entityID
	}
	_ = dc.eventPublisher.Emit(ctx, evt)
}

// resolveProjectScope 读取项目的 family、category_code 和数据权限归属。
func (dc *DecisionCenter) resolveProjectScope(ctx context.Context, projectID int64) (family, categoryCode string, createdBy, deptID int64) {
	if projectID == 0 {
		return "", "", 0, 0
	}
	record, err := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).
		Fields("project_family, project_category, created_by, dept_id").
		One()
	if err != nil || record.IsEmpty() {
		return "", "", 0, 0
	}
	return record["project_family"].String(), record["project_category"].String(),
		record["created_by"].Int64(), record["dept_id"].Int64()
}

// rebuildRequest 从审计记录重建 DecisionRequest（用于人工审批后执行）。
func (dc *DecisionCenter) rebuildRequest(action g.Map) *DecisionRequest {
	req := &DecisionRequest{
		WorkflowRunID: mapInt64(action, "workflow_run_id"),
		ProjectID:     mapInt64(action, "project_id"),
		StageRunID:    mapInt64(action, "stage_run_id"),
		DomainTaskID:  mapInt64(action, "domain_task_id"),
		TriggerSource: mapString(action, "trigger_source"),
	}
	if tc := mapString(action, "trigger_context"); tc != "" {
		req.TriggerContext = parseJSONMap(tc)
	}
	return req
}
