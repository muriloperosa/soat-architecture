package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/stretchr/testify/require"
)

func setupEngineComSecurityHeaders() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.SecurityHeadersMiddleware())
	engine.GET("/qualquer", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return engine
}

func TestSecurityHeadersMiddleware_SetaXContentTypeOptions(t *testing.T) {
	engine := setupEngineComSecurityHeaders()

	req := httptest.NewRequest(http.MethodGet, "/qualquer", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
}

func TestSecurityHeadersMiddleware_SetaCrossOriginResourcePolicy(t *testing.T) {
	engine := setupEngineComSecurityHeaders()

	req := httptest.NewRequest(http.MethodGet, "/qualquer", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, "same-origin", rec.Header().Get("Cross-Origin-Resource-Policy"))
}

func TestSecurityHeadersMiddleware_NaoInterrompeOFluxo(t *testing.T) {
	engine := setupEngineComSecurityHeaders()

	req := httptest.NewRequest(http.MethodGet, "/qualquer", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
