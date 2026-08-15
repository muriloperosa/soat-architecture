package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/stretchr/testify/require"
)

func setupEngineComClaims(claims *domainauth.AppClaims, tipoEsperado domainauth.TipoUsuario) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/restrito", func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, claims)
		c.Next()
	}, middleware.AuthorizationMiddleware(tipoEsperado), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return engine
}

func TestAuthorizationMiddleware_TipoCorreto_PermitePassagem(t *testing.T) {
	engine := setupEngineComClaims(&domainauth.AppClaims{Tipo: domainauth.TipoInterno}, domainauth.TipoInterno)

	req := httptest.NewRequest(http.MethodGet, "/restrito", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthorizationMiddleware_TipoErrado_Retorna403(t *testing.T) {
	engine := setupEngineComClaims(&domainauth.AppClaims{Tipo: domainauth.TipoCliente}, domainauth.TipoInterno)

	req := httptest.NewRequest(http.MethodGet, "/restrito", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}
