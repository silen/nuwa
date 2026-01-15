package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/silen/nuwa/pkg/logs"
)

// PerformanceMonitor 性能监控中间件
func PerformanceMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 处理请求
		c.Next()
		
		// 计算处理时间
		duration := time.Since(start)
		
		// 记录性能指标
		logs.WithGinContext(c).Infof(
			"Performance Metrics - Path: %s, Method: %s, Status: %d, Duration: %v, ClientIP: %s",
			c.Request.URL.Path,
			c.Request.Method,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
		)
		
		// 如果处理时间超过阈值，记录警告
		if duration > 1*time.Second {
			logs.WithGinContext(c).Warnf(
				"SLOW REQUEST - Path: %s, Method: %s, Duration: %v",
				c.Request.URL.Path,
				c.Request.Method,
				duration,
			)
		}
	}
}

// MetricsCollector 收集性能指标
type MetricsCollector struct {
	requestCount   int64
	errorCount     int64
	totalDuration  time.Duration
	maxDuration    time.Duration
	minDuration    time.Duration
}

// GlobalMetrics 全局指标收集器
var GlobalMetrics = &MetricsCollector{
	minDuration: time.Duration(1<<63 - 1), // 初始化为最大值
}

// MetricsMiddleware 指标收集中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 处理请求
		c.Next()
		
		// 计算处理时间
		duration := time.Since(start)
		
		// 更新指标
		GlobalMetrics.updateMetrics(duration, c.Writer.Status())
	}
}

// updateMetrics 更新指标
func (m *MetricsCollector) updateMetrics(duration time.Duration, statusCode int) {
	m.requestCount++
	
	if statusCode >= 400 {
		m.errorCount++
	}
	
	m.totalDuration += duration
	
	if duration > m.maxDuration {
		m.maxDuration = duration
	}
	
	if duration < m.minDuration {
		m.minDuration = duration
	}
}

// GetMetrics 获取指标统计
func (m *MetricsCollector) GetMetrics() map[string]interface{} {
	avgDuration := time.Duration(0)
	if m.requestCount > 0 {
		avgDuration = m.totalDuration / time.Duration(m.requestCount)
	}
	
	errorRate := 0.0
	if m.requestCount > 0 {
		errorRate = float64(m.errorCount) / float64(m.requestCount) * 100
	}
	
	return map[string]interface{}{
		"request_count":        m.requestCount,
		"error_count":          m.errorCount,
		"error_rate_percent":   fmt.Sprintf("%.2f", errorRate),
		"total_duration":       m.totalDuration,
		"average_duration":     avgDuration,
		"max_duration":         m.maxDuration,
		"min_duration":         m.minDuration,
		"uptime":               time.Since(startTime).String(),
	}
}

// startTime 服务启动时间
var startTime = time.Now()

// HealthCheck 健康检查端点
func HealthCheck(c *gin.Context) {
	// 这里可以添加对数据库、缓存等依赖服务的检查
	healthStatus := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
		"metrics":   GlobalMetrics.GetMetrics(),
	}
	
	c.JSON(200, healthStatus)
}

// MetricsEndpoint 指标端点
func MetricsEndpoint(c *gin.Context) {
	metrics := GlobalMetrics.GetMetrics()
	
	c.JSON(200, map[string]interface{}{
		"status": "success",
		"data":   metrics,
	})
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	type RequestRecord struct {
		count int
		time  time.Time
	}
	
	clients := make(map[string]*RequestRecord)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		now := time.Now()
		record, exists := clients[clientIP]
		
		if !exists || now.Sub(record.time) > window {
			clients[clientIP] = &RequestRecord{
				count: 1,
				time:  now,
			}
			c.Next()
			return
		}
		
		if record.count >= maxRequests {
			c.JSON(429, map[string]interface{}{
				"status":  429,
				"message": "Too many requests",
			})
			c.Abort()
			return
		}
		
		record.count++
		c.Next()
	}
}