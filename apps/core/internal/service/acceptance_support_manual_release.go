package service

import (
	"context"
	"database/sql"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func applyManualRelease(ctx context.Context, req ApplyManualReleaseCommand) (string, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return "", err
	}
	defer closeFn()

	aggregate, err := loadAcceptanceAggregate(ctx, db, req.ProjectID)
	if err != nil {
		return "", err
	}
	if aggregate.LatestAcceptanceRun == nil {
		return "", gerror.New("latest acceptance run is required")
	}
	if aggregate.LatestAcceptanceRun.ManualReleaseRequired != 1 {
		return "", gerror.New("manual release is not required for the latest acceptance run")
	}
	if hasHumanReleaseApproval(aggregate) {
		return aggregate.LatestAcceptanceRun.Id, nil
	}

	now := nowText()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", gerror.Wrap(err, "begin manual release transaction failed")
	}

	if aggregate.LatestAcceptanceRun.TaskId != "" {
		if err = upsertTaskManualReleaseGate(ctx, tx, aggregate.Project.Id, aggregate.LatestAcceptanceRun.TaskId, req.Comment, now); err != nil {
			_ = tx.Rollback()
			return "", err
		}
	}
	if err = insertAcceptanceJudgementRecord(ctx, tx, entity.AcceptanceJudgements{
		Id:              newResourceID("judge"),
		ProjectId:       aggregate.Project.Id,
		AcceptanceRunId: aggregate.LatestAcceptanceRun.Id,
		JudgementKind:   "release_gate",
		JudgementResult: "approved",
		Summary:         firstNonEmpty(req.Comment, "Manual release approved"),
		DetailJson: mustMarshalJSONString(map[string]any{
			"manual_release_required":  true,
			"manual_release_completed": true,
			"approval_source":          "manual_release_command",
		}, "{}"),
		CreatedAt: now,
	}); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	if err = updateAcceptanceRunAfterManualRelease(ctx, tx, aggregate.LatestAcceptanceRun.Id, now); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	if err = updateProjectAfterManualRelease(ctx, tx, aggregate.Project.Id, now); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	if err = insertManualReleaseAuditLog(ctx, tx, aggregate.Project.Id, aggregate.LatestAcceptanceRun.Id, aggregate.LatestAcceptanceRun.TaskId, req.Comment, now); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", gerror.Wrap(err, "commit manual release transaction failed")
	}
	return aggregate.LatestAcceptanceRun.Id, nil
}

func upsertTaskManualReleaseGate(ctx context.Context, tx *sql.Tx, projectID string, taskID string, comment string, now string) error {
	if _, err := tx.ExecContext(
		ctx,
		`DELETE FROM `+dao.TaskManualGates.Table()+` WHERE project_id = ? AND task_id = ? AND gate_kind = ?`,
		projectID,
		taskID,
		"manual_release",
	); err != nil {
		return gerror.Wrap(err, "delete existing manual release gate failed")
	}

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.TaskManualGates.Table()+` (
id, project_id, task_id, gate_kind, gate_status, comment, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		newResourceID("gate"),
		projectID,
		taskID,
		"manual_release",
		"approved",
		nullIfEmpty(comment),
		now,
		now,
	)
	if err != nil {
		return gerror.Wrap(err, "insert manual release gate failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert manual release gate affected unexpected rows")
	}
	return nil
}

func updateAcceptanceRunAfterManualRelease(ctx context.Context, tx *sql.Tx, acceptanceRunID string, now string) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE `+dao.AcceptanceRuns.Table()+` SET status = ?, production_status = ?, finished_at = ? WHERE id = ?`,
		"completed",
		"production_passed",
		now,
		acceptanceRunID,
	); err != nil {
		return gerror.Wrap(err, "update acceptance run after manual release failed")
	}
	return nil
}

func updateProjectAfterManualRelease(ctx context.Context, tx *sql.Tx, projectID string, now string) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE `+dao.Projects.Table()+` SET status = ?, production_status = ?, updated_at = ? WHERE id = ?`,
		"completed",
		"production_passed",
		now,
		projectID,
	); err != nil {
		return gerror.Wrap(err, "update project after manual release failed")
	}
	return nil
}

func insertManualReleaseAuditLog(ctx context.Context, tx *sql.Tx, projectID string, acceptanceRunID string, taskID string, comment string, now string) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.AuditLogs.Table()+` (
id, project_id, event_type, actor_kind, summary, payload_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		newResourceID("audit"),
		projectID,
		"manual_release.approved",
		"user:local_operator",
		firstNonEmpty(comment, "Manual release approved"),
		mustMarshalJSONString(map[string]any{
			"acceptance_run_id": acceptanceRunID,
			"task_id":           taskID,
			"comment":           comment,
		}, "{}"),
		now,
	)
	if err != nil {
		return gerror.Wrap(err, "insert manual release audit log failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert manual release audit log affected unexpected rows")
	}
	return nil
}
