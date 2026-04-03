package service

import (
	"context"

	"easymvp/app/ai/internal/model"
)

type IEngine interface {
	List(ctx context.Context) (list []*model.EngineListOutput, err error)
	Detail(ctx context.Context, engineCode string) (out *model.EngineDetailOutput, err error)
	Update(ctx context.Context, in *model.EngineUpdateInput) error
	TestConnection(ctx context.Context, engineCode string) (*model.EngineTestOutput, error)
}

var localEngine IEngine

func Engine() IEngine {
	return localEngine
}

func RegisterEngine(i IEngine) {
	localEngine = i
}
