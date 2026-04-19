package runtime

import (
	"context"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) StartRun(ctx context.Context, req *runtimev1.StartRunReq) (res *runtimev1.StartRunRes, err error) {
	return service.Runtime().StartRunCommand(ctx, service.StartBrainRunCommand{
		ProjectID: req.ProjectID,
		TaskID:    req.TaskID,
		BrainKind: req.BrainKind,
		Prompt:    req.Prompt,
		Workdir:   req.Workdir,
		MaxTurns:  req.MaxTurns,
		Provider:  req.Provider,
	})
}
