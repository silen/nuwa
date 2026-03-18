package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	nuwa "github.com/silen/nuwa"
	"github.com/silen/nuwa/logs"
	tk "github.com/silen/nuwa/token"
)

func Token() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenString := c.Request.Header.Get("token")
		if tokenString == "" {
			nuwa.Message(c, "缺失token", nuwa.UNAUTHORIZED)
			return
		}

		res, err := tk.NewToken(c).Check(tokenString)
		if errors.Is(err, tk.ErrTokenNotFound) {
			nuwa.Message(c, "token错误", nuwa.UNAUTHORIZED)
			return
		}
		if err != nil {
			logs.WithContext(c).Error("token check failed:", err)
			nuwa.Message(c, "token校验失败", nuwa.SYSTEM_ERROR)
			return
		}
		if res == "" {
			nuwa.Message(c, "token错误", nuwa.UNAUTHORIZED)
			return
		}
		c.Set("userInfo", res)
		c.Set("token", tokenString)
		c.Next()
	}
}
