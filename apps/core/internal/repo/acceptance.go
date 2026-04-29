package repo

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	acceptancev1 "github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func LoadPersistedCompletionVerdict(
	ctx context.Context,
	projectID string,
	acceptanceRunID string,
) (*acceptancev1.CompletionVerdictView, error) {
	m := dao.CompletionVerdicts.Ctx(ctx).Where(g.Map{"project_id": projectID})
	if strings.TrimSpace(acceptanceRunID) != "" {
		m = m.Where(g.Map{"acceptance_run_id": acceptanceRunID})
	}
	var view acceptancev1.CompletionVerdictView
	if err := m.Order("updated_at DESC").Limit(1).Scan(&view); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query persisted completion verdict failed")
	}
	return &view, nil
}

func GetAcceptanceRunByID(ctx context.Context, acceptanceRunID string) (*entity.AcceptanceRuns, error) {
	var run entity.AcceptanceRuns
	if err := dao.AcceptanceRuns.Ctx(ctx).Where(g.Map{"id": acceptanceRunID}).Scan(&run); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query acceptance run by id failed")
	}
	return &run, nil
}

func ListAcceptanceIssuesByRun(ctx context.Context, acceptanceRunID string, limit int) ([]entity.AcceptanceIssues, error) {
	var items []entity.AcceptanceIssues
	if err := dao.AcceptanceIssues.Ctx(ctx).
		Where(g.Map{"acceptance_run_id": acceptanceRunID}).
		Order("created_at DESC").
		Limit(limit).
		Scan(&items); err != nil {
		return nil, gerror.Wrap(err, "query acceptance issues by run failed")
	}
	return items, nil
}

func ListAcceptanceSurfaceCoverageByRun(ctx context.Context, acceptanceRunID string) ([]entity.AcceptanceSurfaceCoverage, error) {
	var items []entity.AcceptanceSurfaceCoverage
	if err := dao.AcceptanceSurfaceCoverage.Ctx(ctx).
		Where(g.Map{"acceptance_run_id": acceptanceRunID}).
		Order("surface ASC").
		Scan(&items); err != nil {
		return nil, gerror.Wrap(err, "query acceptance surface coverage failed")
	}
	return items, nil
}

func ListAcceptanceJourneyCoverageByRun(ctx context.Context, acceptanceRunID string) ([]entity.AcceptanceJourneyCoverage, error) {
	var items []entity.AcceptanceJourneyCoverage
	if err := dao.AcceptanceJourneyCoverage.Ctx(ctx).
		Where(g.Map{"acceptance_run_id": acceptanceRunID}).
		Order("journey ASC").
		Scan(&items); err != nil {
		return nil, gerror.Wrap(err, "query acceptance journey coverage failed")
	}
	return items, nil
}

func ListAcceptanceJudgementsByRun(ctx context.Context, acceptanceRunID string) ([]entity.AcceptanceJudgements, error) {
	var items []entity.AcceptanceJudgements
	if err := dao.AcceptanceJudgements.Ctx(ctx).
		Where(g.Map{"acceptance_run_id": acceptanceRunID}).
		Order("created_at DESC").
		Scan(&items); err != nil {
		return nil, gerror.Wrap(err, "query acceptance judgements failed")
	}
	return items, nil
}

func ListProjectEvidenceItems(ctx context.Context, projectID string, limit int) ([]entity.EvidenceItems, error) {
	var items []entity.EvidenceItems
	if err := dao.EvidenceItems.Ctx(ctx).
		Where(g.Map{"project_id": projectID}).
		Order(gdb.Raw("COALESCE(captured_at, created_at) DESC"), "created_at DESC").
		Limit(limit).
		Scan(&items); err != nil {
		return nil, gerror.Wrap(err, "query project evidence items failed")
	}
	return items, nil
}
