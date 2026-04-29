package service

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// LoadGuardStatus reflects the current server-load protection state.
// Engineering Cybernetics ch.15 / 铁律 5: server load protection.
type LoadGuardStatus string

const (
	LoadGuardHealthy   LoadGuardStatus = "healthy"   // CPU < resumeThreshold
	LoadGuardThrottled LoadGuardStatus = "throttled" // resumeThreshold <= CPU < stopThreshold
	LoadGuardStopped   LoadGuardStatus = "stopped"   // CPU >= stopThreshold
)

// SystemLoadGuard implements 铁律 5: server load protection.
// - Stop when CPU busy >= 80
// - Resume only when CPU busy < 50
// - Re-check before every recovery attempt.
type SystemLoadGuard struct {
	mu sync.RWMutex

	stopThreshold   float64 // default 80
	resumeThreshold float64 // default 50

	lastCheckTime  time.Time
	lastCPUPercent float64
	status         LoadGuardStatus
}

func NewSystemLoadGuard() *SystemLoadGuard {
	return &SystemLoadGuard{
		stopThreshold:   80.0,
		resumeThreshold: 50.0,
		status:          LoadGuardHealthy,
	}
}

// Check queries the current system load and updates the guard status.
// It is safe for concurrent use.
func (guard *SystemLoadGuard) Check(ctx context.Context) LoadGuardStatus {
	cpuPercent, err := guard.queryCPUPercent(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "load guard CPU check failed: %v; assuming healthy to avoid deadlock", err)
		return LoadGuardHealthy
	}

	guard.mu.Lock()
	defer guard.mu.Unlock()

	guard.lastCheckTime = time.Now()
	guard.lastCPUPercent = cpuPercent

	switch guard.status {
	case LoadGuardHealthy, LoadGuardThrottled:
		if cpuPercent >= guard.stopThreshold {
			guard.status = LoadGuardStopped
			g.Log().Warningf(ctx, "load guard STOPPED: CPU %.1f%% >= stopThreshold %.1f%%", cpuPercent, guard.stopThreshold)
		} else if cpuPercent >= guard.resumeThreshold {
			guard.status = LoadGuardThrottled
		} else {
			guard.status = LoadGuardHealthy
		}
	case LoadGuardStopped:
		if cpuPercent < guard.resumeThreshold {
			guard.status = LoadGuardHealthy
			g.Log().Infof(ctx, "load guard RESUMED: CPU %.1f%% < resumeThreshold %.1f%%", cpuPercent, guard.resumeThreshold)
		}
	}

	return guard.status
}

// AllowRun returns true if the guard permits a new resource-intensive run.
func (guard *SystemLoadGuard) AllowRun() bool {
	guard.mu.RLock()
	defer guard.mu.RUnlock()
	return guard.status != LoadGuardStopped
}

// Status returns the current guard status without triggering a new check.
func (guard *SystemLoadGuard) Status() LoadGuardStatus {
	guard.mu.RLock()
	defer guard.mu.RUnlock()
	return guard.status
}

// LastCPUPercent returns the most recently sampled CPU percentage.
func (guard *SystemLoadGuard) LastCPUPercent() float64 {
	guard.mu.RLock()
	defer guard.mu.RUnlock()
	return guard.lastCPUPercent
}

// LastCheckTime returns the timestamp of the most recent check.
func (guard *SystemLoadGuard) LastCheckTime() time.Time {
	guard.mu.RLock()
	defer guard.mu.RUnlock()
	return guard.lastCheckTime
}

func (guard *SystemLoadGuard) queryCPUPercent(ctx context.Context) (float64, error) {
	if runtime.GOOS == "windows" {
		return guard.queryWindowsCPU(ctx)
	}
	return guard.queryUnixCPU(ctx)
}

func (guard *SystemLoadGuard) queryWindowsCPU(ctx context.Context) (float64, error) {
	cmd := exec.CommandContext(ctx, "wmic", "cpu", "get", "loadpercentage", "/value")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		const prefix = "LoadPercentage="
		if strings.HasPrefix(line, prefix) {
			val := strings.TrimSpace(line[len(prefix):])
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				return v, nil
			}
		}
	}
	return 0, fmt.Errorf("could not parse CPU load from wmic output: %s", string(out))
}

func (guard *SystemLoadGuard) queryUnixCPU(ctx context.Context) (float64, error) {
	// top gives instantaneous usage; awk extracts the user+system percentage.
	cmd := exec.CommandContext(ctx, "sh", "-c",
		"top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}


