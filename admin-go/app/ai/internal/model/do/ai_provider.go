// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// AiProvider is the golang structure of table ai_provider for DAO operations like Where/Data.
type AiProvider struct {
	g.Meta             `orm:"table:ai_provider, do:true"`
	Id                 any         // 雪花ID
	Name               any         // 供应商名称
	Code               any         // 供应商代码：openai/anthropic/deepseek/qwen/doubao/ernie/spark/glm/moonshot/yi/google/ollama
	ProviderType       any         // 供应商主类型/默认路由类型
	SupportedProtocols any         // 支持的协议类型(JSON)：anthropic/openai_compatible/google 等
	BaseUrl            any         // API基础地址
	Icon               any         // 图标URL
	Status             any         // 状态:0=禁用,1=启用
	Sort               any         // 排序
	CreatedBy          any         // 创建人ID
	DeptId             any         // 所属部门ID
	CreatedAt          *gtime.Time // 创建时间
	UpdatedAt          *gtime.Time // 更新时间
	DeletedAt          *gtime.Time // 软删除时间
}
