package model

type HealthStatus struct {
	Status       string `json:"status"`
	Service      string `json:"service"`
	Version      string `json:"version"`
	StorageReady bool   `json:"storageReady"`
}

type GameConfig struct {
	GridSize         int    `json:"gridSize"`
	InitialSpeedMs   int    `json:"initialSpeedMs"`
	MinSpeedMs       int    `json:"minSpeedMs"`
	SpeedStepMs      int    `json:"speedStepMs"`
	ScoreStep        int    `json:"scoreStep"`
	LeaderboardLimit int    `json:"leaderboardLimit"`
	DefaultPlayer    string `json:"defaultPlayer"`
}

type ScoreEntry struct {
	PlayerName      string `json:"playerName"`
	Score           int    `json:"score"`
	DurationSeconds int    `json:"durationSeconds"`
	ApplesEaten     int    `json:"applesEaten"`
	RecordedAt      string `json:"recordedAt"`
}

type Scoreboard struct {
	Items []ScoreEntry `json:"items"`
	Total int          `json:"total"`
}

type SubmitScoreInput struct {
	PlayerName      string `json:"playerName"`
	Score           int    `json:"score"`
	DurationSeconds int    `json:"durationSeconds"`
	ApplesEaten     int    `json:"applesEaten"`
}

type SubmitScoreOutput struct {
	Rank  int        `json:"rank"`
	Entry ScoreEntry `json:"entry"`
}
