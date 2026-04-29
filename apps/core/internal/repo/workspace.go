package repo

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func GetProjectByID(ctx context.Context, projectID string) (*entity.Projects, error) {
	var project entity.Projects
	err := dao.Projects.Ctx(ctx).Where("id", projectID).Scan(&project)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.Newf("project not found: %s", projectID)
		}
		return nil, gerror.Wrap(err, "query project by id failed")
	}
	return &project, nil
}

func ListProjectDomainTasks(ctx context.Context, projectID string) ([]entity.DomainTasks, error) {
	project, err := GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	m := dao.DomainTasks.Ctx(ctx).Where("project_id", projectID)
	if strings.TrimSpace(project.CurrentCompiledPlanId) != "" {
		m = m.Where("source_compiled_plan_id", project.CurrentCompiledPlanId)
	}

	var items []entity.DomainTasks
	if err = m.Order("updated_at DESC, created_at DESC").Scan(&items); err != nil {
		return nil, gerror.Wrap(err, "query project domain tasks failed")
	}
	return items, nil
}

func ListProjectBrainRunBindings(ctx context.Context, projectID string, limit int) ([]entity.BrainRunBindings, error) {
	var items []entity.BrainRunBindings
	err := dao.BrainRunBindings.Ctx(ctx).Where("project_id", projectID).
		Order("COALESCE(last_sync_at, updated_at, started_at, created_at) DESC").
		Limit(limit).
		Scan(&items)
	if err != nil {
		return nil, gerror.Wrap(err, "query project brain run bindings failed")
	}
	return items, nil
}

func ListProjectAcceptanceRuns(ctx context.Context, projectID string, limit int) ([]entity.AcceptanceRuns, error) {
	var items []entity.AcceptanceRuns
	err := dao.AcceptanceRuns.Ctx(ctx).Where("project_id", projectID).
		Order("created_at DESC").
		Limit(limit).
		Scan(&items)
	if err != nil {
		return nil, gerror.Wrap(err, "query project acceptance runs failed")
	}
	return items, nil
}

func ListProjectAcceptanceIssues(ctx context.Context, projectID string, limit int) ([]entity.AcceptanceIssues, error) {
	var items []entity.AcceptanceIssues
	err := dao.AcceptanceIssues.Ctx(ctx).Where("project_id", projectID).
		Order("created_at DESC").
		Limit(limit).
		Scan(&items)
	if err != nil {
		return nil, gerror.Wrap(err, "query project acceptance issues failed")
	}
	return items, nil
}

func ListProjectAuditLogs(ctx context.Context, projectID string, limit int) ([]entity.AuditLogs, error) {
	var items []entity.AuditLogs
	err := dao.AuditLogs.Ctx(ctx).Where("project_id", projectID).
		Order("created_at DESC").
		Limit(limit).
		Scan(&items)
	if err != nil {
		return nil, gerror.Wrap(err, "query project audit logs failed")
	}
	return items, nil
}

func GetLatestAcceptanceRunByProjectID(ctx context.Context, projectID string) (*entity.AcceptanceRuns, error) {
	var run entity.AcceptanceRuns
	err := dao.AcceptanceRuns.Ctx(ctx).Where("project_id", projectID).
		Order("created_at DESC").
		Limit(1).
		Scan(&run)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query latest acceptance run failed")
	}
	return &run, nil
}
