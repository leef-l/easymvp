package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

type EngineListOutput struct {
	ID             snowflake.JsonInt64 `json:"id"`
	Code           string              `json:"code"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Status         int                 `json:"status"`
	ConfigStatus   int                 `json:"configStatus"`
	DefaultModelID snowflake.JsonInt64 `json:"defaultModelID"`
	CreatedAt      *gtime.Time         `json:"createdAt"`
	UpdatedAt      *gtime.Time         `json:"updatedAt"`
}

type EngineDetailOutput struct {
	ID              snowflake.JsonInt64 `json:"id"`
	EngineCode      string              `json:"engineCode"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	BaseURL         string              `json:"baseURL"`
	APIKeyMasked    string              `json:"apiKeyMasked"`
	DefaultModelID  snowflake.JsonInt64 `json:"defaultModelID"`
	TimeoutSeconds  int                 `json:"timeoutSeconds"`
	MaxSteps        int                 `json:"maxSteps"`
	WorkspaceRoot   string              `json:"workspaceRoot"`
	CommandTemplate string              `json:"commandTemplate"`
	CallbackURL     string              `json:"callbackURL"`
	CallbackSecret  string              `json:"callbackSecret"`
	ExtraConfig     string              `json:"extraConfig"`
	Status          int                 `json:"status"`
	ConfigStatus    int                 `json:"configStatus"`
}

type EngineUpdateInput struct {
	EngineCode      string
	DefaultModelID  snowflake.JsonInt64
	TimeoutSeconds  int
	MaxSteps        int
	WorkspaceRoot   string
	CommandTemplate string
	CallbackURL     string
	CallbackSecret  string
	ExtraConfig     string
	Status          int
}

type EngineTestOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
