package main

import (
	_ "easymvp/app/ai/internal/packed"

	_ "easymvp/app/ai/internal/logic/engine"
	_ "easymvp/app/ai/internal/logic/model"
	_ "easymvp/app/ai/internal/logic/plan"
	_ "easymvp/app/ai/internal/logic/provider"
	_ "easymvp/app/ai/internal/logic/task"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"easymvp/app/ai/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
