package main

import (
	_ "easymvp/app/svc-template/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"easymvp/app/svc-template/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
