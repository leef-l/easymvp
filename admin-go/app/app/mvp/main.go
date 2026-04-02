package main

import (
	_ "easymvp/app/app/mvp/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"easymvp/app/app/mvp/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
