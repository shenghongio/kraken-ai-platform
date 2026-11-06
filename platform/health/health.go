package health

import (
	"context"
	"fmt"
)

// Probe 就绪探测接口，由具体依赖实现
type Probe interface {
	Ready(ctx context.Context) error
}

// Checker 聚合多个依赖的就绪状态
type Checker struct {
	probes map[string]Probe
}

// NewChecker 创建健康检查器
func NewChecker() *Checker {
	return &Checker{probes: make(map[string]Probe)}
}

// Register 注册一个命名探测项
func (c *Checker) Register(name string, p Probe) {
	c.probes[name] = p
}

// Liveness 存活检查，不触发任何依赖探测，仅表示进程可响应
func (c *Checker) Liveness() error {
	return nil
}

// Readiness 就绪检查，任一探测失败即返回错误
func (c *Checker) Readiness(ctx context.Context) error {
	for name, p := range c.probes {
		if p == nil {
			continue
		}
		if err := p.Ready(ctx); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}
