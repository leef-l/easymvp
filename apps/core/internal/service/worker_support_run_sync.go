package service

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type runSyncWorker struct {
	interval  time.Duration
	batchSize int
}

func newRunSyncWorker() backgroundWorker {
	cfgInterval := g.Cfg().MustGet(context.Background(), "easymvp.workers.runSyncInterval", "5s").Duration()
	if cfgInterval <= 0 {
		cfgInterval = 5 * time.Second
	}
	batchSize := g.Cfg().MustGet(context.Background(), "easymvp.workers.runSyncBatchSize", 20).Int()
	if batchSize <= 0 {
		batchSize = 20
	}
	return &runSyncWorker{
		interval:  cfgInterval,
		batchSize: batchSize,
	}
}

func (w *runSyncWorker) Name() string {
	return "run_sync_worker"
}

func (w *runSyncWorker) Interval() time.Duration {
	return w.interval
}

func (w *runSyncWorker) RunOnce(ctx context.Context) error {
	bindingIDs, err := listPendingRunBindingIDs(ctx, w.batchSize)
	if err != nil {
		return err
	}

	for _, bindingID := range bindingIDs {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if _, err = Runtime().SyncRunBindingCommand(ctx, bindingID); err != nil {
			handleWorkerFailure(
				ctx,
				w.Name(),
				"",
				"WORKER_RUN_SYNC",
				"run sync worker failed to sync binding",
				map[string]any{
					"binding_id": bindingID,
					"error":      err.Error(),
				},
			)
			continue
		}
	}

	if len(bindingIDs) > 0 {
		g.Log().Debugf(ctx, "[worker:%s] synced %d bindings", w.Name(), len(bindingIDs))
	}
	return nil
}
