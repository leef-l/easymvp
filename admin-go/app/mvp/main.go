package main

import (
	_ "easymvp/app/mvp/internal/packed"

	_ "easymvp/app/mvp/internal/logic/config"
	_ "easymvp/app/mvp/internal/logic/conversation"
	_ "easymvp/app/mvp/internal/logic/message"
	_ "easymvp/app/mvp/internal/logic/project"
	_ "easymvp/app/mvp/internal/logic/project_role"
	_ "easymvp/app/mvp/internal/logic/role_preset"
	_ "easymvp/app/mvp/internal/logic/task"
	_ "easymvp/app/mvp/internal/logic/task_log"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"easymvp/app/mvp/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
