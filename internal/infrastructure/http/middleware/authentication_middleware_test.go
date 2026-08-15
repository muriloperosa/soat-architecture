package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/stretchr/testify/require"
)

func setupEngine(jwtAuth *infraauth.AutenticadorJWT) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/protegido", middleware.AuthenticationMiddleware(jwtAuth), func(c *gin.Context) {
		claims, _ := c.Get(middleware.ClaimsContextKey)
		appClaims := claims.(*domainauth.AppClaims)
		c.JSON(http.StatusOK, gin.H{"tipo": appClaims.Tipo})
	})
	return engine
}

func TestAuthenticationMiddleware_TokenValido_PermitePassagem(t *testing.T) {
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute)
	token, _ := jwtAuth.GerarAccessToken("user-1", domainauth.TipoInterno, domainauth.PapelAdmin)
	engine := setupEngine(jwtAuth)

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthenticationMiddleware_SemHeader_Retorna401(t *testing.T) {
	engine := setupEngine(infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute))

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticationMiddleware_TokenInvalido_Retorna401(t *testing.T) {
	engine := setupEngine(infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute))

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer token-invalido")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
