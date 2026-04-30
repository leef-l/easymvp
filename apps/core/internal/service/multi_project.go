package service

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

// ProjectPriority defines the execution priority level for a project slot.
type ProjectPriority int

const (
	PriorityLow      ProjectPriority = 0
	PriorityNormal   ProjectPriority = 1
	PriorityHigh     ProjectPriority = 2
	PriorityCritical ProjectPriority = 3
)

func (p ProjectPriority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// SlotStatus represents the lifecycle state of a project execution slot.
type SlotStatus string

const (
	SlotIdle    SlotStatus = "idle"
	SlotRunning SlotStatus = "running"
	SlotPaused  SlotStatus = "paused"
	SlotQueued  SlotStatus = "queued"
)

// ---------------------------------------------------------------------------
// ProjectSlot
// ---------------------------------------------------------------------------

// ProjectSlot holds the runtime state of a single project execution slot.
type ProjectSlot struct {
	ProjectID   string          `json:"projectId"`
	Status      SlotStatus      `json:"status"`
	Priority    ProjectPriority `json:"priority"`
	StartedAt   time.Time       `json:"startedAt"`
	ResumedAt   time.Time       `json:"resumedAt,omitempty"`
	QueuedAt    time.Time       `json:"queuedAt,omitempty"`
	CPUPercent  float64         `json:"cpuPercent"`
	MemoryBytes int64           `json:"memoryBytes"`
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

// IMultiProjectManager defines the contract for multi-project concurrency control.
type IMultiProjectManager interface {
	AcquireSlot(ctx context.Context, projectID string, priority ProjectPriority) (*ProjectSlot, error)
	ReleaseSlot(ctx context.Context, projectID string) error
	GetSlotStatus(ctx context.Context, projectID string) (*ProjectSlot, error)
	ListActiveSlots(ctx context.Context) []ProjectSlot
	SetConcurrencyLimit(limit int)
	GetConcurrencyLimit() int
	SetPriority(ctx context.Context, projectID string, priority ProjectPriority) error
	PauseProject(ctx context.Context, projectID string) error
	ResumeProject(ctx context.Context, projectID string) error
	ArbitrateConflict(ctx context.Context, projectIDs []string) (string, error)
}

// ---------------------------------------------------------------------------
// Singleton
// ---------------------------------------------------------------------------

var (
	localMultiProjectManager     IMultiProjectManager
	localMultiProjectManagerOnce sync.Once
)

// MultiProjectManager returns the singleton multi-project manager instance.
func MultiProjectManager() IMultiProjectManager {
	localMultiProjectManagerOnce.Do(func() {
		localMultiProjectManager = newMultiProjectManager()
	})
	return localMultiProjectManager
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type sMultiProjectManager struct {
	mu               sync.RWMutex
	slots            map[string]*ProjectSlot
	concurrencyLimit int
}

func newMultiProjectManager() *sMultiProjectManager {
	ctx := gctx.New()
	limit := g.Cfg().MustGet(ctx, "easymvp.multiProject.concurrencyLimit", 3).Int()
	if limit < 1 {
		limit = 1
	}
	return &sMultiProjectManager{
		slots:            make(map[string]*ProjectSlot),
		concurrencyLimit: limit,
	}
}

// AcquireSlot requests an execution slot for the given project.
// If the concurrency limit is reached, the slot is placed in queued state.
func (m *sMultiProjectManager) AcquireSlot(ctx context.Context, projectID string, priority ProjectPriority) (*ProjectSlot, error) {
	if projectID == "" {
		return nil, gerror.New("projectID is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Already has a slot.
	if existing, ok := m.slots[projectID]; ok {
		if priority > existing.Priority {
			existing.Priority = priority
			g.Log().Infof(ctx, "multi-project: priority for %s upgraded to %s", projectID, priority)
			// 如果 existing 是 queued，可能因为 priority 升高需要立即晋升
			if existing.Status == SlotQueued {
				m.promoteNextLocked(ctx)
			}
		}
		cp := *existing
		return &cp, nil
	}

	now := time.Now()
	slot := &ProjectSlot{
		ProjectID: projectID,
		Priority:  priority,
		QueuedAt:  now,
	}

	running := m.countRunningLocked()
	if running < m.concurrencyLimit {
		slot.Status = SlotRunning
		slot.StartedAt = now
		g.Log().Infof(ctx, "multi-project: slot acquired for %s (priority=%s, running=%d/%d)",
			projectID, priority, running+1, m.concurrencyLimit)
	} else {
		slot.Status = SlotQueued
		g.Log().Infof(ctx, "multi-project: slot queued for %s (priority=%s, running=%d/%d)",
			projectID, priority, running, m.concurrencyLimit)
	}

	m.slots[projectID] = slot
	cp := *slot
	return &cp, nil
}

// ReleaseSlot frees the execution slot held by a project and promotes queued
// projects if capacity becomes available.
func (m *sMultiProjectManager) ReleaseSlot(ctx context.Context, projectID string) error {
	if projectID == "" {
		return gerror.New("projectID is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.slots[projectID]; !ok {
		return gerror.Newf("no slot found for project %s", projectID)
	}

	delete(m.slots, projectID)
	g.Log().Infof(ctx, "multi-project: slot released for %s", projectID)

	// Promote highest-priority queued project.
	m.promoteNextLocked(ctx)
	return nil
}

// GetSlotStatus returns the current slot state for a project.
func (m *sMultiProjectManager) GetSlotStatus(ctx context.Context, projectID string) (*ProjectSlot, error) {
	if projectID == "" {
		return nil, gerror.New("projectID is required")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	slot, ok := m.slots[projectID]
	if !ok {
		return nil, gerror.Newf("no slot found for project %s", projectID)
	}

	// Return a copy to avoid data races on the caller side.
	cp := *slot
	return &cp, nil
}

// ListActiveSlots returns a snapshot of all non-idle slots.
func (m *sMultiProjectManager) ListActiveSlots(ctx context.Context) []ProjectSlot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]ProjectSlot, 0, len(m.slots))
	for _, s := range m.slots {
		result = append(result, *s)
	}
	return result
}

// SetConcurrencyLimit updates the maximum number of concurrently running projects.
func (m *sMultiProjectManager) SetConcurrencyLimit(limit int) {
	if limit < 1 {
		limit = 1
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.concurrencyLimit = limit
}

// GetConcurrencyLimit returns the current concurrency limit.
func (m *sMultiProjectManager) GetConcurrencyLimit() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.concurrencyLimit
}

// SetPriority updates the priority of an existing project slot. If the slot is
// in queued state, it may be promoted to running when capacity allows.
func (m *sMultiProjectManager) SetPriority(ctx context.Context, projectID string, priority ProjectPriority) error {
	if projectID == "" {
		return gerror.New("projectID is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	slot, ok := m.slots[projectID]
	if !ok {
		return gerror.Newf("no slot found for project %s", projectID)
	}

	slot.Priority = priority
	g.Log().Infof(ctx, "multi-project: priority for %s set to %s", projectID, priority)

	if slot.Status == SlotQueued {
		m.promoteNextLocked(ctx)
	}
	return nil
}

// PauseProject transitions a running project to the paused state.
func (m *sMultiProjectManager) PauseProject(ctx context.Context, projectID string) error {
	if projectID == "" {
		return gerror.New("projectID is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	slot, ok := m.slots[projectID]
	if !ok {
		return gerror.Newf("no slot found for project %s", projectID)
	}
	if slot.Status != SlotRunning {
		return gerror.Newf("project %s is not running (current: %s)", projectID, slot.Status)
	}

	slot.Status = SlotPaused
	g.Log().Infof(ctx, "multi-project: project %s paused", projectID)

	// Pausing frees a running slot; promote next in queue.
	m.promoteNextLocked(ctx)
	return nil
}

// ResumeProject transitions a paused project back to running, provided there
// is capacity. If not, the project is queued.
func (m *sMultiProjectManager) ResumeProject(ctx context.Context, projectID string) error {
	if projectID == "" {
		return gerror.New("projectID is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	slot, ok := m.slots[projectID]
	if !ok {
		return gerror.Newf("no slot found for project %s", projectID)
	}
	if slot.Status != SlotPaused {
		return gerror.Newf("project %s is not paused (current: %s)", projectID, slot.Status)
	}

	running := m.countRunningLocked()
	if running < m.concurrencyLimit {
		slot.Status = SlotRunning
		slot.ResumedAt = time.Now()
		g.Log().Infof(ctx, "multi-project: project %s resumed (running=%d/%d)",
			projectID, running+1, m.concurrencyLimit)
	} else {
		slot.Status = SlotQueued
		slot.QueuedAt = time.Now()
		g.Log().Infof(ctx, "multi-project: project %s re-queued (running=%d/%d)",
			projectID, running, m.concurrencyLimit)
	}
	return nil
}

// ArbitrateConflict resolves resource contention among the given project IDs
// by returning the one that should execute first. Strategy: highest priority
// wins; ties broken by longest wait time.
func (m *sMultiProjectManager) ArbitrateConflict(ctx context.Context, projectIDs []string) (string, error) {
	if len(projectIDs) == 0 {
		return "", gerror.New("at least one projectID is required")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	type candidate struct {
		id       string
		priority ProjectPriority
		queued   time.Time
	}

	var candidates []candidate
	for _, id := range projectIDs {
		slot, ok := m.slots[id]
		if !ok {
			continue
		}
		candidates = append(candidates, candidate{
			id:       id,
			priority: slot.Priority,
			queued:   slot.QueuedAt,
		})
	}

	if len(candidates) == 0 {
		// None of the IDs have slots; fall back to first in list.
		return projectIDs[0], nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority > candidates[j].priority // higher priority first
		}
		return candidates[i].queued.Before(candidates[j].queued) // longer wait first
	})

	winner := candidates[0].id
	g.Log().Infof(ctx, "multi-project: arbitration winner=%s among %v", winner, projectIDs)
	return winner, nil
}

// ---------------------------------------------------------------------------
// Internal helpers (caller must hold m.mu)
// ---------------------------------------------------------------------------

// countRunningLocked returns the number of slots in Running state.
func (m *sMultiProjectManager) countRunningLocked() int {
	count := 0
	for _, s := range m.slots {
		if s.Status == SlotRunning {
			count++
		}
	}
	return count
}

// promoteNextLocked picks the highest-priority queued slot and promotes it to
// running if capacity is available.
func (m *sMultiProjectManager) promoteNextLocked(ctx context.Context) {
	if m.countRunningLocked() >= m.concurrencyLimit {
		return
	}

	var best *ProjectSlot
	for _, s := range m.slots {
		if s.Status != SlotQueued {
			continue
		}
		if best == nil || s.Priority > best.Priority ||
			(s.Priority == best.Priority && s.QueuedAt.Before(best.QueuedAt)) {
			best = s
		}
	}

	if best != nil {
		best.Status = SlotRunning
		best.StartedAt = time.Now()
		g.Log().Infof(ctx, "multi-project: promoted %s from queue (priority=%s)", best.ProjectID, best.Priority)
	}
}
