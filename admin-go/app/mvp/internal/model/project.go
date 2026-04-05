package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Project DTO 模型

// ProjectCreateInput 创建MVP项目表输入
type ProjectCreateInput struct {
	Name string `json:"name"`
	ProjectCategory string `json:"projectCategory"`
	Description string `json:"description"`
	Status string `json:"status"`
	PauseReason string `json:"pauseReason"`
	GlobalContext string `json:"globalContext"`
	ArchitectModelID snowflake.JsonInt64 `json:"architectModelID"`
	WorkDir string `json:"workDir"`
}

// ProjectUpdateInput 更新MVP项目表输入（不包含 status，状态由工作流管理）
type ProjectUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	Name string `json:"name"`
	ProjectCategory string `json:"projectCategory"`
	Description string `json:"description"`
	PauseReason string `json:"pauseReason"`
	GlobalContext string `json:"globalContext"`
	ArchitectModelID snowflake.JsonInt64 `json:"architectModelID"`
	WorkDir string `json:"workDir"`
}

// ProjectDetailOutput MVP项目表详情输出
type ProjectDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	Name string `json:"name"`
	ProjectCategory string `json:"projectCategory"`
	Description string `json:"description"`
	Status string `json:"status"`
	PauseReason string `json:"pauseReason"`
	GlobalContext string `json:"globalContext"`
	ArchitectModelID snowflake.JsonInt64 `json:"architectModelID"`
	ArchitectModelName string `json:"architectModelName"`
	WorkDir string `json:"workDir"`
	EngineVersion string `json:"engineVersion"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ProjectListOutput MVP项目表列表输出
type ProjectListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	Name string `json:"name"`
	ProjectCategory string `json:"projectCategory"`
	Description string `json:"description"`
	Status string `json:"status"`
	PauseReason string `json:"pauseReason"`
	GlobalContext string `json:"globalContext"`
	ArchitectModelID snowflake.JsonInt64 `json:"architectModelID"`
	ArchitectModelName string `json:"architectModelName"`
	WorkDir string `json:"workDir"`
	EngineVersion string `json:"engineVersion"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ProjectListInput MVP项目表列表查询输入
type ProjectListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Name string `json:"name"`
}


// ProjectBatchUpdateInput 批量编辑MVP项目表输入
type ProjectBatchUpdateInput struct {
	IDs    []snowflake.JsonInt64 `json:"ids"`
	Status *string               `json:"status"`
}

