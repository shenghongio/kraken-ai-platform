package config

import (
	"errors"
	"fmt"
	"time"
)

type SchedulerConfig struct {
	App   AppConfig   `koanf:"app"`
	Log   LogConfig   `koanf:"log"`
	HTTP  HTTPConfig  `koanf:"http"`
	DB    DBConfig    `koanf:"db"`
	Redis RedisConfig `koanf:"redis"`

	LeaseTTL            time.Duration `koanf:"lease.ttl"`
	LeaseReaperInterval time.Duration `koanf:"lease.reaper_interval"`
	HeartbeatInterval   time.Duration `koanf:"heartbeat.interval"`
	OutboxInterval      time.Duration `koanf:"outbox.interval"`
}

// GetHTTPConfig 实现 http.HTTPConfigProvider接口
func (s *SchedulerConfig) GetHTTPConfig() HTTPConfig {
	return s.HTTP
}
func (s *SchedulerConfig) GetLeaseTTL() time.Duration {
	return s.LeaseTTL
}

type AppConfig struct {
	Role string `koanf:"role"`
}

func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		App: AppConfig{
			Role: "all",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		HTTP: HTTPConfig{
			Addr:         ":8080",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
		DB: DBConfig{
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 10 * time.Minute,
		},
		LeaseTTL:            60 * time.Second,
		LeaseReaperInterval: 5 * time.Second,
		HeartbeatInterval:   15 * time.Second,
		OutboxInterval:      5 * time.Second,
	}
}
func LoadSchedulerConfig(configFile string) (*SchedulerConfig, error) {
	lc, err := Load(configFile)
	if err != nil {
		return nil, err
	}
	cfg := DefaultSchedulerConfig()
	if err := lc.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal scheduler config: %w", err)
	}
	if cfg.DB.DSN == "" {
		return nil, errors.New("db.dsn is required")
	}
	return &cfg, nil
}

// Validate 补齐通用默认值并校验配置
// 已显示设置的字段不会被覆盖
func (s *SchedulerConfig) Validate() error {
	if s.DB.DSN == "" {
		return errors.New("db dsn is required")
	}
	if s.Log.Level == "" {
		s.Log.Level = "info"
	}
	if s.Log.Format == "" {
		s.Log.Format = "json"
	}
	if s.DB.MaxOpenConns == 0 {
		s.DB.MaxOpenConns = 25
	}
	if s.DB.MaxIdleConns == 0 {
		s.DB.MaxIdleConns = 5
	}
	if s.DB.ConnMaxLifetime == 0 {
		s.DB.ConnMaxLifetime = 10 * time.Minute
	}
	if s.LeaseTTL == 0 {
		s.LeaseTTL = 60 * time.Second
	}
	if s.LeaseReaperInterval == 0 {
		s.LeaseReaperInterval = 5 * time.Second
	}
	if s.HeartbeatInterval == 0 {
		s.HeartbeatInterval = 15 * time.Second
	}
	if s.OutboxInterval == 0 {
		s.OutboxInterval = 5 * time.Second
	}
	return nil
}

// ApplyServiceDefaults 补齐服务专属默认值
func (s *SchedulerConfig) ApplyServiceDefaults(serviceName string) {
	if serviceName != "kraken-api" {
		return
	}
}
