package service

import (
	"context"
	"strings"

	"github.com/leef-l/easymvp/apps/core/api/system/v1"
)

const (
	startupSeverityInfo    = "info"
	startupSeverityWarning = "warning"
	startupSeverityError   = "error"
)

func buildStartupSnapshot(ctx context.Context, startup StartupConfig, workerStatus WorkerManagerStatus, runtimeErr error) *v1.HealthStartupSnapshot {
	var (
		mode          = StartupMode(ctx)
		runtimeStatus = "ok"
		workerState   = "stopped"
	)

	switch {
	case startup.SafeMode:
		workerState = "disabled-safe-mode"
	case workerStatus.Started:
		workerState = "running"
	}
	if runtimeErr != nil {
		runtimeStatus = "degraded"
	}

	diagnostics := make([]*v1.HealthDiagnosticItem, 0, 8)
	if startup.SafeMode {
		diagnostics = append(diagnostics, newStartupDiagnostic(
			"STARTUP_SAFE_MODE",
			startupSeverityWarning,
			"startup",
			"safe_mode",
			"core is running in safe-mode",
			"Background workers are intentionally disabled so recovery and inspection APIs can stay available.",
			[]string{
				"Restart without --safe-mode after recovery is complete.",
				"Use diagnostics data before re-enabling background workers.",
			},
		))
	}
	if !startup.SafeMode && !workerStatus.Started {
		diagnostics = append(diagnostics, newStartupDiagnostic(
			"STARTUP_WORKERS_NOT_RUNNING",
			startupSeverityWarning,
			"workers",
			"worker_status",
			"background workers are not running",
			"Core is in normal mode but the worker manager has not reported a started state.",
			[]string{
				"Check startup logs for worker manager initialization failures.",
				"Restart core if workers should be active in this environment.",
			},
		))
	}

	diagnostics = append(diagnostics, collectStartupOptionDiagnostics(startup)...)

	if strings.TrimSpace(startup.BrainServeBaseURL) == "" {
		diagnostics = append(diagnostics, newStartupDiagnostic(
			"STARTUP_RUNTIME_BASE_URL_MISSING",
			startupSeverityError,
			"runtime",
			"brain_serve_base_url",
			"brain serve base URL is missing",
			"The runtime integration cannot reach easymvp-brain because the startup configuration resolved an empty base URL.",
			[]string{
				"Set easymvp.brainServeBaseURL in config or pass --brain-serve-base-url.",
				"Restart core after updating the runtime base URL.",
			},
		))
	} else if runtimeErr != nil {
		diagnostics = append(diagnostics, newStartupDiagnostic(
			"STARTUP_RUNTIME_UNAVAILABLE",
			startupSeverityWarning,
			"runtime",
			"brain_serve_base_url",
			"brain runtime health check failed",
			runtimeErr.Error(),
			[]string{
				"Verify the easymvp-brain service is reachable from the core host.",
				"Confirm the configured brain serve base URL points to the correct endpoint.",
			},
		))
	}

	status := "ok"
	summary := "core startup configuration is healthy"
	ready := true
	if startup.SafeMode {
		status = "recovery"
		summary = "core is serving recovery mode with workers disabled"
		ready = false
	}
	if hasStartupSeverity(diagnostics, startupSeverityError) {
		status = "degraded"
		summary = "core startup has blocking diagnostics that require intervention"
		ready = false
	} else if hasStartupSeverity(diagnostics, startupSeverityWarning) && status == "ok" {
		status = "attention"
		summary = "core startup is available with warnings that may affect recovery or automation"
	}

	return &v1.HealthStartupSnapshot{
		Status:         status,
		Ready:          ready,
		Summary:        summary,
		Mode:           mode,
		SafeMode:       startup.SafeMode,
		WorkerEnabled:  !startup.SafeMode,
		WorkerStatus:   workerState,
		RuntimeStatus:  runtimeStatus,
		RuntimeBaseURL: startup.BrainServeBaseURL,
		Config: &v1.HealthStartupConfigState{
			DataRoot:          toHealthStartupConfigField(startup.Options.DataRoot),
			DBPath:            toHealthStartupConfigField(startup.Options.DBPath),
			MigrationPath:     toHealthStartupConfigField(startup.Options.MigrationPath),
			BrainServeBaseURL: toHealthStartupConfigField(startup.Options.BrainServeBaseURL),
			ServerAddress:     toHealthStartupConfigField(startup.Options.ServerAddress),
			SafeMode:          toHealthStartupConfigField(startup.Options.SafeMode),
		},
		Diagnostics: diagnostics,
	}
}

func collectStartupOptionDiagnostics(startup StartupConfig) []*v1.HealthDiagnosticItem {
	options := []struct {
		codePrefix string
		component  string
		field      string
		summary    string
		actions    []string
		option     StartupOption
	}{
		{
			codePrefix: "STARTUP_DATA_ROOT",
			component:  "startup",
			field:      "data_root",
			summary:    "data root is using the built-in default path",
			actions: []string{
				"Set easymvp.dataRoot in config when the default workspace is not appropriate.",
				"Pass --data-root to override the data directory for this boot.",
			},
			option: startup.Options.DataRoot,
		},
		{
			codePrefix: "STARTUP_DB_PATH",
			component:  "startup",
			field:      "db_path",
			summary:    "database path is using the built-in default path",
			actions: []string{
				"Set easymvp.dbPath in config if SQLite should live outside the default data directory.",
				"Pass --db-path to override the database file path for this boot.",
			},
			option: startup.Options.DBPath,
		},
		{
			codePrefix: "STARTUP_MIGRATION_PATH",
			component:  "startup",
			field:      "migration_path",
			summary:    "migration path is using the built-in default path",
			actions: []string{
				"Set easymvp.migrationPath in config when migrations are packaged elsewhere.",
				"Pass --migration-path to override the migration directory for this boot.",
			},
			option: startup.Options.MigrationPath,
		},
		{
			codePrefix: "STARTUP_RUNTIME_BASE_URL",
			component:  "runtime",
			field:      "brain_serve_base_url",
			summary:    "brain serve base URL is using the built-in default value",
			actions: []string{
				"Set easymvp.brainServeBaseURL in config if easymvp-brain is hosted elsewhere.",
				"Pass --brain-serve-base-url to override the runtime endpoint for this boot.",
			},
			option: startup.Options.BrainServeBaseURL,
		},
		{
			codePrefix: "STARTUP_SERVER_ADDRESS",
			component:  "server",
			field:      "server_address",
			summary:    "HTTP listen address is using the built-in default value",
			actions: []string{
				"Set server.address in config when the default port is not appropriate.",
				"Pass --port to override the HTTP listen address for this boot.",
			},
			option: startup.Options.ServerAddress,
		},
	}

	diagnostics := make([]*v1.HealthDiagnosticItem, 0, len(options))
	for _, item := range options {
		if !item.option.UsingDefault {
			continue
		}
		diagnostics = append(diagnostics, newStartupDiagnostic(
			item.codePrefix+"_DEFAULT",
			startupSeverityInfo,
			item.component,
			item.field,
			item.summary,
			"Current value: "+item.option.Value,
			item.actions,
		))
	}
	return diagnostics
}

func hasStartupSeverity(items []*v1.HealthDiagnosticItem, severity string) bool {
	for _, item := range items {
		if item != nil && item.Severity == severity {
			return true
		}
	}
	return false
}

func toHealthStartupConfigField(option StartupOption) *v1.HealthStartupConfigField {
	return &v1.HealthStartupConfigField{
		Key:          option.Key,
		Value:        option.Value,
		Source:       option.Source,
		DefaultValue: option.DefaultValue,
		Configured:   option.Configured,
		UsingDefault: option.UsingDefault,
	}
}

func newStartupDiagnostic(code string, severity string, component string, field string, summary string, detail string, actions []string) *v1.HealthDiagnosticItem {
	return &v1.HealthDiagnosticItem{
		Code:      code,
		Severity:  severity,
		Component: component,
		Field:     field,
		Summary:   summary,
		Detail:    detail,
		Actions:   actions,
	}
}
