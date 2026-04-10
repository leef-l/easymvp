package eventstream

import (
	"context"

	"github.com/gogf/gf/v2/container/gvar"
)

// RedisCommander 约束事件流对 Redis 的最小依赖。
type RedisCommander interface {
	Do(ctx context.Context, command string, args ...any) (*gvar.Var, error)
}
