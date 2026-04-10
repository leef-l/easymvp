package event

import "time"

// StreamStatus 表示事件流桥接状态。
type StreamStatus struct {
	Enabled    bool      `json:"enabled"`
	Degraded   bool      `json:"degraded"`
	LastError  string    `json:"last_error,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
	StreamName string    `json:"stream_name,omitempty"`
}

// StreamRuntimeSnapshot 表示 workflow event stream consumer 的运行态快照。
type StreamRuntimeSnapshot struct {
	Enabled           bool      `json:"enabled"`
	ConsumerEnabled   bool      `json:"consumer_enabled"`
	ConsumerCreated   bool      `json:"consumer_created"`
	ConsumerStarted   bool      `json:"consumer_started"`
	GroupReady        bool      `json:"group_ready"`
	Degraded          bool      `json:"degraded"`
	LastError         string    `json:"last_error,omitempty"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
	StreamName        string    `json:"stream_name,omitempty"`
	ConsumerGroup     string    `json:"consumer_group,omitempty"`
	ConsumerName      string    `json:"consumer_name,omitempty"`
	Pending           int64     `json:"pending,omitempty"`
	PendingKnown      bool      `json:"pending_known"`
	Lag               int64     `json:"lag,omitempty"`
	LagKnown          bool      `json:"lag_known"`
	ReclaimAttempts   int64     `json:"reclaim_attempts,omitempty"`
	ReclaimedMessages int64     `json:"reclaimed_messages,omitempty"`
	LastConsumeAt     time.Time `json:"last_consume_at,omitempty"`
	LastAckAt         time.Time `json:"last_ack_at,omitempty"`
	LastReclaimAt     time.Time `json:"last_reclaim_at,omitempty"`
	WorkerHeartbeatAt time.Time `json:"worker_heartbeat_at,omitempty"`
	StartedAt         time.Time `json:"started_at,omitempty"`
}

// StreamSink 抽象持久化事件流投递端。
// 接口保持在 event 包，避免上层依赖具体 Redis Stream 实现。
type StreamSink interface {
	Publish(event Event) error
	Status() StreamStatus
}
