package chat

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

func TestMapToDecisionActionDTO(t *testing.T) {
	t.Parallel()

	executedAt := gtime.NewFromTime(time.Now())
	createdAt := gtime.NewFromTime(time.Now().Add(-time.Minute))
	dto := mapToDecisionActionDTO(g.Map{
		"id":              int64(1),
		"workflow_run_id": json.Number("2"),
		"project_id":      float64(3),
		"stage_run_id":    int64(4),
		"domain_task_id":  int64(5),
		"decision_type":   "risk_gate",
		"decision_level":  "B",
		"trigger_source":  "system",
		"trigger_context": map[string]interface{}{"risk": "high"},
		"matched_rule_id": json.Number("6"),
		"matched_gate_ids": []int64{
			7, 8,
		},
		"action_type":     "pause",
		"recommendation":  map[string]interface{}{"action": "pause"},
		"final_action":    "paused",
		"action_status":   "done",
		"auto_executable": json.Number("1"),
		"human_required":  float64(0),
		"executed_at":     executedAt,
		"result":          map[string]interface{}{"ok": true},
		"created_at":      createdAt,
	})

	if dto.ID != 1 || dto.WorkflowRunID != 2 || dto.ProjectID != 3 || dto.StageRunID != 4 || dto.DomainTaskID != 5 {
		t.Fatalf("unexpected id mapping: %+v", dto)
	}
	if dto.DecisionType != "risk_gate" || dto.ActionType != "pause" || dto.ActionStatus != "done" {
		t.Fatalf("unexpected action mapping: %+v", dto)
	}
	if dto.AutoExecutable != 1 || dto.HumanRequired != 0 {
		t.Fatalf("unexpected int mapping: %+v", dto)
	}
	if dto.ExecutedAt == nil || dto.CreatedAt == nil {
		t.Fatalf("expected times to be mapped: %+v", dto)
	}
	if dto.TriggerContext != `{"risk":"high"}` {
		t.Fatalf("unexpected trigger context: %s", dto.TriggerContext)
	}
	if dto.MatchedGateIDs != `[7,8]` {
		t.Fatalf("unexpected matched gate ids: %s", dto.MatchedGateIDs)
	}
	if dto.Result != `{"ok":true}` {
		t.Fatalf("unexpected result json: %s", dto.Result)
	}
}

func TestMapToCheckpointDTO(t *testing.T) {
	t.Parallel()

	handledAt := gtime.NewFromTime(time.Now())
	dto := mapToCheckpointDTO(g.Map{
		"id":                 int64(11),
		"workflow_run_id":    int64(12),
		"project_id":         int64(13),
		"decision_action_id": int64(14),
		"checkpoint_type":    "manual_review",
		"title":              "人工审核",
		"description":        "需要人工确认",
		"status":             "open",
		"assigned_to":        int64(15),
		"handled_by":         int64(16),
		"handle_action":      "approve",
		"handle_reason":      "确认通过",
		"handled_at":         handledAt,
	})

	if dto.ID != 11 || dto.ProjectID != 13 || dto.DecisionActionID != 14 {
		t.Fatalf("unexpected checkpoint mapping: %+v", dto)
	}
	if dto.CheckpointType != "manual_review" || dto.HandleAction != "approve" {
		t.Fatalf("unexpected checkpoint fields: %+v", dto)
	}
	if dto.HandledAt == nil {
		t.Fatalf("expected handled_at mapping: %+v", dto)
	}
}

func TestMapRuleDTOs(t *testing.T) {
	t.Parallel()

	gate := mapToRiskGateRuleDTO(g.Map{
		"id":                    int64(21),
		"gate_code":             "high_risk_block",
		"gate_name":             "高风险阻断",
		"gate_type":             "risk",
		"project_family":        "coding",
		"project_category_code": "software_dev",
		"trigger_expression":    map[string]interface{}{"risk": "high"},
		"block_action":          "pause",
		"fallback_action":       "manual_review",
		"enabled":               json.Number("1"),
		"priority":              float64(90),
	})
	if gate.ID != 21 || gate.Enabled != 1 || gate.Priority != 90 {
		t.Fatalf("unexpected gate mapping: %+v", gate)
	}
	if gate.TriggerExpression != `{"risk":"high"}` {
		t.Fatalf("unexpected gate trigger expression: %s", gate.TriggerExpression)
	}

	policy := mapToPolicyRuleDTO(g.Map{
		"id":                    int64(31),
		"rule_code":             "cost_guard",
		"rule_name":             "成本守卫",
		"decision_type":         "budget",
		"decision_level":        "C",
		"trigger_source":        "observer",
		"project_family":        "coding",
		"project_category_code": "software_dev",
		"config_json":           map[string]interface{}{"threshold": 0.8},
		"enabled":               1,
		"priority":              100,
	})
	if policy.ID != 31 || policy.RuleCode != "cost_guard" || policy.Enabled != 1 || policy.Priority != 100 {
		t.Fatalf("unexpected policy mapping: %+v", policy)
	}
	if policy.ConfigJSON != `{"threshold":0.8}` {
		t.Fatalf("unexpected policy config json: %s", policy.ConfigJSON)
	}
}
