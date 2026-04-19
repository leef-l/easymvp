package runtime

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/runtime/v1"
)

type IRuntimeV1 interface {
	Healthz(ctx context.Context, req *v1.HealthzReq) (res *v1.HealthzRes, err error)
	ExecutionView(ctx context.Context, req *v1.ExecutionViewReq) (res *v1.ExecutionViewRes, err error)
	StartRun(ctx context.Context, req *v1.StartRunReq) (res *v1.StartRunRes, err error)
	GetRunBinding(ctx context.Context, req *v1.GetRunBindingReq) (res *v1.GetRunBindingRes, err error)
	GetRunBindingDetail(ctx context.Context, req *v1.GetRunBindingDetailReq) (res *v1.GetRunBindingDetailRes, err error)
	ListRunBindingEvents(ctx context.Context, req *v1.ListRunBindingEventsReq) (res *v1.ListRunBindingEventsRes, err error)
	SyncRunBinding(ctx context.Context, req *v1.SyncRunBindingReq) (res *v1.SyncRunBindingRes, err error)
	ResumeRunBinding(ctx context.Context, req *v1.ResumeRunBindingReq) (res *v1.ResumeRunBindingRes, err error)
	CancelRunBinding(ctx context.Context, req *v1.CancelRunBindingReq) (res *v1.CancelRunBindingRes, err error)
}
