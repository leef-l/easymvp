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
	_ = ctx

	return &v1.HealthRes{
		Status:    "ok",
		Service:   "easymvp-core",
		Version:   "v3-bootstrap",
		Timestamp: gtime.Now().Format("Y-m-d H:i:s"),
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
