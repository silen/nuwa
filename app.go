package nuwa

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-module/carbon/v2"
	"github.com/spf13/cast"
	"go.uber.org/zap"

	"github.com/silen/nuwa/pkg/conf"
	"github.com/silen/nuwa/pkg/logs"
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

	serverConf := conf.Config.GetStringMapString("server")
	if cast.ToString(serverConf["port"]) == "" {
		logs.Fatal("配置文件缺失服务端口")
		return
	}

	addr := serverConf["host"] + ":" + serverConf["port"]

	ginEngine.GET("/checkHealth", func(ginC *gin.Context) {
		ginC.String(200, "ok")
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

	logs.Info("服务启动成功，PID：", os.Getegid(), " address：", addr, " 当前环境：", os.Getenv("environment"))
	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logs.Info("Shutdown Server ...")

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
			xLogs.Warn(errString)
			xLogs.Warn("the broken pipe request url is ", ctx.Request.URL.RequestURI())
			return
		}

		//var stack string
		xLogs.Error("the request url is ", ctx.Request.URL.RequestURI(), ctx.Request.Method)
		xLogs.Error("the request params are ", ctx.Request.Form)
		xLogs.Error("Handler crashed with error", err)
		for i := 1; ; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			xLogs.Error(fmt.Sprintf("%s:%d", file, line))
			//因为kibana的日志是按换行符分隔的，所以stack打在生产也用不了
			//stack = fmt.Sprintln(stack + fmt.Sprintf("%s:%d", file, line))
		}
		//xLogs.Error(ctx, stack)
		stack := string(debug.Stack())
		xLogs.Error(zap.String("debug_stack", stack))
		ctx.JSON(500, map[string]any{
			"status":  SYSTEM_ERROR,
			"message": "internal server error",
		})

	}
}
