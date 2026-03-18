package nuwa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMessageWritesNuwaEnvelope(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/", func(c *gin.Context) {
		Message(c, "bad request", PARAMETER_ERROR)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if response.Status != PARAMETER_ERROR {
		t.Fatalf("expected status %d, got %d", PARAMETER_ERROR, response.Status)
	}
	if response.Message != "bad request" {
		t.Fatalf("expected message to be preserved")
	}
}

func TestSuccessWritesData(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/", func(c *gin.Context) {
		Success(c, map[string]string{"name": "nuwa"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if response.Status != SUCCESSFUL {
		t.Fatalf("expected successful status, got %d", response.Status)
	}
	data, ok := response.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", response.Data)
	}
	if data["name"] != "nuwa" {
		t.Fatalf("expected response data to include payload")
	}
}
