package nuwa

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon/v2"
	"github.com/spf13/cast"

	"github.com/silen/nuwa/config"
	"github.com/silen/nuwa/logs"
)

var healthRouteRegistry sync.Map

// Run starts the shared Nuwa engine and exits the process on failure.
func Run() {
	if err := RunE(); err != nil {
		logs.Fatal(err)
	}
}

// RunE starts the shared Nuwa engine with signal-aware shutdown handling.
func RunE() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return RunContext(ctx)
}

// RunContext starts the shared Nuwa engine with the provided context.
func RunContext(ctx context.Context) error {
	return RunEngineContext(ctx, Default())
}

// RunEngineContext starts the provided engine with the provided context.
func RunEngineContext(ctx context.Context, engine *gin.Engine) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if engine == nil {
		return errors.New("engine is nil")
	}

	initTime()

	if err := config.Load(); err != nil {
		return fmt.Errorf("配置初始化失败: %w", err)
	}

	addr, err := serverAddress()
	if err != nil {
		return err
	}

	attachHealthRoute(engine)

	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen server failed: %w", err)
	}

	serveErrCh := make(chan error, 1)
	go func() {
		if serveErr := srv.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			serveErrCh <- serveErr
			return
		}
		close(serveErrCh)
	}()

	logs.Info("服务启动成功，PID：", os.Getpid(), " address：", addr, " 当前环境：", os.Getenv(environmentKey))

	select {
	case <-ctx.Done():
	case serveErr := <-serveErrCh:
		if serveErr != nil {
			return fmt.Errorf("serve server failed: %w", serveErr)
		}
		return nil
	}

	logs.Info("Shutdown Server ...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}

func initTime() {
	carbon.SetDefault(carbon.Default{
		Layout:       carbon.DateTimeLayout,
		Timezone:     carbon.PRC,
		WeekStartsAt: carbon.Sunday,
		Locale:       "zh-CN",
	})
}

func serverAddress() (string, error) {
	serverConf := config.Config.GetStringMapString("server")
	port := cast.ToString(serverConf["port"])
	if port == "" {
		return "", errors.New("配置文件缺失服务端口")
	}

	return net.JoinHostPort(serverConf["host"], port), nil
}

func attachHealthRoute(engine *gin.Engine) {
	if _, loaded := healthRouteRegistry.LoadOrStore(engine, struct{}{}); loaded {
		return
	}

	engine.GET("/checkHealth", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
}
