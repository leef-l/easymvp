// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// AiProvider is the golang structure for table ai_provider.
type AiProvider struct {
	Id           uint64      `orm:"id"            description:"雪花ID"`                                                                                  // 雪花ID
	Name         string      `orm:"name"          description:"供应商名称"`                                                                                 // 供应商名称
	Code         string      `orm:"code"          description:"供应商代码：openai/anthropic/deepseek/qwen/doubao/ernie/spark/glm/moonshot/yi/google/ollama"` // 供应商代码：openai/anthropic/deepseek/qwen/doubao/ernie/spark/glm/moonshot/yi/google/ollama
	ProviderType string      `orm:"provider_type" description:"Provider类型：openai_compatible/anthropic/baidu/xfyun/google"`                             // Provider类型：openai_compatible/anthropic/baidu/xfyun/google
	BaseUrl      string      `orm:"base_url"      description:"API基础地址"`                                                                               // API基础地址
	Icon         string      `orm:"icon"          description:"图标URL"`                                                                                 // 图标URL
	Status       int         `orm:"status"        description:"状态:0=禁用,1=启用"`                                                                          // 状态:0=禁用,1=启用
	Sort         int         `orm:"sort"          description:"排序"`                                                                                    // 排序
	CreatedBy    uint64      `orm:"created_by"    description:"创建人ID"`                                                                                 // 创建人ID
	DeptId       uint64      `orm:"dept_id"       description:"所属部门ID"`                                                                                // 所属部门ID
	CreatedAt    *gtime.Time `orm:"created_at"    description:"创建时间"`                                                                                  // 创建时间
	UpdatedAt    *gtime.Time `orm:"updated_at"    description:"更新时间"`                                                                                  // 更新时间
	DeletedAt    *gtime.Time `orm:"deleted_at"    description:"软删除时间"`                                                                                 // 软删除时间
}
