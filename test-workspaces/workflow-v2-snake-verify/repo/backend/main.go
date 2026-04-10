package main

import (
	_ "workflowv2snake/backend/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"workflowv2snake/backend/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
