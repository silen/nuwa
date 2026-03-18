package nuwa

import (
	"os"

	"github.com/gin-gonic/gin"
)

const environmentKey = "environment"

// Mode returns the Gin mode derived from the current process environment.
func Mode() string {
	switch os.Getenv(environmentKey) {
	case "dev":
		return gin.DebugMode
	case "test":
		return gin.TestMode
	default:
		return gin.ReleaseMode
	}
}

// SetMode applies the provided Gin mode globally.
func SetMode(mode string) {
	gin.SetMode(mode)
}
