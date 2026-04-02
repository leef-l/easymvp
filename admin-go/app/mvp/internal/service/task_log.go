package service

import (
	"context"
	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
	"github.com/gogf/gf/v2/net/ghttp"
)

type ITaskLog interface {
	Create(ctx context.Context, in *model.TaskLogCreateInput) error
	Update(ctx context.Context, in *model.TaskLogUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.TaskLogDetailOutput, err error)
	List(ctx context.Context, in *model.TaskLogListInput) (list []*model.TaskLogListOutput, total int, err error)
	Export(ctx context.Context, in *model.TaskLogListInput) (list []*model.TaskLogListOutput, err error)
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
}

var localTaskLog ITaskLog

func TaskLog() ITaskLog {
	return localTaskLog
}

func RegisterTaskLog(i ITaskLog) {
	localTaskLog = i
}
