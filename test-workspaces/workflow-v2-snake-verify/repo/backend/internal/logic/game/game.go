package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"workflowv2snake/backend/internal/model"
	"workflowv2snake/backend/internal/service"
)

type sGame struct {
	mu          sync.Mutex
	storagePath string
	config      model.GameConfig
	nowFn       func() time.Time
	loaded      bool
	scores      []model.ScoreEntry
}

func init() {
	service.RegisterGame(New())
}

func New() *sGame {
	ctx := context.Background()

	storagePath := "storage/scores.json"
	if value, err := g.Cfg().Get(ctx, "game.scoreStoragePath"); err == nil && strings.TrimSpace(value.String()) != "" {
		storagePath = strings.TrimSpace(value.String())
	}

	leaderboardLimit := 8
	if value, err := g.Cfg().Get(ctx, "game.leaderboardLimit"); err == nil && value.Int() > 0 {
		leaderboardLimit = value.Int()
	}

	return newGame(storagePath, defaultConfig(leaderboardLimit), time.Now)
}

func newGame(storagePath string, config model.GameConfig, nowFn func() time.Time) *sGame {
	return &sGame{
		storagePath: storagePath,
		config:      config,
		nowFn:       nowFn,
		scores:      make([]model.ScoreEntry, 0, config.LeaderboardLimit),
	}
}

func defaultConfig(leaderboardLimit int) model.GameConfig {
	return model.GameConfig{
		GridSize:         20,
		InitialSpeedMs:   160,
		MinSpeedMs:       80,
		SpeedStepMs:      6,
		ScoreStep:        10,
		LeaderboardLimit: leaderboardLimit,
		DefaultPlayer:    "Player 1",
	}
}

func (s *sGame) Health(_ context.Context) *model.HealthStatus {
	err := s.ensureLoaded()
	return &model.HealthStatus{
		Status:       "ok",
		Service:      "snake-backend",
		Version:      "goframe-v2",
		StorageReady: err == nil,
	}
}

func (s *sGame) Config(_ context.Context) *model.GameConfig {
	config := s.config
	return &config
}

func (s *sGame) ListScores(_ context.Context) (*model.Scoreboard, error) {
	if err := s.ensureLoaded(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	items := make([]model.ScoreEntry, len(s.scores))
	copy(items, s.scores)
	return &model.Scoreboard{Items: items, Total: len(items)}, nil
}

func (s *sGame) SubmitScore(_ context.Context, in model.SubmitScoreInput) (*model.SubmitScoreOutput, error) {
	if err := s.ensureLoaded(); err != nil {
		return nil, err
	}

	entry, err := s.buildEntry(in)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.scores = append(s.scores, entry)
	sortScores(s.scores)
	if len(s.scores) > s.config.LeaderboardLimit {
		s.scores = append([]model.ScoreEntry(nil), s.scores[:s.config.LeaderboardLimit]...)
	}
	if err := s.persistLocked(); err != nil {
		return nil, err
	}

	return &model.SubmitScoreOutput{
		Rank:  findRank(s.scores, entry),
		Entry: entry,
	}, nil
}

func (s *sGame) ensureLoaded() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.loaded {
		return nil
	}
	if err := s.loadLocked(); err != nil {
		return err
	}
	s.loaded = true
	return nil
}

func (s *sGame) loadLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.storagePath), 0o755); err != nil {
		return fmt.Errorf("create storage directory: %w", err)
	}

	payload, err := os.ReadFile(s.storagePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.scores = make([]model.ScoreEntry, 0, s.config.LeaderboardLimit)
			return s.persistLocked()
		}
		return fmt.Errorf("read scores file: %w", err)
	}
	if len(payload) == 0 {
		s.scores = make([]model.ScoreEntry, 0, s.config.LeaderboardLimit)
		return nil
	}

	var items []model.ScoreEntry
	if err := json.Unmarshal(payload, &items); err != nil {
		return fmt.Errorf("decode scores file: %w", err)
	}
	sortScores(items)
	if len(items) > s.config.LeaderboardLimit {
		items = items[:s.config.LeaderboardLimit]
	}
	s.scores = items
	return nil
}

func (s *sGame) persistLocked() error {
	payload, err := json.MarshalIndent(s.scores, "", "  ")
	if err != nil {
		return fmt.Errorf("encode scores file: %w", err)
	}
	if err := os.WriteFile(s.storagePath, payload, 0o644); err != nil {
		return fmt.Errorf("write scores file: %w", err)
	}
	return nil
}

func (s *sGame) buildEntry(in model.SubmitScoreInput) (model.ScoreEntry, error) {
	playerName := strings.TrimSpace(in.PlayerName)
	if playerName == "" {
		playerName = s.config.DefaultPlayer
	}
	playerName = truncateRunes(playerName, 24)

	if in.Score <= 0 {
		return model.ScoreEntry{}, errors.New("score must be greater than zero")
	}
	if in.DurationSeconds <= 0 {
		return model.ScoreEntry{}, errors.New("durationSeconds must be greater than zero")
	}
	if in.ApplesEaten < 0 {
		return model.ScoreEntry{}, errors.New("applesEaten cannot be negative")
	}

	return model.ScoreEntry{
		PlayerName:      playerName,
		Score:           in.Score,
		DurationSeconds: in.DurationSeconds,
		ApplesEaten:     in.ApplesEaten,
		RecordedAt:      s.nowFn().UTC().Format(time.RFC3339),
	}, nil
}

func sortScores(items []model.ScoreEntry) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Score != items[j].Score {
			return items[i].Score > items[j].Score
		}
		if items[i].ApplesEaten != items[j].ApplesEaten {
			return items[i].ApplesEaten > items[j].ApplesEaten
		}
		if items[i].DurationSeconds != items[j].DurationSeconds {
			return items[i].DurationSeconds < items[j].DurationSeconds
		}
		return items[i].RecordedAt < items[j].RecordedAt
	})
}

func findRank(items []model.ScoreEntry, target model.ScoreEntry) int {
	for index, item := range items {
		if item == target {
			return index + 1
		}
	}
	return len(items)
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
