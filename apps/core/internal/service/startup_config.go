package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"

	"github.com/leef-l/easymvp/apps/core/internal/consts"
)

type StartupConfig struct {
	DataRoot          string
	DBPath            string
	MigrationPath     string
	BrainServeBaseURL string
	ServerAddress     string
	SafeMode          bool
	Options           StartupConfigOptions
}

type StartupConfigOptions struct {
	DataRoot          StartupOption
	DBPath            StartupOption
	MigrationPath     StartupOption
	BrainServeBaseURL StartupOption
	ServerAddress     StartupOption
	SafeMode          StartupOption
}

type StartupOption struct {
	Key          string
	Value        string
	Source       string
	DefaultValue string
	Configured   bool
	UsingDefault bool
}

var startupConfigStore = struct {
	mu    sync.RWMutex
	value *StartupConfig
}{}

func ResolveStartupConfig(ctx context.Context, parser *gcmd.Parser) StartupConfig {
	var (
		dataRootValue, dataRootOption                   = loadStartupStringOption(ctx, "easymvp.dataRoot", "./var", strings.TrimSpace)
		dbPathValue, dbPathOption                       = loadStartupStringOption(ctx, "easymvp.dbPath", "./var/data/easymvp.db", strings.TrimSpace)
		migrationPathValue, migrationPathOption         = loadStartupStringOption(ctx, "easymvp.migrationPath", "./manifest/migrations", strings.TrimSpace)
		brainServeBaseURLValue, brainServeBaseURLOption = loadStartupStringOption(ctx, "easymvp.brainServeBaseURL", "http://127.0.0.1:7701", normalizeRuntimeBaseURL)
		serverAddressValue, serverAddressOption         = loadStartupStringOption(ctx, "server.address", ":8000", strings.TrimSpace)
		safeModeValue, safeModeOption                   = loadStartupBoolOption(ctx, "easymvp.safeMode", false)
	)

	cfg := StartupConfig{
		DataRoot:          dataRootValue,
		DBPath:            dbPathValue,
		MigrationPath:     migrationPathValue,
		BrainServeBaseURL: brainServeBaseURLValue,
		ServerAddress:     serverAddressValue,
		SafeMode:          safeModeValue,
		Options: StartupConfigOptions{
			DataRoot:          dataRootOption,
			DBPath:            dbPathOption,
			MigrationPath:     migrationPathOption,
			BrainServeBaseURL: brainServeBaseURLOption,
			ServerAddress:     serverAddressOption,
			SafeMode:          safeModeOption,
		},
	}
	if parser == nil {
		return cfg
	}

	opts := parser.GetOptAll()
	if value := strings.TrimSpace(opts["data-root"]); value != "" {
		cfg.DataRoot = value
		cfg.Options.DataRoot = overrideStartupStringOption(cfg.Options.DataRoot, value, strings.TrimSpace)
	}
	if value := strings.TrimSpace(opts["db-path"]); value != "" {
		cfg.DBPath = value
		cfg.Options.DBPath = overrideStartupStringOption(cfg.Options.DBPath, value, strings.TrimSpace)
	}
	if value := strings.TrimSpace(opts["migration-path"]); value != "" {
		cfg.MigrationPath = value
		cfg.Options.MigrationPath = overrideStartupStringOption(cfg.Options.MigrationPath, value, strings.TrimSpace)
	}
	if value := strings.TrimSpace(opts["brain-serve-base-url"]); value != "" {
		cfg.BrainServeBaseURL = normalizeRuntimeBaseURL(value)
		cfg.Options.BrainServeBaseURL = overrideStartupStringOption(cfg.Options.BrainServeBaseURL, value, normalizeRuntimeBaseURL)
	}
	if value := strings.TrimSpace(opts["port"]); value != "" {
		cfg.ServerAddress = normalizeServerAddress(value)
		cfg.Options.ServerAddress = overrideStartupStringOption(cfg.Options.ServerAddress, value, normalizeServerAddress)
	}
	if _, ok := opts["safe-mode"]; ok {
		cfg.SafeMode = parseBoolOption(opts["safe-mode"])
		cfg.Options.SafeMode = overrideStartupBoolOption(cfg.Options.SafeMode, cfg.SafeMode)
	}
	return cfg
}

func SetStartupConfig(cfg StartupConfig) {
	startupConfigStore.mu.Lock()
	defer startupConfigStore.mu.Unlock()

	copied := cfg
	startupConfigStore.value = &copied
}

func CurrentStartupConfig(ctx context.Context) StartupConfig {
	startupConfigStore.mu.RLock()
	if startupConfigStore.value != nil {
		copied := *startupConfigStore.value
		startupConfigStore.mu.RUnlock()
		return copied
	}
	startupConfigStore.mu.RUnlock()
	return ResolveStartupConfig(ctx, nil)
}

func StartupMode(ctx context.Context) string {
	if CurrentStartupConfig(ctx).SafeMode {
		return "safe-mode"
	}
	return "normal"
}

func loadStartupStringOption(ctx context.Context, key string, def string, normalize func(string) string) (string, StartupOption) {
	value := def
	option := StartupOption{
		Key:          key,
		Source:       consts.StartupConfigSourceDefault,
		DefaultValue: def,
		Configured:   false,
		UsingDefault: true,
	}
	cfgValue := g.Cfg().MustGet(ctx, key).String()
	if strings.TrimSpace(cfgValue) != "" {
		value = cfgValue
		option.Source = consts.StartupConfigSourceConfig
		option.Configured = true
		option.UsingDefault = false
	}
	if normalize != nil {
		value = normalize(value)
		option.DefaultValue = normalize(option.DefaultValue)
	}
	option.Value = value
	return value, option
}

func loadStartupBoolOption(ctx context.Context, key string, def bool) (bool, StartupOption) {
	value := def
	option := StartupOption{
		Key:          key,
		Source:       consts.StartupConfigSourceDefault,
		DefaultValue: formatStartupBool(def),
		Configured:   false,
		UsingDefault: true,
	}
	if g.Cfg().Available(ctx, key) {
		value = g.Cfg().MustGet(ctx, key).Bool()
		option.Source = consts.StartupConfigSourceConfig
		option.Configured = true
		option.UsingDefault = false
	}
	option.Value = formatStartupBool(value)
	return value, option
}

func overrideStartupStringOption(option StartupOption, value string, normalize func(string) string) StartupOption {
	option.Source = consts.StartupConfigSourceCLI
	option.Configured = true
	option.UsingDefault = false
	if normalize != nil {
		value = normalize(value)
	}
	option.Value = value
	return option
}

func overrideStartupBoolOption(option StartupOption, value bool) StartupOption {
	option.Source = consts.StartupConfigSourceCLI
	option.Configured = true
	option.UsingDefault = false
	option.Value = formatStartupBool(value)
	return option
}

func parseBoolOption(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func normalizeServerAddress(port string) string {
	value := strings.TrimSpace(port)
	if value == "" {
		return ":8000"
	}
	if strings.Contains(value, ":") {
		return value
	}
	return ":" + value
}

func normalizeRuntimeBaseURL(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func formatStartupBool(value bool) string {
	return fmt.Sprintf("%t", value)
}
