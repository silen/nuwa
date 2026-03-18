package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	nuwa "github.com/silen/nuwa"
)

func TestTokenMiddlewareMissingToken(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Token())
	engine.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if int(body["status"].(float64)) != nuwa.UNAUTHORIZED {
		t.Fatalf("expected unauthorized, got %#v", body)
	}
}

func TestTokenMiddlewareBackendFailure(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Token())
	engine.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("token", "abc")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if int(body["status"].(float64)) != nuwa.SYSTEM_ERROR {
		t.Fatalf("expected system error, got %#v", body)
	}
}
