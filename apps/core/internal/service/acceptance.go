package service

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	acceptancev1 "github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
)

type StartAcceptanceCommand struct {
	ProjectID      string
	TaskID         string
	ProfileVersion string
	Mode           string
}

type ApplyManualReleaseCommand struct {
	ProjectID string
	Comment   string
}

type IAcceptance interface {
	MapAcceptanceProfiles(ctx context.Context, projectID string) (res *braincontracts.AcceptanceMappingResult, err error)
	RefreshAcceptanceProfiles(ctx context.Context, projectID string) (res *acceptancev1.RefreshAcceptanceProfilesRes, err error)
	StartAcceptanceRun(ctx context.Context, req StartAcceptanceCommand) (res *acceptancev1.StartAcceptanceRes, err error)
	AdjudicateAcceptanceRun(ctx context.Context, projectID string) (res *acceptancev1.AdjudicateAcceptanceRes, err error)
	ApplyManualRelease(ctx context.Context, req ApplyManualReleaseCommand) (res *acceptancev1.ApplyManualReleaseRes, err error)
	AdjudicateLatestAcceptanceRun(ctx context.Context, projectID string) (res *braincontracts.CompletionAdjudicationResult, err error)
	AdjudicateAcceptanceRunByID(ctx context.Context, projectID string, acceptanceRunID string) (res *braincontracts.CompletionAdjudicationResult, err error)
	GetAcceptanceView(ctx context.Context, projectID string) (res *acceptancev1.AcceptanceViewRes, err error)
}

var localAcceptance IAcceptance = (*sAcceptance)(nil)

type sAcceptance struct{}

func Acceptance() IAcceptance {
	if localAcceptance == nil {
		localAcceptance = &sAcceptance{}
	}
	return localAcceptance
}

func (s *sAcceptance) MapAcceptanceProfiles(ctx context.Context, projectID string) (res *braincontracts.AcceptanceMappingResult, err error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	return mapAcceptanceProfiles(ctx, projectID)
}

func (s *sAcceptance) RefreshAcceptanceProfiles(ctx context.Context, projectID string) (res *acceptancev1.RefreshAcceptanceProfilesRes, err error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	result, err := mapAcceptanceProfiles(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if auditErr := insertAuditLog(ctx, projectID, "acceptance.profiles.refreshed", "user:local_operator", "Acceptance profiles refreshed", nil); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	return &acceptancev1.RefreshAcceptanceProfilesRes{
		CommandID:  newResourceID("cmd"),
		Accepted:   true,
		ResourceID: strings.TrimSpace(result.AcceptanceProfileID),
		NextAction: "refresh_acceptance_view",
	}, nil
}

func (s *sAcceptance) StartAcceptanceRun(ctx context.Context, req StartAcceptanceCommand) (res *acceptancev1.StartAcceptanceRes, err error) {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.ProfileVersion = strings.TrimSpace(req.ProfileVersion)
	req.Mode = strings.TrimSpace(req.Mode)
	if req.ProjectID == "" {
		return nil, gerror.New("project id is required")
	}
	runID, err := startAcceptanceRun(ctx, req)
	if err != nil {
		return nil, err
	}
	return &acceptancev1.StartAcceptanceRes{
		CommandID:  newResourceID("cmd"),
		Accepted:   true,
		ResourceID: runID,
		NextAction: "refresh_acceptance_view",
	}, nil
}

func (s *sAcceptance) AdjudicateLatestAcceptanceRun(ctx context.Context, projectID string) (res *braincontracts.CompletionAdjudicationResult, err error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	return adjudicateLatestAcceptanceRun(ctx, projectID)
}

func (s *sAcceptance) AdjudicateAcceptanceRunByID(ctx context.Context, projectID string, acceptanceRunID string) (res *braincontracts.CompletionAdjudicationResult, err error) {
	projectID = strings.TrimSpace(projectID)
	acceptanceRunID = strings.TrimSpace(acceptanceRunID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	if acceptanceRunID == "" {
		return nil, gerror.New("acceptance run id is required")
	}
	return adjudicateAcceptanceRunByID(ctx, projectID, acceptanceRunID)
}

func (s *sAcceptance) AdjudicateAcceptanceRun(ctx context.Context, projectID string) (res *acceptancev1.AdjudicateAcceptanceRes, err error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	result, err := adjudicateLatestAcceptanceRun(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &acceptancev1.AdjudicateAcceptanceRes{
		CommandID:   newResourceID("cmd"),
		Accepted:    true,
		ResourceID:  projectID,
		NextAction:  "refresh_acceptance_view",
		FinalStatus: strings.TrimSpace(result.FinalStatus),
	}, nil
}

func (s *sAcceptance) ApplyManualRelease(ctx context.Context, req ApplyManualReleaseCommand) (res *acceptancev1.ApplyManualReleaseRes, err error) {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.Comment = strings.TrimSpace(req.Comment)
	if req.ProjectID == "" {
		return nil, gerror.New("project id is required")
	}

	runID, err := applyManualRelease(ctx, req)
	if err != nil {
		return nil, err
	}

	return &acceptancev1.ApplyManualReleaseRes{
		CommandID:   newResourceID("cmd"),
		Accepted:    true,
		ResourceID:  runID,
		NextAction:  "refresh_acceptance_view",
		FinalStatus: "released_by_human",
	}, nil
}

func (s *sAcceptance) GetAcceptanceView(ctx context.Context, projectID string) (res *acceptancev1.AcceptanceViewRes, err error) {
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	data, err := loadAcceptanceViewData(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &acceptancev1.AcceptanceViewRes{
		Overview:           data.Overview,
		AcceptanceRun:      data.AcceptanceRun,
		CoverageMatrix:     data.CoverageMatrix,
		Issues:             data.Issues,
		EvidenceCards:      data.EvidenceCards,
		ReleaseGate:        data.ReleaseGate,
		VerificationResult: data.VerificationResult,
		CompletionVerdict:  data.CompletionVerdict,
		RuntimeEscalation:  data.RuntimeEscalation,
		FaultSummary:       data.FaultSummary,
		RepairPlanDraft:    data.RepairPlanDraft,
		ContractGap:        data.ContractGap,
	}, nil
}
