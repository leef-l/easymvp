package main

import (
	_ "easymvp/app/app/ai/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"easymvp/app/app/ai/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
