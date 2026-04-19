package service

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	auditv1 "github.com/leef-l/easymvp/apps/core/api/audit/v1"
)

type IAudit interface {
	ListProjectAuditLogs(ctx context.Context, projectID string, limit int) (*auditv1.ListProjectAuditLogsRes, error)
}

var localAudit IAudit = (*sAudit)(nil)

type sAudit struct{}

func Audit() IAudit {
	if localAudit == nil {
		localAudit = &sAudit{}
	}
	return localAudit
}

func (s *sAudit) ListProjectAuditLogs(ctx context.Context, projectID string, limit int) (*auditv1.ListProjectAuditLogsRes, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	if limit <= 0 {
		limit = 50
	}
	return listProjectAuditLogsView(ctx, projectID, limit)
}
