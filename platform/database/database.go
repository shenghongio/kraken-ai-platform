package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shenghongio/kraken-ai-platform/orchestrator/persistence/postgres"
)

// DB 封装 PostgreSQL 连接池生命周期
type DB struct {
	pool *pgxpool.Pool
}

// Open 创建连接池，参数与行为完全沿用 postgres.NewPool
func Open(ctx context.Context, dsn string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*DB, error) {
	pool, err := postgres.NewPool(ctx, dsn, maxOpenConns, maxIdleConns, connMaxLifetime)
	if err != nil {
		return nil, err
	}
	return &DB{pool: pool}, nil
}

// Pool 返回底层连接池
func (d *DB) Pool() *pgxpool.Pool {
	return d.pool
}

// Ready 实现 health.Probe
func (d *DB) Ready(ctx context.Context) error {
	if d.pool == nil {
		return nil
	}
	return d.pool.Ping(ctx)
}

// Close 释放连接池
func (d *DB) Close() {
	if d.pool != nil {
		d.pool.Close()
	}
}
