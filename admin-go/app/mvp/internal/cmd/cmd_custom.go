package cmd

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"easymvp/app/mvp/internal/controller/chat"
)

// registerCustomRoutes 注册手写的自定义路由（不会被 codegen 覆盖）
func registerCustomRoutes(group *ghttp.RouterGroup) {
	group.Bind(
		chat.Chat,
		chat.Workflow,
	)
}
