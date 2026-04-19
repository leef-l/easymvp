package service

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type IWorkerManager interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type backgroundWorker interface {
	Name() string
	Interval() time.Duration
	RunOnce(ctx context.Context) error
}

var localWorkerManager IWorkerManager = (*sWorkerManager)(nil)

type sWorkerManager struct {
	mu      sync.Mutex
	started bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	workers []backgroundWorker
}

func Workers() IWorkerManager {
	if localWorkerManager == nil {
		localWorkerManager = &sWorkerManager{
			workers: []backgroundWorker{
				newRunSyncWorker(),
				newWorkspaceSnapshotRefreshWorker(),
			},
		}
	}
	return localWorkerManager
}

func (m *sWorkerManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return nil
	}

	runCtx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.started = true

	for _, worker := range m.workers {
		currentWorker := worker
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.runWorkerLoop(runCtx, currentWorker)
		}()
	}

	g.Log().Infof(ctx, "worker manager started with %d workers", len(m.workers))
	return nil
}

func (m *sWorkerManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	if !m.started {
		m.mu.Unlock()
		return nil
	}
	cancel := m.cancel
	m.cancel = nil
	m.started = false
	m.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	waitDone := make(chan struct{})
	go func() {
		defer close(waitDone)
		m.wg.Wait()
	}()

	timeout := 5 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > 0 && remaining < timeout {
			timeout = remaining
		}
	}

	select {
	case <-waitDone:
		g.Log().Info(ctx, "worker manager stopped")
		return nil
	case <-time.After(timeout):
		g.Log().Warning(ctx, "worker manager stop timed out")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *sWorkerManager) runWorkerLoop(ctx context.Context, worker backgroundWorker) {
	m.runWorkerSafely(ctx, worker)

	ticker := time.NewTicker(worker.Interval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.runWorkerSafely(ctx, worker)
		}
	}
}

func (m *sWorkerManager) runWorkerSafely(ctx context.Context, worker backgroundWorker) {
	defer func() {
		if recovered := recover(); recovered != nil {
			handleWorkerFailure(
				ctx,
				worker.Name(),
				"",
				"panic",
				"worker panic recovered",
				map[string]any{
					"panic": recovered,
				},
			)
		}
	}()

	if err := worker.RunOnce(ctx); err != nil {
		handleWorkerFailure(
			ctx,
			worker.Name(),
			"",
			"run_failed",
			"worker run failed",
			map[string]any{
				"error": err.Error(),
			},
		)
	}
}
