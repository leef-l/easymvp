package requirements

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/requirements/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Get(ctx context.Context, req *v1.GetRequirementReq) (res *v1.GetRequirementRes, err error) {
	entity, err := service.Requirement().GetRequirement(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &v1.GetRequirementRes{
		Requirement: &v1.RequirementDetail{
			ID:                 entity.Id,
			ProjectID:          entity.ProjectId,
			RawInput:           entity.RawInput,
			Status:             entity.Status,
			RequirementDocJSON: entity.RequirementDocJson,
			UserConfirmed:      entity.UserConfirmed,
			ConfirmedAt:        entity.ConfirmedAt,
			CreatedAt:          entity.CreatedAt,
			UpdatedAt:          entity.UpdatedAt,
		},
	}, nil
}

func (c *ControllerV1) GetByProject(ctx context.Context, req *v1.GetProjectRequirementReq) (res *v1.GetProjectRequirementRes, err error) {
	entity, err := service.Requirement().GetProjectRequirement(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}
	return &v1.GetProjectRequirementRes{
		Requirement: &v1.RequirementDetail{
			ID:                 entity.Id,
			ProjectID:          entity.ProjectId,
			RawInput:           entity.RawInput,
			Status:             entity.Status,
			RequirementDocJSON: entity.RequirementDocJson,
			UserConfirmed:      entity.UserConfirmed,
			ConfirmedAt:        entity.ConfirmedAt,
			CreatedAt:          entity.CreatedAt,
			UpdatedAt:          entity.UpdatedAt,
		},
	}, nil
}
