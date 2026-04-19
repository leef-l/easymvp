package service

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	planv1 "github.com/leef-l/easymvp/apps/core/api/plan/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
)

type CompilePlanCommand struct {
	ProjectID      string
	PlanDraftID    string
	ForceRecompile bool
}

type CreateRepairDraftCommand struct {
	ProjectID             string
	FailedTaskContextJSON string
	FailureReasonJSON     string
	OriginalContractsJSON string
	RuntimeSummaryJSON    string
	ArtifactRefs          []braincontracts.ArtifactRef
	CreatedBy             string
}

type CreateRepairDraftResult struct {
	CommandID  string
	Accepted   bool
	ResourceID string
	NextAction string
}

type IPlan interface {
	CreateInitialDraft(ctx context.Context, projectID string) error
	CreateRepairDraft(ctx context.Context, req CreateRepairDraftCommand) (res *CreateRepairDraftResult, err error)
	CompilePlan(ctx context.Context, req CompilePlanCommand) (res *planv1.CompilePlanRes, err error)
	GetPlanView(ctx context.Context, projectID string) (res *planv1.PlanViewRes, err error)
	GetRepairDraftView(ctx context.Context, projectID string) (res *planv1.RepairDraftRes, err error)
}

var localPlan IPlan = (*sPlan)(nil)

type sPlan struct{}

func Plan() IPlan {
	if localPlan == nil {
		localPlan = &sPlan{}
	}
	return localPlan
}

func (s *sPlan) CreateInitialDraft(ctx context.Context, projectID string) error {
	return createInitialDraftIfNeeded(ctx, projectID)
}

func (s *sPlan) CreateRepairDraft(ctx context.Context, req CreateRepairDraftCommand) (res *CreateRepairDraftResult, err error) {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.CreatedBy = strings.TrimSpace(req.CreatedBy)
	if req.ProjectID == "" {
		return nil, gerror.New("project id is required")
	}

	repairDraftID, err := createRepairDraftForProject(ctx, req)
	if err != nil {
		return nil, err
	}
	return &CreateRepairDraftResult{
		CommandID:  newResourceID("cmd"),
		Accepted:   true,
		ResourceID: repairDraftID,
		NextAction: "refresh_plan_view",
	}, nil
}

func (s *sPlan) CompilePlan(ctx context.Context, req CompilePlanCommand) (res *planv1.CompilePlanRes, err error) {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.PlanDraftID = strings.TrimSpace(req.PlanDraftID)
	if req.ProjectID == "" {
		return nil, gerror.New("project id is required")
	}
	compiledPlanID, err := compilePlanForProject(ctx, req)
	if err != nil {
		return nil, err
	}
	return &planv1.CompilePlanRes{
		CommandID:  newResourceID("cmd"),
		Accepted:   true,
		ResourceID: compiledPlanID,
		NextAction: "refresh_plan_view",
	}, nil
}

func (s *sPlan) GetPlanView(ctx context.Context, projectID string) (res *planv1.PlanViewRes, err error) {
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	data, err := loadPlanViewData(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &planv1.PlanViewRes{
		Overview:       data.Overview,
		Draft:          data.Draft,
		Review:         data.Review,
		Compiled:       data.Compiled,
		RepairDraft:    data.RepairDraft,
		TaskProjection: data.TaskProjection,
		DiffSummary:    data.DiffSummary,
	}, nil
}

func (s *sPlan) GetRepairDraftView(ctx context.Context, projectID string) (res *planv1.RepairDraftRes, err error) {
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	view, err := loadRepairDraftView(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &planv1.RepairDraftRes{
		RepairDraft: *view,
	}, nil
}
