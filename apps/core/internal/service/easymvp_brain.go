package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
)

const (
	easyMVPBrainErrorCodeContractInvalid = "BRN_001"
	easyMVPBrainErrorCodeUnsupportedKind = "BRN_002"
)

type IEasyMVPBrain interface {
	ResolveClientConfig(ctx context.Context) (*EasyMVPBrainClientConfig, error)
	ExecuteContract(ctx context.Context, cmd EasyMVPBrainExecuteCommand) (*EasyMVPBrainExecuteResult, error)
	CallPlanReview(ctx context.Context, input braincontracts.PlanReviewInput) (*braincontracts.BrainContractEnvelope, *braincontracts.PlanReviewResult, error)
	CallPlanCompile(ctx context.Context, input braincontracts.PlanCompileInput) (*braincontracts.BrainContractEnvelope, *braincontracts.PlanCompileResult, error)
	CallAcceptanceMapping(ctx context.Context, input braincontracts.AcceptanceMappingInput) (*braincontracts.BrainContractEnvelope, *braincontracts.AcceptanceMappingResult, error)
	CallCompletionAdjudication(ctx context.Context, input braincontracts.CompletionAdjudicationInput) (*braincontracts.BrainContractEnvelope, *braincontracts.CompletionAdjudicationResult, error)
	CallRepairDesign(ctx context.Context, input braincontracts.RepairDesignInput) (*braincontracts.BrainContractEnvelope, *braincontracts.RepairDesignResult, error)
	CallWorkspaceExplanation(ctx context.Context, input braincontracts.WorkspaceExplanationInput) (*braincontracts.BrainContractEnvelope, *braincontracts.WorkspaceExplanationResult, error)
	ValidateEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope) error
	ValidatePlanReviewEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.PlanReviewResult) error
	ValidatePlanCompileEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.PlanCompileResult) error
	ValidateAcceptanceMappingEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.AcceptanceMappingResult) error
	ValidateCompletionAdjudicationEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.CompletionAdjudicationResult) error
	ValidateRepairDesignEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.RepairDesignResult) error
	ValidateWorkspaceExplanationEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.WorkspaceExplanationResult) error
}

var localEasyMVPBrain IEasyMVPBrain = (*sEasyMVPBrain)(nil)

type sEasyMVPBrain struct{}

func EasyMVPBrain() IEasyMVPBrain {
	if localEasyMVPBrain == nil {
		localEasyMVPBrain = &sEasyMVPBrain{}
	}
	return localEasyMVPBrain
}

func (s *sEasyMVPBrain) CallPlanReview(ctx context.Context, input braincontracts.PlanReviewInput) (*braincontracts.BrainContractEnvelope, *braincontracts.PlanReviewResult, error) {
	return executeTypedContract(ctx, s, "plan_review", input, s.ValidatePlanReviewEnvelope)
}

func (s *sEasyMVPBrain) CallPlanCompile(ctx context.Context, input braincontracts.PlanCompileInput) (*braincontracts.BrainContractEnvelope, *braincontracts.PlanCompileResult, error) {
	return executeTypedContract(ctx, s, "plan_compile", input, s.ValidatePlanCompileEnvelope)
}

func (s *sEasyMVPBrain) CallAcceptanceMapping(ctx context.Context, input braincontracts.AcceptanceMappingInput) (*braincontracts.BrainContractEnvelope, *braincontracts.AcceptanceMappingResult, error) {
	return executeTypedContract(ctx, s, "acceptance_mapping", input, s.ValidateAcceptanceMappingEnvelope)
}

func (s *sEasyMVPBrain) CallCompletionAdjudication(ctx context.Context, input braincontracts.CompletionAdjudicationInput) (*braincontracts.BrainContractEnvelope, *braincontracts.CompletionAdjudicationResult, error) {
	return executeTypedContract(ctx, s, "completion_adjudication", input, s.ValidateCompletionAdjudicationEnvelope)
}

func (s *sEasyMVPBrain) CallRepairDesign(ctx context.Context, input braincontracts.RepairDesignInput) (*braincontracts.BrainContractEnvelope, *braincontracts.RepairDesignResult, error) {
	return executeTypedContract(ctx, s, "repair_design", input, s.ValidateRepairDesignEnvelope)
}

func (s *sEasyMVPBrain) CallWorkspaceExplanation(ctx context.Context, input braincontracts.WorkspaceExplanationInput) (*braincontracts.BrainContractEnvelope, *braincontracts.WorkspaceExplanationResult, error) {
	return executeTypedContract(ctx, s, "workspace_explanation", input, s.ValidateWorkspaceExplanationEnvelope)
}

func (s *sEasyMVPBrain) ValidateEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope) error {
	_ = ctx
	if envelope == nil {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "contract envelope is required", nil)
	}
	if envelope.SchemaVersion < 1 {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "schema_version must be greater than or equal to 1", nil)
	}
	if strings.TrimSpace(envelope.ResultKind) == "" {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "result_kind is required", nil)
	}
	if envelope.ResultVersion < 1 {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "result_version must be greater than or equal to 1", nil)
	}
	if len(envelope.SourceRefs) == 0 {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "source_refs is required", nil)
	}
	for _, ref := range envelope.SourceRefs {
		if strings.TrimSpace(ref.Kind) == "" || strings.TrimSpace(ref.ID) == "" || ref.Version < 1 {
			return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "source_refs contains invalid item", nil)
		}
	}
	if strings.TrimSpace(envelope.DecisionSummary) == "" {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "decision_summary is required", nil)
	}
	if len(envelope.ResultJSON) == 0 {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "result_json is required", nil)
	}
	if err := validateDeploymentMode(envelope.DeploymentMode); err != nil {
		return err
	}
	return nil
}

func (s *sEasyMVPBrain) ValidatePlanReviewEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.PlanReviewResult) error {
	if err := s.validateTypedEnvelope(ctx, envelope, "plan_review_result"); err != nil {
		return err
	}
	if result == nil {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "plan review result is required", nil)
	}
	if strings.TrimSpace(result.ReviewResultID) == "" || result.ReviewVersion < 1 {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "plan review identifiers are invalid", nil)
	}
	if !isAllowedValue(result.Decision, "approved", "approved_with_advisory", "rejected") {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "plan review decision is invalid", nil)
	}
	return nil
}

func (s *sEasyMVPBrain) ValidatePlanCompileEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.PlanCompileResult) error {
	if err := s.validateTypedEnvelope(ctx, envelope, "compiled_plan"); err != nil {
		return err
	}
	if result == nil {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "plan compile result is required", nil)
	}
	if strings.TrimSpace(result.CompiledPlanID) == "" || result.CompiledVersion < 1 {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "compiled plan identifiers are invalid", nil)
	}
	if len(result.CompiledTasks) == 0 {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "compiled_tasks is required", nil)
	}
	for _, task := range result.CompiledTasks {
		if strings.TrimSpace(task.CompiledTaskID) == "" || strings.TrimSpace(task.Name) == "" {
			return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "compiled_tasks contains invalid identifiers", nil)
		}
		if strings.TrimSpace(task.RoleType) == "" || strings.TrimSpace(task.BrainKind) == "" || strings.TrimSpace(task.RiskLevel) == "" {
			return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "compiled_tasks contains invalid routing fields", nil)
		}
		if len(task.DeliveryContract) == 0 || len(task.VerificationContract) == 0 {
			return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "compiled_tasks is missing contracts", nil)
		}
	}
	return nil
}

func (s *sEasyMVPBrain) ValidateAcceptanceMappingEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.AcceptanceMappingResult) error {
	if err := s.validateTypedEnvelope(ctx, envelope, "acceptance_mapping_result"); err != nil {
		return err
	}
	if result == nil {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "acceptance mapping result is required", nil)
	}
	if strings.TrimSpace(result.AcceptanceProfileID) == "" || strings.TrimSpace(result.ProductionAcceptanceProfileID) == "" {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "acceptance profile identifiers are invalid", nil)
	}
	if len(result.RequiredSurfaces) == 0 || len(result.RequiredJourneys) == 0 || len(result.RequiredEvidence) == 0 {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "acceptance mapping is missing required coverage fields", nil)
	}
	return nil
}

func (s *sEasyMVPBrain) ValidateCompletionAdjudicationEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.CompletionAdjudicationResult) error {
	if err := s.validateTypedEnvelope(ctx, envelope, "completion_decision"); err != nil {
		return err
	}
	if result == nil {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "completion adjudication result is required", nil)
	}
	if !isAllowedValue(result.FinalStatus, "failed", "functional_passed", "production_passed", "released_by_human") {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "completion final_status is invalid", nil)
	}
	if strings.TrimSpace(result.DecisionReason) == "" {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "completion decision_reason is required", nil)
	}
	return nil
}

func (s *sEasyMVPBrain) ValidateRepairDesignEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.RepairDesignResult) error {
	if err := s.validateTypedEnvelope(ctx, envelope, "repair_plan_draft"); err != nil {
		return err
	}
	if result == nil {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "repair design result is required", nil)
	}
	if strings.TrimSpace(result.RepairPlanDraftID) == "" || len(result.RepairPlanJSON) == 0 {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "repair design result is invalid", nil)
	}
	if strings.TrimSpace(result.RepairReasoningSummary) == "" {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "repair reasoning summary is required", nil)
	}
	return nil
}

func (s *sEasyMVPBrain) ValidateWorkspaceExplanationEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, result *braincontracts.WorkspaceExplanationResult) error {
	if err := s.validateTypedEnvelope(ctx, envelope, "workspace_explanation"); err != nil {
		return err
	}
	if result == nil {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "workspace explanation result is required", nil)
	}
	if strings.TrimSpace(result.Headline) == "" || strings.TrimSpace(result.Summary) == "" {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "workspace explanation headline and summary are required", nil)
	}
	for _, item := range result.RecommendedActions {
		if strings.TrimSpace(item.ActionKey) == "" || strings.TrimSpace(item.Label) == "" || strings.TrimSpace(item.Reason) == "" {
			return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "workspace explanation contains invalid recommended action", nil)
		}
	}
	return nil
}

func (s *sEasyMVPBrain) validateTypedEnvelope(ctx context.Context, envelope *braincontracts.BrainContractEnvelope, expectedKind string) error {
	if err := s.ValidateEnvelope(ctx, envelope); err != nil {
		return err
	}
	if envelope.ResultKind != expectedKind {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeUnsupportedKind, "unexpected result_kind: expected "+expectedKind, nil)
	}
	return nil
}

func wrapEasyMVPBrainError(code string, summary string, err error) error {
	summary = strings.TrimSpace(summary)
	if err == nil {
		return gerror.Newf("%s: %s", code, summary)
	}
	return gerror.Wrap(err, code+": "+summary)
}

func validateDeploymentMode(mode string) error {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		return nil
	}
	if isAllowedValue(mode, "local-sidecar", "remote-service") {
		return nil
	}
	return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "deployment_mode is invalid", nil)
}

func isAllowedValue(value string, candidates ...string) bool {
	value = strings.TrimSpace(value)
	for _, item := range candidates {
		if value == item {
			return true
		}
	}
	return false
}

func executeTypedContract[T any](
	ctx context.Context,
	brain IEasyMVPBrain,
	contractKind string,
	input any,
	validate func(context.Context, *braincontracts.BrainContractEnvelope, *T) error,
) (*braincontracts.BrainContractEnvelope, *T, error) {
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "marshal contract input failed", err)
	}
	execResult, err := brain.ExecuteContract(ctx, EasyMVPBrainExecuteCommand{
		ContractKind: contractKind,
		Instruction:  "Execute the requested domain contract and return only the final contract envelope JSON.",
		ContextJSON:  inputJSON,
	})
	if err != nil {
		return nil, nil, err
	}
	if execResult == nil {
		return nil, nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "brain contract envelope is required", nil)
	}
	if execResult.Envelope == nil {
		return nil, nil, deriveEasyMVPBrainExecutionError(execResult)
	}
	var typedResult T
	if err := json.Unmarshal(execResult.Envelope.ResultJSON, &typedResult); err != nil {
		return nil, nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "decode contract result failed", err)
	}
	if err := validate(ctx, execResult.Envelope, &typedResult); err != nil {
		return nil, nil, err
	}
	return execResult.Envelope, &typedResult, nil
}

func deriveEasyMVPBrainExecutionError(result *EasyMVPBrainExecuteResult) error {
	if result == nil {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "brain contract envelope is required", nil)
	}

	status := strings.ToLower(strings.TrimSpace(result.Status))
	detail := strings.TrimSpace(result.Error)
	if detail == "" {
		detail = strings.TrimSpace(result.Summary)
	}

	switch status {
	case "unsupported", "tool_unsupported", "not_supported", "run_unsupported":
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "brain execution reported unsupported capability", gerror.New(detail))
	case "denied", "permission_denied", "tool_denied", "forbidden", "run_denied":
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "brain execution was denied by runtime policy", gerror.New(detail))
	default:
		if detail == "" {
			return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "brain contract envelope is required", nil)
		}
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeContractInvalid, "brain contract envelope is required", gerror.New(detail))
	}
}
