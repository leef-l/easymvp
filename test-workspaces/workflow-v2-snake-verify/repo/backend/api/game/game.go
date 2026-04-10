package game

import (
	"context"

	v1 "workflowv2snake/backend/api/game/v1"
)

type IGameV1 interface {
	Health(ctx context.Context, req *v1.HealthReq) (res *v1.HealthRes, err error)
	Config(ctx context.Context, req *v1.ConfigReq) (res *v1.ConfigRes, err error)
	Scores(ctx context.Context, req *v1.ScoresReq) (res *v1.ScoresRes, err error)
	SubmitScore(ctx context.Context, req *v1.SubmitScoreReq) (res *v1.SubmitScoreRes, err error)
}
