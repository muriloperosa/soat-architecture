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
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/servico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	httpservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/servico"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func comSubjectAutenticado(engine *gin.Engine) {
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "1", Tipo: domainauth.TipoInterno, Papel: shared.PapelAdmin})
		c.Next()
	})
}

func novoQueryParser() *httpquery.Parser {
	return httpquery.NewParser()
}

func servicoExistente(t *testing.T) *domainservico.Servico {
	t.Helper()
	s, err := domainservico.NewServico("Troca de óleo", "Troca de óleo e filtro", 150.50, 60, 1)
	require.NoError(t, err)
	s.AtribuirID(1)
	return s
}

func precoBase(v float64) *float64 { return &v }

func TestHandler_Criar_RequestValido_Retorna201(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*servico.Servico")).
		Run(func(_ context.Context, s *domainservico.Servico) { s.AtribuirID(1) }).
		Return(nil)

	h := httpservico.NewHandler(appservico.NewCriarServicoUseCase(repo), nil, nil, nil, nil, nil, nil)
	engine := gin.New()
	comSubjectAutenticado(engine)
	engine.POST("/v1/servicos", h.Criar)

	body, _ := json.Marshal(httpservico.CriarServicoRequest{Nome: "Troca de óleo", Descricao: "Troca de óleo e filtro", PrecoBase: precoBase(150.5), TempoEstimadoMinutos: 60})
	req := httptest.NewRequest(http.MethodPost, "/v1/servicos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp httpservico.ServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, uint64(1), resp.ID)
	require.Equal(t, uint64(1), resp.CriadoPor)
	require.True(t, resp.Ativo)
}

func TestHandler_Criar_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)

	h := httpservico.NewHandler(appservico.NewCriarServicoUseCase(repo), nil, nil, nil, nil, nil, nil)
	engine := gin.New()
	comSubjectAutenticado(engine)
	engine.POST("/v1/servicos", h.Criar)

	req := httptest.NewRequest(http.MethodPost, "/v1/servicos", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Criar_ErroDeValidacaoDoDominio_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)

	h := httpservico.NewHandler(appservico.NewCriarServicoUseCase(repo), nil, nil, nil, nil, nil, nil)
	engine := gin.New()
	comSubjectAutenticado(engine)
	engine.POST("/v1/servicos", h.Criar)

	body, _ := json.Marshal(httpservico.CriarServicoRequest{Nome: "Troca de óleo", Descricao: "Troca de óleo e filtro", PrecoBase: precoBase(-1), TempoEstimadoMinutos: 60})
	req := httptest.NewRequest(http.MethodPost, "/v1/servicos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Atualizar_RequestValido_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(servicoExistente(t), nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(nil)

	h := httpservico.NewHandler(nil, appservico.NewAtualizarServicoUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/servicos/:id", h.Atualizar)

	body, _ := json.Marshal(httpservico.AtualizarServicoRequest{Nome: "Alinhamento", Descricao: "Alinhamento e balanceamento", PrecoBase: precoBase(200.75), TempoEstimadoMinutos: 90})
	req := httptest.NewRequest(http.MethodPut, "/v1/servicos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpservico.ServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "Alinhamento", resp.Nome)
}

func TestHandler_Atualizar_ServicoNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	h := httpservico.NewHandler(nil, appservico.NewAtualizarServicoUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/servicos/:id", h.Atualizar)

	body, _ := json.Marshal(httpservico.AtualizarServicoRequest{Nome: "Alinhamento", Descricao: "Alinhamento e balanceamento", PrecoBase: precoBase(200.75), TempoEstimadoMinutos: 90})
	req := httptest.NewRequest(http.MethodPut, "/v1/servicos/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Buscar_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(servicoExistente(t), nil)

	h := httpservico.NewHandler(nil, nil, nil, appservico.NewBuscarServicoUseCase(repo), nil, nil, nil)
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

func TestHandler_Buscar_ServicoNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	h := httpservico.NewHandler(nil, nil, nil, appservico.NewBuscarServicoUseCase(repo), nil, nil, nil)
	engine := gin.New()
	engine.GET("/v1/servicos/:id", h.Buscar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos/999", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Ativar_ComSucesso_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	s := servicoExistente(t)
	s.Inativar()
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(s, nil)
	repo.EXPECT().Atualizar(mock.Anything, s).Return(nil)

	h := httpservico.NewHandler(nil, nil, nil, nil, appservico.NewAtivarServicoUseCase(repo), nil, nil)
	engine := gin.New()
	engine.PATCH("/v1/servicos/:id/ativar", h.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/servicos/1/ativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_Ativar_ServicoNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	h := httpservico.NewHandler(nil, nil, nil, nil, appservico.NewAtivarServicoUseCase(repo), nil, nil)
	engine := gin.New()
	engine.PATCH("/v1/servicos/:id/ativar", h.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/servicos/999/ativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Inativar_ComSucesso_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	s := servicoExistente(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(s, nil)
	repo.EXPECT().Atualizar(mock.Anything, s).Return(nil)

	h := httpservico.NewHandler(nil, nil, nil, nil, nil, appservico.NewInativarServicoUseCase(repo), nil)
	engine := gin.New()
	engine.PATCH("/v1/servicos/:id/inativar", h.Inativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/servicos/1/inativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_Inativar_ServicoNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	h := httpservico.NewHandler(nil, nil, nil, nil, nil, appservico.NewInativarServicoUseCase(repo), nil)
	engine := gin.New()
	engine.PATCH("/v1/servicos/:id/inativar", h.Inativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/servicos/999/inativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Listar_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	s := servicoExistente(t)

	repo.EXPECT().
		Listar(mock.Anything, mock.MatchedBy(func(params domainquery.Params) bool {
			return params.Page == 2 && params.Order == "preco_base" && params.Direction == domainquery.DirectionDESC
		})).
		Return(domainquery.Page[*domainservico.Servico]{
			Items:      []*domainservico.Servico{s},
			Total:      20,
			Page:       2,
			PageSize:   20,
			TotalPages: 1,
			Order:      "preco_base",
			Direction:  domainquery.DirectionDESC,
		}, nil)

	h := httpservico.NewHandler(nil, nil, appservico.NewListarServicosUseCase(repo), nil, nil, nil, novoQueryParser())
	engine := gin.New()
	engine.GET("/v1/servicos", h.Listar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos?page=2&order=preco_base&direction=DESC", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpservico.ListarServicosResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, int64(20), resp.Total)
	require.Len(t, resp.Items, 1)
	require.Equal(t, "Troca de óleo", resp.Items[0].Nome)
}

func TestHandler_Listar_ComFiltros_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)
	s := servicoExistente(t)

	repo.EXPECT().
		Listar(mock.Anything, mock.MatchedBy(func(params domainquery.Params) bool {
			for _, filter := range params.Filters {
				if filter.Field == "nome" && filter.Value == "Troca" {
					return true
				}
			}
			return false
		})).
		Return(domainquery.Page[*domainservico.Servico]{
			Items: []*domainservico.Servico{s}, Total: 1, Page: 1, PageSize: 20, TotalPages: 1,
			Order: "id", Direction: domainquery.DirectionASC,
		}, nil)

	h := httpservico.NewHandler(nil, nil, appservico.NewListarServicosUseCase(repo), nil, nil, nil, novoQueryParser())
	engine := gin.New()
	engine.GET("/v1/servicos", h.Listar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos?nome=Troca", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpservico.ListarServicosResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Items, 1)
}

func TestHandler_Listar_ListaVazia_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)

	repo.EXPECT().
		Listar(mock.Anything, mock.MatchedBy(func(params domainquery.Params) bool { return params.Page == 1 })).
		Return(domainquery.Page[*domainservico.Servico]{
			Items: []*domainservico.Servico{}, Total: 0, Page: 1, PageSize: 20, TotalPages: 0,
			Order: "id", Direction: domainquery.DirectionASC,
		}, nil)

	h := httpservico.NewHandler(nil, nil, appservico.NewListarServicosUseCase(repo), nil, nil, nil, novoQueryParser())
	engine := gin.New()
	engine.GET("/v1/servicos", h.Listar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpservico.ListarServicosResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotNil(t, resp.Items)
	require.Empty(t, resp.Items)
}

func TestHandler_Listar_QueryInvalida_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)

	h := httpservico.NewHandler(nil, nil, appservico.NewListarServicosUseCase(repo), nil, nil, nil, novoQueryParser())
	engine := gin.New()
	engine.GET("/v1/servicos", h.Listar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos?page=abc", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Listar_ErroInterno_Retorna500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewServicoRepository(t)

	repo.EXPECT().
		Listar(mock.Anything, mock.Anything).
		Return(domainquery.Page[*domainservico.Servico]{}, errors.New("conexao recusada"))

	h := httpservico.NewHandler(nil, nil, appservico.NewListarServicosUseCase(repo), nil, nil, nil, novoQueryParser())
	engine := gin.New()
	engine.GET("/v1/servicos", h.Listar)

	req := httptest.NewRequest(http.MethodGet, "/v1/servicos", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
