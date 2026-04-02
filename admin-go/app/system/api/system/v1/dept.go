package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/system/internal/model"
	"easymvp/utility/snowflake"
)

// Dept API

// DeptCreateReq 创建部门表请求
type DeptCreateReq struct {
	g.Meta `path:"/dept/create" method:"post" tags:"部门表" summary:"创建部门表"`
	ParentID snowflake.JsonInt64 `json:"parentID"  dc:"上级部门ID，0 表示顶级部门"`
	Title string `json:"title" v:"required#部门名称不能为空" dc:"部门名称"`
	Username string `json:"username"  dc:"部门负责人姓名"`
	Email string `json:"email"  dc:"负责人邮箱"`
	Sort int `json:"sort"  dc:"排序（升序）"`
	Status int `json:"status"  dc:"状态"`
}

// DeptCreateRes 创建部门表响应
type DeptCreateRes struct {
	g.Meta `mime:"application/json"`
}

// DeptUpdateReq 更新部门表请求
type DeptUpdateReq struct {
	g.Meta `path:"/dept/update" method:"put" tags:"部门表" summary:"更新部门表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"部门表ID"`
	ParentID snowflake.JsonInt64 `json:"parentID" dc:"上级部门ID，0 表示顶级部门"`
	Title string `json:"title" dc:"部门名称"`
	Username string `json:"username" dc:"部门负责人姓名"`
	Email string `json:"email" dc:"负责人邮箱"`
	Sort int `json:"sort" dc:"排序（升序）"`
	Status int `json:"status" dc:"状态"`
}

// DeptUpdateRes 更新部门表响应
type DeptUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// DeptDeleteReq 删除部门表请求
type DeptDeleteReq struct {
	g.Meta `path:"/dept/delete" method:"delete" tags:"部门表" summary:"删除部门表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"部门表ID"`
}

// DeptDeleteRes 删除部门表响应
type DeptDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// DeptDetailReq 获取部门表详情请求
type DeptDetailReq struct {
	g.Meta `path:"/dept/detail" method:"get" tags:"部门表" summary:"获取部门表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"部门表ID"`
}

// DeptDetailRes 获取部门表详情响应
type DeptDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.DeptDetailOutput
}

// DeptListReq 获取部门表列表请求
type DeptListReq struct {
	g.Meta   `path:"/dept/list" method:"get" tags:"部门表" summary:"获取部门表列表"`
	PageNum  int `json:"pageNum" d:"1" dc:"页码"`
	PageSize int `json:"pageSize" d:"10" dc:"每页数量"`
	Status int `json:"status" dc:"状态"`
}

// DeptListRes 获取部门表列表响应
type DeptListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.DeptListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}

// DeptTreeReq 获取部门表树形结构请求
type DeptTreeReq struct {
	g.Meta `path:"/dept/tree" method:"get" tags:"部门表" summary:"获取部门表树形结构"`
	Status int `json:"status" dc:"状态"`
}

// DeptTreeRes 获取部门表树形结构响应
type DeptTreeRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.DeptTreeOutput `json:"list" dc:"树形数据"`
}

