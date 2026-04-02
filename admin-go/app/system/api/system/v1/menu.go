package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/system/internal/model"
	"easymvp/utility/snowflake"
)

// Menu API

// MenuCreateReq 创建菜单表请求
type MenuCreateReq struct {
	g.Meta `path:"/menu/create" method:"post" tags:"菜单表" summary:"创建菜单表"`
	ParentID snowflake.JsonInt64 `json:"parentID"  dc:"上级菜单ID，0 表示顶级菜单"`
	Title string `json:"title" v:"required#菜单名称不能为空" dc:"菜单名称"`
	Type int `json:"type"  dc:"类型"`
	Path string `json:"path"  dc:"前端路由路径"`
	Component string `json:"component"  dc:"前端组件路径"`
	Permission string `json:"permission"  dc:"权限标识（如 system"`
	Icon string `json:"icon"  dc:"菜单图标（图标名称）"`
	Sort int `json:"sort"  dc:"排序（升序）"`
	IsShow int `json:"isShow"  dc:"是否显示"`
	IsCache int `json:"isCache"  dc:"是否缓存"`
	LinkURL string `json:"linkURL"  dc:"外链/内链地址（type=4或5时有效）"`
	Status int `json:"status"  dc:"状态"`
}

// MenuCreateRes 创建菜单表响应
type MenuCreateRes struct {
	g.Meta `mime:"application/json"`
}

// MenuUpdateReq 更新菜单表请求
type MenuUpdateReq struct {
	g.Meta `path:"/menu/update" method:"put" tags:"菜单表" summary:"更新菜单表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"菜单表ID"`
	ParentID snowflake.JsonInt64 `json:"parentID" dc:"上级菜单ID，0 表示顶级菜单"`
	Title string `json:"title" dc:"菜单名称"`
	Type int `json:"type" dc:"类型"`
	Path string `json:"path" dc:"前端路由路径"`
	Component string `json:"component" dc:"前端组件路径"`
	Permission string `json:"permission" dc:"权限标识（如 system"`
	Icon string `json:"icon" dc:"菜单图标（图标名称）"`
	Sort int `json:"sort" dc:"排序（升序）"`
	IsShow int `json:"isShow" dc:"是否显示"`
	IsCache int `json:"isCache" dc:"是否缓存"`
	LinkURL string `json:"linkURL" dc:"外链/内链地址（type=4或5时有效）"`
	Status int `json:"status" dc:"状态"`
}

// MenuUpdateRes 更新菜单表响应
type MenuUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// MenuDeleteReq 删除菜单表请求
type MenuDeleteReq struct {
	g.Meta `path:"/menu/delete" method:"delete" tags:"菜单表" summary:"删除菜单表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"菜单表ID"`
}

// MenuDeleteRes 删除菜单表响应
type MenuDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// MenuDetailReq 获取菜单表详情请求
type MenuDetailReq struct {
	g.Meta `path:"/menu/detail" method:"get" tags:"菜单表" summary:"获取菜单表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"菜单表ID"`
}

// MenuDetailRes 获取菜单表详情响应
type MenuDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.MenuDetailOutput
}

// MenuListReq 获取菜单表列表请求
type MenuListReq struct {
	g.Meta   `path:"/menu/list" method:"get" tags:"菜单表" summary:"获取菜单表列表"`
	PageNum  int `json:"pageNum" d:"1" dc:"页码"`
	PageSize int `json:"pageSize" d:"10" dc:"每页数量"`
	Type int `json:"type" dc:"类型"`
	IsShow int `json:"isShow" dc:"是否显示"`
	IsCache int `json:"isCache" dc:"是否缓存"`
	Status int `json:"status" dc:"状态"`
}

// MenuListRes 获取菜单表列表响应
type MenuListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.MenuListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}

// MenuTreeReq 获取菜单表树形结构请求
type MenuTreeReq struct {
	g.Meta `path:"/menu/tree" method:"get" tags:"菜单表" summary:"获取菜单表树形结构"`
	Type int `json:"type" dc:"类型"`
	IsShow int `json:"isShow" dc:"是否显示"`
	IsCache int `json:"isCache" dc:"是否缓存"`
	Status int `json:"status" dc:"状态"`
}

// MenuTreeRes 获取菜单表树形结构响应
type MenuTreeRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.MenuTreeOutput `json:"list" dc:"树形数据"`
}

