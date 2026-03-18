package database

import (
	"context"
	"time"

	gormclickhouse "gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	nuwaconfig "github.com/silen/nuwa/config"
	"github.com/silen/nuwa/database/clickhouse"
	"github.com/silen/nuwa/logs"
)

/*
Config.Master = "root:123456@tcp(127.0.0.1:3306)/a0001_chat?charset=utf8mb4&parseTime=True&loc=Local"

		Config.Slave = []string{
		"root:123456@tcp(127.0.0.1:3307)/a0001_chat?charset=utf8mb4&parseTime=True&loc=Local",
		"root:123456@tcp(127.0.0.1:3308)/a0001_chat?charset=utf8mb4&parseTime=True&loc=Local"
	}
*/
type Config struct {
	Master   string
	Slave    []string
	PrintSql bool
}

func withCtxLogs(printSql bool) gormlogger.Interface {
	var (
		logConfig = gormlogger.Config{
			SlowThreshold:             time.Second,       // 慢 SQL 阈值
			LogLevel:                  gormlogger.Silent, // Log level
			IgnoreRecordNotFoundError: true,
			Colorful:                  false, // 禁用彩色打印
		}
	)
	logConfig.LogLevel = gormlogger.Warn
	if printSql {
		logConfig.LogLevel = gormlogger.Info
	}
	return New(logConfig)
}

// mysql 连接
func Mysql(ctx context.Context, cfg Config) (db *gorm.DB, err error) {

	db, err = gorm.Open(mysql.Open(cfg.Master), &gorm.Config{
		Logger:                 withCtxLogs(cfg.PrintSql),
		SkipDefaultTransaction: false,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   nuwaconfig.Config.GetString("mysql.prefix"),
			SingularTable: true,
		},
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		logs.WithContext(ctx).Error("mysql conn error :", err.Error())
		return
	}

	if len(cfg.Slave) > 0 {
		replicas := []gorm.Dialector{}
		for _, s := range cfg.Slave {
			cfg := mysql.Config{
				DSN: s,
			}
			replicas = append(replicas, mysql.New(cfg))
		}

		err = db.Use(
			dbresolver.Register(dbresolver.Config{
				Sources: []gorm.Dialector{mysql.New(mysql.Config{
					DSN: cfg.Master,
				})},
				Replicas: replicas,
				Policy:   dbresolver.RandomPolicy{},
			}).
				SetMaxIdleConns(10).
				SetConnMaxLifetime(time.Hour).
				SetMaxOpenConns(200),
		)
		if err != nil {
			logs.WithContext(ctx).Error("mysql replica resolver error :", err.Error())
			return nil, err
		}
	}

	return
}

// mysql 连接
func SQLServer(ctx context.Context, cfg Config) (db *gorm.DB, err error) {

	db, err = gorm.Open(sqlserver.Open(cfg.Master), &gorm.Config{
		Logger:                 withCtxLogs(cfg.PrintSql),
		SkipDefaultTransaction: false,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   nuwaconfig.Config.GetString("mysql.prefix"),
			SingularTable: true,
		},
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		logs.WithContext(ctx).Error("sqlserver conn error :", err.Error())
		return
	}

	if len(cfg.Slave) > 0 {
		replicas := []gorm.Dialector{}
		for _, s := range cfg.Slave {
			cfg := sqlserver.Config{
				DSN: s,
			}
			replicas = append(replicas, sqlserver.New(cfg))
		}

		err = db.Use(
			dbresolver.Register(dbresolver.Config{
				Sources: []gorm.Dialector{sqlserver.New(sqlserver.Config{
					DSN: cfg.Master,
				})},
				Replicas: replicas,
				Policy:   dbresolver.RandomPolicy{},
			}).
				SetMaxIdleConns(10).
				SetConnMaxLifetime(time.Hour).
				SetMaxOpenConns(200),
		)
		if err != nil {
			logs.WithContext(ctx).Error("sqlserver replica resolver error :", err.Error())
			return nil, err
		}
	}

	return
}

// ClickHouse opens a GORM database using an existing ClickHouse SQL connection.
func ClickHouse(ctx context.Context, cfg clickhouse.Config) (db *gorm.DB, err error) {
	conn, err := clickhouse.Open(ctx, cfg)
	if err != nil {
		return
	}

	db, err = gorm.Open(gormclickhouse.New(gormclickhouse.Config{
		Conn: conn,
	}), &gorm.Config{
		Logger:                 withCtxLogs(cfg.PrintSQL),
		SkipDefaultTransaction: false,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   nuwaconfig.Config.GetString("mysql.prefix"),
			SingularTable: true,
		},
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})

	return
}
