package main

import (
	_ "github.com/leef-l/easymvp/apps/core/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"github.com/leef-l/easymvp/apps/core/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
