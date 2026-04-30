package retrospectives

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/retrospectives/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Generate(ctx context.Context, req *v1.GenerateRetrospectiveReq) (res *v1.GenerateRetrospectiveRes, err error) {
	result, err := service.Retrospective().GenerateRetrospective(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}
	return &v1.GenerateRetrospectiveRes{
		RetrospectiveID: result.RetrospectiveID,
		TotalTasks:      result.TotalTasks,
		CompletedTasks:  result.CompletedTasks,
		FailedTasks:     result.FailedTasks,
	}, nil
}
