package game

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"workflowv2snake/backend/internal/model"
)

func TestSubmitScoreOrdersAndTrimsLeaderboard(t *testing.T) {
	t.Parallel()

	current := time.Date(2026, 4, 9, 8, 0, 0, 0, time.UTC)
	service := newGame(
		filepath.Join(t.TempDir(), "scores.json"),
		defaultConfig(3),
		func() time.Time {
			current = current.Add(time.Second)
			return current
		},
	)

	inputs := []model.SubmitScoreInput{
		{PlayerName: "Nova", Score: 120, DurationSeconds: 40, ApplesEaten: 12},
		{PlayerName: "Iris", Score: 140, DurationSeconds: 38, ApplesEaten: 14},
		{PlayerName: "Milo", Score: 140, DurationSeconds: 45, ApplesEaten: 14},
		{PlayerName: "Kai", Score: 110, DurationSeconds: 30, ApplesEaten: 11},
	}

	for _, input := range inputs {
		if _, err := service.SubmitScore(context.Background(), input); err != nil {
			t.Fatalf("SubmitScore() error = %v", err)
		}
	}

	board, err := service.ListScores(context.Background())
	if err != nil {
		t.Fatalf("ListScores() error = %v", err)
	}

	if board.Total != 3 {
		t.Fatalf("expected top 3 scores, got %d", board.Total)
	}
	if board.Items[0].PlayerName != "Iris" {
		t.Fatalf("expected Iris to rank #1, got %s", board.Items[0].PlayerName)
	}
	if board.Items[1].PlayerName != "Milo" {
		t.Fatalf("expected Milo to rank #2, got %s", board.Items[1].PlayerName)
	}
	if board.Items[2].PlayerName != "Nova" {
		t.Fatalf("expected Nova to rank #3, got %s", board.Items[2].PlayerName)
	}
}

func TestSubmitScoreFallsBackToDefaultPlayerName(t *testing.T) {
	t.Parallel()

	service := newGame(
		filepath.Join(t.TempDir(), "scores.json"),
		defaultConfig(5),
		func() time.Time { return time.Date(2026, 4, 9, 8, 0, 0, 0, time.UTC) },
	)

	result, err := service.SubmitScore(context.Background(), model.SubmitScoreInput{
		PlayerName:      "   ",
		Score:           90,
		DurationSeconds: 21,
		ApplesEaten:     9,
	})
	if err != nil {
		t.Fatalf("SubmitScore() error = %v", err)
	}
	if result.Entry.PlayerName != service.config.DefaultPlayer {
		t.Fatalf("expected default player name %q, got %q", service.config.DefaultPlayer, result.Entry.PlayerName)
	}
}
