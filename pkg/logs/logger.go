package logs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

var (
	defaultLogger = &logrus.Logger{}
	//DefaultLogger = &logrus.Logger{}
)

// ContextKey 用于在上下文中存储键值
type ContextKey string

const (
	RequestIDKey ContextKey = "X-Request-Id"
)

// MyFormatter  log format definition
type MyFormatter struct {
}

// Format ...
func (s *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Local().Format("2006-01-02 15:04:05")
	var line int
	var file string

	if entry.HasCaller() {
		file = filepath.Base(entry.Caller.File)
		line = entry.Caller.Line
	}

	var requestId = "-"
	if entry.Context != nil {
		requestId = cast.ToString(entry.Context.Value(RequestIDKey))
		if requestId == "" {
			requestId = "-"
		}
	}

	msg := fmt.Sprintf("%s [%s] [%s:%d] [%s] %s\n", timestamp, strings.ToUpper(entry.Level.String()), file, line, requestId, entry.Message)
	return []byte(msg), nil
}

func init() {
	defaultLogger = &logrus.Logger{
		Out:   os.Stderr,
		Hooks: make(logrus.LevelHooks),
		Level: logrus.TraceLevel,
	}

	defaultLogger.SetReportCaller(true)
	defaultLogger.SetFormatter(&MyFormatter{})
	// defaultLogger.SetOutput(os.Stdout)
}

// WithContext 从上下文创建带请求ID的日志条目
func WithContext(ctx context.Context) *logrus.Entry {
	return defaultLogger.WithContext(ctx)
}

// WithGinContext 从Gin上下文创建带请求ID的日志条目
func WithGinContext(c *gin.Context) *logrus.Entry {
	requestId := GetRequestIDFromContext(c)
	ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestId)
	c.Request = c.Request.WithContext(ctx)
	return defaultLogger.WithContext(ctx)
}

// GetRequestIDFromContext 从Gin上下文获取请求ID
func GetRequestIDFromContext(c *gin.Context) string {
	// 首先检查上下文中是否已存在请求ID
	if reqID, exists := c.Get(string(RequestIDKey)); exists {
		if id, ok := reqID.(string); ok {
			return id
		}
	}

	// 检查请求头中是否有请求ID
	if reqID := c.GetHeader(string(RequestIDKey)); reqID != "" {
		return reqID
	}

	// 生成新的请求ID
	reqID := generateRequestID()
	c.Set(string(RequestIDKey), reqID)
	return reqID
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return uuid.New().String()
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取或生成请求ID
		requestID := GetRequestIDFromContext(c)

		// 将请求ID添加到响应头中，便于调试
		c.Header(string(RequestIDKey), requestID)

		c.Next()
	}
}

// Trace ..
func Trace(args ...interface{}) {
	defaultLogger.Trace(args...)
}

// Tracef ..
func Tracef(format string, args ...interface{}) {
	defaultLogger.Tracef(format, args...)
}

// Debug ..
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

// Debugf ..
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Print ..
func Print(args ...interface{}) {
	defaultLogger.Print(args...)
}

// Printf ..
func Printf(format string, args ...interface{}) {
	defaultLogger.Printf(format, args...)
}

// Info ..
func Info(args ...interface{}) {
	//log.Println(args)
	defaultLogger.Info(args...)
}

// Infof ..
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Warn ..
func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

// Warnf ..
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Warning ..
func Warning(args ...interface{}) {
	defaultLogger.Warning(args...)
}

// Warningf ..
func Warningf(format string, args ...interface{}) {
	defaultLogger.Warningf(format, args...)
}

// Error ..
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Errorf ..
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatal ..
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Fatalf ..
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// Panic ..
func Panic(args ...interface{}) {
	defaultLogger.Panic(args...)
}

// Panicf ..
func Panicf(format string, args ...interface{}) {
	defaultLogger.Panicf(format, args...)
}
