package usuario_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	appusuario "github.com/muriloperosa/soat-architecture/internal/application/usuario"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
	"github.com/muriloperosa/soat-architecture/internal/domain/usuario/mocks"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	httpusuario "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/usuario"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_Criar_RequestValido_Retorna201(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	repo.EXPECT().BuscarPorEmail(mock.Anything, "ana@oficina.com").Return(nil, domainusuario.ErrUsuarioNaoEncontrado)
	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).
		Run(func(ctx context.Context, u *domainusuario.Usuario) { u.AtribuirID(1) }).
		Return(nil)

	h := httpusuario.NewHandler(appusuario.NewCriarUsuarioUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.POST("/v1/usuarios", h.Criar)

	body, _ := json.Marshal(httpusuario.CriarUsuarioRequest{Nome: "Ana Souza", Email: "ana@oficina.com", Senha: "senha123", Papel: string(shared.PapelMecanico)})
	req := httptest.NewRequest(http.MethodPost, "/v1/usuarios", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp httpusuario.UsuarioResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, uint64(1), resp.ID)
	require.True(t, resp.RequerAlterarSenha)
}

func TestHandler_Criar_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)

	h := httpusuario.NewHandler(appusuario.NewCriarUsuarioUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.POST("/v1/usuarios", h.Criar)

	req := httptest.NewRequest(http.MethodPost, "/v1/usuarios", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Criar_ErroInternoDoUseCase_Retorna500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	repo.EXPECT().BuscarPorEmail(mock.Anything, "ana@oficina.com").Return(nil, errors.New("conexao recusada"))

	h := httpusuario.NewHandler(appusuario.NewCriarUsuarioUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.POST("/v1/usuarios", h.Criar)

	body, _ := json.Marshal(httpusuario.CriarUsuarioRequest{Nome: "Ana Souza", Email: "ana@oficina.com", Senha: "senha123", Papel: string(shared.PapelMecanico)})
	req := httptest.NewRequest(http.MethodPost, "/v1/usuarios", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_Criar_EmailJaExiste_Retorna409(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	repo.EXPECT().BuscarPorEmail(mock.Anything, "ana@oficina.com").Return(existente, nil)

	h := httpusuario.NewHandler(appusuario.NewCriarUsuarioUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.POST("/v1/usuarios", h.Criar)

	body, _ := json.Marshal(httpusuario.CriarUsuarioRequest{Nome: "Ana Souza", Email: "ana@oficina.com", Senha: "senha123", Papel: string(shared.PapelMecanico)})
	req := httptest.NewRequest(http.MethodPost, "/v1/usuarios", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestHandler_Atualizar_UsuarioExiste_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).Return(nil)

	h := httpusuario.NewHandler(nil, appusuario.NewAtualizarUsuarioUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/usuarios/:id", h.Atualizar)

	body, _ := json.Marshal(httpusuario.AtualizarUsuarioRequest{Nome: "Ana S. Costa", Email: "ana@oficina.com", Papel: string(shared.PapelAtendente)})
	req := httptest.NewRequest(http.MethodPut, "/v1/usuarios/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_Atualizar_IDInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)

	h := httpusuario.NewHandler(nil, appusuario.NewAtualizarUsuarioUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/usuarios/:id", h.Atualizar)

	body, _ := json.Marshal(httpusuario.AtualizarUsuarioRequest{Nome: "Ana S. Costa", Email: "ana@oficina.com", Papel: string(shared.PapelAtendente)})
	req := httptest.NewRequest(http.MethodPut, "/v1/usuarios/nao-e-numero", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Atualizar_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)

	h := httpusuario.NewHandler(nil, appusuario.NewAtualizarUsuarioUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/usuarios/:id", h.Atualizar)

	req := httptest.NewRequest(http.MethodPut, "/v1/usuarios/1", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Atualizar_UsuarioNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	h := httpusuario.NewHandler(nil, appusuario.NewAtualizarUsuarioUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/usuarios/:id", h.Atualizar)

	body, _ := json.Marshal(httpusuario.AtualizarUsuarioRequest{Nome: "Ana S. Costa", Email: "ana@oficina.com", Papel: string(shared.PapelAtendente)})
	req := httptest.NewRequest(http.MethodPut, "/v1/usuarios/99", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Atualizar_ComSenhaNova_RedefineEForcaTroca(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	require.NoError(t, existente.AlterarSenha("senhaAntiga123"))
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).Return(nil)

	h := httpusuario.NewHandler(nil, appusuario.NewAtualizarUsuarioUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/usuarios/:id", h.Atualizar)

	body, _ := json.Marshal(httpusuario.AtualizarUsuarioRequest{Nome: "Ana Souza", Email: "ana@oficina.com", SenhaNova: "senhaDoAdmin123", Papel: string(shared.PapelMecanico)})
	req := httptest.NewRequest(http.MethodPut, "/v1/usuarios/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpusuario.UsuarioResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.True(t, resp.RequerAlterarSenha)
}

func TestHandler_Ativar_UsuarioExiste_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	existente.Inativar()
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).Return(nil)

	h := httpusuario.NewHandler(nil, nil, nil, appusuario.NewAtivarUsuarioUseCase(repo), nil, nil)
	engine := gin.New()
	engine.PATCH("/v1/usuarios/:id/ativar", h.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/usuarios/1/ativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_Ativar_IDInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)

	h := httpusuario.NewHandler(nil, nil, nil, appusuario.NewAtivarUsuarioUseCase(repo), nil, nil)
	engine := gin.New()
	engine.PATCH("/v1/usuarios/:id/ativar", h.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/usuarios/nao-e-numero/ativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Ativar_UsuarioNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	h := httpusuario.NewHandler(nil, nil, nil, appusuario.NewAtivarUsuarioUseCase(repo), nil, nil)
	engine := gin.New()
	engine.PATCH("/v1/usuarios/:id/ativar", h.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/usuarios/99/ativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Inativar_UsuarioExiste_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).Return(nil)

	h := httpusuario.NewHandler(nil, nil, nil, nil, appusuario.NewInativarUsuarioUseCase(repo), nil)
	engine := gin.New()
	engine.PATCH("/v1/usuarios/:id/inativar", h.Inativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/usuarios/1/inativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_Inativar_IDInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)

	h := httpusuario.NewHandler(nil, nil, nil, nil, appusuario.NewInativarUsuarioUseCase(repo), nil)
	engine := gin.New()
	engine.PATCH("/v1/usuarios/:id/inativar", h.Inativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/usuarios/nao-e-numero/inativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Inativar_UsuarioNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	h := httpusuario.NewHandler(nil, nil, nil, nil, appusuario.NewInativarUsuarioUseCase(repo), nil)
	engine := gin.New()
	engine.PATCH("/v1/usuarios/:id/inativar", h.Inativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/usuarios/99/inativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_AlterarSenha_ClaimsInjetados_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).Return(nil)

	h := httpusuario.NewHandler(nil, nil, appusuario.NewAlterarSenhaUseCase(repo), nil, nil, nil)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "1", Tipo: domainauth.TipoInterno, Papel: shared.PapelMecanico})
		c.Next()
	})
	engine.PUT("/v1/usuarios/me/senha", h.AlterarSenha)

	body, _ := json.Marshal(httpusuario.AlterarSenhaRequest{SenhaNova: "senhaNova123"})
	req := httptest.NewRequest(http.MethodPut, "/v1/usuarios/me/senha", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_AlterarSenha_SemClaims_Retorna401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)

	h := httpusuario.NewHandler(nil, nil, appusuario.NewAlterarSenhaUseCase(repo), nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/usuarios/me/senha", h.AlterarSenha)

	body, _ := json.Marshal(httpusuario.AlterarSenhaRequest{SenhaNova: "senhaNova123"})
	req := httptest.NewRequest(http.MethodPut, "/v1/usuarios/me/senha", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_AlterarSenha_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)

	h := httpusuario.NewHandler(nil, nil, appusuario.NewAlterarSenhaUseCase(repo), nil, nil, nil)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "1", Tipo: domainauth.TipoInterno, Papel: shared.PapelMecanico})
		c.Next()
	})
	engine.PUT("/v1/usuarios/me/senha", h.AlterarSenha)

	req := httptest.NewRequest(http.MethodPut, "/v1/usuarios/me/senha", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_AlterarSenha_ErroDoUseCase_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)

	h := httpusuario.NewHandler(nil, nil, appusuario.NewAlterarSenhaUseCase(repo), nil, nil, nil)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "1", Tipo: domainauth.TipoInterno, Papel: shared.PapelMecanico})
		c.Next()
	})
	engine.PUT("/v1/usuarios/me/senha", h.AlterarSenha)

	body, _ := json.Marshal(httpusuario.AlterarSenhaRequest{SenhaNova: "curta"})
	req := httptest.NewRequest(http.MethodPut, "/v1/usuarios/me/senha", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Me_SemClaims_Retorna401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)

	h := httpusuario.NewHandler(nil, nil, nil, nil, nil, appusuario.NewBuscarUsuarioLogadoUseCase(repo))
	engine := gin.New()
	engine.GET("/v1/usuarios/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/v1/usuarios/me", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_Me_UsuarioNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	h := httpusuario.NewHandler(nil, nil, nil, nil, nil, appusuario.NewBuscarUsuarioLogadoUseCase(repo))
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "1", Tipo: domainauth.TipoInterno, Papel: shared.PapelMecanico})
		c.Next()
	})
	engine.GET("/v1/usuarios/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/v1/usuarios/me", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Me_ClaimsInjetados_RetornaDadosDoUsuarioLogado(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewUsuarioRepository(t)
	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)

	h := httpusuario.NewHandler(nil, nil, nil, nil, nil, appusuario.NewBuscarUsuarioLogadoUseCase(repo))
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "1", Tipo: domainauth.TipoInterno, Papel: shared.PapelMecanico})
		c.Next()
	})
	engine.GET("/v1/usuarios/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/v1/usuarios/me", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpusuario.UsuarioResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "ana@oficina.com", resp.Email)
}
