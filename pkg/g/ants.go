package g

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/spf13/cast"
	"go.uber.org/zap"

	"github.com/silen/nuwa/pkg/logs"
)

//多协程执行器ExecTaskByGoroutine
//内置 错误捕捉
//github.com/panjf2000/ants
//第四个参数可设置 并发执行个数
/*
case:
	cc := []any{1, 2, 3, 4}
	var ctx context.Context
	if c.Ctx != nil {
		ctx = c.Ctx
	}
	g.ExecTaskByGoroutine(ctx, cc, func(i any) {
		logs.Notice(cast.ToInt(i))
		//cast.ToStringMap(i)
		//cast.ToStringMapString(i)
		logs.Notice(i)
	}, 20)

	var ccc []any
	ccc = append(ccc, map[string]any{
		"token":     "fasdfasdfa",
		"pageSize":  10,
		"pageIndex": 1,
		"userId":    111,
	})
	ccc = append(ccc, map[string]any{
		"token":     "fasdfasdfa",
		"pageSize":  10,
		"pageIndex": 2,
		"userId":    111,
	})
	ccc = append(ccc, map[string]any{
		"token":     "fasdfasdfa",
		"pageSize":  10,
		"pageIndex": 3,
		"userId":    111,
	})

	//Map并发安全简易解决
	//var maps sync.Map{}
	var rows []any
	var mutex sync.Mutex
	g.ExecTaskByGoroutine(ctx, ccc, func(i T) {
		mutex.Lock()
		rows = append(rows, i)
		mutex.Unlock()
		//maps.Store(, value any)
	})
*/
func ExecTaskByGoroutine[T any](ctx context.Context, params []T, execFunc func(T), args ...any) (err error) {
	//pool .... how do you do!!!!
	var wg sync.WaitGroup
	size := 10
	if len(args) > 0 {
		size = cast.ToInt(args[0])
	}

	options := ants.Options{
		ExpiryDuration: 30 * time.Second,
		PanicHandler: func(i any) {
			logs.WithContext(ctx).Error("ants goroutine fatal error========", i)
			stack := string(debug.Stack())
			logs.WithContext(ctx).Error(zap.String("debug_stack", stack))
		},
	}

	var p *ants.PoolWithFunc
	// 这里池子是否应该做成全局的
	p, err = ants.NewPoolWithFunc(size, func(i any) {
		defer wg.Done()
		execFunc(params[cast.ToInt(i)])
	}, ants.WithOptions(options))

	if err != nil {
		return err
	}

	defer p.Release()
	for i := 0; i < len(params); i++ {
		wg.Add(1)
		if err := p.Invoke(i); err != nil {
			return err
		}
	}
	wg.Wait()

	return
}

// 某个任务失败 那么不等待其他任务 推出返回主线程
func ExecTaskByGoroutineErrorEnd[T any](ctx context.Context, params []T, execFunc func(T) error, args ...any) (err error) {
	poolNum := 15
	if len(args) > 0 {
		poolNum = cast.ToInt(args[0])
	}
	ch := make(chan struct{})

	var wg sync.WaitGroup
	options := ants.Options{
		ExpiryDuration: 30 * time.Second,
		PanicHandler: func(i any) {
			logs.WithContext(ctx).Error("ants goroutine fatal error========", i)
			stack := string(debug.Stack())
			logs.WithContext(ctx).Error(zap.String("debug_stack", stack))
		},
	}
	p, err := ants.NewPoolWithFunc(poolNum, func(i any) {
		defer wg.Done()
		if execFunc(params[cast.ToInt(i)]) != nil {
			ch <- struct{}{}
		}

	}, ants.WithOptions(options))

	defer p.Release()
	for i := 0; i < len(params); i++ {
		wg.Add(1)
		_ = p.Invoke(i)
	}
	go func() {
		wg.Wait()
		ch <- struct{}{}
		close(ch)
	}()
	<-ch
	return
}

// Go 协程创建封装
func Go(ctx context.Context, f func()) {
	New().Submit(f)
	/*
		go func() {
			//子协程recover，打印子协程的报错并防止因为子协程没recover错误导致的应用崩溃
			defer func() {
				if err := recover(); err != nil {
					logs.WithContext(ctx).Error("goroutine fatal error========", err)
					//生产环境不打印
					//if os.Getenv("environment") != "prod" {
					stack := string(debug.Stack())
					logs.WithContext(ctx).Error(zap.String("debug_stack", stack))
					//}
				}
			}()
			f()
		}()*/
}
