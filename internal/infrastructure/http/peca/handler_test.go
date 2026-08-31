package peca_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	apppeca "github.com/muriloperosa/soat-architecture/internal/application/peca"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	reservamocks "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	httppeca "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/peca"
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

func TestHandler_Cadastrar_RequestValido_Retorna201(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*peca.Peca")).
		Run(func(ctx context.Context, p *domainpeca.Peca) { p.AtribuirID(1) }).
		Return(nil)

	h := httppeca.NewHandler(apppeca.NewCadastrarPecaUseCase(repo), nil, nil, nil, nil, nil, nil, nil)
	engine := gin.New()
	comSubjectAutenticado(engine)
	engine.POST("/v1/pecas", h.Cadastrar)

	body, _ := json.Marshal(httppeca.CadastrarPecaRequest{Nome: "Pastilha", Marca: "Bosch", Descricao: "Pastilha dianteira", Preco: 89.9, QuantidadeEmEstoque: 20, EstoqueMinimo: 5})
	req := httptest.NewRequest(http.MethodPost, "/v1/pecas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp httppeca.PecaResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, uint64(1), resp.ID)
	require.Equal(t, uint64(1), resp.CriadoPor)
	require.True(t, resp.Ativo)
}

func TestHandler_Cadastrar_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)

	h := httppeca.NewHandler(apppeca.NewCadastrarPecaUseCase(repo), nil, nil, nil, nil, nil, nil, nil)
	engine := gin.New()
	comSubjectAutenticado(engine)
	engine.POST("/v1/pecas", h.Cadastrar)

	req := httptest.NewRequest(http.MethodPost, "/v1/pecas", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Cadastrar_ErroDeValidacaoDoDominio_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)

	h := httppeca.NewHandler(apppeca.NewCadastrarPecaUseCase(repo), nil, nil, nil, nil, nil, nil, nil)
	engine := gin.New()
	comSubjectAutenticado(engine)
	engine.POST("/v1/pecas", h.Cadastrar)

	body, _ := json.Marshal(httppeca.CadastrarPecaRequest{Nome: "Pastilha", Marca: "Bosch", Descricao: "Pastilha dianteira", Preco: -1, QuantidadeEmEstoque: 20, EstoqueMinimo: 5})
	req := httptest.NewRequest(http.MethodPost, "/v1/pecas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func pecaExistente(t *testing.T) *domainpeca.Peca {
	t.Helper()

	p, err := domainpeca.NewPeca("Pastilha", "Bosch", "Pastilha dianteira", 89.9, 20, 5, 1)
	require.NoError(t, err)
	p.AtribuirID(1)

	return p
}

func TestHandler_Atualizar_RequestValido_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(pecaExistente(t), nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*peca.Peca")).Return(nil)

	h := httppeca.NewHandler(nil, apppeca.NewAtualizarPecaUseCase(repo), nil, nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/pecas/:id", h.Atualizar)

	body, _ := json.Marshal(httppeca.AtualizarPecaRequest{Nome: "Pastilha nova", Marca: "Bosch", Descricao: "Pastilha dianteira", Preco: 99.9, EstoqueMinimo: 8})
	req := httptest.NewRequest(http.MethodPut, "/v1/pecas/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httppeca.PecaResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "Pastilha nova", resp.Nome)
}

func TestHandler_Atualizar_PecaNaoEncontrada_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainpeca.ErrPecaNaoEncontrada)

	h := httppeca.NewHandler(nil, apppeca.NewAtualizarPecaUseCase(repo), nil, nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/pecas/:id", h.Atualizar)

	body, _ := json.Marshal(httppeca.AtualizarPecaRequest{Nome: "Pastilha nova", Marca: "Bosch", Descricao: "Pastilha dianteira", Preco: 99.9, EstoqueMinimo: 8})
	req := httptest.NewRequest(http.MethodPut, "/v1/pecas/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Ativar_ComSucesso_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	p := pecaExistente(t)
	p.Inativar()
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(p, nil)
	repo.EXPECT().Atualizar(mock.Anything, p).Return(nil)

	h := httppeca.NewHandler(nil, nil, apppeca.NewAtivarPecaUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.PATCH("/v1/pecas/:id/ativar", h.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/pecas/1/ativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_Ativar_PecaNaoEncontrada_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainpeca.ErrPecaNaoEncontrada)

	h := httppeca.NewHandler(nil, nil, apppeca.NewAtivarPecaUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	engine.PATCH("/v1/pecas/:id/ativar", h.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/pecas/999/ativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Inativar_ComSucesso_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	p := pecaExistente(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(p, nil)
	repo.EXPECT().Atualizar(mock.Anything, p).Return(nil)

	h := httppeca.NewHandler(nil, nil, nil, apppeca.NewInativarPecaUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PATCH("/v1/pecas/:id/inativar", h.Inativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/pecas/1/inativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_ConsultarPorID_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(pecaExistente(t), nil)
	reservaRepo.EXPECT().SomarQuantidadeReservada(mock.Anything, uint64(1)).Return(4, nil)

	h := httppeca.NewHandler(nil, nil, nil, nil, apppeca.NewConsultarPecaPorIDUseCase(repo, reservaRepo), nil, nil, nil)
	engine := gin.New()
	engine.GET("/v1/pecas/:id", h.ConsultarPorID)

	req := httptest.NewRequest(http.MethodGet, "/v1/pecas/1", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httppeca.PecaResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, uint64(1), resp.ID)
	require.Equal(t, 4, resp.QuantidadeReservada)
	require.Equal(t, resp.QuantidadeEmEstoque-4, resp.QuantidadeDisponivel)
}

func TestHandler_ConsultarPorID_PecaNaoEncontrada_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainpeca.ErrPecaNaoEncontrada)

	h := httppeca.NewHandler(nil, nil, nil, nil, apppeca.NewConsultarPecaPorIDUseCase(repo, reservaRepo), nil, nil, nil)
	engine := gin.New()
	engine.GET("/v1/pecas/:id", h.ConsultarPorID)

	req := httptest.NewRequest(http.MethodGet, "/v1/pecas/999", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_ReporEstoque_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	p := pecaExistente(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(p, nil)
	repo.EXPECT().Atualizar(mock.Anything, p).Return(nil)

	h := httppeca.NewHandler(nil, nil, nil, nil, nil, nil, apppeca.NewReporEstoqueUseCase(repo), nil)
	engine := gin.New()
	engine.PATCH("/v1/pecas/:id/repor-estoque", h.ReporEstoque)

	body, _ := json.Marshal(httppeca.ReporEstoqueRequest{Quantidade: 10})
	req := httptest.NewRequest(http.MethodPatch, "/v1/pecas/1/repor-estoque", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httppeca.PecaResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 30, resp.QuantidadeEmEstoque)
}

func TestHandler_ReporEstoque_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)

	h := httppeca.NewHandler(nil, nil, nil, nil, nil, nil, apppeca.NewReporEstoqueUseCase(repo), nil)
	engine := gin.New()
	engine.PATCH("/v1/pecas/:id/repor-estoque", h.ReporEstoque)

	req := httptest.NewRequest(http.MethodPatch, "/v1/pecas/1/repor-estoque", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_ReporEstoque_ErroInternoDoUseCase_Retorna500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	p := pecaExistente(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(p, nil)
	repo.EXPECT().Atualizar(mock.Anything, p).Return(errors.New("conexao recusada"))

	h := httppeca.NewHandler(nil, nil, nil, nil, nil, nil, apppeca.NewReporEstoqueUseCase(repo), nil)
	engine := gin.New()
	engine.PATCH("/v1/pecas/:id/repor-estoque", h.ReporEstoque)

	body, _ := json.Marshal(httppeca.ReporEstoqueRequest{Quantidade: 10})
	req := httptest.NewRequest(http.MethodPatch, "/v1/pecas/1/repor-estoque", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_Listar_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)
	p := pecaExistente(t)

	repo.
		EXPECT().
		Listar(
			mock.Anything,
			mock.MatchedBy(func(params domainquery.Params) bool {
				return params.Page == 2 &&
					params.Order == "preco" &&
					params.Direction == domainquery.DirectionDESC
			}),
		).
		Return(
			domainquery.Page[*domainpeca.Peca]{
				Items: []*domainpeca.Peca{
					p,
				},
				Total:      20,
				Page:       2,
				PageSize:   20,
				TotalPages: 1,
				Order:      "preco",
				Direction:  domainquery.DirectionDESC,
			},
			nil,
		)

	h := httppeca.NewHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		apppeca.NewListarPecasUseCase(repo),
		nil,
		novoQueryParser(),
	)

	engine := gin.New()
	engine.GET("/v1/pecas", h.Listar)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/pecas?page=2&order=preco&direction=DESC",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httppeca.ListarPecasResponse

	require.NoError(
		t,
		json.Unmarshal(rec.Body.Bytes(), &resp),
	)

	require.Equal(t, int64(20), resp.Total)
	require.Equal(t, 2, resp.Page)
	require.Equal(t, 20, resp.PageSize)
	require.Equal(t, 1, resp.TotalPages)
	require.Equal(t, "preco", resp.Order)
	require.Equal(t, "DESC", resp.Direction)

	require.Len(t, resp.Items, 1)

	require.Equal(t, uint64(1), resp.Items[0].ID)
	require.Equal(t, "Pastilha", resp.Items[0].Nome)
	require.Equal(t, "Bosch", resp.Items[0].Marca)
	require.Equal(t, 89.9, resp.Items[0].Preco)
	require.Equal(t, 20, resp.Items[0].QuantidadeEmEstoque)
	require.Equal(t, 5, resp.Items[0].EstoqueMinimo)
}

func TestHandler_Listar_ListaVazia_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		Listar(
			mock.Anything,
			mock.MatchedBy(func(params domainquery.Params) bool {
				return params.Page == 1
			}),
		).
		Return(
			domainquery.Page[*domainpeca.Peca]{
				Items:      []*domainpeca.Peca{},
				Total:      0,
				Page:       1,
				PageSize:   20,
				TotalPages: 0,
				Order:      "id",
				Direction:  domainquery.DirectionASC,
			},
			nil,
		)

	h := httppeca.NewHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		apppeca.NewListarPecasUseCase(repo),
		nil,
		novoQueryParser(),
	)

	engine := gin.New()
	engine.GET("/v1/pecas", h.Listar)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/pecas",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httppeca.ListarPecasResponse

	require.NoError(
		t,
		json.Unmarshal(rec.Body.Bytes(), &resp),
	)

	require.NotNil(t, resp.Items)
	require.Empty(t, resp.Items)

	require.Equal(t, int64(0), resp.Total)
	require.Equal(t, 1, resp.Page)
	require.Equal(t, 20, resp.PageSize)
	require.Zero(t, resp.TotalPages)
	require.Equal(t, "id", resp.Order)
	require.Equal(t, "ASC", resp.Direction)
}

func TestHandler_Listar_ComFiltro_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)
	p := pecaExistente(t)

	repo.
		EXPECT().
		Listar(
			mock.Anything,
			mock.MatchedBy(func(params domainquery.Params) bool {
				if len(params.Filters) != 1 {
					return false
				}

				filter := params.Filters[0]

				return filter.Field == "marca" &&
					filter.Value == "Bosch"
			}),
		).
		Return(
			domainquery.Page[*domainpeca.Peca]{
				Items: []*domainpeca.Peca{
					p,
				},
				Total:     1,
				Page:      2,
				Order:     "id",
				Direction: domainquery.DirectionASC,
			},
			nil,
		)

	h := httppeca.NewHandler(nil, nil, nil, nil, nil, apppeca.NewListarPecasUseCase(repo), nil, novoQueryParser())

	engine := gin.New()
	engine.GET("/v1/pecas", h.Listar)

	req := httptest.NewRequest(http.MethodGet, "/v1/pecas?marca=Bosch", nil)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httppeca.ListarPecasResponse

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.Equal(t, int64(1), resp.Total)
	require.Len(t, resp.Items, 1)

	require.Equal(t, "Bosch", resp.Items[0].Marca)
}

func TestHandler_Listar_QueryInvalida_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	h := httppeca.NewHandler(nil, nil, nil, nil, nil, apppeca.NewListarPecasUseCase(repo), nil, novoQueryParser())

	engine := gin.New()
	engine.GET("/v1/pecas", h.Listar)

	req := httptest.NewRequest(http.MethodGet, "/v1/pecas?page=abc", nil)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Listar_ErroInterno_Retorna500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		Listar(mock.Anything, mock.Anything).
		Return(domainquery.Page[*domainpeca.Peca]{}, errors.New("conexao recusada"))

	h := httppeca.NewHandler(nil, nil, nil, nil, nil, apppeca.NewListarPecasUseCase(repo), nil, novoQueryParser())

	engine := gin.New()
	engine.GET("/v1/pecas", h.Listar)

	req := httptest.NewRequest(http.MethodGet, "/v1/pecas", nil)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
