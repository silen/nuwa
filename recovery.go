package nuwa

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/silen/nuwa/logs"
)

// Recovery returns the standard Nuwa panic recovery middleware.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		handlePanic(c, recovered)
	})
}

// RecoverPanic recovers panics when Nuwa helpers are used outside Nuwa's engine.
func RecoverPanic(c *gin.Context) {
	if recovered := recover(); recovered != nil {
		handlePanic(c, recovered)
	}
}

func handlePanic(c *gin.Context, recovered any) {
	if recovered == nil {
		return
	}

	xLogs := logs.WithContext(c)
	errString := fmt.Sprintf("%v", recovered)
	if strings.Contains(errString, "write: broken pipe") {
		xLogs.Warn(errString)
		xLogs.Warn("the broken pipe request url is ", c.Request.URL.RequestURI())
		c.Abort()
		return
	}

	xLogs.Error("the request url is ", c.Request.URL.RequestURI(), c.Request.Method)
	xLogs.Error("the request params are ", c.Request.Form)
	xLogs.Error("handler crashed with error", recovered)

	for i := 1; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		xLogs.Error(fmt.Sprintf("%s:%d", file, line))
	}

	xLogs.Error(zap.String("debug_stack", string(debug.Stack())))
	writeJSON(c, http.StatusInternalServerError, Response{
		Status:  SYSTEM_ERROR,
		Message: "internal server error",
	})
}
