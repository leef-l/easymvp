package audit

import (
	"context"

	auditv1 "github.com/leef-l/easymvp/apps/core/api/audit/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) ListProjectAuditLogs(ctx context.Context, req *auditv1.ListProjectAuditLogsReq) (res *auditv1.ListProjectAuditLogsRes, err error) {
	return service.Audit().ListProjectAuditLogs(ctx, req.ProjectID, req.Limit)
}
