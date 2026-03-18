package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	nuwa "github.com/silen/nuwa"
	httpclient "github.com/silen/nuwa/httpclient"
	runtimeinfo "github.com/silen/nuwa/internal/runtime"
	"github.com/silen/nuwa/logs"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		timeStart := time.Now()
		requestId := c.Request.Header.Get("X-Request-Id")
		if requestId == "" {
			u := uuid.Must(uuid.NewUUID())
			requestId = u.String()
		}

		c.Set("X-Request-Id", requestId)
		c.Writer.Header().Set("X-Request-Id", requestId)

		c.Writer.Header().Set("serverIp", maskedServerIP(runtimeinfo.ServerIP()))
		c.Writer.Header().Set("start-server-time", nuwa.StartServerTime)

		str, _ := c.GetRawData()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(str))

		c.Next()

		logs.WithContext(c).Trace(fmt.Sprintf(
			"[%f] URL:[%s]%s%s | status:%d | clientIP:%s | header:%v | query:%s | bodyBytes:%d",
			time.Since(timeStart).Seconds(),
			c.Request.Method,
			c.Request.Host,
			c.Request.URL.Path,
			c.Writer.Status(),
			c.ClientIP(),
			sanitizeHeaders(c.Request.Header),
			httpclient.CompressStr(c.Request.URL.RawQuery),
			len(str),
		))

	}
}

func sanitizeHeaders(header http.Header) http.Header {
	safeHeader := make(http.Header, len(header))
	for key, values := range header {
		if isSensitiveHeader(key) {
			safeHeader[key] = []string{"[REDACTED]"}
			continue
		}

		copied := make([]string, len(values))
		copy(copied, values)
		safeHeader[key] = copied
	}
	return safeHeader
}

func isSensitiveHeader(key string) bool {
	switch strings.ToLower(key) {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "token", "x-api-key", "x-auth-token":
		return true
	default:
		return false
	}
}

func maskedServerIP(serverIP string) string {
	serverIPs := strings.Split(serverIP, ".")
	if len(serverIPs) > 3 {
		serverIPs[2] = "*"
		return strings.Join(serverIPs, ".")
	}
	return serverIP
}
