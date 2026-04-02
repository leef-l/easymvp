package service

import (
	"context"
	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
	"github.com/gogf/gf/v2/net/ghttp"
)

type IConversation interface {
	Create(ctx context.Context, in *model.ConversationCreateInput) error
	Update(ctx context.Context, in *model.ConversationUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ConversationDetailOutput, err error)
	List(ctx context.Context, in *model.ConversationListInput) (list []*model.ConversationListOutput, total int, err error)
	Export(ctx context.Context, in *model.ConversationListInput) (list []*model.ConversationListOutput, err error)
	BatchUpdate(ctx context.Context, in *model.ConversationBatchUpdateInput) error
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
}

var localConversation IConversation

func Conversation() IConversation {
	return localConversation
}

func RegisterConversation(i IConversation) {
	localConversation = i
}
