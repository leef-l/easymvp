package autonomy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

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
	sensor           *Sensor
	objectiveSvc     *ObjectiveService
	planner          *Planner
	actuator         *Actuator
	observer         *MetaObserver
	learner          *Learner
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

// SetPhaseADeps 注入 Phase A 的地基组件。
func (dc *DecisionCenter) SetPhaseADeps(sensor *Sensor, objectiveSvc *ObjectiveService) {
	dc.sensor = sensor
	dc.objectiveSvc = objectiveSvc
}

// SetPhaseBDeps 注入 Phase B 的策略组件。
func (dc *DecisionCenter) SetPhaseBDeps(planner *Planner, actuator *Actuator) {
	dc.planner = planner
	dc.actuator = actuator
}

// SetPhaseDDeps 注入 Phase D 的元认知组件。
func (dc *DecisionCenter) SetPhaseDDeps(observer *MetaObserver, learner *Learner) {
	dc.observer = observer
	dc.learner = learner
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

// isPatrolEnabled 态势采集灰度。
func (dc *DecisionCenter) isPatrolEnabled(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.autonomy.patrol_enabled", "workflow.autonomy.patrolEnabled", 0) == 1
}

// isObjectiveEnabled 目标层灰度。
func (dc *DecisionCenter) isObjectiveEnabled(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.autonomy.objective_enabled", "workflow.autonomy.objectiveEnabled", 0) == 1
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

	// 2. Phase A：统一态势感知 + 目标层准入
	var sit *Situation
	if dc.sensor != nil && req.WorkflowRunID > 0 && (dc.isPatrolEnabled(ctx) || dc.isObjectiveEnabled(ctx)) {
		if perceived, err := dc.sensor.Perceive(ctx, req.WorkflowRunID); err != nil {
			g.Log().Warningf(ctx, "[DecisionCenter] Situation 感知失败: wfRun=%d err=%v", req.WorkflowRunID, err)
		} else {
			sit = perceived
			_ = dc.sensor.RecordSnapshot(ctx, sit)
		}
	}
	if dc.objectiveSvc != nil && dc.isObjectiveEnabled(ctx) && sit != nil && req.ProjectID > 0 {
		obj, err := dc.objectiveSvc.Load(ctx, req.ProjectID)
		if err != nil {
			g.Log().Warningf(ctx, "[DecisionCenter] Objective 加载失败: project=%d err=%v", req.ProjectID, err)
		} else if admission, _ := dc.objectiveSvc.Check(ctx, sit, obj, req.TriggerSource); admission != nil && !admission.Allowed {
			return dc.handleAdmissionDenied(ctx, req, admission, sit, createdBy, deptID)
		}
	}

	// 3. Phase B：Planner 策略函数评估（受 strategy_enabled 灰度控制）
	var plan *ActionPlan
	if dc.planner != nil && dc.planner.IsEnabled(ctx) && sit != nil {
		plan = dc.planner.Plan(ctx, sit, req)
		if plan != nil {
			g.Log().Infof(ctx, "[DecisionCenter] Planner 输出计划: strategy=%s action=%s level=%s",
				plan.StrategyName, plan.ActionType, plan.DecisionLevel)
		}
	}

	// 4. 策略规则匹配（受 policy_engine_enabled 灰度控制）
	// Planner 有结果时优先使用，否则走 PolicyEngine
	var match *PolicyMatch
	if plan != nil {
		// Planner 结果转为 PolicyMatch 格式，统一后续流程
		match = &PolicyMatch{
			DecisionLevel:  plan.DecisionLevel,
			ActionType:     plan.ActionType,
			AutoExecutable: plan.DecisionLevel == consts.DecisionLevelA,
		}
	} else if dc.isPolicyEngineEnabled(ctx) {
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

	// 4. 闸门检查（受 risk_gate_enabled 灰度控制）
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

	// 5. 写审计记录 (mvp_decision_action)
	triggerJSON, tjErr := json.Marshal(req.TriggerContext)
	if tjErr != nil {
		triggerJSON = []byte("{}")
	}
	// 闸门命中时把原动作写入 recommendation，gate ID 列表写入 matched_gate_ids
	var recommendationJSON string
	var matchedGateIDsJSON string
	if gateResult != nil && gateResult.Blocked {
		recMap := map[string]interface{}{
			"original_action": originalActionType,
			"blocked_by":      gateResult.BlockedGates,
		}
		recBytes, recMErr := json.Marshal(recMap)
		if recMErr != nil {
			recBytes = []byte("{}")
		}
		recommendationJSON = string(recBytes)

		gateIDs := make([]int64, 0, len(gateResult.BlockedGates))
		for _, bg := range gateResult.BlockedGates {
			gateIDs = append(gateIDs, bg.GateID)
		}
		gateIDBytes, gidErr := json.Marshal(gateIDs)
		if gidErr != nil {
			gateIDBytes = []byte("[]")
		}
		matchedGateIDsJSON = string(gateIDBytes)
	}

	// 来源判断：Planner 策略 vs PolicyEngine 规则
	var decisionType string
	var matchedRuleID int64
	if plan != nil {
		decisionType = "strategy:" + plan.StrategyName
	} else if match.Rule != nil {
		decisionType = match.Rule.DecisionType
		matchedRuleID = match.Rule.ID
	}

	actionData := g.Map{
		"workflow_run_id":  req.WorkflowRunID,
		"project_id":       req.ProjectID,
		"stage_run_id":     req.StageRunID,
		"domain_task_id":   req.DomainTaskID,
		"decision_type":    decisionType,
		"decision_level":   resp.DecisionLevel,
		"trigger_source":   req.TriggerSource,
		"trigger_context":  string(triggerJSON),
		"matched_rule_id":  matchedRuleID,
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
	} else if plan != nil {
		// Planner 策略的 reasoning 写入 recommendation
		recMap := map[string]interface{}{
			"strategy":        plan.StrategyName,
			"reasoning":       plan.Reasoning,
			"expected_outcome": plan.ExpectedOutcome,
			"rollback_action": plan.RollbackAction,
		}
		if plan.Meta != nil {
			recMap["confidence"] = plan.Meta.Confidence
			recMap["blast_radius"] = plan.Meta.BlastRadius
		}
		recBytes, recMarshalErr := json.Marshal(recMap)
		if recMarshalErr != nil {
			recBytes = []byte("{}")
		}
		actionData["recommendation"] = string(recBytes)
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

	// 6. 审计模式：只写记录，不执行，不接管
	if dc.isAuditOnly(ctx) {
		g.Log().Infof(ctx, "[DecisionCenter] audit_only 模式，仅记录: actionID=%d type=%s",
			actionID, resp.ActionType)
		// Phase D：观测记录（审计模式也记录）
		dc.recordObservation(ctx, actionID, req, resp, decisionType, plan, sit, createdBy, deptID)
		// Handled=false → 调用方继续执行原逻辑
		return resp
	}

	// 7. 分流执行
	// 将 Planner 策略的 Parameters 注入 TriggerContext，供 ActionDispatcher 回调使用
	if plan != nil && len(plan.Parameters) > 0 {
		if req.TriggerContext == nil {
			req.TriggerContext = make(map[string]interface{})
		}
		for k, v := range plan.Parameters {
			req.TriggerContext[k] = v
		}
	}
	if resp.AutoExecutable && !resp.HumanRequired {
		// Actuator: 记录执行前态势
		var outcomeID int64
		strategyName := ""
		if plan != nil {
			strategyName = plan.StrategyName
		}
		if dc.actuator != nil && sit != nil && strategyName != "" {
			outcomeID = dc.actuator.RecordBefore(ctx, actionID, req.WorkflowRunID, req.ProjectID,
				strategyName, resp.ActionType, resp.DecisionLevel, sit, createdBy, deptID)
		}

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
			if upErr := dc.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusAutoExecuted, g.Map{
				"final_action": resp.ActionType,
				"executed_at":  gtime.Now(),
				"result":       "ok",
			}); upErr != nil {
				g.Log().Warningf(ctx, "[DecisionCenter] 更新 action 状态失败: actionID=%d err=%v", actionID, upErr)
			}
			dc.emitEvent(ctx, event.EventAutonomyActionExecuted, event.EntityDecisionAction,
				req.WorkflowRunID, actionID, g.Map{"action_type": resp.ActionType})

			// Actuator: 延迟评估效果（执行后重新采集态势）
			if dc.actuator != nil && dc.sensor != nil && outcomeID > 0 {
				if sitAfter, err := dc.sensor.Perceive(ctx, req.WorkflowRunID); err == nil {
					dc.actuator.EvaluateAfter(ctx, outcomeID, sitAfter)
				}
			}
		}
	} else {
		// B/C 级：等待人工
		resp.HumanRequired = true
		resp.Handled = true // 创建人工节点 = 已接管
		if upErr := dc.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusWaitingHuman, nil); upErr != nil {
			g.Log().Warningf(ctx, "[DecisionCenter] 更新 action 为 waiting_human 失败: actionID=%d err=%v", actionID, upErr)
		}

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

	// Phase D：观测记录
	outcome := "pending"
	if resp.Executed {
		outcome = "success"
	} else if resp.Error != nil {
		outcome = "failure"
	}
	dc.recordObservation(ctx, actionID, req, resp, decisionType, plan, sit, createdBy, deptID)
	// Phase D：如果已有确定结果，回填观测 + 喂给 Learner
	if outcome != "pending" && dc.observer != nil && dc.observer.IsEnabled(ctx) {
		dc.observer.UpdateOutcome(ctx, actionID, outcome, 0)
		if dc.learner != nil && dc.learner.IsEnabled(ctx) {
			dc.learner.FeedFromObservation(ctx, &ObservationInput{
				DecisionActionID: actionID,
				ProjectID:        req.ProjectID,
				DecisionType:     decisionType,
				Outcome:          outcome,
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

	if upErr := dc.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusApproved, g.Map{
		"executed_at": gtime.Now(),
	}); upErr != nil {
		g.Log().Warningf(ctx, "[DecisionCenter] 审批更新 action 状态失败: actionID=%d err=%v", actionID, upErr)
	}

	// 更新人工节点
	wfRunID := mapInt64(action, "workflow_run_id")
	cp, _ := dc.checkpointRepo.GetByActionID(ctx, actionID)
	if cp != nil {
		cpID := mapInt64(cp, "id")
		if upErr := dc.checkpointRepo.UpdateHandle(ctx, cpID, g.Map{
			"status":        consts.CheckpointStatusHandled,
			"handle_action": consts.HandleActionApprove,
			"handled_at":    gtime.Now(),
		}); upErr != nil {
			g.Log().Warningf(ctx, "[DecisionCenter] 更新 checkpoint 状态失败: cpID=%d err=%v", cpID, upErr)
		}
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
	if upErr := dc.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusApproved, g.Map{
		"final_action": actionType,
		"result":       "ok",
	}); upErr != nil {
		g.Log().Warningf(ctx, "[DecisionCenter] 审批后更新 action 失败: actionID=%d err=%v", actionID, upErr)
	}

	dc.emitEvent(ctx, event.EventAutonomyActionExecuted, event.EntityDecisionAction,
		wfRunID, actionID, g.Map{"action_type": actionType})

	// Phase D：记录人工批准事件
	if dc.observer != nil && dc.observer.IsEnabled(ctx) {
		dc.observer.UpdateHumanOverride(ctx, actionID, "approve", "")
		if dc.learner != nil && dc.learner.IsEnabled(ctx) {
			dc.learner.FeedFromObservation(ctx, &ObservationInput{
				DecisionActionID: actionID,
				ProjectID:        mapInt64(action, "project_id"),
				DecisionType:     mapString(action, "decision_type"),
				Outcome:          "success",
				HumanOverride:    true,
			})
		}
	}

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

	if upErr := dc.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusRejected, g.Map{
		"result": reason,
	}); upErr != nil {
		g.Log().Warningf(ctx, "[DecisionCenter] 驳回更新 action 状态失败: actionID=%d err=%v", actionID, upErr)
	}

	wfRunID := mapInt64(action, "workflow_run_id")
	cp, _ := dc.checkpointRepo.GetByActionID(ctx, actionID)
	if cp != nil {
		cpID := mapInt64(cp, "id")
		if upErr := dc.checkpointRepo.UpdateHandle(ctx, cpID, g.Map{
			"status":        consts.CheckpointStatusHandled,
			"handle_action": consts.HandleActionReject,
			"handle_reason":  reason,
			"handled_at":    gtime.Now(),
		}); upErr != nil {
			g.Log().Warningf(ctx, "[DecisionCenter] 驳回更新 checkpoint 状态失败: cpID=%d err=%v", cpID, upErr)
		}
		dc.emitEvent(ctx, event.EventAutonomyCheckpointHandled, event.EntityHumanCheckpoint,
			wfRunID, cpID, g.Map{"action": consts.HandleActionReject, "reason": reason})
	}

	// Phase D：记录人工驳回事件（强信号 1.0）
	if dc.observer != nil && dc.observer.IsEnabled(ctx) {
		dc.observer.UpdateHumanOverride(ctx, actionID, "reject", reason)
		if dc.learner != nil && dc.learner.IsEnabled(ctx) {
			dc.learner.FeedFromObservation(ctx, &ObservationInput{
				DecisionActionID: actionID,
				ProjectID:        mapInt64(action, "project_id"),
				DecisionType:     mapString(action, "decision_type"),
				Outcome:          "failure",
				HumanOverride:    true,
				OverrideReason:   reason,
			})
		}
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

// ListAllActions 查询项目全量决策记录（支持可选的状态和类型过滤）。
func (dc *DecisionCenter) ListAllActions(ctx context.Context, projectID int64, actionStatus, decisionType string) ([]g.Map, error) {
	return dc.actionRepo.ListByProjectFiltered(ctx, projectID, actionStatus, decisionType)
}

// ListGateRules 查询项目适用的风险闸门规则。
func (dc *DecisionCenter) ListGateRules(ctx context.Context, projectID int64) ([]g.Map, error) {
	family, categoryCode, _, _ := dc.resolveProjectScope(ctx, projectID)
	return dc.riskGate.ListRules(ctx, family, categoryCode)
}

// ListPolicyRules 查询项目适用的策略规则。
func (dc *DecisionCenter) ListPolicyRules(ctx context.Context, projectID int64) ([]g.Map, error) {
	family, categoryCode, _, _ := dc.resolveProjectScope(ctx, projectID)
	return dc.policyEngine.ListRules(ctx, family, categoryCode)
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
		Fields("category_code, project_category, created_by, dept_id").
		One()
	if err != nil || record.IsEmpty() {
		return "", "", 0, 0
	}
	categoryCode = record["category_code"].String()
	if categoryCode == "" {
		categoryCode = record["project_category"].String()
	}
	family = dc.resolveProjectFamily(ctx, categoryCode)
	return family, categoryCode, record["created_by"].Int64(), record["dept_id"].Int64()
}

func (dc *DecisionCenter) resolveProjectFamily(ctx context.Context, categoryCode string) string {
	if categoryCode == "" {
		return ""
	}
	row, err := repo.NewProjectCategoryRepo().GetByCode(ctx, categoryCode)
	if err == nil && row != nil {
		return gconv.String(row["family_code"])
	}
	row, err = repo.NewProjectCategoryRepo().GetByDisplayName(ctx, categoryCode)
	if err == nil && row != nil {
		return gconv.String(row["family_code"])
	}
	return ""
}

func (dc *DecisionCenter) handleAdmissionDenied(
	ctx context.Context,
	req *DecisionRequest,
	admission *AdmissionResult,
	sit *Situation,
	createdBy, deptID int64,
) *DecisionResponse {
	resp := &DecisionResponse{
		DecisionLevel: consts.DecisionLevelC,
		ActionType:    consts.ActionTypeNotifyHuman,
		HumanRequired: true,
		Handled:       true,
		DenyReason:    admission.DenyReason,
	}
	triggerJSON, tErr := json.Marshal(g.Map{
		"trigger_context": req.TriggerContext,
		"admission":       admission,
		"situation":       sit,
	})
	if tErr != nil {
		triggerJSON = []byte("{}")
	}
	actionID, err := dc.actionRepo.Create(ctx, g.Map{
		"workflow_run_id": req.WorkflowRunID,
		"project_id":      req.ProjectID,
		"stage_run_id":    req.StageRunID,
		"domain_task_id":  req.DomainTaskID,
		"decision_type":   "objective_guard",
		"decision_level":  consts.DecisionLevelC,
		"trigger_source":  req.TriggerSource,
		"trigger_context": string(triggerJSON),
		"action_type":     admission.SuggestedAction,
		"auto_executable": 0,
		"human_required":  1,
		"action_status":   consts.ActionStatusWaitingHuman,
		"result":          admission.DenyReason,
		"created_by":      createdBy,
		"dept_id":         deptID,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	})
	if err != nil {
		g.Log().Errorf(ctx, "[DecisionCenter] 准入拒绝记录失败: err=%v", err)
		resp.Error = err
		return resp
	}
	resp.ActionID = actionID
	if _, cpErr := dc.checkpointRepo.Create(ctx, g.Map{
		"decision_action_id": actionID,
		"project_id":         req.ProjectID,
		"workflow_run_id":    req.WorkflowRunID,
		"checkpoint_type":    consts.CheckpointEscalation,
		"status":             consts.CheckpointStatusOpen,
		"title":              "[C] objective_guard",
		"description":        admission.DenyReason,
		"created_by":         createdBy,
		"dept_id":            deptID,
		"created_at":         gtime.Now(),
		"updated_at":         gtime.Now(),
	}); cpErr != nil {
		g.Log().Errorf(ctx, "[DecisionCenter] 创建检查点失败: err=%v", cpErr)
	}
	dc.emitEvent(ctx, event.EventAutonomyCheckpointOpened, event.EntityHumanCheckpoint,
		req.WorkflowRunID, 0, g.Map{"action_id": actionID, "reason": admission.DenyReason})
	return resp
}

// recordObservation Phase D 观测记录的便捷方法。
func (dc *DecisionCenter) recordObservation(ctx context.Context, actionID int64, req *DecisionRequest, resp *DecisionResponse, decisionType string, plan *ActionPlan, sit *Situation, createdBy, deptID int64) {
	if dc.observer == nil || !dc.observer.IsEnabled(ctx) {
		return
	}

	input := &ObservationInput{
		DecisionActionID: actionID,
		WorkflowRunID:    req.WorkflowRunID,
		ProjectID:        req.ProjectID,
		DecisionType:     decisionType,
		TriggerSource:    req.TriggerSource,
		DecisionLevel:    resp.DecisionLevel,
		ActionType:       resp.ActionType,
		InputSnapshot:    req.TriggerContext,
		OutputSnapshot: map[string]interface{}{
			"action_type":    resp.ActionType,
			"decision_level": resp.DecisionLevel,
			"auto_executable": resp.AutoExecutable,
			"human_required": resp.HumanRequired,
			"handled":        resp.Handled,
		},
		CreatedBy: createdBy,
		DeptID:    deptID,
	}

	if plan != nil && plan.Meta != nil {
		input.Meta = plan.Meta
	}

	dc.observer.Record(ctx, input)
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

// isEventDrivenEnabled 事件驱动灰度开关（默认关闭，需显式设置 workflow.autonomy.event_driven=1 开启）。
func (dc *DecisionCenter) isEventDrivenEnabled(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.autonomy.event_driven", "workflow.autonomy.eventDriven", 0) == 1
}

// stageTypeToTrigger 将阶段类型映射为自治触发源。
func (dc *DecisionCenter) stageTypeToTrigger(stageType string) string {
	switch stageType {
	case "review":
		return consts.TriggerReviewCompleted
	case "execute":
		return consts.TriggerExecuteCompleted
	case "accept":
		return consts.TriggerAcceptPassed
	case "rework":
		return consts.TriggerReworkCompleted
	default:
		return ""
	}
}

// SubscribeEvents 订阅工作流事件总线，实现事件驱动的自治决策。
// 监听关键转换事件，自动触发 Decide 流程，与现有硬编码回调并存。
// 默认关闭，需设置配置项 workflow.autonomy.event_driven=1 后生效。
func (dc *DecisionCenter) SubscribeEvents(bus *event.Bus) {
	// 监听工作流阶段完成事件
	bus.Subscribe(event.EventStageCompleted, func(evt event.Event) {
		ctx := context.Background()
		if !dc.IsEnabled(ctx) || !dc.isEventDrivenEnabled(ctx) {
			return
		}
		ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		// 从事件中提取 stageRunID
		stageRunID := int64(0)
		if evt.StageRunID != nil {
			stageRunID = *evt.StageRunID
		}

		// 查询阶段类型
		stageType := ""
		if stageRunID > 0 {
			row, err := g.DB().Model("mvp_stage_run").Ctx(ctx).
				Where("id", stageRunID).Fields("stage_type").One()
			if err == nil && !row.IsEmpty() {
				stageType = row["stage_type"].String()
			}
		}

		triggerSource := dc.stageTypeToTrigger(stageType)
		if triggerSource == "" {
			return
		}

		// 查询 project_id
		projectID := int64(0)
		if evt.WorkflowRunID > 0 {
			pid, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
				Where("id", evt.WorkflowRunID).Value("project_id")
			if err == nil {
				projectID = pid.Int64()
			}
		}

		g.Log().Debugf(ctx, "[DecisionCenter] 事件驱动触发: wfRun=%d stage=%s trigger=%s",
			evt.WorkflowRunID, stageType, triggerSource)
		dc.Decide(ctx, &DecisionRequest{
			WorkflowRunID: evt.WorkflowRunID,
			ProjectID:     projectID,
			StageRunID:    stageRunID,
			TriggerSource: triggerSource,
		})
	})

	// 监听任务失败事件
	bus.Subscribe(event.EventTaskFailed, func(evt event.Event) {
		ctx := context.Background()
		if !dc.IsEnabled(ctx) || !dc.isEventDrivenEnabled(ctx) {
			return
		}
		ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		taskID := int64(0)
		if evt.EntityID != nil {
			taskID = *evt.EntityID
		}

		projectID := int64(0)
		if evt.WorkflowRunID > 0 {
			pid, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
				Where("id", evt.WorkflowRunID).Value("project_id")
			if err == nil {
				projectID = pid.Int64()
			}
		}

		g.Log().Debugf(ctx, "[DecisionCenter] 事件驱动触发: wfRun=%d task=%d trigger=%s",
			evt.WorkflowRunID, taskID, consts.TriggerTaskFailed)
		dc.Decide(ctx, &DecisionRequest{
			WorkflowRunID: evt.WorkflowRunID,
			ProjectID:     projectID,
			DomainTaskID:  taskID,
			TriggerSource: consts.TriggerTaskFailed,
		})
	})
}
