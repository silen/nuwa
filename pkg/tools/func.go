package tools

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"github.com/spf13/cast"

	"github.com/silen/nuwa/pkg/conf"
	"github.com/silen/nuwa/pkg/logs"
)

const TimeTemp = "2006-01-02 15:04:05"

var (
	shanghaiLocation *time.Location
)

// SHLocal 获取时区
func SHLocal() *time.Location {
	if shanghaiLocation == nil {
		shanghaiLocation, _ = time.LoadLocation("Asia/Shanghai")
	}
	return shanghaiLocation
}

// 服务器IP
func ServerIP() (ip string) {
	ip = "127.0.0.1"
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logs.Error("ServerIP====", err)
		return ip
	}

	defer conn.Close()
	if localAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		ip = localAddr.IP.String()
	}

	return
}

// 创建随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// MD5 ...
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// MapToMD5Key map转string
func MapToMD5Key(data any) string {
	content, _ := json.Marshal(data)
	return MD5(string(content))

}

// 服务名称 如 cdp-system
func InternalDomain(app string) string {
	if os.Getenv("environment") == "" {
		host := conf.Config.GetString("apiHost")
		if host == "" {
			host = "https://cdp-api-dev.1.cc"
		}

		return host + "/" + app
	}
	//k8s 使用！
	return "http://" + app + ":80"
}

// 小数点几位 使用方法
// Decimal(1000.3134123, 3)
func Decimal(value float64, args ...any) float64 {

	if value != value {
		return 0
	}
	truncate := int32(2)
	if len(args) > 0 {
		truncate = cast.ToInt32(args[0])
	}

	return cast.ToFloat64(decimal.NewFromFloat(value).Truncate(truncate).String())
}
