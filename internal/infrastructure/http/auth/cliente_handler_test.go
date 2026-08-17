package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	httpauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/auth"
	"github.com/stretchr/testify/require"
)

func TestAuthClienteHandler_Login_CredencialValida_Retorna200ComTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	credenciais := &credenciaisFakeHTTP{credencial: &domainauth.Credencial{ID: 1, SenhaHash: hashSenhaHTTP(t, "senha123"), Papel: shared.PapelCliente, Ativo: true}}
	loginUC := appauth.NewLoginUseCase(credenciais, &refreshTokensFakeHTTP{}, infraauth.NewAuthenticatorJWT("s", 15*time.Minute), domainauth.TipoCliente, time.Hour, time.Hour)
	h := httpauth.NewAuthClienteHandler(loginUC, nil, nil)

	engine := gin.New()
	engine.POST("/v1/auth/cliente/login", h.Login)

	body, _ := json.Marshal(httpauth.LoginRequest{Email: "cliente@a.com", Senha: "senha123"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cliente/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpauth.TokenResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.AccessToken)
	require.NotEmpty(t, resp.RefreshToken)
}

func TestAuthClienteHandler_Login_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	loginUC := appauth.NewLoginUseCase(&credenciaisFakeHTTP{}, &refreshTokensFakeHTTP{}, infraauth.NewAuthenticatorJWT("s", 15*time.Minute), domainauth.TipoCliente, time.Hour, time.Hour)
	h := httpauth.NewAuthClienteHandler(loginUC, nil, nil)

	engine := gin.New()
	engine.POST("/v1/auth/cliente/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cliente/login", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthClienteHandler_Login_CredencialInvalida_Retorna401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	loginUC := appauth.NewLoginUseCase(&credenciaisFakeHTTP{}, &refreshTokensFakeHTTP{}, infraauth.NewAuthenticatorJWT("s", 15*time.Minute), domainauth.TipoCliente, time.Hour, time.Hour)
	h := httpauth.NewAuthClienteHandler(loginUC, nil, nil)

	engine := gin.New()
	engine.POST("/v1/auth/cliente/login", h.Login)

	body, _ := json.Marshal(httpauth.LoginRequest{Email: "naoexiste@a.com", Senha: "x"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cliente/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthClienteHandler_Refresh_TokenValido_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtAuth := infraauth.NewAuthenticatorJWT("s", 15*time.Minute)
	refreshTokens := &refreshTokensFakeHTTP{}
	credenciais := &credenciaisFakeHTTP{credencial: &domainauth.Credencial{ID: 1, SenhaHash: hashSenhaHTTP(t, "senha123"), Papel: shared.PapelCliente, Ativo: true}}
	loginUC := appauth.NewLoginUseCase(credenciais, refreshTokens, jwtAuth, domainauth.TipoCliente, time.Hour, time.Hour)
	refreshUC := appauth.NewRefreshUseCase(refreshTokens, jwtAuth, time.Hour, time.Hour)
	h := httpauth.NewAuthClienteHandler(loginUC, refreshUC, nil)

	out, err := loginUC.Executar(context.Background(), appauth.LoginInput{Email: "cliente@a.com", Senha: "senha123"})
	require.NoError(t, err)

	engine := gin.New()
	engine.POST("/v1/auth/cliente/refresh", h.Refresh)

	body, _ := json.Marshal(httpauth.RefreshRequest{RefreshToken: out.RefreshToken})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cliente/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthClienteHandler_Refresh_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtAuth := infraauth.NewAuthenticatorJWT("s", 15*time.Minute)
	refreshUC := appauth.NewRefreshUseCase(&refreshTokensFakeHTTP{}, jwtAuth, time.Hour, time.Hour)
	h := httpauth.NewAuthClienteHandler(nil, refreshUC, nil)

	engine := gin.New()
	engine.POST("/v1/auth/cliente/refresh", h.Refresh)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cliente/refresh", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthClienteHandler_Refresh_TokenInvalido_Retorna401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtAuth := infraauth.NewAuthenticatorJWT("s", 15*time.Minute)
	refreshUC := appauth.NewRefreshUseCase(&refreshTokensFakeHTTP{}, jwtAuth, time.Hour, time.Hour)
	h := httpauth.NewAuthClienteHandler(nil, refreshUC, nil)

	engine := gin.New()
	engine.POST("/v1/auth/cliente/refresh", h.Refresh)

	body, _ := json.Marshal(httpauth.RefreshRequest{RefreshToken: "token-inexistente"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cliente/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthClienteHandler_Logout_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logoutUC := appauth.NewLogoutUseCase(&refreshTokensFakeHTTP{})
	h := httpauth.NewAuthClienteHandler(nil, nil, logoutUC)

	engine := gin.New()
	engine.POST("/v1/auth/cliente/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cliente/logout", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthClienteHandler_Logout_ErroAoRevogar_Retorna500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtAuth := infraauth.NewAuthenticatorJWT("s", 15*time.Minute)
	refreshTokens := &refreshTokensFakeHTTP{}
	credenciais := &credenciaisFakeHTTP{credencial: &domainauth.Credencial{ID: 1, SenhaHash: hashSenhaHTTP(t, "senha123"), Papel: shared.PapelCliente, Ativo: true}}
	loginUC := appauth.NewLoginUseCase(credenciais, refreshTokens, jwtAuth, domainauth.TipoCliente, time.Hour, time.Hour)

	out, err := loginUC.Executar(context.Background(), appauth.LoginInput{Email: "cliente@a.com", Senha: "senha123"})
	require.NoError(t, err)

	refreshTokens.revogarErr = errors.New("conexao recusada")
	logoutUC := appauth.NewLogoutUseCase(refreshTokens)
	h := httpauth.NewAuthClienteHandler(nil, nil, logoutUC)

	engine := gin.New()
	engine.POST("/v1/auth/cliente/logout", h.Logout)

	body, _ := json.Marshal(httpauth.RefreshRequest{RefreshToken: out.RefreshToken})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cliente/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAuthClienteHandler_Logout_TokenValido_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtAuth := infraauth.NewAuthenticatorJWT("s", 15*time.Minute)
	refreshTokens := &refreshTokensFakeHTTP{}
	credenciais := &credenciaisFakeHTTP{credencial: &domainauth.Credencial{ID: 1, SenhaHash: hashSenhaHTTP(t, "senha123"), Papel: shared.PapelCliente, Ativo: true}}
	loginUC := appauth.NewLoginUseCase(credenciais, refreshTokens, jwtAuth, domainauth.TipoCliente, time.Hour, time.Hour)
	logoutUC := appauth.NewLogoutUseCase(refreshTokens)
	h := httpauth.NewAuthClienteHandler(loginUC, nil, logoutUC)

	out, err := loginUC.Executar(context.Background(), appauth.LoginInput{Email: "cliente@a.com", Senha: "senha123"})
	require.NoError(t, err)

	engine := gin.New()
	engine.POST("/v1/auth/cliente/logout", h.Logout)

	body, _ := json.Marshal(httpauth.RefreshRequest{RefreshToken: out.RefreshToken})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cliente/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}
