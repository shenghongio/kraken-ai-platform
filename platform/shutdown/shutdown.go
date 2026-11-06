package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Hook 关闭钩子，按注册顺序的逆序执行
type Hook struct {
	Name string
	Fn   func(ctx context.Context) error
}

// Notify 返回接收 SIGINT / SIGTERM 的通道
func Notify() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return ch
}

// Drain 逆序执行关闭钩子，任一钩子失败即中断并返回错误
func Drain(ctx context.Context, hooks ...Hook) error {
	for i := len(hooks) - 1; i >= 0; i-- {
		if hooks[i].Fn == nil {
			continue
		}
		if err := hooks[i].Fn(ctx); err != nil {
			return err
		}
	}
	return nil
}
