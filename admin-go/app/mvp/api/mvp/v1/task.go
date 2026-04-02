package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// Task API

// TaskCreateReq 创建MVP任务表请求
type TaskCreateReq struct {
	g.Meta `path:"/task/create" method:"post" tags:"MVP任务表" summary:"创建MVP任务表"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	ParentID snowflake.JsonInt64 `json:"parentID"  dc:"父任务ID，0=顶级"`
	Name string `json:"name" v:"required|max-length:500" dc:"任务名称"`
	Description string `json:"description" v:"max-length:65535" dc:"任务描述"`
	RoleType string `json:"roleType" v:"required|max-length:20" dc:"角色类型"`
	RoleLevel string `json:"roleLevel" v:"max-length:10" dc:"角色等级"`
	ModelID snowflake.JsonInt64 `json:"modelID"  dc:"使用的AI模型ID"`
	Status string `json:"status" v:"max-length:20" dc:"状态"`
	Sort int `json:"sort"  dc:"排序"`
	BatchNo int `json:"batchNo"  dc:"执行批次号，同批次内可并行，批次间串行"`
	AffectedResources string `json:"affectedResources"  dc:"涉及的资源范围（文件/模块），用于并发冲突检测"`
	DependsOn string `json:"dependsOn"  dc:"依赖的任务ID列表"`
	Result string `json:"result" v:"max-length:4294967295" dc:"任务执行结果"`
	ContextSummary string `json:"contextSummary" v:"max-length:65535" dc:"任务完成后的上下文压缩摘要，供后续AI读取"`
	ErrorMessage string `json:"errorMessage" v:"max-length:65535" dc:"错误信息"`
	StartedAt *gtime.Time `json:"startedAt"  dc:"开始时间"`
	CompletedAt *gtime.Time `json:"completedAt"  dc:"完成时间"`
}

// TaskCreateRes 创建MVP任务表响应
type TaskCreateRes struct {
	g.Meta `mime:"application/json"`
}

// TaskUpdateReq 更新MVP任务表请求
type TaskUpdateReq struct {
	g.Meta `path:"/task/update" method:"put" tags:"MVP任务表" summary:"更新MVP任务表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP任务表ID"`
	ProjectID snowflake.JsonInt64 `json:"projectID" dc:"项目ID"`
	ParentID snowflake.JsonInt64 `json:"parentID" dc:"父任务ID，0=顶级"`
	Name string `json:"name" dc:"任务名称"`
	Description string `json:"description" dc:"任务描述"`
	RoleType string `json:"roleType" dc:"角色类型"`
	RoleLevel string `json:"roleLevel" dc:"角色等级"`
	ModelID snowflake.JsonInt64 `json:"modelID" dc:"使用的AI模型ID"`
	Status string `json:"status" dc:"状态"`
	Sort int `json:"sort" dc:"排序"`
	BatchNo int `json:"batchNo" dc:"执行批次号，同批次内可并行，批次间串行"`
	AffectedResources string `json:"affectedResources" dc:"涉及的资源范围（文件/模块），用于并发冲突检测"`
	DependsOn string `json:"dependsOn" dc:"依赖的任务ID列表"`
	Result string `json:"result" dc:"任务执行结果"`
	ContextSummary string `json:"contextSummary" dc:"任务完成后的上下文压缩摘要，供后续AI读取"`
	ErrorMessage string `json:"errorMessage" dc:"错误信息"`
	StartedAt *gtime.Time `json:"startedAt" dc:"开始时间"`
	CompletedAt *gtime.Time `json:"completedAt" dc:"完成时间"`
}

// TaskUpdateRes 更新MVP任务表响应
type TaskUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// TaskDeleteReq 删除MVP任务表请求
type TaskDeleteReq struct {
	g.Meta `path:"/task/delete" method:"delete" tags:"MVP任务表" summary:"删除MVP任务表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP任务表ID"`
}

// TaskDeleteRes 删除MVP任务表响应
type TaskDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// TaskBatchDeleteReq 批量删除MVP任务表请求
type TaskBatchDeleteReq struct {
	g.Meta `path:"/task/batch-delete" method:"delete" tags:"MVP任务表" summary:"批量删除MVP任务表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"MVP任务表ID列表"`
}

// TaskBatchDeleteRes 批量删除MVP任务表响应
type TaskBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// TaskBatchUpdateReq 批量编辑MVP任务表请求
type TaskBatchUpdateReq struct {
	g.Meta `path:"/task/batch-update" method:"put" tags:"MVP任务表" summary:"批量编辑MVP任务表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"MVP任务表ID列表"`
	Status *int                  `json:"status" dc:"状态"`
}

// TaskBatchUpdateRes 批量编辑MVP任务表响应
type TaskBatchUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// TaskDetailReq 获取MVP任务表详情请求
type TaskDetailReq struct {
	g.Meta `path:"/task/detail" method:"get" tags:"MVP任务表" summary:"获取MVP任务表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP任务表ID"`
}

// TaskDetailRes 获取MVP任务表详情响应
type TaskDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.TaskDetailOutput
}

// TaskListReq 获取MVP任务表列表请求
type TaskListReq struct {
	g.Meta    `path:"/task/list" method:"get" tags:"MVP任务表" summary:"获取MVP任务表列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Name string `json:"name" dc:"任务名称"`
}

// TaskListRes 获取MVP任务表列表响应
type TaskListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.TaskListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// TaskExportReq 导出MVP任务表请求
type TaskExportReq struct {
	g.Meta    `path:"/task/export" method:"get" tags:"MVP任务表" summary:"导出MVP任务表"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Name string `json:"name" dc:"任务名称"`
}

// TaskExportRes 导出MVP任务表响应
type TaskExportRes struct {
	g.Meta `mime:"text/csv"`
}


// TaskTreeReq 获取MVP任务表树形结构请求
type TaskTreeReq struct {
	g.Meta    `path:"/task/tree" method:"get" tags:"MVP任务表" summary:"获取MVP任务表树形结构"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Name      string `json:"name" dc:"任务名称"`
	ProjectID int64  `json:"projectID" dc:"项目ID"`
	Status    string `json:"status" dc:"任务状态"`
	BatchNo   int    `json:"batchNo" dc:"执行批次"`
	RoleType  string `json:"roleType" dc:"角色类型"`
}

// TaskTreeRes 获取MVP任务表树形结构响应
type TaskTreeRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.TaskTreeOutput `json:"list" dc:"树形数据"`
}


