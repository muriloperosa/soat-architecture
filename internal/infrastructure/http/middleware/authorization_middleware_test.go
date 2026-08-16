package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
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

func setupEngineComPapeis(claims *domainauth.AppClaims, tipoEsperado domainauth.TipoUsuario, papeisPermitidos ...shared.PapelUsuario) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/restrito", func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, claims)
		c.Next()
	}, middleware.AuthorizationMiddleware(tipoEsperado, papeisPermitidos...), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return engine
}

func TestAuthorizationMiddleware_PapelNaoPermitido_Retorna403(t *testing.T) {
	engine := setupEngineComPapeis(&domainauth.AppClaims{Tipo: domainauth.TipoInterno, Papel: shared.PapelMecanico}, domainauth.TipoInterno, shared.PapelAdmin)

	req := httptest.NewRequest(http.MethodGet, "/restrito", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAuthorizationMiddleware_PapelPermitido_Retorna200(t *testing.T) {
	engine := setupEngineComPapeis(&domainauth.AppClaims{Tipo: domainauth.TipoInterno, Papel: shared.PapelAdmin}, domainauth.TipoInterno, shared.PapelAdmin)

	req := httptest.NewRequest(http.MethodGet, "/restrito", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthorizationMiddleware_SemPapelInformado_SoChecaTipo(t *testing.T) {
	engine := setupEngineComPapeis(&domainauth.AppClaims{Tipo: domainauth.TipoInterno, Papel: shared.PapelAtendente}, domainauth.TipoInterno)

	req := httptest.NewRequest(http.MethodGet, "/restrito", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
