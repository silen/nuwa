package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	nuwa "github.com/silen/nuwa"
	tk "github.com/silen/nuwa/pkg/token"
)

func Token() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenString := c.Request.Header.Get("token")
		if tokenString == "" {
			c.JSON(http.StatusOK, map[string]any{
				"status":  nuwa.UNAUTHORIZED,
				"message": "缺失token",
			})
			c.Abort()
			return
		}

		res, err := tk.NewToken(c).Check(tokenString)
		if err != nil || res == "" {
			c.JSON(http.StatusOK, map[string]any{
				"status":  nuwa.UNAUTHORIZED,
				"message": "token错误",
			})
			c.Abort()
			return
		}
		c.Set("userInfo", res)
		c.Set("token", tokenString)
		c.Next()
	}
}
