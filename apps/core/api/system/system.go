// =================================================================================
// Code generated and maintained by GoFrame CLI tool style.
// =================================================================================

package system

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/system/v1"
)

type ISystemV1 interface {
	Health(ctx context.Context, req *v1.HealthReq) (res *v1.HealthRes, err error)
	ListProjectDiagnostics(ctx context.Context, req *v1.ListProjectDiagnosticsReq) (res *v1.ListProjectDiagnosticsRes, err error)
}
