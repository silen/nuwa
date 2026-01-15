package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/silen/nuwa/pkg/logs"
)

// DBPoolConfig 数据库连接池配置
type DBPoolConfig struct {
	MaxIdleConns int           // 最大空闲连接数
	MaxOpenConns int           // 最大打开连接数
	ConnLifetime time.Duration // 连接最大生命周期
}

// DefaultDBPoolConfig 默认数据库连接池配置
var DefaultDBPoolConfig = DBPoolConfig{
	MaxIdleConns: 10,
	MaxOpenConns: 100,
	ConnLifetime: time.Hour,
}

// ApplyPoolConfig 应用连接池配置到数据库实例
func ApplyPoolConfig(db *gorm.DB, config DBPoolConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)

	// 设置最大打开连接数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)

	// 设置连接最大生命周期
	sqlDB.SetConnMaxLifetime(config.ConnLifetime)

	return nil
}

// ApplyDefaultPoolConfig 应用默认连接池配置
func ApplyDefaultPoolConfig(db *gorm.DB) error {
	return ApplyPoolConfig(db, DefaultDBPoolConfig)
}

// GetDBStats 获取数据库连接池统计信息
func GetDBStats(db *gorm.DB) (*DBStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return &DBStats{
		MaxOpenConnections: int64(stats.MaxOpenConnections),
		OpenConnections:    int64(stats.OpenConnections),
		InUse:              int64(stats.InUse),
		Idle:               int64(stats.Idle),
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}, nil
}

// DBStats 数据库连接池统计信息
type DBStats struct {
	MaxOpenConnections int64         // 最大打开连接数
	OpenConnections    int64         // 当前打开连接数
	InUse              int64         // 正在使用的连接数
	Idle               int64         // 空闲连接数
	WaitCount          int64         // 等待连接次数
	WaitDuration       time.Duration // 等待连接总时间
	MaxIdleClosed      int64         // 因空闲而关闭的连接数
	MaxLifetimeClosed  int64         // 因达到最大生命周期而关闭的连接数
}

// MonitorDBPool 监控数据库连接池状态
func MonitorDBPool(ctx context.Context, db *gorm.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats, err := GetDBStats(db)
			if err != nil {
				continue
			}

			// 记录连接池状态
			logs.Info("Database Pool Stats:",
				"MaxOpenConnections:", stats.MaxOpenConnections,
				"OpenConnections:", stats.OpenConnections,
				"InUse:", stats.InUse,
				"Idle:", stats.Idle,
				"WaitCount:", stats.WaitCount,
				"WaitDuration:", stats.WaitDuration,
			)

			// 如果等待次数过多，发出警告
			if stats.WaitCount > 100 {
				logs.Warn("High database connection wait count:", stats.WaitCount)
			}

			// 如果正在使用的连接接近最大连接数，发出警告
			if float64(stats.InUse)/float64(stats.MaxOpenConnections) > 0.8 {
				logs.Warn("High database connection usage:", float64(stats.InUse)/float64(stats.MaxOpenConnections)*100, "%")
			}
		}
	}
}
