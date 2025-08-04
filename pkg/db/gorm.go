package db

import (
	"context"
	"log"
	"os"
	"time"

	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	ck "github.com/silen/nuwa/pkg/cilickhouse"
	"github.com/silen/nuwa/pkg/conf"
	"github.com/silen/nuwa/pkg/logs"
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

func withCtxLogs(printSql bool) logger.Interface {
	var (
		logConfig = logger.Config{
			SlowThreshold:             time.Second,   // 慢 SQL 阈值
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,
			Colorful:                  false, // 禁用彩色打印
		}
	)
	logConfig.LogLevel = logger.Warn
	if printSql {
		logConfig.LogLevel = logger.Info
	}
	return New(
		log.New(os.Stdout, "", log.LstdFlags), // io writer
		logConfig,
	)
}

// mysql 连接
func Mysql(ctx context.Context, config Config) (db *gorm.DB, err error) {

	db, err = gorm.Open(mysql.Open(config.Master), &gorm.Config{
		Logger:                 withCtxLogs(config.PrintSql),
		SkipDefaultTransaction: false,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.Config.GetString("mysql.prefix"),
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

	if len(config.Slave) > 0 {
		replicas := []gorm.Dialector{}
		for _, s := range config.Slave {
			cfg := mysql.Config{
				DSN: s,
			}
			replicas = append(replicas, mysql.New(cfg))
		}

		db.Use(
			dbresolver.Register(dbresolver.Config{
				Sources: []gorm.Dialector{mysql.New(mysql.Config{
					DSN: config.Master,
				})},
				Replicas: replicas,
				Policy:   dbresolver.RandomPolicy{},
			}).
				SetMaxIdleConns(10).
				SetConnMaxLifetime(time.Hour).
				SetMaxOpenConns(200),
		)
	}

	return
}

// mysql 连接
func SQLServer(ctx context.Context, config Config) (db *gorm.DB, err error) {

	db, err = gorm.Open(sqlserver.Open(config.Master), &gorm.Config{
		Logger:                 withCtxLogs(config.PrintSql),
		SkipDefaultTransaction: false,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.Config.GetString("mysql.prefix"),
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

	if len(config.Slave) > 0 {
		replicas := []gorm.Dialector{}
		for _, s := range config.Slave {
			cfg := mysql.Config{
				DSN: s,
			}
			replicas = append(replicas, mysql.New(cfg))
		}

		db.Use(
			dbresolver.Register(dbresolver.Config{
				Sources: []gorm.Dialector{mysql.New(mysql.Config{
					DSN: config.Master,
				})},
				Replicas: replicas,
				Policy:   dbresolver.RandomPolicy{},
			}).
				SetMaxIdleConns(10).
				SetConnMaxLifetime(time.Hour).
				SetMaxOpenConns(200),
		)
	}

	return
}

// clickHouse 连接！
func ClickHouse(ctx context.Context, config ck.ClickHouseConfig) (db *gorm.DB, err error) {
	ck, err := ck.Conn(ctx, config)
	if err != nil {
		return
	}

	db, err = gorm.Open(clickhouse.New(clickhouse.Config{
		Conn: ck, // initialize with existing database conn
	}), &gorm.Config{
		Logger:                 withCtxLogs(config.PrintSql),
		SkipDefaultTransaction: false,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.Config.GetString("mysql.prefix"),
			SingularTable: true,
		},
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})

	return
}
