package shutdown

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
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

// Context 返回在收到 SIGINT / SIGTERM 或父上下文取消时被取消的上下文。
// 调用发需要调用返回的stop函数释放信号监听
func Context(parent context.Context) (context.Context, context.CancelFunc, <-chan os.Signal) {
	ctx, cancelFunc := context.WithCancel(parent)
	ch := Notify()
	go func() {
		select {
		case <-ch:
			cancelFunc()
		case <-ctx.Done():
		}
	}()
	return ctx, cancelFunc, ch
}

// Drain 逆序执行关闭钩子，任一钩子失败即中断并返回错误
func Drain(ctx context.Context, hooks ...Hook) error {
	var errs []error
	for i := len(hooks) - 1; i >= 0; i-- {
		if hooks[i].Fn == nil {
			continue
		}
		if err := hooks[i].Fn(ctx); err != nil {
			errs = append(errs, fmt.Errorf("hook %s: %w", hooks[i].Name, err))
			return err
		}
	}
	return errors.Join(errs...)
}

// DrainWithTimeout 在限定时间内逆序执行关闭钩子
// 超时后返回 context.DeadlineExceeded，已执行与未执行的钩子均不补偿
func DrainWithTimeout(timeout time.Duration, hooks ...Hook) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return Drain(ctx, hooks...)
}
