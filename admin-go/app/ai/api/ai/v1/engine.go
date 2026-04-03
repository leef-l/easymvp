package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/ai/internal/model"
	"easymvp/utility/snowflake"
)

// EngineListReq 获取执行引擎列表
type EngineListReq struct {
	g.Meta `path:"/engine/list" method:"get" tags:"AI执行引擎" summary:"获取AI执行引擎列表"`
}

type EngineListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.EngineListOutput `json:"list"`
}

// EngineDetailReq 获取执行引擎详情
type EngineDetailReq struct {
	g.Meta     `path:"/engine/detail" method:"get" tags:"AI执行引擎" summary:"获取AI执行引擎详情"`
	EngineCode string `json:"engineCode" v:"required#引擎编码不能为空"`
}

type EngineDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.EngineDetailOutput
}

// EngineUpdateReq 更新执行引擎配置
type EngineUpdateReq struct {
	g.Meta          `path:"/engine/update" method:"post" tags:"AI执行引擎" summary:"更新AI执行引擎配置"`
	EngineCode      string                 `json:"engineCode" v:"required#引擎编码不能为空"`
	DefaultModelID  snowflake.JsonInt64    `json:"defaultModelID"`
	TimeoutSeconds  int                    `json:"timeoutSeconds"`
	MaxSteps        int                    `json:"maxSteps"`
	WorkspaceRoot   string                 `json:"workspaceRoot"`
	CommandTemplate string                 `json:"commandTemplate"`
	CallbackURL     string                 `json:"callbackURL"`
	CallbackSecret  string                 `json:"callbackSecret"`
	ExtraConfig     map[string]interface{} `json:"extraConfig"`
	Status          int                    `json:"status"`
}

type EngineUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// EngineTestConnectionReq 测试引擎连接
type EngineTestConnectionReq struct {
	g.Meta     `path:"/engine/test-connection" method:"post" tags:"AI执行引擎" summary:"测试AI执行引擎连接"`
	EngineCode string `json:"engineCode" v:"required#引擎编码不能为空"`
}

type EngineTestConnectionRes struct {
	g.Meta  `mime:"application/json"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}
