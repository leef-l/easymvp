package service

import (
	"context"
	"testing"

	projectsv1 "github.com/leef-l/easymvp/apps/core/api/projects/v1"
)

type workspaceTestProjectsStub struct {
	lastProjectID string
	result        *projectsv1.ProjectWorkspaceViewRes
}

func (s *workspaceTestProjectsStub) CreateProject(ctx context.Context, req CreateProjectCommand) (*projectsv1.CreateProjectRes, error) {
	_ = ctx
	_ = req
	return nil, nil
}

func (s *workspaceTestProjectsStub) GetProjectWorkspaceView(ctx context.Context, projectID string) (*projectsv1.ProjectWorkspaceViewRes, error) {
	_ = ctx
	s.lastProjectID = projectID
	return s.result, nil
}

func TestWorkspaceGetProjectWorkspaceViewDelegatesToProjects(t *testing.T) {
	t.Parallel()

	original := localProjects
	defer func() {
		localProjects = original
	}()

	expected := &projectsv1.ProjectWorkspaceViewRes{
		ProjectSnapshot: projectsv1.ProjectSnapshot{
			ProjectID: "proj_workspace_1",
			Name:      "Workspace Delegate",
		},
	}
	stub := &workspaceTestProjectsStub{result: expected}
	localProjects = stub

	got, err := Workspace().GetProjectWorkspaceView(context.Background(), "proj_workspace_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.lastProjectID != "proj_workspace_1" {
		t.Fatalf("unexpected delegated project id: got %s want %s", stub.lastProjectID, "proj_workspace_1")
	}
	if got != expected {
		t.Fatalf("unexpected workspace view pointer: got %#v want %#v", got, expected)
	}
}
