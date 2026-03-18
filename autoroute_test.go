package nuwa

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type restOnlyContextHandler struct{}

func (restOnlyContextHandler) GET(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

type routeWithArgHandler struct{}

func (routeWithArgHandler) Show(c *gin.Context, id int) {
	c.String(http.StatusOK, "id")
}

func TestRESTEmptyPathInvokesMethod(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/*path", REST(restOnlyContextHandler{}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAutoRouteRejectsExtraArgsForNonVariadicHandler(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/*path", AutoRouteAny(routeWithArgHandler{}))

	req := httptest.NewRequest(http.MethodGet, "/show/1/extra", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
