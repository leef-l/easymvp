package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type BackendConfig struct {
	Output string `yaml:"output"`
}

type FrontendConfig struct {
	Output string `yaml:"output"`
}

// MenuAppConfig 菜单应用目录配置
type MenuAppConfig struct {
	Title string `yaml:"title"`
	Icon  string `yaml:"icon"`
	Sort  int    `yaml:"sort"` // 目录排序，默认 50
}

// MenuModuleConfig 模块级菜单配置
type MenuModuleConfig struct {
	Sort   int  `yaml:"sort"`    // 菜单排序
	IsShow *int `yaml:"is_show"` // 菜单是否显示（nil=默认1）
}

type Config struct {
	Database     DatabaseConfig                      `yaml:"database"`
	Backend      BackendConfig                       `yaml:"backend"`
	Frontend     FrontendConfig                      `yaml:"frontend"`
	SkipFields   []string                            `yaml:"skip_fields"`
	MenuApps     map[string]MenuAppConfig            `yaml:"menu_apps"`
	MenuModules  map[string]MenuModuleConfig         `yaml:"menu_modules"`  // key: "appName/moduleName"
	OperationLog bool                                `yaml:"operation_log"` // 全局操作日志开关
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	// 支持环境变量语法 ${ENV_VAR}，用于避免密码明文存储
	cfg.Database.Password = expandEnvVar(cfg.Database.Password)
	return &cfg, nil
}

// expandEnvVar 如果值形如 ${VAR_NAME}，则从环境变量读取替换
func expandEnvVar(val string) string {
	if len(val) > 3 && val[:2] == "${" && val[len(val)-1] == '}' {
		envName := val[2 : len(val)-1]
		if envVal := os.Getenv(envName); envVal != "" {
			return envVal
		}
	}
	return val
}

// DSN 生成数据库连接字符串
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

// DSNForHack 生成 hack/config.yaml 中 gf gen dao 使用的 link 格式
// 格式: mysql:user:password@tcp(host:port)/dbname?charset=utf8mb4
func (c *DatabaseConfig) DSNForHack() string {
	return fmt.Sprintf("mysql:%s:%s@tcp(%s:%d)/%s?charset=utf8mb4",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}
