package executor

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestWaitForChatReplyReturnsCompletedContent(t *testing.T) {
	originalLoad := loadChatReplySnapshot
	originalInterval := chatReplyPollInterval
	t.Cleanup(func() {
		loadChatReplySnapshot = originalLoad
		chatReplyPollInterval = originalInterval
	})

	chatReplyPollInterval = time.Millisecond
	attempts := 0
	loadChatReplySnapshot = func(ctx context.Context, replyID int64) (chatReplySnapshot, error) {
		attempts++
		if attempts < 3 {
			return chatReplySnapshot{Status: "streaming"}, nil
		}
		return chatReplySnapshot{
			Status:  "completed",
			Content: "{\"task_repair\":{\"task_name\":\"demo\",\"description\":\"ready\"}}",
		}, nil
	}

	content, err := waitForChatReply(context.Background(), 1)
	if err != nil {
		t.Fatalf("waitForChatReply() error = %v", err)
	}
	if !strings.Contains(content, "\"task_repair\"") {
		t.Fatalf("unexpected content: %s", content)
	}
	if attempts < 3 {
		t.Fatalf("expected polling before completion, attempts=%d", attempts)
	}
}

func TestWaitForChatReplyReturnsFailureContent(t *testing.T) {
	originalLoad := loadChatReplySnapshot
	originalInterval := chatReplyPollInterval
	t.Cleanup(func() {
		loadChatReplySnapshot = originalLoad
		chatReplyPollInterval = originalInterval
	})

	chatReplyPollInterval = time.Millisecond
	loadChatReplySnapshot = func(ctx context.Context, replyID int64) (chatReplySnapshot, error) {
		return chatReplySnapshot{
			Status:  "failed",
			Content: "AI 调用失败: rate limit",
		}, nil
	}

	_, err := waitForChatReply(context.Background(), 1)
	if err == nil {
		t.Fatal("expected failed reply to return error")
	}
	if !strings.Contains(err.Error(), "rate limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitForChatReplyHonorsContextCancellation(t *testing.T) {
	originalLoad := loadChatReplySnapshot
	originalInterval := chatReplyPollInterval
	t.Cleanup(func() {
		loadChatReplySnapshot = originalLoad
		chatReplyPollInterval = originalInterval
	})

	chatReplyPollInterval = 5 * time.Millisecond
	loadChatReplySnapshot = func(ctx context.Context, replyID int64) (chatReplySnapshot, error) {
		return chatReplySnapshot{Status: "streaming"}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()

	_, err := waitForChatReply(ctx, 1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}
