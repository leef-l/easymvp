package main

import (
	_ "easymvp/app/{{.AppName}}/internal/packed"
{{range .Modules}}
	_ "easymvp/app/{{$.AppName}}/internal/logic/{{.}}"{{end}}

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"easymvp/app/{{.AppName}}/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
