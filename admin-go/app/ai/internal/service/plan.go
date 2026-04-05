package service

import (
	"context"
	"easymvp/app/ai/internal/model"
	"easymvp/utility/snowflake"
	"github.com/gogf/gf/v2/net/ghttp"
)

type IPlan interface {
	Create(ctx context.Context, in *model.PlanCreateInput) (int, error)
	Update(ctx context.Context, in *model.PlanUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.PlanDetailOutput, err error)
	List(ctx context.Context, in *model.PlanListInput) (list []*model.PlanListOutput, total int, err error)
	Export(ctx context.Context, in *model.PlanListInput) (list []*model.PlanListOutput, err error)
	BatchUpdate(ctx context.Context, in *model.PlanBatchUpdateInput) error
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
}

var localPlan IPlan

func Plan() IPlan {
	return localPlan
}

func RegisterPlan(i IPlan) {
	localPlan = i
}
