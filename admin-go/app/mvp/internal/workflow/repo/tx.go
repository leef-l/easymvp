package repo

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WithTx 在 repo 层统一打开事务，避免业务层直接依赖 g.DB().
func WithTx(ctx context.Context, fn func(ctx context.Context, tx gdb.TX) error) error {
	return g.DB().Transaction(ctx, fn)
}
