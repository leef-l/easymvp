package chat

import (
	"sync"
	"time"
)

type feishuSession struct {
	ConversationID int64
	LastActive     time.Time
}

var (
	feishuSessions   = map[string]*feishuSession{}
	feishuSessionsMu sync.Mutex
)

func setFeishuSession(openID string, conversationID int64) {
	feishuSessionsMu.Lock()
	defer feishuSessionsMu.Unlock()
	feishuSessions[openID] = &feishuSession{
		ConversationID: conversationID,
		LastActive:     time.Now(),
	}
}

func getFeishuSession(openID string) (int64, bool) {
	feishuSessionsMu.Lock()
	defer feishuSessionsMu.Unlock()
	s, ok := feishuSessions[openID]
	if !ok {
		return 0, false
	}
	if time.Since(s.LastActive) > 30*time.Minute {
		delete(feishuSessions, openID)
		return 0, false
	}
	s.LastActive = time.Now()
	return s.ConversationID, true
}

func clearFeishuSession(openID string) {
	feishuSessionsMu.Lock()
	defer feishuSessionsMu.Unlock()
	delete(feishuSessions, openID)
}
