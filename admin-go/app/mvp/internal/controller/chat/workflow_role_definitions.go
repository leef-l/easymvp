package chat

import (
	"context"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/workflow/configstore"
	"easymvp/app/mvp/internal/workflow/rolecatalog"
)

func (c *cWorkflow) RoleDefinitions(ctx context.Context, req *v1.WorkflowRoleDefinitionsReq) (res *v1.WorkflowRoleDefinitionsRes, err error) {
	definitions := rolecatalog.DefaultService().List(ctx)
	list := make([]v1.RoleDefinitionItem, 0, len(definitions))
	for _, definition := range definitions {
		list = append(list, v1.RoleDefinitionItem{
			RoleType:            definition.RoleType,
			DisplayName:         definition.DisplayName,
			Color:               definition.Color,
			Description:         definition.Description,
			PreferredLevels:     definition.PreferredLevels,
			DefaultSystemPrompt: definition.DefaultSystemPrompt,
			AcceptanceJudge:     definition.AcceptanceJudge,
			Sort:                definition.Sort,
		})
	}
	return &v1.WorkflowRoleDefinitionsRes{List: list}, nil
}

func (c *cWorkflow) SaveRoleDefinitions(ctx context.Context, req *v1.WorkflowSaveRoleDefinitionsReq) (res *v1.WorkflowSaveRoleDefinitionsRes, err error) {
	definitions := make([]rolecatalog.Definition, 0, len(req.List))
	for _, item := range req.List {
		definitions = append(definitions, rolecatalog.Definition{
			RoleType:            item.RoleType,
			DisplayName:         item.DisplayName,
			Color:               item.Color,
			Description:         item.Description,
			PreferredLevels:     item.PreferredLevels,
			DefaultSystemPrompt: item.DefaultSystemPrompt,
			AcceptanceJudge:     item.AcceptanceJudge,
			Sort:                item.Sort,
		})
	}
	svc := rolecatalog.NewService(configstore.NewStore())
	if err := svc.Save(ctx, definitions, middleware.GetUserID(ctx), middleware.GetDeptID(ctx)); err != nil {
		return nil, err
	}
	return &v1.WorkflowSaveRoleDefinitionsRes{Message: "角色定义配置已保存"}, nil
}

func (c *cWorkflow) ResetRoleDefinitions(ctx context.Context, req *v1.WorkflowResetRoleDefinitionsReq) (res *v1.WorkflowResetRoleDefinitionsRes, err error) {
	svc := rolecatalog.NewService(configstore.NewStore())
	if err := svc.Reset(ctx); err != nil {
		return nil, err
	}
	return &v1.WorkflowResetRoleDefinitionsRes{Message: "角色定义配置已重置为系统默认"}, nil
}
