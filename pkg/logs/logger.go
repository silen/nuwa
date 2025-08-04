package logs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

var (
	defaultLogger = &logrus.Logger{}
	//DefaultLogger = &logrus.Logger{}
)

// MyFormatter  log format definition
type MyFormatter struct {
}

// Format ...
func (s *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Local().Format("2006-01-02 15:04:05")
	var len int
	var file string

	if entry.HasCaller() {
		file = filepath.Base(entry.Caller.File)
		len = entry.Caller.Line
	}

	var requestId = "-"
	if entry.Context != nil {
		requestId = cast.ToString(entry.Context.Value("X-Request-Id"))
	}

	msg := fmt.Sprintf("%s [%s] [%s:%d] [%s] %s\n", timestamp, strings.ToUpper(entry.Level.String()), file, len, requestId, entry.Message)
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

func WithContext(ctx context.Context) *logrus.Entry {
	return defaultLogger.WithContext(ctx)
}

// Trace ..
func Trace(args ...interface{}) {
	defaultLogger.Trace(args...)
}

// Debug ..
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

// Print ..
func Print(args ...interface{}) {
	defaultLogger.Print(args...)
}

// Info ..
func Info(args ...interface{}) {
	//log.Println(args)
	defaultLogger.Info(args...)
}

// Warn ..
func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

// Warning ..
func Warning(args ...interface{}) {
	defaultLogger.Warning(args...)
}

// Error ..
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Fatal ..
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Panic ..
func Panic(args ...interface{}) {
	defaultLogger.Panic(args...)
}
