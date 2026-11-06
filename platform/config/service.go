package config

import (
	"flag"
	"fmt"
	"os"
)

// DefaultConfigPath 默认配置文件，未匹配到服务名时使用
const DefaultConfigPath = "./config/api.example.yaml"

// serviceConfigPaths 各服务的默认配置文件
var serviceConfigPaths = map[string]string{
	"kraken-api":    "./configs/api.example.yaml",
	"kraken-reaper": "./configs/reaper.example.yaml",
	"kraken-outbox": "./configs/outbox.example.yaml",
}

// DefaultConfigPathFor 返回服务对应的默认配置文件路径
func DefaultConfigPathFor(serviceName string) string {
	if path, ok := serviceConfigPaths[serviceName]; ok {
		return path
	}
	return DefaultConfigPath
}

// LoadServiceConfig 加载并校验服务配置
// 加载顺序：默认值 -> YAML 文件 -> 环境变量 -> 命令行 flag (-config)
// 环境变量 KRAKEN_ 前缀，双下划线 __ 表示层级，例如 KRAKEN_DB__DSN 对应 db.dsn
// 返回解析后的配置与实际使用的配置文件路径
func LoadServiceConfig(serviceName, defaultPath string) (*SchedulerConfig, string, error) {
	if defaultPath == "" {
		defaultPath = DefaultConfigPathFor(serviceName)
	}
	cfgPath := flag.String("config", defaultPath, "path to kraken-ai-platform config file")
	flag.Parse()

	if !isFlagPassed("config") {
		if envPath := os.Getenv("KRAKEN_SCHEDULER_CONFIG_PATH"); envPath != "" {
			*cfgPath = envPath
		}
	}

	cfg, err := LoadSchedulerConfig(*cfgPath)
	if err != nil {
		return nil, *cfgPath, fmt.Errorf("load config from %s: %w", *cfgPath, err)
	}

	cfg.ApplyServiceDefaults(serviceName)
	if err := cfg.Validate(); err != nil {
		return nil, *cfgPath, fmt.Errorf("validate config from %s: %w", *cfgPath, err)
	}
	return cfg, *cfgPath, nil
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
