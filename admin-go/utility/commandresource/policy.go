package commandresource

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// Policy 描述命令执行时的默认资源约束。
// 主要用于限制低配服务器上的 npm/pnpm/go/turbo/compose 类命令资源占用。
type Policy struct {
	Enabled                  bool
	NodeMaxOldSpaceMB        int
	NPMJobs                  int
	NPMNetworkConcurrency    int
	PNPMChildConcurrency     int
	PNPMWorkspaceConcurrency int
	GoMaxProcs               int
	GoMemLimitMB             int
	ComposeParallelLimit     int
	TurboConcurrency         int
	MakeJobs                 int
	CargoBuildJobs           int
	ProcessAddressSpaceMB    int
}

// Get 读取命令资源约束配置。
func Get(ctx context.Context) Policy {
	policy := Policy{
		Enabled:                  getConfigInt(ctx, "command_resource.enabled", "engine.commandResource.enabled", 1) == 1,
		NodeMaxOldSpaceMB:        getConfigInt(ctx, "command_resource.node_max_old_space_mb", "engine.commandResource.nodeMaxOldSpaceMB", 1024),
		NPMJobs:                  getConfigInt(ctx, "command_resource.npm_jobs", "engine.commandResource.npmJobs", 1),
		NPMNetworkConcurrency:    getConfigInt(ctx, "command_resource.npm_network_concurrency", "engine.commandResource.npmNetworkConcurrency", 4),
		PNPMChildConcurrency:     getConfigInt(ctx, "command_resource.pnpm_child_concurrency", "engine.commandResource.pnpmChildConcurrency", 1),
		PNPMWorkspaceConcurrency: getConfigInt(ctx, "command_resource.pnpm_workspace_concurrency", "engine.commandResource.pnpmWorkspaceConcurrency", 1),
		GoMaxProcs:               getConfigInt(ctx, "command_resource.go_max_procs", "engine.commandResource.goMaxProcs", 1),
		GoMemLimitMB:             getConfigInt(ctx, "command_resource.go_mem_limit_mb", "engine.commandResource.goMemLimitMB", 768),
		ComposeParallelLimit:     getConfigInt(ctx, "command_resource.compose_parallel_limit", "engine.commandResource.composeParallelLimit", 1),
		TurboConcurrency:         getConfigInt(ctx, "command_resource.turbo_concurrency", "engine.commandResource.turboConcurrency", 1),
		MakeJobs:                 getConfigInt(ctx, "command_resource.make_jobs", "engine.commandResource.makeJobs", 1),
		CargoBuildJobs:           getConfigInt(ctx, "command_resource.cargo_build_jobs", "engine.commandResource.cargoBuildJobs", 1),
		ProcessAddressSpaceMB:    getConfigInt(ctx, "command_resource.process_address_space_mb", "engine.commandResource.processAddressSpaceMB", 1536),
	}

	if policy.NodeMaxOldSpaceMB < 0 {
		policy.NodeMaxOldSpaceMB = 0
	}
	if policy.NPMJobs < 1 {
		policy.NPMJobs = 1
	}
	if policy.NPMNetworkConcurrency < 1 {
		policy.NPMNetworkConcurrency = 1
	}
	if policy.PNPMChildConcurrency < 1 {
		policy.PNPMChildConcurrency = 1
	}
	if policy.PNPMWorkspaceConcurrency < 1 {
		policy.PNPMWorkspaceConcurrency = 1
	}
	if policy.GoMaxProcs < 1 {
		policy.GoMaxProcs = 1
	}
	if policy.GoMemLimitMB < 0 {
		policy.GoMemLimitMB = 0
	}
	if policy.ComposeParallelLimit < 1 {
		policy.ComposeParallelLimit = 1
	}
	if policy.TurboConcurrency < 1 {
		policy.TurboConcurrency = 1
	}
	if policy.MakeJobs < 1 {
		policy.MakeJobs = 1
	}
	if policy.CargoBuildJobs < 1 {
		policy.CargoBuildJobs = 1
	}
	if policy.ProcessAddressSpaceMB < 0 {
		policy.ProcessAddressSpaceMB = 0
	}

	return policy
}

// Summary 返回适合展示在系统检查里的摘要。
func (p Policy) Summary() string {
	if !p.Enabled {
		return "命令资源限制已禁用，安装/构建命令将不再自动注入资源约束。"
	}
	parts := []string{
		fmt.Sprintf("NODE_OPTIONS=--max-old-space-size=%d", p.NodeMaxOldSpaceMB),
		fmt.Sprintf("npm_maxsockets=%d", p.NPMNetworkConcurrency),
		fmt.Sprintf("pnpm_child=%d", p.PNPMChildConcurrency),
		fmt.Sprintf("pnpm_workspace=%d", p.PNPMWorkspaceConcurrency),
		fmt.Sprintf("go_max_procs=%d", p.GoMaxProcs),
		fmt.Sprintf("go_mem_limit=%dMiB", p.GoMemLimitMB),
		fmt.Sprintf("compose_parallel=%d", p.ComposeParallelLimit),
		fmt.Sprintf("turbo=%d", p.TurboConcurrency),
	}
	if p.ProcessAddressSpaceMB > 0 {
		parts = append(parts, fmt.Sprintf("process_as=%dMiB", p.ProcessAddressSpaceMB))
	}
	return "已启用命令资源限制：" + strings.Join(parts, ", ")
}

// MergeEnv 把资源约束合并到命令环境变量中。
// base 中已有值时，默认以系统约束为准，避免项目在低配服务器上把并发和内存重新拉高。
func (p Policy) MergeEnv(base map[string]string) map[string]string {
	env := make(map[string]string, len(base)+12)
	for key, value := range base {
		env[key] = value
	}
	if !p.Enabled {
		return env
	}

	p.mergeCommonEnv(env)
	return env
}

// EnvPairs 以 key=value 形式输出资源约束环境变量。
func (p Policy) EnvPairs(base map[string]string) []string {
	env := p.MergeEnv(base)
	return mapToEnvPairs(env)
}

// EnvSlice 以现有环境变量切片为基础，返回应用资源约束后的去重结果。
func (p Policy) EnvSlice(base []string) []string {
	return mapToEnvPairs(p.MergeEnv(envSliceToMap(base)))
}

// Apply 按命令类型合并环境变量，并在 Linux 上使用 prlimit 包一层地址空间上限。
func (p Policy) Apply(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}

	baseEnv := cmd.Env
	if len(baseEnv) == 0 {
		baseEnv = os.Environ()
	}
	commandName := detectCommandName(cmd)
	env := envSliceToMap(baseEnv)
	p.mergeCommonEnv(env)
	p.mergeCommandSpecificEnv(env, commandName)
	cmd.Env = mapToEnvPairs(env)

	if !p.Enabled || p.ProcessAddressSpaceMB <= 0 || runtime.GOOS != "linux" || skipPrlimitCommand(commandName) {
		return
	}
	if base := normalizeCommandName(filepath.Base(cmd.Path)); base == "prlimit" {
		return
	}
	if len(cmd.Args) > 0 && normalizeCommandName(filepath.Base(cmd.Args[0])) == "prlimit" {
		return
	}
	prlimitPath, err := exec.LookPath("prlimit")
	if err != nil {
		return
	}

	target := strings.TrimSpace(cmd.Path)
	if target == "" && len(cmd.Args) > 0 {
		target = cmd.Args[0]
	}
	if target == "" {
		return
	}

	limitBytes := int64(p.ProcessAddressSpaceMB) * 1024 * 1024
	originalArgs := append([]string(nil), cmd.Args[1:]...)
	cmd.Path = prlimitPath
	cmd.Args = append([]string{prlimitPath, fmt.Sprintf("--as=%d", limitBytes), "--", target}, originalArgs...)
}

func getConfigInt(ctx context.Context, key string, yamlPath string, defaultVal int) int {
	if ctx == nil {
		ctx = context.Background()
	}

	row, err := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Fields("config_value").
		One()
	if err == nil && !row.IsEmpty() {
		if v, convErr := strconv.Atoi(row["config_value"].String()); convErr == nil {
			return v
		}
	}

	cfgVal := g.Cfg().MustGet(ctx, yamlPath)
	if cfgVal != nil && !cfgVal.IsEmpty() {
		return cfgVal.Int()
	}
	return defaultVal
}

func envSliceToMap(values []string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	env := make(map[string]string, len(values))
	for _, pair := range values {
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		env[key] = value
	}
	return env
}

func mapToEnvPairs(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+env[key])
	}
	return pairs
}

func (p Policy) mergeCommonEnv(env map[string]string) {
	if env == nil || !p.Enabled {
		return
	}
	if p.NodeMaxOldSpaceMB > 0 {
		env["NODE_OPTIONS"] = mergeNodeMaxOldSpace(env["NODE_OPTIONS"], p.NodeMaxOldSpaceMB)
	}
	env["GOMAXPROCS"] = strconv.Itoa(p.GoMaxProcs)
	if p.GoMemLimitMB > 0 {
		env["GOMEMLIMIT"] = fmt.Sprintf("%dMiB", p.GoMemLimitMB)
	}
	env["COMPOSE_PARALLEL_LIMIT"] = strconv.Itoa(p.ComposeParallelLimit)
	env["TURBO_CONCURRENCY"] = strconv.Itoa(p.TurboConcurrency)
	env["MAKEFLAGS"] = fmt.Sprintf("-j%d", p.MakeJobs)
	env["CARGO_BUILD_JOBS"] = strconv.Itoa(p.CargoBuildJobs)
}

func (p Policy) mergeCommandSpecificEnv(env map[string]string, commandName string) {
	if env == nil || !p.Enabled {
		return
	}

	switch normalizeCommandName(commandName) {
	case "npm", "npx":
		env["npm_config_audit"] = "false"
		env["npm_config_fund"] = "false"
		env["npm_config_update_notifier"] = "false"
		env["npm_config_loglevel"] = "warn"
		env["npm_config_maxsockets"] = strconv.Itoa(p.NPMNetworkConcurrency)
	case "pnpm", "pnpx":
		env["npm_config_audit"] = "false"
		env["npm_config_fund"] = "false"
		env["npm_config_update_notifier"] = "false"
		env["npm_config_loglevel"] = "warn"
		env["npm_config_maxsockets"] = strconv.Itoa(p.NPMNetworkConcurrency)
		env["npm_config_network_concurrency"] = strconv.Itoa(p.NPMNetworkConcurrency)
		env["npm_config_child_concurrency"] = strconv.Itoa(p.PNPMChildConcurrency)
		env["npm_config_workspace_concurrency"] = strconv.Itoa(p.PNPMWorkspaceConcurrency)
	}
}

func detectCommandName(cmd *exec.Cmd) string {
	if cmd == nil {
		return ""
	}
	if len(cmd.Args) == 0 {
		return filepath.Base(strings.TrimSpace(cmd.Path))
	}

	base := normalizeCommandName(filepath.Base(cmd.Args[0]))
	if base == "docker" && len(cmd.Args) > 1 && normalizeCommandName(cmd.Args[1]) == "compose" {
		return "docker-compose"
	}
	if !isShellCommand(base) || len(cmd.Args) < 3 {
		return cmd.Args[0]
	}
	return detectShellCommandName(cmd.Args[2])
}

func detectShellCommandName(script string) string {
	fields := strings.Fields(strings.TrimSpace(script))
	if len(fields) == 0 {
		return ""
	}

	for i := 0; i < len(fields); i++ {
		token := strings.Trim(fields[i], "'\"")
		if token == "" {
			continue
		}
		if token == "exec" {
			continue
		}
		if isShellEnvAssignment(token) {
			continue
		}
		token = normalizeCommandName(filepath.Base(token))
		if token == "docker" && i+1 < len(fields) && normalizeCommandName(strings.Trim(fields[i+1], "'\"")) == "compose" {
			return "docker-compose"
		}
		if token == "corepack" && i+1 < len(fields) {
			next := normalizeCommandName(strings.Trim(fields[i+1], "'\""))
			switch next {
			case "pnpm", "pnpx", "npm", "npx":
				return next
			}
		}
		return token
	}
	return ""
}

func isShellCommand(name string) bool {
	switch normalizeCommandName(name) {
	case "sh", "bash", "zsh", "dash":
		return true
	default:
		return false
	}
}

func isShellEnvAssignment(token string) bool {
	if token == "" || strings.HasPrefix(token, "-") {
		return false
	}
	eq := strings.Index(token, "=")
	if eq <= 0 {
		return false
	}
	key := token[:eq]
	for i, ch := range key {
		if ch == '_' || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
			continue
		}
		if i > 0 && ch >= '0' && ch <= '9' {
			continue
		}
		return false
	}
	return true
}

func normalizeCommandName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func skipPrlimitCommand(commandName string) bool {
	switch normalizeCommandName(commandName) {
	case "npm", "npx", "pnpm", "pnpx", "node", "vite", "vitest", "tsc", "tsx", "eslint", "next", "nuxt", "react-scripts", "turbo":
		return true
	default:
		return false
	}
}

func mergeNodeMaxOldSpace(existing string, limitMB int) string {
	if limitMB <= 0 {
		return strings.TrimSpace(existing)
	}

	fields := strings.Fields(existing)
	result := make([]string, 0, len(fields)+1)
	replaced := false

	for i := 0; i < len(fields); i++ {
		part := fields[i]
		switch {
		case strings.HasPrefix(part, "--max-old-space-size="):
			result = append(result, fmt.Sprintf("--max-old-space-size=%d", limitMB))
			replaced = true
		case part == "--max-old-space-size":
			result = append(result, fmt.Sprintf("--max-old-space-size=%d", limitMB))
			replaced = true
			if i+1 < len(fields) {
				i++
			}
		default:
			result = append(result, part)
		}
	}

	if !replaced {
		result = append(result, fmt.Sprintf("--max-old-space-size=%d", limitMB))
	}
	return strings.TrimSpace(strings.Join(result, " "))
}
