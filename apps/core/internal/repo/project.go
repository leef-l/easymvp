package repo

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/do"
)

func InsertProjectRow(ctx context.Context, tx gdb.TX, row *do.Projects) error {
	result, err := tx.Model(dao.Projects.Table()).Data(g.Map{
		"id":                       row.Id,
		"name":                     row.Name,
		"project_category":         row.ProjectCategory,
		"goal_summary":             row.GoalSummary,
		"status":                   row.Status,
		"production_status":        row.ProductionStatus,
		"workspace_root":           row.WorkspaceRoot,
		"repo_root":                nullIfEmpty(row.RepoRoot),
		"current_plan_draft_id":    nullIfEmpty(row.CurrentPlanDraftId),
		"current_compiled_plan_id": nullIfEmpty(row.CurrentCompiledPlanId),
		"created_at":               row.CreatedAt,
		"updated_at":               row.UpdatedAt,
	}).Insert()
	if err != nil {
		return gerror.Wrap(err, "insert project failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert project affected unexpected rows")
	}
	return nil
}

func InsertProjectProfileRow(ctx context.Context, tx gdb.TX, row *do.ProjectProfiles) error {
	result, err := tx.Model(dao.ProjectProfiles.Table()).Data(g.Map{
		"id":                         row.Id,
		"project_id":                 row.ProjectId,
		"category_profile_version":   row.CategoryProfileVersion,
		"acceptance_profile_version": row.AcceptanceProfileVersion,
		"role_profile_version":       nullIfEmpty(row.RoleProfileVersion),
		"created_at":                 row.CreatedAt,
		"updated_at":                 row.UpdatedAt,
	}).Insert()
	if err != nil {
		return gerror.Wrap(err, "insert project profile failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert project profile affected unexpected rows")
	}
	return nil
}

func InsertProjectWorkspaceRow(ctx context.Context, tx gdb.TX, row *do.ProjectWorkspaces) error {
	result, err := tx.Model(dao.ProjectWorkspaces.Table()).Data(g.Map{
		"id":               row.Id,
		"project_id":       row.ProjectId,
		"workspace_root":   row.WorkspaceRoot,
		"evidence_root":    row.EvidenceRoot,
		"runs_root":        row.RunsRoot,
		"replay_root":      row.ReplayRoot,
		"diagnostics_root": row.DiagnosticsRoot,
		"created_at":       row.CreatedAt,
		"updated_at":       row.UpdatedAt,
	}).Insert()
	if err != nil {
		return gerror.Wrap(err, "insert project workspace failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert project workspace affected unexpected rows")
	}
	return nil
}

func UpdateProjectRow(ctx context.Context, projectID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	result, err := dao.Projects.Ctx(ctx).Data(g.Map(updates)).Where("id", projectID).Update()
	if err != nil {
		return gerror.Wrap(err, "update project failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.Newf("update project affected unexpected rows: %d", affected)
	}
	return nil
}

func UpdateProjectStatus(ctx context.Context, projectID string, status string) error {
	_, err := dao.Projects.Ctx(ctx).Data(g.Map{
		"status":     status,
		"updated_at": nowText(),
	}).Where("id", projectID).Update()
	if err != nil {
		return gerror.Wrapf(err, "update project status to %s failed", status)
	}
	return nil
}

func UpdateProjectStatusTx(ctx context.Context, tx gdb.TX, projectID string, status string) error {
	_, err := tx.Model(dao.Projects.Table()).Data(g.Map{
		"status":     status,
		"updated_at": nowText(),
	}).Where("id", projectID).Update()
	if err != nil {
		return gerror.Wrapf(err, "update project status to %s failed", status)
	}
	return nil
}

func DeleteProjectByID(ctx context.Context, projectID string) error {
	result, err := dao.Projects.Ctx(ctx).Where("id", projectID).Delete()
	if err != nil {
		return gerror.Wrap(err, "delete project failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.Newf("delete project affected unexpected rows: %d", affected)
	}
	return nil
}

func InsertAuditLog(ctx context.Context, projectID, eventType, actorKind, summary string, payload map[string]any) error {
	payloadJSON := "{}"
	if payload != nil {
		payloadJSON = mustMarshalJSONString(payload, "{}")
	}
	_, err := dao.AuditLogs.Ctx(ctx).Data(g.Map{
		"id":           newResourceID("audit"),
		"project_id":   projectID,
		"event_type":   eventType,
		"actor_kind":   actorKind,
		"summary":      summary,
		"payload_json": payloadJSON,
		"created_at":   nowText(),
	}).Insert()
	if err != nil {
		return gerror.Wrap(err, "insert audit log failed")
	}
	return nil
}

func InsertAuditLogTx(ctx context.Context, tx gdb.TX, projectID, eventType, actorKind, summary string, payload map[string]any) error {
	payloadJSON := "{}"
	if payload != nil {
		payloadJSON = mustMarshalJSONString(payload, "{}")
	}
	_, err := tx.Model(dao.AuditLogs.Table()).Data(g.Map{
		"id":           newResourceID("audit"),
		"project_id":   projectID,
		"event_type":   eventType,
		"actor_kind":   actorKind,
		"summary":      summary,
		"payload_json": payloadJSON,
		"created_at":   nowText(),
	}).Insert()
	if err != nil {
		return gerror.Wrap(err, "insert audit log failed")
	}
	return nil
}
