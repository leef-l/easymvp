package fsm

// StageConfig 阶段编排配置（后续从 YAML 加载）。
type StageConfig struct {
	// Code 阶段代码，如 "design"/"review"/"execute"/"accept"/"rework"/"complete"。
	Code string `yaml:"code"`
	// Timeout 阶段超时时间，如 "30m"、"2h"。
	Timeout string `yaml:"timeout"`
	// OnTimeout 超时后的处理策略，如 "fail"/"escalate"/"skip"。
	OnTimeout string `yaml:"on_timeout"`
	// MaxRetry 阶段内任务最大重试次数。
	MaxRetry int `yaml:"max_retry"`
	// AllowSkip 是否允许人工跳过。
	AllowSkip bool `yaml:"allow_skip"`
}

// Config FSM 全局配置，包含各阶段编排参数。
type Config struct {
	Stages []StageConfig
}

// defaultConfig 内置默认配置，后续 PR 接入 YAML 文件时替换此处逻辑。
var defaultConfig = &Config{
	Stages: []StageConfig{
		{Code: "design", Timeout: "0", OnTimeout: "fail", MaxRetry: 0, AllowSkip: false},
		{Code: "review", Timeout: "30m", OnTimeout: "escalate", MaxRetry: 1, AllowSkip: true},
		{Code: "execute", Timeout: "2h", OnTimeout: "fail", MaxRetry: 3, AllowSkip: false},
		{Code: "accept", Timeout: "1h", OnTimeout: "escalate", MaxRetry: 2, AllowSkip: true},
		{Code: "rework", Timeout: "2h", OnTimeout: "fail", MaxRetry: 3, AllowSkip: false},
		{Code: "complete", Timeout: "10m", OnTimeout: "fail", MaxRetry: 0, AllowSkip: false},
	},
}

// LoadConfig 加载 FSM 配置。
// 当前版本返回内置默认配置，后续 PR 将支持从 YAML 文件加载并合并覆盖。
func LoadConfig() *Config {
	// TODO(PR-next): 读取 manifest/fsm/stages.yaml，合并到 defaultConfig。
	return defaultConfig
}

// StageConfigByCode 按阶段代码查找配置，未找到返回 nil。
func (c *Config) StageConfigByCode(code string) *StageConfig {
	for i := range c.Stages {
		if c.Stages[i].Code == code {
			return &c.Stages[i]
		}
	}
	return nil
}
