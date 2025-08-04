package db

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
	"gorm.io/gorm/logger"
)

// ErrRecordNotFound record not found error
var ErrRecordNotFound = errors.New("record not found")

// LogLevel log level
type LogLevel int

const (
	// Silent silent log level
	Silent LogLevel = iota + 1
	// Error error log level
	Error
	// Warn warn log level
	Warn
	// Info info log level
	Info
)

// Writer log writer interface
type Writer interface {
	Printf(string, ...any)
}

// New initialize logger
func New(writer Writer, config logger.Config) logger.Interface {

	var (
		infoStr      = "[INFO] [%s] [%s] [%s]"
		warnStr      = "[WARN] [%s] [%s] [%s]"
		errStr       = "[ERROR] [%s] [%s] [%s]"
		traceStr     = "[INFO] [%s] [%s] [%.3fms] [rows:%v] %s"
		traceWarnStr = "[WARN] [%s] [%s] [%s] [%.3fms] [rows:%v] %s"
		traceErrStr  = "[ERROR] [%s] [%s] [%.3fms] [rows:%v] %s ,%s"
	)

	return &xlogger{
		Writer:       writer,
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type xlogger struct {
	Writer
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *xlogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l *xlogger) Info(ctx context.Context, msg string, data ...any) {

	return
	if l.LogLevel >= logger.Info {

		infoStr := "[INFO] [%s]"
		if l.LogLevel >= logger.Info {
			l.Printf(infoStr+msg, append([]any{FileWithLineNum()}, data...)...)
		}
	}
}

// Warn print warn messages
func (l *xlogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Warn {
		l.Printf(l.warnStr+msg, append([]any{FileWithLineNum()}, data...)...)
	}
}

// FileWithLineNum return the file name and line number of the current file
func FileWithLineNum() string {
	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.Contains(file, "gorm.io") || strings.HasSuffix(file, "_test.go")) &&
			!strings.HasSuffix(file, ".gen.go") {

			files := strings.Split(file, "/")
			return files[len(files)-1] + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}

// Error print error messages
func (l *xlogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Error {
		l.Printf(l.errStr+msg, append([]any{FileWithLineNum()}, data...)...)
	}
}

// Trace print sql message
//
//nolint:cyclop
func (l *xlogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {

	var requestId = "-"
	if ctx != nil {
		requestId = cast.ToString(ctx.Value("X-Request-Id"))
	}
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceErrStr, FileWithLineNum(), requestId, float64(elapsed.Nanoseconds())/1e6, "-", err, sql)
		} else {
			l.Printf(l.traceErrStr, FileWithLineNum(), requestId, float64(elapsed.Nanoseconds())/1e6, rows, err, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf(l.traceWarnStr, FileWithLineNum(), requestId, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceWarnStr, FileWithLineNum(), requestId, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceStr, FileWithLineNum(), requestId, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceStr, FileWithLineNum(), requestId, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

// ParamsFilter filter params
func (l *xlogger) ParamsFilter(ctx context.Context, sql string, params ...any) (string, []any) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
