package workspace

import (
	"context"
	"fmt"
	"strings"

	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// repo 封装 mvp_task_workspace 表的数据库操作。
type repo struct{}

var defaultRepo = &repo{}

var workspaceDeliveryColumns = map[string]struct{}{
	"delivery_mode":   {},
	"delivery_status": {},
	"sync_strategy":   {},
	"sync_status":     {},
	"risk_level":      {},
	"patch_ref":       {},
	"delivery_ref":    {},
	"delivery_title":  {},
}

var workspaceDeliveryRefColumns = map[string]struct{}{
	"delivery_ref":   {},
	"delivery_title": {},
}

// create 插入工作空间记录。
func (r *repo) create(ctx context.Context, ws *TaskWorkspace) error {
	now := gtime.Now()
	ws.ID = int64(snowflake.Generate())

	sql := `
INSERT INTO mvp_task_workspace
	(id, task_id, workflow_run_id, project_id, workspace_type, workspace_path, base_ref, status, cleanup_status, delivery_mode, delivery_status, sync_strategy, sync_status, risk_level, delivery_ref, delivery_title, created_at, updated_at)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	workflow_run_id = VALUES(workflow_run_id),
	project_id = VALUES(project_id),
	workspace_type = VALUES(workspace_type),
	workspace_path = VALUES(workspace_path),
	base_ref = VALUES(base_ref),
	status = VALUES(status),
	cleanup_status = VALUES(cleanup_status),
	delivery_mode = VALUES(delivery_mode),
	delivery_status = VALUES(delivery_status),
	sync_strategy = VALUES(sync_strategy),
	sync_status = VALUES(sync_status),
	risk_level = VALUES(risk_level),
	patch_ref = NULL,
	delivery_ref = VALUES(delivery_ref),
	delivery_title = VALUES(delivery_title),
	diff_summary = NULL,
	error_message = NULL,
	deleted_at = NULL,
	created_at = VALUES(created_at),
	updated_at = VALUES(updated_at)
`
	if _, err := g.DB().Exec(ctx, sql,
		ws.ID,
		ws.TaskID,
		ws.WorkflowRunID,
		ws.ProjectID,
		ws.WorkspaceType,
		ws.WorkspacePath,
		ws.BaseRef,
		ws.Status,
		ws.CleanupStatus,
		ws.DeliveryMode,
		ws.DeliveryStatus,
		ws.SyncStrategy,
		ws.SyncStatus,
		ws.RiskLevel,
		ws.DeliveryRef,
		ws.DeliveryTitle,
		now,
		now,
	); err != nil {
		if !isUnknownColumnErr(err) {
			return err
		}
		if isDeliveryReferenceColumnErr(err) {
			midSQL := `
INSERT INTO mvp_task_workspace
	(id, task_id, workflow_run_id, project_id, workspace_type, workspace_path, base_ref, status, cleanup_status, delivery_mode, delivery_status, sync_strategy, sync_status, risk_level, created_at, updated_at)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	workflow_run_id = VALUES(workflow_run_id),
	project_id = VALUES(project_id),
	workspace_type = VALUES(workspace_type),
	workspace_path = VALUES(workspace_path),
	base_ref = VALUES(base_ref),
	status = VALUES(status),
	cleanup_status = VALUES(cleanup_status),
	delivery_mode = VALUES(delivery_mode),
	delivery_status = VALUES(delivery_status),
	sync_strategy = VALUES(sync_strategy),
	sync_status = VALUES(sync_status),
	risk_level = VALUES(risk_level),
	patch_ref = NULL,
	diff_summary = NULL,
	error_message = NULL,
	deleted_at = NULL,
	created_at = VALUES(created_at),
	updated_at = VALUES(updated_at)
`
			if _, midErr := g.DB().Exec(ctx, midSQL,
				ws.ID,
				ws.TaskID,
				ws.WorkflowRunID,
				ws.ProjectID,
				ws.WorkspaceType,
				ws.WorkspacePath,
				ws.BaseRef,
				ws.Status,
				ws.CleanupStatus,
				ws.DeliveryMode,
				ws.DeliveryStatus,
				ws.SyncStrategy,
				ws.SyncStatus,
				ws.RiskLevel,
				now,
				now,
			); midErr == nil {
				goto queryID
			}
		}
		legacySQL := `
INSERT INTO mvp_task_workspace
	(id, task_id, workflow_run_id, project_id, workspace_type, workspace_path, base_ref, status, cleanup_status, created_at, updated_at)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	workflow_run_id = VALUES(workflow_run_id),
	project_id = VALUES(project_id),
	workspace_type = VALUES(workspace_type),
	workspace_path = VALUES(workspace_path),
	base_ref = VALUES(base_ref),
	status = VALUES(status),
	cleanup_status = VALUES(cleanup_status),
	diff_summary = NULL,
	error_message = NULL,
	deleted_at = NULL,
	created_at = VALUES(created_at),
	updated_at = VALUES(updated_at)
`
		if _, legacyErr := g.DB().Exec(ctx, legacySQL,
			ws.ID,
			ws.TaskID,
			ws.WorkflowRunID,
			ws.ProjectID,
			ws.WorkspaceType,
			ws.WorkspacePath,
			ws.BaseRef,
			ws.Status,
			ws.CleanupStatus,
			now,
			now,
		); legacyErr != nil {
			return legacyErr
		}
	}

queryID:
	record, err := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		Where("task_id", ws.TaskID).
		Fields("id").
		One()
	if err != nil {
		return err
	}
	if record.IsEmpty() {
		return fmt.Errorf("workspace upsert 后查询 task_id=%d 失败", ws.TaskID)
	}
	ws.ID = record["id"].Int64()
	return nil
}

// getByTaskID 按任务 ID 查询工作空间。
func (r *repo) getByTaskID(ctx context.Context, taskID int64) (*TaskWorkspace, error) {
	record, err := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		Where("task_id", taskID).
		WhereNull("deleted_at").
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, nil
	}
	return &TaskWorkspace{
		ID:             record["id"].Int64(),
		TaskID:         record["task_id"].Int64(),
		WorkflowRunID:  record["workflow_run_id"].Int64(),
		ProjectID:      record["project_id"].Int64(),
		WorkspaceType:  record["workspace_type"].String(),
		WorkspacePath:  record["workspace_path"].String(),
		BaseRef:        record["base_ref"].String(),
		Status:         record["status"].String(),
		CleanupStatus:  record["cleanup_status"].String(),
		DeliveryMode:   record["delivery_mode"].String(),
		DeliveryStatus: record["delivery_status"].String(),
		SyncStrategy:   record["sync_strategy"].String(),
		SyncStatus:     record["sync_status"].String(),
		RiskLevel:      record["risk_level"].String(),
		PatchRef:       record["patch_ref"].String(),
		DeliveryRef:    record["delivery_ref"].String(),
		DeliveryTitle:  record["delivery_title"].String(),
		DiffSummary:    record["diff_summary"].String(),
	}, nil
}

// updateStatus 更新工作空间状态。
func (r *repo) updateStatus(ctx context.Context, id int64, status string, extra g.Map) error {
	data := g.Map{
		"status":     status,
		"updated_at": gtime.Now(),
	}
	for k, v := range extra {
		data[k] = v
	}
	result, err := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		Where("id", id).
		WhereNull("deleted_at").
		Data(data).
		Update()
	if err != nil {
		if isUnknownColumnErr(err) {
			legacyData := filterWorkspaceDataByError(data, err)
			result, err = g.DB().Model("mvp_task_workspace").Ctx(ctx).
				Where("id", id).
				WhereNull("deleted_at").
				Data(legacyData).
				Update()
		}
	}
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("workspace %d 不存在或已删除", id)
	}
	return nil
}

// softDelete 软删除记录（用于幂等重建前清理旧记录）。
func (r *repo) softDelete(ctx context.Context, id int64) error {
	_, err := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		Where("id", id).
		Data(g.Map{
			"deleted_at": gtime.Now(),
			"updated_at": gtime.Now(),
		}).
		Update()
	return err
}

// updateCleanupStatus 更新清理状态。
func (r *repo) updateCleanupStatus(ctx context.Context, id int64, cleanupStatus string) error {
	_, err := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		Where("id", id).
		Data(g.Map{
			"cleanup_status": cleanupStatus,
			"updated_at":     gtime.Now(),
		}).
		Update()
	return err
}

// listPendingCleanup 查询待清理的工作空间。
func (r *repo) listPendingCleanup(ctx context.Context, olderThanHours int) ([]*TaskWorkspace, error) {
	var result []*TaskWorkspace
	records, err := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		Where("cleanup_status", CleanupPending).
		WhereIn("status", g.Slice{StatusCompleted, StatusFailed, StatusCanceled}).
		WhereLT("updated_at", gtime.Now().Add(time.Duration(-olderThanHours)*time.Hour)).
		WhereNull("deleted_at").
		All()
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		result = append(result, &TaskWorkspace{
			ID:             record["id"].Int64(),
			TaskID:         record["task_id"].Int64(),
			ProjectID:      record["project_id"].Int64(),
			WorkspaceType:  record["workspace_type"].String(),
			WorkspacePath:  record["workspace_path"].String(),
			Status:         record["status"].String(),
			CleanupStatus:  record["cleanup_status"].String(),
			DeliveryMode:   record["delivery_mode"].String(),
			DeliveryStatus: record["delivery_status"].String(),
			SyncStrategy:   record["sync_strategy"].String(),
			SyncStatus:     record["sync_status"].String(),
			RiskLevel:      record["risk_level"].String(),
			PatchRef:       record["patch_ref"].String(),
			DeliveryRef:    record["delivery_ref"].String(),
			DeliveryTitle:  record["delivery_title"].String(),
			DiffSummary:    record["diff_summary"].String(),
		})
	}
	return result, nil
}

func isUnknownColumnErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unknown column")
}

func filterLegacyWorkspaceData(data g.Map) g.Map {
	legacy := g.Map{}
	for k, v := range data {
		if _, blocked := workspaceDeliveryColumns[k]; blocked {
			continue
		}
		legacy[k] = v
	}
	return legacy
}

func filterWorkspaceDataByError(data g.Map, err error) g.Map {
	if isDeliveryReferenceColumnErr(err) {
		partial := g.Map{}
		for k, v := range data {
			if _, blocked := workspaceDeliveryRefColumns[k]; blocked {
				continue
			}
			partial[k] = v
		}
		return partial
	}
	return filterLegacyWorkspaceData(data)
}

func isDeliveryReferenceColumnErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "delivery_ref") || strings.Contains(msg, "delivery_title")
}
