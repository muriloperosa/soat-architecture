package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/auth/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupEngine(jwtAuth domainauth.JWTProvider, refreshTokens domainauth.RefreshTokenRepository, usuarios domainauth.UsuarioStatusRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/protegido", middleware.AuthenticationMiddleware(jwtAuth, refreshTokens, usuarios), func(c *gin.Context) {
		claims, _ := c.Get(middleware.ClaimsContextKey)
		appClaims := claims.(*domainauth.AppClaims)
		c.JSON(http.StatusOK, gin.H{"tipo": appClaims.Tipo})
	})
	return engine
}

func TestAuthenticationMiddleware_TokenValido_PermitePassagem(t *testing.T) {
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute)
	token, jti, _ := jwtAuth.GerarAccessToken("1", domainauth.TipoInterno, shared.PapelAdmin)
	refreshTokens := mocks.NewRefreshTokenRepository(t)
	refreshTokens.EXPECT().AccessTokenRevogado(mock.Anything, jti).Return(false, nil)
	usuarios := mocks.NewUsuarioStatusRepository(t)
	usuarios.EXPECT().EstaAtivo(mock.Anything, uint64(1)).Return(true, nil)
	engine := setupEngine(jwtAuth, refreshTokens, usuarios)

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthenticationMiddleware_TokenRevogado_Retorna401(t *testing.T) {
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute)
	token, jti, _ := jwtAuth.GerarAccessToken("1", domainauth.TipoInterno, shared.PapelAdmin)
	refreshTokens := mocks.NewRefreshTokenRepository(t)
	refreshTokens.EXPECT().AccessTokenRevogado(mock.Anything, jti).Return(true, nil)
	engine := setupEngine(jwtAuth, refreshTokens, mocks.NewUsuarioStatusRepository(t))

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticationMiddleware_SemHeader_Retorna401(t *testing.T) {
	engine := setupEngine(infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute), mocks.NewRefreshTokenRepository(t), mocks.NewUsuarioStatusRepository(t))

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticationMiddleware_TokenInvalido_Retorna401(t *testing.T) {
	engine := setupEngine(infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute), mocks.NewRefreshTokenRepository(t), mocks.NewUsuarioStatusRepository(t))

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer token-invalido")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticationMiddleware_UsuarioInativo_Retorna401(t *testing.T) {
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute)
	token, jti, _ := jwtAuth.GerarAccessToken("1", domainauth.TipoInterno, shared.PapelAdmin)
	refreshTokens := mocks.NewRefreshTokenRepository(t)
	refreshTokens.EXPECT().AccessTokenRevogado(mock.Anything, jti).Return(false, nil)
	usuarios := mocks.NewUsuarioStatusRepository(t)
	usuarios.EXPECT().EstaAtivo(mock.Anything, uint64(1)).Return(false, nil)
	engine := setupEngine(jwtAuth, refreshTokens, usuarios)

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticationMiddleware_SubjectNaoNumerico_Retorna401(t *testing.T) {
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute)
	token, jti, _ := jwtAuth.GerarAccessToken("nao-e-numero", domainauth.TipoInterno, shared.PapelAdmin)
	refreshTokens := mocks.NewRefreshTokenRepository(t)
	refreshTokens.EXPECT().AccessTokenRevogado(mock.Anything, jti).Return(false, nil)
	engine := setupEngine(jwtAuth, refreshTokens, mocks.NewUsuarioStatusRepository(t))

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticationMiddleware_ErroAoChecarStatus_Retorna401(t *testing.T) {
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute)
	token, jti, _ := jwtAuth.GerarAccessToken("1", domainauth.TipoInterno, shared.PapelAdmin)
	refreshTokens := mocks.NewRefreshTokenRepository(t)
	refreshTokens.EXPECT().AccessTokenRevogado(mock.Anything, jti).Return(false, nil)
	usuarios := mocks.NewUsuarioStatusRepository(t)
	usuarios.EXPECT().EstaAtivo(mock.Anything, uint64(1)).Return(false, errors.New("conexao recusada"))
	engine := setupEngine(jwtAuth, refreshTokens, usuarios)

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
