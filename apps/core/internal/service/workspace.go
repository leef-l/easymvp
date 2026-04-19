package service

import (
	"context"
	"database/sql"

	projectsv1 "github.com/leef-l/easymvp/apps/core/api/projects/v1"
	workspacev1 "github.com/leef-l/easymvp/apps/core/api/workspace/v1"
)

type IWorkspace interface {
	GetHomeView(ctx context.Context) (res *workspacev1.HomeViewRes, err error)
	GetProjectWorkspaceView(ctx context.Context, projectID string) (res *projectsv1.ProjectWorkspaceViewRes, err error)
	ProjectEventsStream(ctx context.Context, req *workspacev1.ProjectEventsStreamReq) error
}

var localWorkspace IWorkspace = (*sWorkspace)(nil)

type sWorkspace struct{}

func Workspace() IWorkspace {
	if localWorkspace == nil {
		localWorkspace = &sWorkspace{}
	}
	return localWorkspace
}

func (s *sWorkspace) GetHomeView(ctx context.Context) (res *workspacev1.HomeViewRes, err error) {
	data, err := loadWorkspaceHomeData(ctx)
	if err != nil {
		snapshot, snapshotErr := loadWorkspaceHomeSnapshot(ctx, "home_view")
		if snapshotErr == nil {
			return snapshot, nil
		}
		if snapshotErr != nil && snapshotErr != sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	res = &workspacev1.HomeViewRes{
		Summary: workspacev1.HomeSummary{
			TotalProjects:   data.TotalProjects,
			ActiveProjects:  data.ActiveProjects,
			BlockedProjects: data.BlockedProjects,
			PendingActions:  data.PendingActions,
		},
		ActiveProjects:   data.ProjectCards,
		NeedAttention:    data.NeedAttention,
		RecentActivity:   data.RecentActivity,
		ReleaseReadiness: data.ReleaseReadiness,
	}
	_ = persistWorkspaceSnapshot(ctx, "home_view", res)
	return res, nil
}

func (s *sWorkspace) GetProjectWorkspaceView(ctx context.Context, projectID string) (res *projectsv1.ProjectWorkspaceViewRes, err error) {
	return Projects().GetProjectWorkspaceView(ctx, projectID)
}

func (s *sWorkspace) ProjectEventsStream(ctx context.Context, req *workspacev1.ProjectEventsStreamReq) error {
	return streamWorkspaceProjectEvents(ctx, req)
}
