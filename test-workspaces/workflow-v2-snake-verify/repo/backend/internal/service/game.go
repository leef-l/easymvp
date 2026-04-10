package service

import (
	"context"

	"workflowv2snake/backend/internal/model"
)

type IGame interface {
	Health(ctx context.Context) *model.HealthStatus
	Config(ctx context.Context) *model.GameConfig
	ListScores(ctx context.Context) (*model.Scoreboard, error)
	SubmitScore(ctx context.Context, in model.SubmitScoreInput) (*model.SubmitScoreOutput, error)
}

var localGame IGame

func RegisterGame(game IGame) {
	localGame = game
}

func Game() IGame {
	if localGame == nil {
		panic("service IGame implementation not registered")
	}
	return localGame
}
