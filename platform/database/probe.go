package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolProbe 连接池满足health.Probe,用于apiserver，这类支持有 *pgxpool.Pool的场景
type PoolProbe struct {
	pool *pgxpool.Pool
}

// NewPoolProbe 包装
func NewPoolProbe(pool *pgxpool.Pool) *PoolProbe {
	return &PoolProbe{pool: pool}
}

// Ready实现 health.Probe
func (p *PoolProbe) Ready(ctx context.Context) error {
	if p.pool == nil {
		return nil
	}
	return p.pool.Ping(ctx)
}
