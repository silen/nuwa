package db

import (
	"context"
	"fmt"
	"time"

	"github.com/silen/nuwa/pkg/logs"
	"gorm.io/gorm"
)

// ConnectionPoolMonitor 连接池监控器
type ConnectionPoolMonitor struct {
	db            *gorm.DB
	interval      time.Duration
	alertCallback func(*DBStats)
}

// NewConnectionPoolMonitor 创建新的连接池监控器
func NewConnectionPoolMonitor(db *gorm.DB, interval time.Duration, alertCallback func(*DBStats)) *ConnectionPoolMonitor {
	return &ConnectionPoolMonitor{
		db:            db,
		interval:      interval,
		alertCallback: alertCallback,
	}
}

// Start 开始监控
func (m *ConnectionPoolMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logs.Info("Database connection pool monitoring stopped")
			return
		case <-ticker.C:
			m.checkPoolStatus()
		}
	}
}

// checkPoolStatus 检查连接池状态
func (m *ConnectionPoolMonitor) checkPoolStatus() {
	stats, err := GetDBStats(m.db)
	if err != nil {
		logs.Error("Failed to get database stats: ", err)
		return
	}

	// 记录连接池状态
	logs.WithContext(context.Background()).Infof(
		"DB Pool Stats - MaxOpen: %d, Open: %d, InUse: %d, Idle: %d, WaitCount: %d, WaitDuration: %v",
		stats.MaxOpenConnections,
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		stats.WaitCount,
		stats.WaitDuration,
	)

	// 检查是否需要告警
	m.evaluateAlerts(stats)
}

// evaluateAlerts 评估是否需要触发告警
func (m *ConnectionPoolMonitor) evaluateAlerts(stats *DBStats) {
	thresholds := map[string]float64{
		"usage_ratio": 0.8, // 使用率超过80%时告警
		"wait_count":  100, // 等待次数超过100次时告警
	}

	alerts := []string{}

	// 检查连接使用率
	usageRatio := float64(stats.InUse) / float64(stats.MaxOpenConnections)
	if usageRatio > thresholds["usage_ratio"] {
		alerts = append(alerts, fmt.Sprintf("High connection usage: %.2f%%", usageRatio*100))
	}

	// 检查等待次数
	if stats.WaitCount > int64(thresholds["wait_count"]) {
		alerts = append(alerts, fmt.Sprintf("High wait count: %d", stats.WaitCount))
	}

	// 检查等待时间
	if stats.WaitDuration > 10*time.Second {
		alerts = append(alerts, fmt.Sprintf("Long wait duration: %v", stats.WaitDuration))
	}

	// 检查空闲连接过少
	if stats.Idle < 2 && stats.MaxOpenConnections > 10 {
		alerts = append(alerts, fmt.Sprintf("Low idle connections: %d", stats.Idle))
	}

	// 如果有告警，执行回调
	if len(alerts) > 0 {
		logs.Warn("Database connection pool alerts: ", alerts)
		if m.alertCallback != nil {
			m.alertCallback(stats)
		}
	}
}

// MonitorConnectionPool 便捷函数，开始监控连接池
func MonitorConnectionPool(ctx context.Context, db *gorm.DB, interval time.Duration) {
	monitor := NewConnectionPoolMonitor(db, interval, func(stats *DBStats) {
		// 默认告警回调
		logs.Error("Database connection pool alert triggered!")
		logs.Error("Stats:", *stats)
	})

	// 在goroutine中运行监控
	go monitor.Start(ctx)
}