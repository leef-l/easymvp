package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"workflowv2snake/backend/internal/model"
)

type HealthReq struct {
	g.Meta `path:"/api/health" method:"get" tags:"System" summary:"Backend health check"`
}

type HealthRes struct {
	Status model.HealthStatus `json:"status"`
}

type ConfigReq struct {
	g.Meta `path:"/api/game/config" method:"get" tags:"Game" summary:"Fetch runtime game config"`
}

type ConfigRes struct {
	Config model.GameConfig `json:"config"`
}

type ScoresReq struct {
	g.Meta `path:"/api/game/scores" method:"get" tags:"Game" summary:"List leaderboard scores"`
}

type ScoresRes struct {
	Items []model.ScoreEntry `json:"items"`
	Total int                `json:"total"`
}

type SubmitScoreReq struct {
	g.Meta          `path:"/api/game/scores" method:"post" tags:"Game" summary:"Submit a completed snake run"`
	PlayerName      string `json:"playerName" v:"required|length:1,24"`
	Score           int    `json:"score" v:"required|min:1"`
	DurationSeconds int    `json:"durationSeconds" v:"required|min:1"`
	ApplesEaten     int    `json:"applesEaten" v:"min:0"`
}

type SubmitScoreRes struct {
	Rank    int              `json:"rank"`
	Entry   model.ScoreEntry `json:"entry"`
	Message string           `json:"message,omitempty"`
}
