package coordination

import (
	"context"
	"time"

	redispkg "github.com/shenghongio/kraken-ai-platform/orchestrator/persistence/redis"
)

// coordination 封装Redis 客户端生命周期
// Redis 仅作为通知通道，非权威数据源，不可用时不应该阻塞启动
type Coordination struct {
	client *redispkg.Client
}

// Open 创建客户端，
func Open(addr, password string, db int) *Coordination {
	return &Coordination{
		client: redispkg.NewRedisClient(addr, password, db),
	}
}

// Client 返回底层客户端
func (c *Coordination) Client() *redispkg.Client {
	return c.client
}

// Ready 实现health.Probe 带超时的非阻塞探测
func (c *Coordination) Ready(ctx context.Context) error {
	if c.client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.Ping(ctx)
}

// Close 释放连接
func (c *Coordination) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}
