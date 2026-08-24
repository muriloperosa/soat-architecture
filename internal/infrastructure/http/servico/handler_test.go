package servico_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	appservico "github.com/muriloperosa/soat-architecture/internal/application/servico"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/servico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	httpservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/servico"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func preco(v float64) *float64 { return &v }

func claimsInterno() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "1", Tipo: domainauth.TipoInterno, Papel: shared.PapelAtendente})
		c.Next()
	}
}

func novoServico(t *testing.T) *domainservico.Servico {
	t.Helper()
	s, err := domainservico.NewServico("Troca de óleo", "Troca de óleo e filtro", 150.50, 60, 1)
	require.NoError(t, err)
	return s
}

func TestHandler_Criar_RequestValido_Retorna201(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*servico.Servico")).
		Run(func(ctx context.Context, s *domainservico.Servico) { s.AtribuirID(1) }).
		Return(nil)

	h := httpservico.NewHandler(appservico.NewCriarServicoUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.Use(claimsInterno())
	engine.POST("/v1/servicos", h.Criar)

	body, _ := json.Marshal(httpservico.CriarServicoRequest{
		Nome: "Troca de óleo", Descricao: "Troca de óleo e filtro", PrecoBase: preco(150.50), TempoEstimadoMinutos: 60,
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/servicos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp httpservico.ServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, uint64(1), resp.ID)
	require.True(t, resp.Ativo)
	require.Equal(t, uint64(1), resp.CriadoPor)
	require.Equal(t, 150.50, resp.PrecoBase)
}

func TestHandler_Criar_PrecoZero_Retorna201(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*servico.Servico")).
		Run(func(ctx context.Context, s *domainservico.Servico) { s.AtribuirID(1) }).
		Return(nil)

	h := httpservico.NewHandler(appservico.NewCriarServicoUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.Use(claimsInterno())
	engine.POST("/v1/servicos", h.Criar)

	body, _ := json.Marshal(httpservico.CriarServicoRequest{
		Nome: "Diagnóstico", Descricao: "avaliação gratuita", PrecoBase: preco(0), TempoEstimadoMinutos: 30,
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/servicos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp httpservico.ServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0.0, resp.PrecoBase)
}

func TestHandler_Criar_SemClaims_Retorna401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)

	h := httpservico.NewHandler(appservico.NewCriarServicoUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.POST("/v1/servicos", h.Criar)

	body, _ := json.Marshal(httpservico.CriarServicoRequest{
		Nome: "Troca de óleo", Descricao: "descrição", PrecoBase: preco(100), TempoEstimadoMinutos: 30,
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/servicos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_Criar_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)

	h := httpservico.NewHandler(appservico.NewCriarServicoUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.Use(claimsInterno())
	engine.POST("/v1/servicos", h.Criar)

	req := httptest.NewRequest(http.MethodPost, "/v1/servicos", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Criar_ErroInternoDoUseCase_Retorna500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(errors.New("conexao recusada"))

	h := httpservico.NewHandler(appservico.NewCriarServicoUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.Use(claimsInterno())
	engine.POST("/v1/servicos", h.Criar)

	body, _ := json.Marshal(httpservico.CriarServicoRequest{
		Nome: "Troca de óleo", Descricao: "descrição", PrecoBase: preco(100), TempoEstimadoMinutos: 30,
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/servicos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_Listar_Retorna200ComCatalogo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	s := novoServico(t)
	s.AtribuirID(1)
	repo.EXPECT().Listar(mock.Anything).Return([]*domainservico.Servico{s}, nil)

	h := httpservico.NewHandler(nil, nil, appservico.NewListarServicosUseCase(repo), nil, nil, nil)
	engine := gin.New()
	engine.GET("/v1/servicos", h.Listar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp []httpservico.ServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp, 1)
	require.Equal(t, "Troca de óleo", resp[0].Nome)
}

func TestHandler_Listar_Vazio_RetornaArrayVazio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().Listar(mock.Anything).Return([]*domainservico.Servico{}, nil)

	h := httpservico.NewHandler(nil, nil, appservico.NewListarServicosUseCase(repo), nil, nil, nil)
	engine := gin.New()
	engine.GET("/v1/servicos", h.Listar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, "[]", rec.Body.String())
}

func TestHandler_Listar_ErroDoUseCase_Retorna500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().Listar(mock.Anything).Return(nil, errors.New("conexao recusada"))

	h := httpservico.NewHandler(nil, nil, appservico.NewListarServicosUseCase(repo), nil, nil, nil)
	engine := gin.New()
	engine.GET("/v1/servicos", h.Listar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_Buscar_ServicoExiste_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	s := novoServico(t)
	s.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(s, nil)

	h := httpservico.NewHandler(nil, nil, nil, appservico.NewBuscarServicoUseCase(repo), nil, nil)
	engine := gin.New()
	engine.GET("/v1/servicos/:id", h.Buscar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos/1", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpservico.ServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, uint64(1), resp.ID)
}

func TestHandler_Buscar_IDInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)

	h := httpservico.NewHandler(nil, nil, nil, appservico.NewBuscarServicoUseCase(repo), nil, nil)
	engine := gin.New()
	engine.GET("/v1/servicos/:id", h.Buscar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos/nao-e-numero", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Buscar_NaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	h := httpservico.NewHandler(nil, nil, nil, appservico.NewBuscarServicoUseCase(repo), nil, nil)
	engine := gin.New()
	engine.GET("/v1/servicos/:id", h.Buscar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos/99", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Atualizar_ServicoExiste_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	s := novoServico(t)
	s.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(s, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(nil)

	h := httpservico.NewHandler(nil, appservico.NewAtualizarServicoUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/servicos/:id", h.Atualizar)

	body, _ := json.Marshal(httpservico.AtualizarServicoRequest{
		Nome: "Alinhamento", Descricao: "alinhamento e balanceamento", PrecoBase: preco(200.75), TempoEstimadoMinutos: 90,
	})
	req := httptest.NewRequest(http.MethodPut, "/v1/servicos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpservico.ServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "Alinhamento", resp.Nome)
	require.Equal(t, 200.75, resp.PrecoBase)
}

func TestHandler_Atualizar_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)

	h := httpservico.NewHandler(nil, appservico.NewAtualizarServicoUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/servicos/:id", h.Atualizar)

	req := httptest.NewRequest(http.MethodPut, "/v1/servicos/1", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Atualizar_NaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	h := httpservico.NewHandler(nil, appservico.NewAtualizarServicoUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/servicos/:id", h.Atualizar)

	body, _ := json.Marshal(httpservico.AtualizarServicoRequest{
		Nome: "Alinhamento", Descricao: "descrição", PrecoBase: preco(200), TempoEstimadoMinutos: 90,
	})
	req := httptest.NewRequest(http.MethodPut, "/v1/servicos/99", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Atualizar_IDInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)

	h := httpservico.NewHandler(nil, appservico.NewAtualizarServicoUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/servicos/:id", h.Atualizar)

	body, _ := json.Marshal(httpservico.AtualizarServicoRequest{
		Nome: "Alinhamento", Descricao: "descrição", PrecoBase: preco(200), TempoEstimadoMinutos: 90,
	})
	req := httptest.NewRequest(http.MethodPut, "/v1/servicos/nao-e-numero", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Ativar_ComSucesso_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	s := novoServico(t)
	s.AtribuirID(1)
	s.Inativar()
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(s, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(nil)

	h := httpservico.NewHandler(nil, nil, nil, nil, appservico.NewAtivarServicoUseCase(repo), nil)
	engine := gin.New()
	engine.PATCH("/v1/servicos/:id/ativar", h.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/servicos/1/ativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_Ativar_NaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	h := httpservico.NewHandler(nil, nil, nil, nil, appservico.NewAtivarServicoUseCase(repo), nil)
	engine := gin.New()
	engine.PATCH("/v1/servicos/:id/ativar", h.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/servicos/99/ativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Inativar_ComSucesso_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	s := novoServico(t)
	s.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(s, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(nil)

	h := httpservico.NewHandler(nil, nil, nil, nil, nil, appservico.NewInativarServicoUseCase(repo))
	engine := gin.New()
	engine.PATCH("/v1/servicos/:id/inativar", h.Inativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/servicos/1/inativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_Inativar_NaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	h := httpservico.NewHandler(nil, nil, nil, nil, nil, appservico.NewInativarServicoUseCase(repo))
	engine := gin.New()
	engine.PATCH("/v1/servicos/:id/inativar", h.Inativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/servicos/99/inativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
