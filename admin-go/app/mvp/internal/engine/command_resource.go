package engine

import (
	"context"

	"easymvp/utility/commandresource"
)

// CommandResourcePolicy 是共享命令资源约束策略在 engine 包内的别名。
type CommandResourcePolicy = commandresource.Policy

// GetCommandResourcePolicy 读取命令资源约束配置。
func GetCommandResourcePolicy(ctx context.Context) CommandResourcePolicy {
	return commandresource.Get(ctx)
}
