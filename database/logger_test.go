package database

import (
	"bytes"
	"testing"
	"time"

	gormlogger "gorm.io/gorm/logger"

	"github.com/silen/nuwa/logs"
)

func TestLoggerTraceWritesThroughSharedLogs(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	originalOutput := logs.Logger().Out
	t.Cleanup(func() {
		logs.SetOutput(originalOutput)
	})

	logs.SetOutput(buf)

	logger := New(gormlogger.Config{
		SlowThreshold: time.Second,
		LogLevel:      gormlogger.Info,
	})

	logger.Trace(nil, time.Now().Add(-5*time.Millisecond), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	if buf.Len() == 0 {
		t.Fatalf("expected trace log to be written")
	}
}
