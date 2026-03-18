package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	clickhousedriver "github.com/ClickHouse/clickhouse-go/v2"

	"github.com/silen/nuwa/logs"
)

// Config contains ClickHouse connection settings.
type Config struct {
	Addr             []string
	Database         string
	Username         string
	Password         string
	MaxExecutionTime int
	Debug            bool
	PrintSQL         bool
}

// Open establishes a ClickHouse sql.DB handle.
func Open(ctx context.Context, config Config) (conn *sql.DB, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	conn = clickhousedriver.OpenDB(&clickhousedriver.Options{
		Addr: config.Addr,
		Auth: clickhousedriver.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Settings: clickhousedriver.Settings{
			//"max_execution_time": maxExecutionTime,
		},
		DialTimeout: 30 * time.Second,
		Compression: &clickhousedriver.Compression{
			Method: clickhousedriver.CompressionLZ4,
		},
		Protocol: clickhousedriver.HTTP,
		Debug:    config.Debug,
		Debugf: func(format string, v ...any) {
			logs.Debug(format, v)
		},
		ConnOpenStrategy:     clickhousedriver.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
	})

	err = conn.PingContext(ctx)
	if err != nil {
		if exception, ok := err.(*clickhousedriver.Exception); ok {
			logs.Error(fmt.Sprintf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace))
		}
		return
	}
	conn.SetMaxIdleConns(5)
	conn.SetMaxOpenConns(10)
	conn.SetConnMaxLifetime(time.Hour)

	return
}
