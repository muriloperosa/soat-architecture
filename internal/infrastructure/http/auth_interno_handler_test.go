package http_test

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
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	httphandler "github.com/muriloperosa/soat-architecture/internal/infrastructure/http"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type credenciaisFakeHTTP struct {
	credencial *domainauth.Credencial
}

func (f *credenciaisFakeHTTP) BuscarPorEmail(ctx context.Context, email string) (*domainauth.Credencial, error) {
	if f.credencial == nil {
		return nil, errors.New("nao encontrado")
	}
	return f.credencial, nil
}

type refreshTokensFakeHTTP struct {
	salvos []*domainauth.RefreshToken
}

func (f *refreshTokensFakeHTTP) Salvar(ctx context.Context, rt *domainauth.RefreshToken) error {
	rt.ID = "rt-1"
	f.salvos = append(f.salvos, rt)
	return nil
}
func (f *refreshTokensFakeHTTP) BuscarPorHash(ctx context.Context, hash string) (*domainauth.RefreshToken, error) {
	for _, rt := range f.salvos {
		if rt.TokenHash == hash {
			return rt, nil
		}
	}
	return nil, errors.New("nao encontrado")
}
func (f *refreshTokensFakeHTTP) Revogar(ctx context.Context, id string) error {
	for _, rt := range f.salvos {
		if rt.ID == id {
			now := time.Now()
			rt.RevogadoEm = &now
		}
	}
	return nil
}

func hashSenhaHTTP(t *testing.T, senha string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(h)
}

func credenciaisComUsuarioValido(t *testing.T) *credenciaisFakeHTTP {
	return &credenciaisFakeHTTP{credencial: &domainauth.Credencial{ID: "user-1", SenhaHash: hashSenhaHTTP(t, "senha123"), Papel: domainauth.PapelAdmin}}
}

func TestAuthInternoHandler_Login_CredencialValida_Retorna200ComTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	loginUC := appauth.NewLoginUseCase(credenciaisComUsuarioValido(t), &refreshTokensFakeHTTP{}, infraauth.NewAuthenticatorJWT("s", 15*time.Minute), domainauth.TipoInterno, time.Hour)
	handler := httphandler.NewAuthInternoHandler(loginUC, nil, nil)

	engine := gin.New()
	engine.POST("/v1/auth/login", handler.Login)

	body, _ := json.Marshal(httphandler.LoginRequest{Email: "func@oficina.com", Senha: "senha123"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httphandler.TokenResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.AccessToken)
	require.NotEmpty(t, resp.RefreshToken)
}

func TestAuthInternoHandler_Login_CredencialInvalida_Retorna401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	loginUC := appauth.NewLoginUseCase(&credenciaisFakeHTTP{}, &refreshTokensFakeHTTP{}, infraauth.NewAuthenticatorJWT("s", 15*time.Minute), domainauth.TipoInterno, time.Hour)
	handler := httphandler.NewAuthInternoHandler(loginUC, nil, nil)

	engine := gin.New()
	engine.POST("/v1/auth/login", handler.Login)

	body, _ := json.Marshal(httphandler.LoginRequest{Email: "naoexiste@oficina.com", Senha: "x"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthInternoHandler_Refresh_TokenValido_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtAuth := infraauth.NewAuthenticatorJWT("s", 15*time.Minute)
	refreshTokens := &refreshTokensFakeHTTP{}
	loginUC := appauth.NewLoginUseCase(credenciaisComUsuarioValido(t), refreshTokens, jwtAuth, domainauth.TipoInterno, time.Hour)
	refreshUC := appauth.NewRefreshUseCase(refreshTokens, jwtAuth, time.Hour)
	handler := httphandler.NewAuthInternoHandler(loginUC, refreshUC, nil)

	out, err := loginUC.Executar(context.Background(), appauth.LoginInput{Email: "func@oficina.com", Senha: "senha123"})
	require.NoError(t, err)

	engine := gin.New()
	engine.POST("/v1/auth/refresh", handler.Refresh)

	body, _ := json.Marshal(httphandler.RefreshRequest{RefreshToken: out.RefreshToken})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthInternoHandler_Logout_TokenValido_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtAuth := infraauth.NewAuthenticatorJWT("s", 15*time.Minute)
	refreshTokens := &refreshTokensFakeHTTP{}
	loginUC := appauth.NewLoginUseCase(credenciaisComUsuarioValido(t), refreshTokens, jwtAuth, domainauth.TipoInterno, time.Hour)
	logoutUC := appauth.NewLogoutUseCase(refreshTokens)
	handler := httphandler.NewAuthInternoHandler(loginUC, nil, logoutUC)

	out, err := loginUC.Executar(context.Background(), appauth.LoginInput{Email: "func@oficina.com", Senha: "senha123"})
	require.NoError(t, err)

	engine := gin.New()
	engine.POST("/v1/auth/logout", handler.Logout)

	body, _ := json.Marshal(httphandler.RefreshRequest{RefreshToken: out.RefreshToken})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}
