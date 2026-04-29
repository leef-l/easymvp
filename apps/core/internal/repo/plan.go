package repo

import (
	"context"
	"database/sql"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

const planTaskProjectionLimit = 64

type RepairPlanDraftRecord struct {
	ID                      string
	ProjectID               string
	FailedTaskContextJSON   string
	FailureReasonJSON       string
	OriginalContractsJSON   string
	RuntimeSummaryJSON      string
	RepairPlanJSON          string
	RepairReasoningSummary  string
	ReplacedConstraintsJSON string
	Status                  string
	CreatedBy               string
	CreatedAt               string
	UpdatedAt               string
}

func GetPlanDraftForProject(ctx context.Context, project entity.Projects) (*entity.WorkflowPlanDrafts, error) {
	var item entity.WorkflowPlanDrafts
	var err error

	if project.CurrentPlanDraftId != "" {
		err = dao.WorkflowPlanDrafts.Ctx(ctx).Where("id", project.CurrentPlanDraftId).Scan(&item)
	} else {
		err = dao.WorkflowPlanDrafts.Ctx(ctx).Where("project_id", project.Id).
			Order("version DESC").
			Limit(1).
			Scan(&item)
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func GetPlanReviewForProject(ctx context.Context, project entity.Projects, draft *entity.WorkflowPlanDrafts) (*entity.WorkflowPlanReviewResults, error) {
	if draft == nil && project.CurrentPlanDraftId == "" {
		return nil, nil
	}

	var item entity.WorkflowPlanReviewResults
	var err error

	if draft != nil {
		err = dao.WorkflowPlanReviewResults.Ctx(ctx).Where("plan_draft_id", draft.Id).
			Order("review_version DESC").
			Limit(1).
			Scan(&item)
	} else {
		err = dao.WorkflowPlanReviewResults.Ctx(ctx).Where("project_id", project.Id).
			Order("reviewed_at DESC, review_version DESC").
			Limit(1).
			Scan(&item)
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func GetCompiledPlanForProject(
	ctx context.Context,
	project entity.Projects,
	draft *entity.WorkflowPlanDrafts,
	review *entity.WorkflowPlanReviewResults,
) (*entity.WorkflowCompiledPlans, error) {
	var item entity.WorkflowCompiledPlans
	var err error

	switch {
	case project.CurrentCompiledPlanId != "":
		err = dao.WorkflowCompiledPlans.Ctx(ctx).Where("id", project.CurrentCompiledPlanId).Scan(&item)
	case review != nil:
		err = dao.WorkflowCompiledPlans.Ctx(ctx).Where("plan_review_result_id", review.Id).
			Order("compiled_version DESC").
			Limit(1).
			Scan(&item)
	case draft != nil:
		err = dao.WorkflowCompiledPlans.Ctx(ctx).Where("plan_draft_id", draft.Id).
			Order("compiled_version DESC").
			Limit(1).
			Scan(&item)
	default:
		err = dao.WorkflowCompiledPlans.Ctx(ctx).Where("project_id", project.Id).
			Order("compiled_version DESC").
			Limit(1).
			Scan(&item)
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func ListCompiledTasksForProject(ctx context.Context, project entity.Projects, compiled *entity.WorkflowCompiledPlans) ([]entity.WorkflowCompiledTasks, error) {
	if compiled == nil {
		return nil, nil
	}

	var items []entity.WorkflowCompiledTasks
	err := dao.WorkflowCompiledTasks.Ctx(ctx).Where("compiled_plan_id", compiled.Id).
		Order("created_at ASC, task_key ASC").
		Limit(planTaskProjectionLimit).
		Scan(&items)
	if err != nil {
		return nil, gerror.Wrap(err, "query compiled tasks failed")
	}
	return items, nil
}

func ListDomainTasksForPlan(ctx context.Context, project entity.Projects, compiled *entity.WorkflowCompiledPlans) ([]entity.DomainTasks, error) {
	m := dao.DomainTasks.Ctx(ctx).Where("project_id", project.Id)
	if compiled != nil {
		m = m.Where("source_compiled_plan_id", compiled.Id)
	}

	var items []entity.DomainTasks
	err := m.Order("updated_at DESC, created_at DESC").
		Limit(planTaskProjectionLimit).
		Scan(&items)
	if err != nil {
		return nil, gerror.Wrap(err, "query domain tasks for plan failed")
	}
	return items, nil
}

func GetLatestRepairDraftForProject(ctx context.Context, projectID string) (*RepairPlanDraftRecord, error) {
	var item entity.RepairPlanDrafts
	err := dao.RepairPlanDrafts.Ctx(ctx).Where("project_id", projectID).
		Order("updated_at DESC, created_at DESC").
		Limit(1).
		Scan(&item)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, gerror.Wrap(err, "query latest repair draft failed")
	}

	return &RepairPlanDraftRecord{
		ID:                      item.Id,
		ProjectID:               item.ProjectId,
		FailedTaskContextJSON:   item.FailedTaskContextJson,
		FailureReasonJSON:       item.FailureReasonJson,
		OriginalContractsJSON:   item.OriginalContractsJson,
		RuntimeSummaryJSON:      item.RuntimeSummaryJson,
		RepairPlanJSON:          item.RepairPlanJson,
		RepairReasoningSummary:  item.RepairReasoningSummary,
		ReplacedConstraintsJSON: item.ReplacedConstraintsJson,
		Status:                  item.Status,
		CreatedBy:               item.CreatedBy,
		CreatedAt:               item.CreatedAt,
		UpdatedAt:               item.UpdatedAt,
	}, nil
}
