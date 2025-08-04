package handler

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	nuwa "github.com/silen/nuwa"
	"github.com/silen/nuwa/pkg/logs"
	nwHttp "github.com/silen/nuwa/pkg/nwHttp"
	"github.com/silen/nuwa/pkg/tools"
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

		serverIP := tools.ServerIP()
		serverIPs := strings.Split(serverIP, ".")
		if len(serverIPs) > 3 {
			serverIPs[2] = "*"
			serverIP = strings.Join(serverIPs, ".")
		}
		c.Writer.Header().Set("serverIp", serverIP)
		c.Writer.Header().Set("start-server-time", nuwa.StartServerTime)

		str, _ := c.GetRawData()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(str))

		c.Next()

		logs.WithContext(c).Trace("["+fmt.Sprintf("%f", time.Since(timeStart).Seconds())+"] URL:["+c.Request.Method+"]"+c.Request.Host+c.Request.URL.Path+" | clientIP："+c.ClientIP()+
			" | header：", c.Request.Header, " | params：", c.Request.Form, " | json：", nwHttp.CompressStr(string(str)))

	}
}
