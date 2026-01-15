package middleware

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// StaticCacheMiddleware 静态资源缓存控制中间件
func StaticCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否是静态资源请求
		ext := strings.ToLower(filepath.Ext(c.Request.URL.Path))
		
		// 根据文件类型设置不同的缓存策略
		cacheControl := getCacheControl(ext)
		if cacheControl != "" {
			c.Header("Cache-Control", cacheControl)
			
			// 设置ETag头
			setETagHeader(c, ext)
		}
		
		c.Next()
	}
}

// getCacheControl 根据文件扩展名返回相应的缓存控制策略
func getCacheControl(ext string) string {
	switch ext {
	case ".css", ".js":
		// CSS和JS文件通常有较长的缓存时间，因为它们通常带有版本号
		return "public, max-age=31536000" // 1年
	case ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".webp":
		// 图片文件也有较长的缓存时间
		return "public, max-age=31536000" // 1年
	case ".woff", ".woff2", ".ttf", ".eot":
		// 字体文件
		return "public, max-age=31536000" // 1年
	case ".html", ".htm":
		// HTML文件通常不缓存或短时间缓存
		return "no-cache"
	case ".json", ".xml":
		// API响应文件
		return "no-cache"
	default:
		// 其他文件类型
		return "public, max-age=86400" // 1天
	}
}

// setETagHeader 设置ETag头部
func setETagHeader(c *gin.Context, ext string) {
	// 对于生产环境，可以基于文件内容生成ETag
	// 这里简化处理，基于文件路径和修改时间
	path := c.Request.URL.Path
	
	// 对于长期缓存的资源，使用强ETag
	if isLongTermCached(ext) {
		// 在实际应用中，这里应该基于文件内容生成哈希值
		etag := generateETag(path)
		c.Header("ETag", etag)
	}
}

// isLongTermCached 检查文件类型是否为长期缓存
func isLongTermCached(ext string) bool {
	longTermExts := []string{".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".webp", ".woff", ".woff2", ".ttf", ".eot"}
	for _, e := range longTermExts {
		if ext == e {
			return true
		}
	}
	return false
}

// generateETag 生成简单的ETag（实际应用中应基于文件内容）
func generateETag(path string) string {
	// 这里只是示例，实际应用中应该基于文件内容生成哈希值
	return "\"" + path + "\""
}

// StaticFileServer 静态文件服务器，带缓存控制
func StaticFileServer(root string) gin.HandlerFunc {
	fileServer := http.FileServer(http.Dir(root))
	
	return func(c *gin.Context) {
		// 检查是否是静态资源请求
		ext := strings.ToLower(filepath.Ext(c.Request.URL.Path))
		
		// 设置缓存控制头
		cacheControl := getCacheControl(ext)
		if cacheControl != "" {
			c.Header("Cache-Control", cacheControl)
		}
		
		// 设置ETag
		if isLongTermCached(ext) {
			etag := generateETag(c.Request.URL.Path)
			c.Header("ETag", etag)
			
			// 检查If-None-Match头，实现304响应
			ifMatch := c.GetHeader("If-None-Match")
			if ifMatch != "" && ifMatch == etag {
				c.Status(http.StatusNotModified)
				return
			}
		}
		
		// 设置Last-Modified头
		c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		
		// 检查If-Modified-Since头
		ifModifiedSince := c.GetHeader("If-Modified-Since")
		if ifModifiedSince != "" {
			ifModifiedTime, err := time.Parse(http.TimeFormat, ifModifiedSince)
			if err == nil && !time.Now().After(ifModifiedTime.Add(1*time.Second)) {
				c.Status(http.StatusNotModified)
				return
			}
		}
		
		// 处理文件服务
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}