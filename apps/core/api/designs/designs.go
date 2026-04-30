package designs

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/designs/v1"
)

type IDesignsV1 interface {
	Generate(ctx context.Context, req *v1.GenerateDesignReq) (res *v1.GenerateDesignRes, err error)
	Confirm(ctx context.Context, req *v1.ConfirmDesignReq) (res *v1.ConfirmDesignRes, err error)
	Reject(ctx context.Context, req *v1.RejectDesignReq) (res *v1.RejectDesignRes, err error)
	Get(ctx context.Context, req *v1.GetDesignReq) (res *v1.GetDesignRes, err error)
	GetProjectLatest(ctx context.Context, req *v1.GetProjectLatestDesignReq) (res *v1.GetProjectLatestDesignRes, err error)
}
