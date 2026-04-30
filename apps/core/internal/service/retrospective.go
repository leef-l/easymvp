package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

// ---------------------------------------------------------------------------
// Result types
// ---------------------------------------------------------------------------

// RetrospectiveResult is the result returned after generating a project retrospective.
type RetrospectiveResult struct {
	RetrospectiveID string
	TotalTasks      int
	CompletedTasks  int
	FailedTasks     int
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

// IRetrospective defines the service interface for project retrospective operations.
type IRetrospective interface {
	GenerateRetrospective(ctx context.Context, projectID string) (*RetrospectiveResult, error)
	GetRetrospective(ctx context.Context, retroID string) (*entity.ProjectRetrospectives, error)
	GetProjectRetrospective(ctx context.Context, projectID string) (*entity.ProjectRetrospectives, error)
}

// ---------------------------------------------------------------------------
// Singleton registration (GoFrame pattern)
// ---------------------------------------------------------------------------

var localRetrospective IRetrospective = (*sRetrospective)(nil)

type sRetrospective struct{}

func Retrospective() IRetrospective {
	if localRetrospective == nil {
		localRetrospective = &sRetrospective{}
	}
	return localRetrospective
}

// ---------------------------------------------------------------------------
// GenerateRetrospective
// ---------------------------------------------------------------------------

func (s *sRetrospective) GenerateRetrospective(ctx context.Context, projectID string) (*RetrospectiveResult, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	// 1. Verify the project exists.
	_, err = getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	// 2. Gather task statistics.
	totalTasks, completedTasks, failedTasks, retriedTasks := gatherTaskStats(ctx, db, projectID)

	// 3. Gather runtime statistics (turns, tokens, cost, duration).
	totalTurns, totalTokens, totalCostUsd, durationSeconds := gatherRuntimeStats(ctx, db, projectID)

	// 4. Count review rounds.
	reviewRounds := countReviewRounds(ctx, db, projectID)

	// 5. Gather brains used.
	brainsUsed := gatherBrainsUsed(ctx, db, projectID)
	brainsUsedJSON, _ := json.Marshal(brainsUsed)

	// 6. Build plan vs actual comparison.
	planVsActual := map[string]any{
		"planned_tasks":   totalTasks,
		"completed_tasks": completedTasks,
		"failed_tasks":    failedTasks,
		"retried_tasks":   retriedTasks,
		"completion_rate": calcCompletionRate(totalTasks, completedTasks),
	}
	planVsActualJSON, _ := json.Marshal(planVsActual)

	// 7. Build success factors and failure lessons (basic auto-analysis).
	successFactors := buildSuccessFactors(completedTasks, totalTasks)
	successFactorsJSON, _ := json.Marshal(successFactors)

	failureLessons := buildFailureLessons(failedTasks, retriedTasks)
	failureLessonsJSON, _ := json.Marshal(failureLessons)

	patterns := buildPatterns(completedTasks, failedTasks, retriedTasks, reviewRounds)
	patternsJSON, _ := json.Marshal(patterns)

	// 8. Insert retrospective record.
	retroID := newResourceID("retro")
	now := nowText()
	_, err = db.ExecContext(ctx,
		`INSERT INTO project_retrospectives (id, project_id, plan_vs_actual_json, success_factors_json, failure_lessons_json, patterns_json, total_tasks, completed_tasks, failed_tasks, retried_tasks, total_turns, total_tokens, total_cost_usd, duration_seconds, review_rounds, brains_used_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		retroID, projectID,
		string(planVsActualJSON), string(successFactorsJSON), string(failureLessonsJSON), string(patternsJSON),
		totalTasks, completedTasks, failedTasks, retriedTasks,
		totalTurns, totalTokens, totalCostUsd, durationSeconds,
		reviewRounds, string(brainsUsedJSON), now,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "insert project retrospective failed")
	}

	// 9. Write audit log.
	if auditErr := insertAuditLog(ctx, projectID, "retrospective.generated", "system", "Project retrospective generated", map[string]any{
		"retrospective_id": retroID,
		"total_tasks":      totalTasks,
		"completed_tasks":  completedTasks,
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}

	return &RetrospectiveResult{
		RetrospectiveID: retroID,
		TotalTasks:      totalTasks,
		CompletedTasks:  completedTasks,
		FailedTasks:     failedTasks,
	}, nil
}

// ---------------------------------------------------------------------------
// GetRetrospective
// ---------------------------------------------------------------------------

func (s *sRetrospective) GetRetrospective(ctx context.Context, retroID string) (*entity.ProjectRetrospectives, error) {
	retroID = strings.TrimSpace(retroID)
	if retroID == "" {
		return nil, gerror.New("retrospective id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	return getRetrospectiveByID(ctx, db, retroID)
}

// ---------------------------------------------------------------------------
// GetProjectRetrospective
// ---------------------------------------------------------------------------

func (s *sRetrospective) GetProjectRetrospective(ctx context.Context, projectID string) (*entity.ProjectRetrospectives, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	return getRetrospectiveByProjectID(ctx, db, projectID)
}

// ---------------------------------------------------------------------------
// DB helpers (package-private)
// ---------------------------------------------------------------------------

func getRetrospectiveByID(ctx context.Context, db *sql.DB, retroID string) (*entity.ProjectRetrospectives, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, project_id, plan_vs_actual_json, success_factors_json, failure_lessons_json, patterns_json, total_tasks, completed_tasks, failed_tasks, retried_tasks, total_turns, total_tokens, total_cost_usd, duration_seconds, review_rounds, brains_used_json, created_at FROM project_retrospectives WHERE id = ? LIMIT 1`,
		retroID,
	)
	var r entity.ProjectRetrospectives
	if err := row.Scan(&r.Id, &r.ProjectId, &r.PlanVsActualJson, &r.SuccessFactorsJson, &r.FailureLessonsJson, &r.PatternsJson, &r.TotalTasks, &r.CompletedTasks, &r.FailedTasks, &r.RetriedTasks, &r.TotalTurns, &r.TotalTokens, &r.TotalCostUsd, &r.DurationSeconds, &r.ReviewRounds, &r.BrainsUsedJson, &r.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.Newf("project retrospective not found: %s", retroID)
		}
		return nil, gerror.Wrap(err, "query project retrospective failed")
	}
	return &r, nil
}

func getRetrospectiveByProjectID(ctx context.Context, db *sql.DB, projectID string) (*entity.ProjectRetrospectives, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, project_id, plan_vs_actual_json, success_factors_json, failure_lessons_json, patterns_json, total_tasks, completed_tasks, failed_tasks, retried_tasks, total_turns, total_tokens, total_cost_usd, duration_seconds, review_rounds, brains_used_json, created_at FROM project_retrospectives WHERE project_id = ? ORDER BY created_at DESC LIMIT 1`,
		projectID,
	)
	var r entity.ProjectRetrospectives
	if err := row.Scan(&r.Id, &r.ProjectId, &r.PlanVsActualJson, &r.SuccessFactorsJson, &r.FailureLessonsJson, &r.PatternsJson, &r.TotalTasks, &r.CompletedTasks, &r.FailedTasks, &r.RetriedTasks, &r.TotalTurns, &r.TotalTokens, &r.TotalCostUsd, &r.DurationSeconds, &r.ReviewRounds, &r.BrainsUsedJson, &r.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query project retrospective by project failed")
	}
	return &r, nil
}

// ---------------------------------------------------------------------------
// Statistics helpers
// ---------------------------------------------------------------------------

func gatherTaskStats(ctx context.Context, db *sql.DB, projectID string) (total, completed, failed, retried int) {
	row := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM domain_tasks WHERE project_id = ?`, projectID)
	if err := row.Scan(&total); err != nil {
		g.Log().Warningf(ctx, "count total tasks failed: %v", err)
	}

	row = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM domain_tasks WHERE project_id = ? AND status = 'done'`, projectID)
	if err := row.Scan(&completed); err != nil {
		g.Log().Warningf(ctx, "count completed tasks failed: %v", err)
	}

	row = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM domain_tasks WHERE project_id = ? AND status = 'failed'`, projectID)
	if err := row.Scan(&failed); err != nil {
		g.Log().Warningf(ctx, "count failed tasks failed: %v", err)
	}

	// Retried tasks: tasks that have been re-executed (status changed back to running after failure).
	// Since domain_tasks does not track retry_count directly, we approximate with 0.
	retried = 0

	return
}

func gatherRuntimeStats(ctx context.Context, db *sql.DB, projectID string) (turns, tokens int, costUsd float64, durationSecs int) {
	// Count brain runs as a proxy for turns.
	row := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM brain_run_bindings WHERE project_id = ?`,
		projectID,
	)
	if err := row.Scan(&turns); err != nil {
		g.Log().Warningf(ctx, "gather runtime stats (brain runs count) failed: %v", err)
	}

	// Duration: difference between project creation and now.
	row = db.QueryRowContext(ctx,
		`SELECT CAST((julianday('now') - julianday(created_at)) * 86400 AS INTEGER) FROM projects WHERE id = ?`,
		projectID,
	)
	if err := row.Scan(&durationSecs); err != nil {
		g.Log().Warningf(ctx, "gather project duration failed: %v", err)
	}

	return
}

func countReviewRounds(ctx context.Context, db *sql.DB, projectID string) int {
	var count int
	row := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM design_reviews WHERE project_id = ?`, projectID)
	if err := row.Scan(&count); err != nil {
		g.Log().Warningf(ctx, "count review rounds failed: %v", err)
	}
	return count
}

func gatherBrainsUsed(ctx context.Context, db *sql.DB, projectID string) []string {
	rows, err := db.QueryContext(ctx,
		`SELECT DISTINCT brain_kind FROM brain_run_bindings WHERE project_id = ? AND brain_kind != ''`,
		projectID,
	)
	if err != nil {
		g.Log().Warningf(ctx, "gather brains used failed: %v", err)
		return []string{}
	}
	defer rows.Close()

	var brains []string
	for rows.Next() {
		var kind string
		if err := rows.Scan(&kind); err == nil && kind != "" {
			brains = append(brains, kind)
		}
	}
	if len(brains) == 0 {
		return []string{}
	}
	return brains
}

func calcCompletionRate(total, completed int) float64 {
	if total == 0 {
		return 0
	}
	return float64(completed) / float64(total) * 100
}

func buildSuccessFactors(completed, total int) []string {
	factors := make([]string, 0)
	if total > 0 && completed == total {
		factors = append(factors, "All tasks completed successfully")
	}
	if total > 0 && float64(completed)/float64(total) >= 0.9 {
		factors = append(factors, "High completion rate (>=90%)")
	}
	return factors
}

func buildFailureLessons(failed, retried int) []string {
	lessons := make([]string, 0)
	if failed > 0 {
		lessons = append(lessons, "Some tasks failed and may need investigation")
	}
	if retried > 0 {
		lessons = append(lessons, "Task retries occurred, indicating intermittent issues")
	}
	return lessons
}

func buildPatterns(completed, failed, retried, reviewRounds int) []string {
	patterns := make([]string, 0)
	if reviewRounds > 3 {
		patterns = append(patterns, "Multiple review rounds suggest design iteration is needed")
	}
	if retried > 0 && failed == 0 {
		patterns = append(patterns, "Retries succeeded, indicating transient failures")
	}
	if failed > 0 && retried > 0 {
		patterns = append(patterns, "Some failures persisted despite retries")
	}
	return patterns
}
