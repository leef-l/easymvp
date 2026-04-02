package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// TaskLog API

// TaskLogCreateReq 创建任务日志表请求
type TaskLogCreateReq struct {
	g.Meta `path:"/task_log/create" method:"post" tags:"任务日志表" summary:"创建任务日志表"`
	TaskID snowflake.JsonInt64 `json:"taskID" v:"required" dc:"任务ID"`
	Action string `json:"action" v:"required|max-length:50" dc:"动作"`
	FromStatus string `json:"fromStatus" v:"max-length:20" dc:"原状态"`
	ToStatus string `json:"toStatus" v:"max-length:20" dc:"新状态"`
	Message string `json:"message" v:"max-length:65535" dc:"日志内容"`
	Operator string `json:"operator" v:"max-length:50" dc:"操作者"`
}

// TaskLogCreateRes 创建任务日志表响应
type TaskLogCreateRes struct {
	g.Meta `mime:"application/json"`
}

// TaskLogUpdateReq 更新任务日志表请求
type TaskLogUpdateReq struct {
	g.Meta `path:"/task_log/update" method:"put" tags:"任务日志表" summary:"更新任务日志表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"任务日志表ID"`
	TaskID snowflake.JsonInt64 `json:"taskID" dc:"任务ID"`
	Action string `json:"action" dc:"动作"`
	FromStatus string `json:"fromStatus" dc:"原状态"`
	ToStatus string `json:"toStatus" dc:"新状态"`
	Message string `json:"message" dc:"日志内容"`
	Operator string `json:"operator" dc:"操作者"`
}

// TaskLogUpdateRes 更新任务日志表响应
type TaskLogUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// TaskLogDeleteReq 删除任务日志表请求
type TaskLogDeleteReq struct {
	g.Meta `path:"/task_log/delete" method:"delete" tags:"任务日志表" summary:"删除任务日志表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"任务日志表ID"`
}

// TaskLogDeleteRes 删除任务日志表响应
type TaskLogDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// TaskLogBatchDeleteReq 批量删除任务日志表请求
type TaskLogBatchDeleteReq struct {
	g.Meta `path:"/task_log/batch-delete" method:"delete" tags:"任务日志表" summary:"批量删除任务日志表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"任务日志表ID列表"`
}

// TaskLogBatchDeleteRes 批量删除任务日志表响应
type TaskLogBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// TaskLogDetailReq 获取任务日志表详情请求
type TaskLogDetailReq struct {
	g.Meta `path:"/task_log/detail" method:"get" tags:"任务日志表" summary:"获取任务日志表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"任务日志表ID"`
}

// TaskLogDetailRes 获取任务日志表详情响应
type TaskLogDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.TaskLogDetailOutput
}

// TaskLogListReq 获取任务日志表列表请求
type TaskLogListReq struct {
	g.Meta    `path:"/task_log/list" method:"get" tags:"任务日志表" summary:"获取任务日志表列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
}

// TaskLogListRes 获取任务日志表列表响应
type TaskLogListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.TaskLogListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// TaskLogExportReq 导出任务日志表请求
type TaskLogExportReq struct {
	g.Meta    `path:"/task_log/export" method:"get" tags:"任务日志表" summary:"导出任务日志表"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
}

// TaskLogExportRes 导出任务日志表响应
type TaskLogExportRes struct {
	g.Meta `mime:"text/csv"`
}



// TaskLogImportReq 导入任务日志表请求
type TaskLogImportReq struct {
	g.Meta `path:"/task_log/import" method:"post" mime:"multipart/form-data" tags:"任务日志表" summary:"导入任务日志表"`
}

// TaskLogImportRes 导入任务日志表响应
type TaskLogImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// TaskLogImportTemplateReq 下载任务日志表导入模板
type TaskLogImportTemplateReq struct {
	g.Meta `path:"/task_log/import-template" method:"get" tags:"任务日志表" summary:"下载任务日志表导入模板"`
}

// TaskLogImportTemplateRes 下载任务日志表导入模板响应
type TaskLogImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}

