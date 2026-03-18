package nuwa

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// Response is Nuwa's root-level JSON response envelope.
type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    any    `json:"data,omitempty"`
}

// JSON writes a Nuwa response with an HTTP 200 status code.
func JSON(c *gin.Context, response Response) {
	writeJSON(c, http.StatusOK, response)
}

// Success writes a successful Nuwa response.
func Success(c *gin.Context, data any) {
	JSON(c, Response{
		Data:    data,
		Status:  SUCCESSFUL,
		Message: "ok",
	})
}

// Message writes an error-style Nuwa response with the provided business code.
func Message(c *gin.Context, message string, args ...any) {
	status := EXCEPTION
	if len(args) > 0 {
		status = cast.ToInt(args[0])
	}

	JSON(c, Response{
		Status:  status,
		Message: message,
	})
}

// Bind binds the request into obj and writes a parameter error response on failure.
func Bind(c *gin.Context, obj any) bool {
	if err := c.ShouldBind(obj); err != nil {
		Message(c, err.Error(), PARAMETER_ERROR)
		return false
	}

	return true
}

func writeJSON(c *gin.Context, httpStatus int, response Response) {
	c.JSON(httpStatus, response)
	c.Abort()
}
