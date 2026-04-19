package service

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"github.com/leef-l/easymvp/apps/core/api/system/v1"
)

type ISystem interface {
	Health(ctx context.Context) (res *v1.HealthRes, err error)
	ListProjectDiagnostics(ctx context.Context, projectID string, limit int) (res *v1.ListProjectDiagnosticsRes, err error)
}

var localSystem ISystem = (*sSystem)(nil)

type sSystem struct{}

func System() ISystem {
	if localSystem == nil {
		localSystem = &sSystem{}
	}
	return localSystem
}

func (s *sSystem) Health(ctx context.Context) (res *v1.HealthRes, err error) {
	startup := CurrentStartupConfig(ctx)
	workerStatus := Workers().Status()
	runtimeStatus := "ok"
	runtimeErr := Runtime().CheckHealth(ctx)
	if runtimeErr != nil {
		runtimeStatus = "degraded"
	}

	workerState := "stopped"
	switch {
	case startup.SafeMode:
		workerState = "disabled-safe-mode"
	case workerStatus.Started:
		workerState = "running"
	}

	startupSnapshot := buildStartupSnapshot(ctx, startup, workerStatus, runtimeErr)
	status := "ok"
	if startupSnapshot != nil && startupSnapshot.Status != "" {
		status = startupSnapshot.Status
	}

	return &v1.HealthRes{
		Status:            status,
		Service:           "easymvp-core",
		Version:           "v3-core",
		Timestamp:         gtime.Now().Format("Y-m-d H:i:s"),
		Mode:              StartupMode(ctx),
		SafeMode:          startup.SafeMode,
		WorkerStatus:      workerState,
		RegisteredWorkers: workerStatus.Workers,
		RuntimeStatus:     runtimeStatus,
		ServerAddress:     startup.ServerAddress,
		DataRoot:          startup.DataRoot,
		DBPath:            startup.DBPath,
		BrainServeBaseURL: startup.BrainServeBaseURL,
		Startup:           startupSnapshot,
	}, nil
}

func (s *sSystem) ListProjectDiagnostics(ctx context.Context, projectID string, limit int) (res *v1.ListProjectDiagnosticsRes, err error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	if limit <= 0 {
		limit = 20
	}
	return listProjectDiagnosticsView(ctx, projectID, limit)
}
