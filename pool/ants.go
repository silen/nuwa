package pool

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/silen/nuwa/logs"
)

const (
	defaultPoolSize       = 10
	defaultExpiryDuration = 30 * time.Second
)

// Options 协程池可选配置，通过 WithOptions / WithPoolSize / WithExpiryDuration / WithTimeout 传入；零值项使用默认。
type Options struct {
	PoolSize       int           // 协程池大小，≤0 表示使用默认 10
	ExpiryDuration time.Duration // worker 过期时间，≤0 表示使用默认 30s
	Timeout        time.Duration // 整体超时，>0 时使用；0 表示不设超时（仅“首个失败即取消”）
}

// Option 为可选函数，用于修改 *Options（与 ants 的 Option 模式一致）。
type Option func(opts *Options)

// loadOptions 以默认值初始化，再按顺序应用 options。
func loadOptions(options ...Option) *Options {
	opts := &Options{PoolSize: defaultPoolSize, ExpiryDuration: defaultExpiryDuration}
	for _, fn := range options {
		fn(opts)
	}
	return opts
}

// WithOptions 接受完整 Options，覆盖默认（与 ants 的 WithOptions 一致）。
func WithOptions(options Options) Option {
	return func(opts *Options) {
		if options.PoolSize > 0 {
			opts.PoolSize = options.PoolSize
		}
		if options.ExpiryDuration > 0 {
			opts.ExpiryDuration = options.ExpiryDuration
		}
		if options.Timeout > 0 {
			opts.Timeout = options.Timeout
		}
	}
}

// WithPoolSize 设置协程池大小。
func WithPoolSize(n int) Option {
	return func(opts *Options) {
		if n > 0 {
			opts.PoolSize = n
		}
	}
}

// WithExpiryDuration 设置 worker 过期时间。
func WithExpiryDuration(d time.Duration) Option {
	return func(opts *Options) {
		if d > 0 {
			opts.ExpiryDuration = d
		}
	}
}

// WithTimeout 设置整体超时；用于 ExecTaskByGoroutineErrorEnd 时等价原 ExecTasksWithTimeOut。
func WithTimeout(d time.Duration) Option {
	return func(opts *Options) {
		if d > 0 {
			opts.Timeout = d
		}
	}
}

// ExecTaskByGoroutine 使用 ants 协程池并发执行任务。options 可选：WithPoolSize、WithExpiryDuration、WithOptions。
// 会等待全部任务结束，返回首个执行错误（若有）；不因单个失败而中断其他任务。
func ExecTaskByGoroutine[T any](
	ctx context.Context,
	params []T,
	execFunc func(T) error,
	options ...Option,
) error {
	if execFunc == nil {
		return errors.New("exec func is nil")
	}

	ctx = normalizeContext(ctx)
	opts := loadOptions(options...)
	var wg sync.WaitGroup
	var errOnce sync.Once
	var retErr error

	pool, err := ants.NewPoolWithFunc(opts.PoolSize, func(i any) {
		defer wg.Done()
		param := i.(T)
		if ctx.Err() != nil {
			return
		}
		if err := execFunc(param); err != nil {
			errOnce.Do(func() { retErr = err })
		}
	}, ants.WithOptions(ants.Options{
		ExpiryDuration: opts.ExpiryDuration,
		PanicHandler:   makePanicHandler(ctx, &errOnce, &retErr, nil),
	}))
	if err != nil {
		return err
	}
	defer pool.Release()

	for _, param := range params {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		if err := pool.Invoke(param); err != nil {
			wg.Done()
			errOnce.Do(func() { retErr = err })
			break
		}
	}
	wg.Wait()
	if retErr == nil && ctx.Err() != nil {
		return ctx.Err()
	}
	return retErr
}

// ExecTaskByGoroutineErrorEnd 并发执行任务，首个失败或 panic 时取消 context、停止提交并返回错误。
// 支持 WithTimeout(d)：d>0 时整体超时，超时且无任务错误时返回 ctx.Err()；不设 WithTimeout 时仅“首个失败即取消”。
// options 可选：WithPoolSize、WithExpiryDuration、WithTimeout、WithOptions。
func ExecTaskByGoroutineErrorEnd[T any](
	parentCtx context.Context,
	params []T,
	execFunc func(T) error,
	options ...Option,
) error {
	if execFunc == nil {
		return errors.New("exec func is nil")
	}

	opts := loadOptions(options...)
	var ctx context.Context
	var cancel context.CancelFunc
	parentCtx = normalizeContext(parentCtx)
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(parentCtx, opts.Timeout)
	} else {
		ctx, cancel = context.WithCancel(parentCtx)
	}
	defer cancel()

	var wg sync.WaitGroup
	var errOnce sync.Once
	var retErr error

	pool, err := ants.NewPoolWithFunc(opts.PoolSize, func(i any) {
		defer wg.Done()
		param := i.(T)
		if ctx.Err() != nil {
			return
		}
		if err := execFunc(param); err != nil {
			errOnce.Do(func() {
				retErr = err
				cancel()
			})
		}
	}, ants.WithOptions(ants.Options{
		ExpiryDuration: opts.ExpiryDuration,
		PanicHandler:   makePanicHandler(ctx, &errOnce, &retErr, cancel),
	}))
	if err != nil {
		return err
	}
	defer pool.Release()

OuterLoop:
	for _, param := range params {
		if ctx.Err() != nil {
			break OuterLoop
		}
		wg.Add(1)
		if err := pool.Invoke(param); err != nil {
			wg.Done()
			errOnce.Do(func() {
				retErr = err
				cancel()
			})
			break OuterLoop
		}
	}
	wg.Wait()

	if retErr == nil && ctx.Err() != nil {
		return ctx.Err()
	}
	return retErr
}

// makePanicHandler 返回 ants 的 PanicHandler，将 panic 转为 retErr 并记录日志；若 cancel 非 nil 则同时取消 context。
func makePanicHandler(ctx context.Context, errOnce *sync.Once, retErr *error, cancel func()) func(any) {
	return func(i any) {
		errOnce.Do(func() {
			*retErr = fmt.Errorf("panic: %v", i)
			if cancel != nil {
				cancel()
			}
		})
		logPoolPanic(ctx, "ants panic", i)
	}
}

// Go 在 ants 池中安全执行 task，带 context 与 panic 恢复；panic 会记录日志并返回 error。
func Go(ctx context.Context, task func()) error {
	if task == nil {
		return errors.New("task is nil")
	}

	ctx = normalizeContext(ctx)
	wrapped := func() {
		defer func() {
			if r := recover(); r != nil {
				logPoolPanic(ctx, "panic in ants task", r)
			}
		}()
		task()
	}

	if err := New().Submit(wrapped); err == nil {
		return nil
	}

	go wrapped()
	return nil
}

func logPoolPanic(ctx context.Context, message string, recovered any) {
	logs.WithContext(normalizeContext(ctx)).WithFields(map[string]any{
		"panic": recovered,
		"stack": string(debug.Stack()),
	}).Error(message)
}

func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
