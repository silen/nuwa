package middleware

import (
	"compress/gzip"
	"strings"

	"github.com/gin-gonic/gin"
)

// CompressionMiddleware 响应压缩中间件
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查客户端是否支持gzip压缩
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			// 客户端不支持gzip，直接处理请求
			c.Next()
			return
		}

		// 包装响应writer以支持gzip压缩
		gzipWriter := &gzipResponseWriter{
			ResponseWriter: c.Writer,
			writer:         gzip.NewWriter(c.Writer),
		}
		defer gzipWriter.writer.Close()

		// 替换响应writer
		c.Writer = gzipWriter

		// 设置响应头
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		c.Next()
	}
}

// gzipResponseWriter 实现gin.ResponseWriter接口，支持gzip压缩
type gzipResponseWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// Write 重写Write方法，使用gzip压缩
func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	// 检查内容类型，只对特定类型进行压缩
	contentType := g.Header().Get("Content-Type")
	if contentType == "" {
		// 如果没有设置Content-Type，默认进行压缩
		return g.writer.Write(data)
	}

	// 检查是否为可压缩的内容类型
	if isCompressibleContentType(contentType) {
		return g.writer.Write(data)
	}

	// 不可压缩的内容类型，直接写入原始数据
	return g.ResponseWriter.Write(data)
}

// WriteString 重写WriteString方法
func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	return g.Write([]byte(s))
}

// isCompressibleContentType 检查内容类型是否适合压缩
func isCompressibleContentType(contentType string) bool {
	// 定义可压缩的内容类型
	compressibleTypes := []string{
		"text/plain",
		"text/html",
		"text/css",
		"text/javascript",
		"application/javascript",
		"application/json",
		"application/xml",
		"application/rss+xml",
		"application/atom+xml",
		"image/svg+xml",
		"application/xhtml+xml",
	}

	for _, ct := range compressibleTypes {
		if strings.Contains(strings.ToLower(contentType), ct) {
			return true
		}
	}

	return false
}

// isForcedNotCompress 检查是否强制不压缩
func (g *gzipResponseWriter) isForcedNotCompress() bool {
	// 检查是否设置了禁用压缩的标志
	return false
}