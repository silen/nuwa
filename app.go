package nuwa

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-module/carbon/v2"
	"github.com/spf13/cast"

	"github.com/silen/nuwa/pkg/cache"
	"github.com/silen/nuwa/pkg/conf"
	"github.com/silen/nuwa/pkg/logs"
	"github.com/silen/nuwa/pkg/middleware"
)

var (
	ginEngine       *gin.Engine
	StartServerTime string
)

func init() {
	Engine()
}

func Engine() {
	StartServerTime = time.Now().Format("2006-01-02 15:04:05")

	ginEngine = gin.New()
	ginEngine.Use(gin.Recovery())
	ginEngine.Use(logs.RequestIDMiddleware()) // 添加请求ID中间件
	ginEngine.NoRoute(go404)

	runMode := gin.DebugMode
	if os.Getenv("environment") == "prod" {
		runMode = gin.ReleaseMode
	}

	gin.SetMode(runMode)
	binding.Validator = new(defaultValidator)
	gin.DebugPrintRouteFunc = printRouteFunc
}

func NewEngine() *gin.Engine {

	if ginEngine == nil {
		Engine()
	}

	return ginEngine
}

// 第一个单词首字母小写
func lcfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

var filterRepeat map[string]bool = make(map[string]bool)

// 自定义debug模式路由注册打印
func printRouteFunc(httpMethod, absolutePath, handlerName string, nuHandlers int) {

	//只在开发模式打印
	if !gin.IsDebugging() {
		return
	}

	pathArr := strings.Split(absolutePath, "/")
	for k, v := range pathArr {
		pathArr[k] = lcfirst(v)
	}
	absolutePathLc := strings.Join(pathArr, "/")

	absolutePathTemp := httpMethod + "_" + strings.ToLower(absolutePath)
	if _, exist := filterRepeat[absolutePathTemp]; exist {
		return
	}

	filterRepeat[absolutePathTemp] = true
	format := "%-6s %-25s --> %s (%d handlers)\n"
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(gin.DefaultWriter, "[GIN-debug] "+format, httpMethod, absolutePathLc, handlerName, nuHandlers)

}

// 自定义404日志
func go404(c *gin.Context) {
	c.JSON(http.StatusNotFound, map[string]any{
		"status":  NOT_FOUND,
		"message": "未找到相对应的服务",
	})
}

func initTimeFunc() {
	carbon.SetDefault(carbon.Default{
		Layout:       carbon.DateTimeLayout,
		Timezone:     carbon.PRC,
		WeekStartsAt: carbon.Sunday,
		Locale:       "zh-CN",
	})
}
func Run() {

	initTimeFunc()

	// 初始化缓存
	cache.InitCache()

	serverConf := conf.GetStringMapString("server")
	if cast.ToString(serverConf["port"]) == "" {
		logs.Fatal("配置文件缺失服务端口")
		return
	}

	addr := serverConf["host"] + ":" + serverConf["port"]

	// 注册中间件
	ginEngine.Use(middleware.CompressionMiddleware()) // 压缩中间件应该放在前面
	ginEngine.Use(middleware.PerformanceMonitor())
	ginEngine.Use(middleware.MetricsMiddleware())

	// 注册健康检查和指标端点
	ginEngine.GET("/checkHealth", middleware.HealthCheck)
	ginEngine.GET("/metrics", middleware.MetricsEndpoint)
	ginEngine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "healthy",
		})
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: ginEngine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Fatal("listen: ", err)
		}
	}()

	logs.Info("服务启动成功，PID：", os.Getpid(), " address：", addr, " 当前环境：", os.Getenv("environment"))
	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logs.Info("Shutdown Server ...")

	// 关闭缓存连接
	if cache.GetCache() != nil {
		if err := cache.GetCache().Close(); err != nil {
			logs.Error("Failed to close cache: ", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logs.Error("Server Shutdown:", err)
	}

}

// ErrAbort return json专用错误文本
var errAbort = errors.New("user stop run")

// StopRun stop controller router
func StopRun() {
	panic(errAbort)
}

// RecoverPanic 异常恢复
func RecoverPanic(ctx *gin.Context) {
	xLogs := logs.WithContext(ctx)
	if err := recover(); err != nil {
		if err == errAbort {
			return
		}

		errString := fmt.Sprintf("%s", err)
		if strings.Contains(errString, "write: broken pipe") {
			xLogs.Warn("Broken pipe error: ", errString)
			xLogs.Warn("Request URL: ", ctx.Request.URL.RequestURI())
			return
		}

		// 记录详细的错误信息
		xLogs.Error("Panic occurred in handler")
		xLogs.Error("Request URL: ", ctx.Request.URL.RequestURI())
		xLogs.Error("Request Method: ", ctx.Request.Method)
		xLogs.Error("Request Headers: ", ctx.Request.Header)

		// 获取查询参数和表单参数
		params := ctx.Request.URL.Query()
		if len(params) > 0 {
			xLogs.Error("Query Params: ", params)
		}

		// 注意：这里不直接读取请求体以避免影响后续中间件或处理器
		// 可以记录请求长度等信息
		if ctx.Request.ContentLength > 0 {
			xLogs.Error("Request Content Length: ", ctx.Request.ContentLength)
		}

		xLogs.Error("Panic Error: ", err)

		// 记录调用栈
		stack := string(debug.Stack())
		xLogs.Error("Stack trace: ", stack)

		// 如果上下文已经被写入，则不再尝试写入响应
		if ctx.IsAborted() {
			return
		}

		ctx.JSON(500, map[string]any{
			"status":  SYSTEM_ERROR,
			"message": "Internal server error",
		})
	}
}
