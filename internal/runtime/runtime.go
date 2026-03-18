package runtimeinfo

import (
	"net"
	"sync"

	"github.com/silen/nuwa/logs"
)

var (
	serverIPOnce  sync.Once
	serverIPValue = "127.0.0.1"
)

// ServerIP returns the preferred outbound server IP.
func ServerIP() (ip string) {
	serverIPOnce.Do(func() {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			logs.Error("ServerIP====", err)
			return
		}
		defer conn.Close()

		if localAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			serverIPValue = localAddr.IP.String()
		}
	})

	return serverIPValue
}
