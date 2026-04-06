package chat

import (
	"sync"
	"time"
)

type tgSession struct {
	ConversationID int64
	LastActive     time.Time
}

var (
	tgSessions   = map[string]*tgSession{}
	tgSessionsMu sync.Mutex
)

func setTGSession(chatIDStr string, conversationID int64) {
	tgSessionsMu.Lock()
	defer tgSessionsMu.Unlock()
	tgSessions[chatIDStr] = &tgSession{
		ConversationID: conversationID,
		LastActive:     time.Now(),
	}
}

func getTGSession(chatIDStr string) (int64, bool) {
	tgSessionsMu.Lock()
	defer tgSessionsMu.Unlock()
	s, ok := tgSessions[chatIDStr]
	if !ok {
		return 0, false
	}
	if time.Since(s.LastActive) > 30*time.Minute {
		delete(tgSessions, chatIDStr)
		return 0, false
	}
	s.LastActive = time.Now()
	return s.ConversationID, true
}

func clearTGSession(chatIDStr string) {
	tgSessionsMu.Lock()
	defer tgSessionsMu.Unlock()
	delete(tgSessions, chatIDStr)
}
