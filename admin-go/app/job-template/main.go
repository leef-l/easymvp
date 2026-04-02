package main

import (
	_ "easymvp/app/job-template/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"easymvp/app/job-template/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
