package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type HealthReq struct {
	g.Meta `path:"/api/v3/system/healthz" tags:"System" method:"get" summary:"System health check"`
}

type HealthRes struct {
	Status            string                 `json:"status"`
	Service           string                 `json:"service"`
	Version           string                 `json:"version"`
	Timestamp         string                 `json:"timestamp"`
	Mode              string                 `json:"mode,omitempty"`
	SafeMode          bool                   `json:"safe_mode"`
	WorkerStatus      string                 `json:"worker_status,omitempty"`
	RegisteredWorkers []string               `json:"registered_workers,omitempty"`
	RuntimeStatus     string                 `json:"runtime_status,omitempty"`
	ServerAddress     string                 `json:"server_address,omitempty"`
	DataRoot          string                 `json:"data_root,omitempty"`
	DBPath            string                 `json:"db_path,omitempty"`
	BrainServeBaseURL string                 `json:"brain_serve_base_url,omitempty"`
	Startup           *HealthStartupSnapshot `json:"startup,omitempty"`
}

type HealthStartupSnapshot struct {
	Status         string                    `json:"status"`
	Ready          bool                      `json:"ready"`
	Summary        string                    `json:"summary,omitempty"`
	Mode           string                    `json:"mode"`
	SafeMode       bool                      `json:"safe_mode"`
	WorkerEnabled  bool                      `json:"worker_enabled"`
	WorkerStatus   string                    `json:"worker_status,omitempty"`
	RuntimeStatus  string                    `json:"runtime_status,omitempty"`
	RuntimeBaseURL string                    `json:"runtime_base_url,omitempty"`
	Config         *HealthStartupConfigState `json:"config,omitempty"`
	Diagnostics    []*HealthDiagnosticItem   `json:"diagnostics,omitempty"`
}

type HealthStartupConfigState struct {
	DataRoot          *HealthStartupConfigField `json:"data_root,omitempty"`
	DBPath            *HealthStartupConfigField `json:"db_path,omitempty"`
	MigrationPath     *HealthStartupConfigField `json:"migration_path,omitempty"`
	BrainServeBaseURL *HealthStartupConfigField `json:"brain_serve_base_url,omitempty"`
	ServerAddress     *HealthStartupConfigField `json:"server_address,omitempty"`
	SafeMode          *HealthStartupConfigField `json:"safe_mode,omitempty"`
}

type HealthStartupConfigField struct {
	Key          string `json:"key"`
	Value        string `json:"value,omitempty"`
	Source       string `json:"source"`
	DefaultValue string `json:"default_value,omitempty"`
	Configured   bool   `json:"configured"`
	UsingDefault bool   `json:"using_default"`
}

type HealthDiagnosticItem struct {
	Code      string   `json:"code"`
	Severity  string   `json:"severity"`
	Component string   `json:"component,omitempty"`
	Field     string   `json:"field,omitempty"`
	Summary   string   `json:"summary"`
	Detail    string   `json:"detail,omitempty"`
	Actions   []string `json:"actions,omitempty"`
}
