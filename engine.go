package nuwa

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	ginbinding "github.com/gin-gonic/gin/binding"
)

var (
	defaultEngineMu sync.Mutex
	defaultEngine   *gin.Engine

	// StartServerTime is exposed for packages that still attach runtime metadata
	// to response headers. It is refreshed whenever a new engine is created.
	StartServerTime string

	filterRepeatMu sync.Mutex
	filterRepeat   = make(map[string]bool)
)

// New creates a configured Gin engine for Nuwa's root runtime.
func New() *gin.Engine {
	SetMode(Mode())

	StartServerTime = time.Now().Format("2006-01-02 15:04:05")
	resetRoutePrintFilter()

	gin.DebugPrintRouteFunc = printRouteFunc
	ginbinding.Validator = Validator()

	engine := gin.New()
	engine.Use(Recovery())
	engine.NoRoute(notFound)
	return engine
}

// Default returns the shared Nuwa engine instance.
func Default() *gin.Engine {
	defaultEngineMu.Lock()
	defer defaultEngineMu.Unlock()

	if defaultEngine == nil {
		defaultEngine = New()
	}

	return defaultEngine
}

// Engine returns the shared Nuwa engine instance.
func Engine() *gin.Engine {
	return Default()
}

// SetDefaultEngine replaces the shared Nuwa engine instance.
func SetDefaultEngine(engine *gin.Engine) {
	if engine == nil {
		engine = New()
	}

	defaultEngineMu.Lock()
	defaultEngine = engine
	defaultEngineMu.Unlock()
}

func resetRoutePrintFilter() {
	filterRepeatMu.Lock()
	filterRepeat = make(map[string]bool)
	filterRepeatMu.Unlock()
}

func lcfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

func printRouteFunc(httpMethod, absolutePath, handlerName string, nuHandlers int) {
	if !gin.IsDebugging() {
		return
	}

	pathArr := strings.Split(absolutePath, "/")
	for k, v := range pathArr {
		pathArr[k] = lcfirst(v)
	}

	absolutePathTemp := httpMethod + "_" + strings.ToLower(absolutePath)

	filterRepeatMu.Lock()
	defer filterRepeatMu.Unlock()
	if _, exist := filterRepeat[absolutePathTemp]; exist {
		return
	}

	filterRepeat[absolutePathTemp] = true
	format := "%-6s %-25s --> %s (%d handlers)\n"
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}

	fmt.Fprintf(
		gin.DefaultWriter,
		"[GIN-debug] "+format,
		httpMethod,
		strings.Join(pathArr, "/"),
		handlerName,
		nuHandlers,
	)
}

func notFound(c *gin.Context) {
	writeJSON(c, http.StatusNotFound, Response{
		Status:  NOT_FOUND,
		Message: "未找到相对应的服务",
	})
}
