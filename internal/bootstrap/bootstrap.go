package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shenghongio/kraken-ai-platform/orchestrator/persistence/postgres"
	redispkg "github.com/shenghongio/kraken-ai-platform/orchestrator/persistence/redis"
	"github.com/shenghongio/kraken-ai-platform/platform/config"
	"github.com/shenghongio/kraken-ai-platform/platform/coordination"
	"github.com/shenghongio/kraken-ai-platform/platform/database"
	"github.com/shenghongio/kraken-ai-platform/platform/logging/observability"
)

// 提供三个服务的统一初始化流程
// 配置加载 -> 配置校验 -> 日志初始化 -> 依赖初始化
// 所有服务 (kraken-api,kraken-reaper,kraken-outbox) 保持一致这种初始化格式,避免三个服务各自发展处不同的基础规范

// DefaultConfigPath 默认配置文件
const DefaultConfigPath = config.DefaultConfigPath

// Deps 服务依赖，包含初始化后的资源
type Deps struct {
	// Logger 统一日志
	Logger *slog.Logger
	// Pool PostgreSQL 连接池, 使用 pgxpool.Pool 相比线程安全
	Pool *pgxpool.Pool
	// Redis 客户端
	Redis *redispkg.Client

	// db 与 coordination 保存资源的生命周期封装，用于健康检查与关闭
	db           *database.DB
	coordination *coordination.Coordination
}

// LoadConfig 加载配置文件，委托给 platform/config
// serviceName: 服务名称，用于确定默认配置路径
// defaultPath: 如果不为空，则优先使用此路径
// 加载顺序为: 命令行 flag > 环境变量 > 默认值
func LoadConfig(serviceName, defaultPath string) (*config.SchedulerConfig, string, error) {
	return config.LoadServiceConfig(serviceName, defaultPath)
}

// ValidateConfig 配置校验
// 默认值与校验规则已在 LoadConfig 内部完成，此函数保留签名兼容
func ValidateConfig(cfg *config.SchedulerConfig, serviceName string) error {
	return cfg.Validate()
}

// InitLogger 初始化日志 自动脱敏敏感字段
func InitLogger(cfg *config.SchedulerConfig) *slog.Logger {
	return observability.InitLogger(cfg.Log.Level, cfg.Log.Format)
}

// InitDeps 初始化依赖 [PostgreSQL + Redis]
// PostgreSQL：权威数据源，连接失败则服务启动失败
// Redis：通知通道，连接失败不阻塞启动（Outbox 事件在恢复后补发
func InitDeps(ctx context.Context, cfg *config.SchedulerConfig, logger *slog.Logger, serviceName string) (*Deps, error) {
	deps := &Deps{
		Logger: logger,
	}

	// 初始化 PostgreSQL 连接池,连接失败意味着服务无法正常工作，必须返回错误
	logger.Info("initializing postgres connection pool",
		slog.String("dsn", cfg.DB.DSN),
		slog.Int("max_open_conns", cfg.DB.MaxOpenConns),
		slog.Int("max_idle_conns", cfg.DB.MaxIdleConns),
		slog.Duration("conn_max_lifetime", cfg.DB.ConnMaxLifetime),
	)
	db, err := database.Open(
		ctx,
		cfg.DB.DSN,
		cfg.DB.MaxOpenConns,
		cfg.DB.MaxIdleConns,
		cfg.DB.ConnMaxLifetime,
	)
	if err != nil {
		return nil, fmt.Errorf("init postgres connection pool: %w", err)
	}
	deps.db = db
	deps.Pool = db.Pool()

	// 执行 migrations
	if serviceName == "kraken-api" {
		logger.Debug("migrating database schema")
		migrationsFile, err := postgres.RunMigrations(ctx, deps.Pool)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("run migrations: %w", err)
		}
		logger.Debug("Migrations complete", slog.String("file", migrationsFile))
	}

	logger.Info("initializing redis connection pool")

	// 初始化 redis, 设计决策：Outbox 事件在 Redis 恢复后会自动补发
	if cfg.Redis.Addr != "" {
		coord := coordination.Open(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
		deps.coordination = coord
		deps.Redis = coord.Client()

		// ping检查 redis 连通性,单失败值记录警告,不阻塞启动
		withTimeout, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
		if err := coord.Ready(withTimeout); err != nil {
			logger.Warn("redis unavailable, outbox events will be buffered and retried after redis recovers",
				slog.String("addr", cfg.Redis.Addr),
				slog.Any("error", err),
				slog.String("note", "redis is notification channel only, postgres is authoritative for task state"))
		} else {
			logger.Info("redis connection pool successfully initialized", slog.String("addr", cfg.Redis.Addr))
		}
		cancelFunc()
	}
	return deps, nil
}

// CloseDeps 关闭依赖顺序于初始化顺序相反，保证资源正确释放
func CloseDeps(deps *Deps) {
	// 先关闭 Redis 在关闭 PostgreSQL
	if deps.coordination != nil {
		deps.coordination.Close()
	}
	if deps.db != nil {
		deps.db.Close()
	}
}

// ShutdownTimeout 优雅关闭超时时间。统一生命周期规范，优雅关闭最多等待超时此时间
const ShutdownTimeout = 10 * time.Second
