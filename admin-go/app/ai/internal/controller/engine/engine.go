package engine

import (
	"context"
	"encoding/json"

	v1 "easymvp/app/ai/api/ai/v1"
	"easymvp/app/ai/internal/model"
	"easymvp/app/ai/internal/service"
)

var Engine = cEngine{}

type cEngine struct{}

func (c *cEngine) List(ctx context.Context, req *v1.EngineListReq) (res *v1.EngineListRes, err error) {
	res = &v1.EngineListRes{}
	res.List, err = service.Engine().List(ctx)
	return
}

func (c *cEngine) Detail(ctx context.Context, req *v1.EngineDetailReq) (res *v1.EngineDetailRes, err error) {
	res = &v1.EngineDetailRes{}
	res.EngineDetailOutput, err = service.Engine().Detail(ctx, req.EngineCode)
	return
}

func (c *cEngine) Update(ctx context.Context, req *v1.EngineUpdateReq) (res *v1.EngineUpdateRes, err error) {
	extraConfig := ""
	if len(req.ExtraConfig) > 0 {
		if bytes, marshalErr := json.Marshal(req.ExtraConfig); marshalErr == nil {
			extraConfig = string(bytes)
		}
	}

	err = service.Engine().Update(ctx, &model.EngineUpdateInput{
		EngineCode:      req.EngineCode,
		DefaultModelID:  req.DefaultModelID,
		TimeoutSeconds:  req.TimeoutSeconds,
		MaxSteps:        req.MaxSteps,
		WorkspaceRoot:   req.WorkspaceRoot,
		CommandTemplate: req.CommandTemplate,
		CallbackURL:     req.CallbackURL,
		CallbackSecret:  req.CallbackSecret,
		ExtraConfig:     extraConfig,
		Status:          req.Status,
	})
	return
}

func (c *cEngine) TestConnection(ctx context.Context, req *v1.EngineTestConnectionReq) (res *v1.EngineTestConnectionRes, err error) {
	res = &v1.EngineTestConnectionRes{}
	out, err := service.Engine().TestConnection(ctx, req.EngineCode)
	if err != nil {
		return nil, err
	}
	res.Success = out.Success
	res.Message = out.Message
	return
}
