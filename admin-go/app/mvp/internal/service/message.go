package service

import (
	"context"
	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
	"github.com/gogf/gf/v2/net/ghttp"
)

type IMessage interface {
	Create(ctx context.Context, in *model.MessageCreateInput) error
	Update(ctx context.Context, in *model.MessageUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.MessageDetailOutput, err error)
	List(ctx context.Context, in *model.MessageListInput) (list []*model.MessageListOutput, total int, err error)
	Export(ctx context.Context, in *model.MessageListInput) (list []*model.MessageListOutput, err error)
	BatchUpdate(ctx context.Context, in *model.MessageBatchUpdateInput) error
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
}

var localMessage IMessage

func Message() IMessage {
	return localMessage
}

func RegisterMessage(i IMessage) {
	localMessage = i
}
