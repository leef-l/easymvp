package service

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

// listDesignReviews returns all reviews for a design, ordered by round ascending.
func listDesignReviews(ctx context.Context, designID string) ([]entity.DesignReviews, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	rows, err := db.QueryContext(ctx,
		"SELECT id, design_id, project_id, round, passed, score, dimensions_json, issues_json, suggestions_json, fix_tasks_json, brain_run_id, created_at FROM design_reviews WHERE design_id = ? ORDER BY round ASC",
		designID,
	)
	if err != nil {
		if isSchemaMissingError(err) {
			return nil, nil
		}
		return nil, gerror.Wrapf(err, "query design reviews for %s failed", designID)
	}
	defer rows.Close()

	var reviews []entity.DesignReviews
	for rows.Next() {
		var r entity.DesignReviews
		if err = rows.Scan(
			&r.Id,
			&r.DesignId,
			&r.ProjectId,
			&r.Round,
			&r.Passed,
			&r.Score,
			&r.DimensionsJson,
			&r.IssuesJson,
			&r.SuggestionsJson,
			&r.FixTasksJson,
			&r.BrainRunId,
			&r.CreatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan design review row failed")
		}
		reviews = append(reviews, r)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate design review rows failed")
	}
	return reviews, nil
}

// insertDesignReview inserts a new design review record.
func insertDesignReview(ctx context.Context, reviewID, designID, projectID string, round, passed, score int, dimensionsJSON, issuesJSON, suggestionsJSON, brainRunID string) error {
	now := nowText()

	_, err := g.DB().Exec(ctx,
		"INSERT INTO design_reviews (id, design_id, project_id, round, passed, score, dimensions_json, issues_json, suggestions_json, fix_tasks_json, brain_run_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		reviewID, designID, projectID, round, passed, score, dimensionsJSON, issuesJSON, suggestionsJSON, "", brainRunID, now,
	)
	if err != nil {
		return gerror.Wrap(err, "insert design review failed")
	}
	return nil
}

// updateDesignReviewFixTasks updates the fix_tasks_json field of a design review.
func updateDesignReviewFixTasks(ctx context.Context, reviewID, fixTasksJSON string) error {
	_, err := g.DB().Exec(ctx,
		"UPDATE design_reviews SET fix_tasks_json = ? WHERE id = ?",
		fixTasksJSON, reviewID,
	)
	if err != nil {
		return gerror.Wrapf(err, "update design review %s fix tasks failed", reviewID)
	}
	return nil
}

// updateDesignWithFix updates the solution design with fixed content and bumps the version.
func updateDesignWithFix(ctx context.Context, designID string, newVersion int, fix *braincontracts.DesignFixResult) error {
	if fix == nil {
		return gerror.New("fix result is required")
	}

	now := nowText()

	architecture := fix.Architecture
	modulesJSON := rawJSONToString(fix.ModulesJSON)
	dataModelsJSON := rawJSONToString(fix.DataModelsJSON)
	pagesJSON := rawJSONToString(fix.PagesJSON)
	taskDraftsJSON := rawJSONToString(fix.TaskDraftsJSON)

	_, err := g.DB().Exec(ctx,
		"UPDATE solution_designs SET version = ?, architecture = ?, modules_json = ?, data_models_json = ?, pages_json = ?, task_drafts_json = ?, status = ?, updated_at = ? WHERE id = ?",
		newVersion, architecture, modulesJSON, dataModelsJSON, pagesJSON, taskDraftsJSON, "reviewing", now, designID,
	)
	if err != nil {
		return gerror.Wrapf(err, "update solution design %s with fix failed", designID)
	}
	return nil
}

// rawJSONToString converts json.RawMessage to string, returning empty string for nil/null.
func rawJSONToString(raw json.RawMessage) string {
	if raw == nil {
		return ""
	}
	s := string(raw)
	if s == "null" || s == "" {
		return ""
	}
	return s
}
