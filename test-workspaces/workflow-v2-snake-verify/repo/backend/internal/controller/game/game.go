package game

import (
	"context"

	v1 "workflowv2snake/backend/api/game/v1"
	"workflowv2snake/backend/internal/model"
	"workflowv2snake/backend/internal/service"
)

func (c *ControllerV1) Health(ctx context.Context, _ *v1.HealthReq) (res *v1.HealthRes, err error) {
	return &v1.HealthRes{Status: *service.Game().Health(ctx)}, nil
}

func (c *ControllerV1) Config(ctx context.Context, _ *v1.ConfigReq) (res *v1.ConfigRes, err error) {
	return &v1.ConfigRes{Config: *service.Game().Config(ctx)}, nil
}

func (c *ControllerV1) Scores(ctx context.Context, _ *v1.ScoresReq) (res *v1.ScoresRes, err error) {
	board, err := service.Game().ListScores(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ScoresRes{Items: board.Items, Total: board.Total}, nil
}

func (c *ControllerV1) SubmitScore(ctx context.Context, req *v1.SubmitScoreReq) (res *v1.SubmitScoreRes, err error) {
	output, err := service.Game().SubmitScore(ctx, model.SubmitScoreInput{
		PlayerName:      req.PlayerName,
		Score:           req.Score,
		DurationSeconds: req.DurationSeconds,
		ApplesEaten:     req.ApplesEaten,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SubmitScoreRes{
		Rank:    output.Rank,
		Entry:   output.Entry,
		Message: "score recorded",
	}, nil
}
