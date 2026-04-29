package braincontracts

// StageMatrixEntry defines the brain routing rules for a single workflow stage.
// It is the code-level representation of the 6-stage brain dominance matrix
// described in the Qian Xuesen master design.
type StageMatrixEntry struct {
	StageKey           string   `json:"stage_key"`
	StageName          string   `json:"stage_name"`
	PrimaryBrain       string   `json:"primary_brain"`
	SupportingBrains   []string `json:"supporting_brains"`
	TypicalOutputs     []string `json:"typical_outputs"`
	EscalationTriggers []string `json:"escalation_triggers"`
}

// StageMatrix is the canonical 6-stage brain dominance matrix.
// It answers: at each stage, which brain leads, which assists, and what
// conditions trigger escalation.
var StageMatrix = []StageMatrixEntry{
	{
		StageKey:         "designing",
		StageName:        "方案设计",
		PrimaryBrain:     "easymvp",
		SupportingBrains: []string{"central"},
		TypicalOutputs:   []string{"PlanDraft", "risk_hints", "acceptance_hints"},
		EscalationTriggers: []string{
			"goal_unclear",
			"scope_uncontrolled",
			"category_unknown",
			"input_conflict",
		},
	},
	{
		StageKey:         "reviewing",
		StageName:        "方案审核",
		PrimaryBrain:     "easymvp",
		SupportingBrains: []string{"central", "verifier"},
		TypicalOutputs:   []string{"PlanReviewResult", "blocking_issues", "advisory_issues", "compile_decision"},
		EscalationTriggers: []string{
			"blocking_issue_hit",
			"verification_requirement_missing",
			"role_resolution_failure",
		},
	},
	{
		StageKey:         "executing",
		StageName:        "任务执行",
		PrimaryBrain:     "code",
		SupportingBrains: []string{"central", "browser", "verifier"},
		TypicalOutputs:   []string{"code_changes", "task_logs", "deliverables", "partial_run_results"},
		EscalationTriggers: []string{
			"execution_failure",
			"artifact_missing",
			"environment_unsatisfied",
			"evidence_gap",
		},
	},
	{
		StageKey:         "accepting",
		StageName:        "验证验收",
		PrimaryBrain:     "verifier", // may switch to "browser" per task type
		SupportingBrains: []string{"central", "easymvp"},
		TypicalOutputs:   []string{"VerificationResult", "Evidence", "acceptance_coverage"},
		EscalationTriggers: []string{
			"blocking_gate_failed",
			"critical_journey_evidence_missing",
			"results_contradictory",
		},
	},
	{
		StageKey:         "reworking",
		StageName:        "故障返工",
		PrimaryBrain:     "fault", // fault classifies first, then easymvp redesigns
		SupportingBrains: []string{"central", "code", "verifier", "easymvp"},
		TypicalOutputs:   []string{"FaultSummary", "RepairPlanDraft", "adjusted_contracts"},
		EscalationTriggers: []string{
			"consecutive_failures",
			"risk_escalation",
			"manual_checkpoint_required",
		},
	},
	{
		StageKey:         "completed",
		StageName:        "完成裁决",
		PrimaryBrain:     "easymvp",
		SupportingBrains: []string{"central", "verifier"},
		TypicalOutputs:   []string{"CompletionVerdict", "completion_reason", "archive_pointer"},
		EscalationTriggers: []string{
			"functional_passed_but_production_not",
			"manual_release_missing",
		},
	},
}

// GetStageMatrixEntry returns the matrix entry for a given stage key.
func GetStageMatrixEntry(stageKey string) (StageMatrixEntry, bool) {
	for _, entry := range StageMatrix {
		if entry.StageKey == stageKey {
			return entry, true
		}
	}
	return StageMatrixEntry{}, false
}

// GetPrimaryBrainForStage returns the dominant brain kind for a stage.
func GetPrimaryBrainForStage(stageKey string) string {
	entry, ok := GetStageMatrixEntry(stageKey)
	if !ok {
		return "central"
	}
	return entry.PrimaryBrain
}
