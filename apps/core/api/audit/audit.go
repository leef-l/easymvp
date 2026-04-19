package audit

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/audit/v1"
)

type IAuditV1 interface {
	ListProjectAuditLogs(ctx context.Context, req *v1.ListProjectAuditLogsReq) (res *v1.ListProjectAuditLogsRes, err error)
}
