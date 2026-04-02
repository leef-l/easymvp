package main

import (
	_ "easymvp/app/system/internal/packed"

	_ "easymvp/app/system/internal/logic/auth"
	_ "easymvp/app/system/internal/logic/dept"
	_ "easymvp/app/system/internal/logic/menu"
	_ "easymvp/app/system/internal/logic/role"
	_ "easymvp/app/system/internal/logic/users"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"easymvp/app/system/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
