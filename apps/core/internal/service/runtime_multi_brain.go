package service

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// MultiBrainRunRequest submits a single prompt to multiple specialist brains in
// parallel. This is the EasyMVP-side implementation of "multi-brain delegation"
// that drives the acceptance verification pipeline (browser + verifier + code).
type MultiBrainRunRequest struct {
	ProjectID string   `json:"project_id"`
	TaskID    string   `json:"task_id"`
	Prompt    string   `json:"prompt"`
	Workdir   string   `json:"workdir,omitempty"`
	MaxTurns  int      `json:"max_turns"`
	Provider  string   `json:"provider,omitempty"`
	Brains    []string `json:"brains"` // e.g. ["browser", "verifier", "code"]
}

// MultiBrainRunResult aggregates the parallel execution results.
type MultiBrainRunResult struct {
	Results map[string]*runtimeCreateRunResponse `json:"results"` // brain kind -> response
	Errors  map[string]error                     `json:"errors"`  // brain kind -> error
}

// ExecuteMultiBrain launches parallel brain runs for all requested specialist
// brains and waits for completion. It is non-blocking per brain: each brain
// gets its own HTTP call to /v1/runs.
func ExecuteMultiBrain(ctx context.Context, client *http.Client, baseURL string, req *MultiBrainRunRequest) *MultiBrainRunResult {
	if req == nil || len(req.Brains) == 0 {
		return &MultiBrainRunResult{Results: make(map[string]*runtimeCreateRunResponse), Errors: make(map[string]error)}
	}

	result := &MultiBrainRunResult{
		Results: make(map[string]*runtimeCreateRunResponse, len(req.Brains)),
		Errors:  make(map[string]error, len(req.Brains)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, brainKind := range req.Brains {
		wg.Add(1)
		go func(kind string) {
			defer wg.Done()

			normalized := &normalizedStartBrainRunCommand{
				ProjectID: req.ProjectID,
				TaskID:    req.TaskID,
				BrainKind: kind,
				Prompt:    req.Prompt,
				Workdir:   req.Workdir,
				MaxTurns:  req.MaxTurns,
				Provider:  req.Provider,
			}
			if normalized.MaxTurns <= 0 {
				normalized.MaxTurns = 20
			}

			// Apply a per-brain timeout that is slightly shorter than the
			// parent context so slow brains don't starve fast ones.
			brainCtx, cancel := context.WithTimeout(ctx, 180*time.Second)
			defer cancel()

			res, err := runtimeCreateRun(brainCtx, client, baseURL, normalized)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				result.Errors[kind] = err
				g.Log().Warningf(ctx, "[multi_brain] run failed for brain=%s project=%s task=%s: %v",
					kind, req.ProjectID, req.TaskID, err)
			} else {
				result.Results[kind] = res
				g.Log().Infof(ctx, "[multi_brain] run started for brain=%s project=%s task=%s run_id=%s",
					kind, req.ProjectID, req.TaskID, res.RunID)
			}
		}(brainKind)
	}

	wg.Wait()
	return result
}

// IsAllSucceeded returns true if every requested brain succeeded.
func (r *MultiBrainRunResult) IsAllSucceeded() bool {
	if r == nil || len(r.Errors) > 0 {
		return false
	}
	return len(r.Results) > 0
}

// CollectRunIDs returns a map of brain kind -> run_id for succeeded brains.
func (r *MultiBrainRunResult) CollectRunIDs() map[string]string {
	if r == nil {
		return nil
	}
	out := make(map[string]string, len(r.Results))
	for kind, res := range r.Results {
		if res != nil && res.RunID != "" {
			out[kind] = res.RunID
		}
	}
	return out
}
