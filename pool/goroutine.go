package pool

import (
	"context"
	"time"

	"github.com/panjf2000/ants/v2"
)

const (
	// DefaultAntsPoolSize sets up the capacity of worker pool, 256 * 1024.
	DefaultAntsPoolSize = 1 << 18

	// ExpiryDuration is the interval time to clean up those expired workers.
	ExpiryDuration = 10 * time.Second

	// Nonblocking decides what to do when submitting a new task to a full worker pool: waiting for a available worker
	// or returning nil directly.
	Nonblocking = true
)

// Pool is the alias of ants.Pool.
type Pool = ants.Pool

var (
	globalPool *Pool
)

func init() {
	// It releases the default pool from ants.
	ants.Release()

	//全局定义！
	globalPool = Default()
}

// 自定一个线程池 执行任务！
//
//	gg := Default()
//	gg.Submit(func() {
//		to do...
//	})
//	gg.Release()
func Default() *Pool {
	options := ants.Options{
		ExpiryDuration: ExpiryDuration,
		Nonblocking:    Nonblocking,
		PanicHandler: func(i any) {
			logPoolPanic(context.Background(), "ants goroutine fatal error========", i)
		},
	}

	defaultAntsPool, _ := ants.NewPool(DefaultAntsPoolSize, ants.WithOptions(options))
	return defaultAntsPool
}

// 使用全局线程池 执行任务
//
//	New().Submit(func() {
//		 to do...
//	})
func New() *Pool {
	if globalPool == nil || globalPool.IsClosed() {
		globalPool = Default()
	}
	return globalPool
}
