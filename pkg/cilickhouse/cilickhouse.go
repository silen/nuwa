package ck

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/silen/nuwa/pkg/logs"
)

type ClickHouseConfig struct {
	Addr             []string
	Database         string
	Username         string
	Password         string
	MaxExecutionTime int
	Debug            bool
	PrintSql         bool
}

func Conn(ctx context.Context, config ClickHouseConfig) (conn *sql.DB, err error) {

	conn = clickhouse.OpenDB(&clickhouse.Options{
		Addr: config.Addr,
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Settings: clickhouse.Settings{
			//"max_execution_time": maxExecutionTime,
		},
		DialTimeout: 30 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Protocol: clickhouse.HTTP,
		Debug:    config.Debug,
		Debugf: func(format string, v ...any) {
			logs.Debug(format, v)
		},
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
	})

	err = conn.Ping()
	if err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			logs.Error(fmt.Sprintf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace))
		}
		return
	}
	conn.SetMaxIdleConns(5)
	conn.SetMaxOpenConns(10)
	conn.SetConnMaxLifetime(time.Hour)

	return
}
