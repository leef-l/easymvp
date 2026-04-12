package fsm

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// StageConfig 阶段编排配置。
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
	Stages []StageConfig `yaml:"stages"`
}

// defaultConfig 内置默认配置，YAML 加载失败时的回退。
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

var (
	loadOnce     sync.Once
	loadedConfig *Config
)

// yamlStagesFile YAML 配置文件的搜索路径。
var yamlStagesFile = []string{
	"manifest/fsm/stages.yaml",
	"../manifest/fsm/stages.yaml",
	"../../manifest/fsm/stages.yaml",
}

// LoadConfig 加载 FSM 配置。
// 优先从 manifest/fsm/stages.yaml 加载，失败时回退到内置默认配置。
// 加载结果全局缓存（单次加载）。
func LoadConfig() *Config {
	loadOnce.Do(func() {
		for _, path := range yamlStagesFile {
			cfg, err := loadYAMLConfig(path)
			if err == nil && cfg != nil && len(cfg.Stages) > 0 {
				loadedConfig = cfg
				return
			}
		}
		loadedConfig = defaultConfig
	})
	return loadedConfig
}

// loadYAMLConfig 从指定路径加载 YAML 配置。
func loadYAMLConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 %s 失败: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析 %s 失败: %w", path, err)
	}
	return &cfg, nil
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

// ResetForTest 重置加载状态（仅测试用）。
func ResetForTest() {
	loadOnce = sync.Once{}
	loadedConfig = nil
}
