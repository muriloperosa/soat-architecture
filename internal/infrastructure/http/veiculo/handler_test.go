package veiculo_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainveiculo "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/domain/veiculo/mocks"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	httpveiculo "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/veiculo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func comSubjectAutenticado(engine *gin.Engine) {
	engine.Use(func(c *gin.Context) {
		c.Set(
			middleware.ClaimsContextKey,
			&domainauth.AppClaims{
				Subject: "1",
				Tipo:    domainauth.TipoInterno,
				Papel:   shared.PapelAdmin,
			},
		)
		c.Next()
	})
}

func novoQueryParser() *httpquery.Parser {
	return &httpquery.Parser{}
}

func veiculoExistente(t *testing.T) *domainveiculo.Veiculo {
	t.Helper()

	v, err := domainveiculo.NewVeiculo(
		"ABC1D23",
		"Fiat",
		"Uno",
		15000,
		2020,
		"Prata",
		1,
	)
	require.NoError(t, err)

	v.AtribuirID(1)

	return v
}

func TestHandler_Cadastrar_RequestValido_Retorna201(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		BuscarPorPlaca(
			mock.Anything,
			mock.AnythingOfType("veiculo.Placa"),
		).
		Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)

	repo.
		EXPECT().
		Salvar(
			mock.Anything,
			mock.AnythingOfType("*veiculo.Veiculo"),
		).
		Run(func(ctx context.Context, v *domainveiculo.Veiculo) {
			v.AtribuirID(1)
		}).
		Return(nil)

	h := httpveiculo.NewHandler(
		appveiculo.NewCadastrarVeiculoUseCase(repo),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	comSubjectAutenticado(engine)

	engine.POST("/v1/veiculos", h.Cadastrar)

	body, _ := json.Marshal(httpveiculo.CadastrarVeiculoRequest{
		Placa:              "ABC1D23",
		Marca:              "Fiat",
		Modelo:             "Uno",
		QuilometragemAtual: 15000,
		Ano:                2020,
		Cor:                "Prata",
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/veiculos",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp httpveiculo.VeiculoResponse

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.Equal(t, uint64(1), resp.ID)
	require.Equal(t, uint64(1), resp.CriadoPor)
	require.True(t, resp.Ativo)
}

func TestHandler_Cadastrar_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	h := httpveiculo.NewHandler(
		appveiculo.NewCadastrarVeiculoUseCase(repo),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	comSubjectAutenticado(engine)

	engine.POST("/v1/veiculos", h.Cadastrar)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/veiculos",
		bytes.NewReader([]byte("{invalido")),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Cadastrar_ErroDeValidacaoDoDominio_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	h := httpveiculo.NewHandler(
		appveiculo.NewCadastrarVeiculoUseCase(repo),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	comSubjectAutenticado(engine)

	engine.POST("/v1/veiculos", h.Cadastrar)

	body, _ := json.Marshal(httpveiculo.CadastrarVeiculoRequest{
		Placa:              "PLACAINVALIDA",
		Marca:              "Fiat",
		Modelo:             "Uno",
		QuilometragemAtual: 15000,
		Ano:                2020,
		Cor:                "Prata",
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/veiculos",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Cadastrar_PlacaJaCadastrada_Retorna409(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		BuscarPorPlaca(
			mock.Anything,
			mock.AnythingOfType("veiculo.Placa"),
		).
		Return(veiculoExistente(t), nil)

	h := httpveiculo.NewHandler(
		appveiculo.NewCadastrarVeiculoUseCase(repo),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	comSubjectAutenticado(engine)

	engine.POST("/v1/veiculos", h.Cadastrar)

	body, _ := json.Marshal(httpveiculo.CadastrarVeiculoRequest{
		Placa:              "ABC1D23",
		Marca:              "Fiat",
		Modelo:             "Uno",
		QuilometragemAtual: 15000,
		Ano:                2020,
		Cor:                "Prata",
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/veiculos",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestHandler_Atualizar_RequestValido_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		BuscarPorID(mock.Anything, uint64(1)).
		Return(veiculoExistente(t), nil)

	repo.
		EXPECT().
		Atualizar(
			mock.Anything,
			mock.AnythingOfType("*veiculo.Veiculo"),
		).
		Return(nil)

	h := httpveiculo.NewHandler(
		nil,
		appveiculo.NewAtualizarVeiculoUseCase(repo),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	engine.PUT("/v1/veiculos/:id", h.Atualizar)

	body, _ := json.Marshal(httpveiculo.AtualizarVeiculoRequest{
		Marca:              "Volkswagen",
		Modelo:             "Gol",
		Cor:                "Preto",
		QuilometragemAtual: 16000,
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/veiculos/1",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpveiculo.VeiculoResponse

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.Equal(t, "Volkswagen", resp.Marca)
	require.Equal(t, uint32(16000), resp.QuilometragemAtual)
}

func TestHandler_Atualizar_VeiculoNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		BuscarPorID(mock.Anything, uint64(999)).
		Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)

	h := httpveiculo.NewHandler(
		nil,
		appveiculo.NewAtualizarVeiculoUseCase(repo),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	engine.PUT("/v1/veiculos/:id", h.Atualizar)

	body, _ := json.Marshal(httpveiculo.AtualizarVeiculoRequest{
		Marca:              "Volkswagen",
		Modelo:             "Gol",
		Cor:                "Preto",
		QuilometragemAtual: 16000,
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/veiculos/999",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Atualizar_QuilometragemMenorQueAtual_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		BuscarPorID(mock.Anything, uint64(1)).
		Return(veiculoExistente(t), nil)

	h := httpveiculo.NewHandler(
		nil,
		appveiculo.NewAtualizarVeiculoUseCase(repo),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	engine.PUT("/v1/veiculos/:id", h.Atualizar)

	body, _ := json.Marshal(httpveiculo.AtualizarVeiculoRequest{
		Marca:              "Volkswagen",
		Modelo:             "Gol",
		Cor:                "Preto",
		QuilometragemAtual: 14000,
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/veiculos/1",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Ativar_ComSucesso_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	v := veiculoExistente(t)
	v.Inativar()

	repo.
		EXPECT().
		BuscarPorID(mock.Anything, uint64(1)).
		Return(v, nil)

	repo.
		EXPECT().
		Atualizar(mock.Anything, v).
		Return(nil)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		appveiculo.NewAtivarVeiculoUseCase(repo),
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	engine.PATCH("/v1/veiculos/:id/ativar", h.Ativar)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/v1/veiculos/1/ativar",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_Ativar_VeiculoNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		BuscarPorID(mock.Anything, uint64(999)).
		Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		appveiculo.NewAtivarVeiculoUseCase(repo),
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	engine.PATCH("/v1/veiculos/:id/ativar", h.Ativar)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/v1/veiculos/999/ativar",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Inativar_ComSucesso_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	v := veiculoExistente(t)

	repo.
		EXPECT().
		BuscarPorID(mock.Anything, uint64(1)).
		Return(v, nil)

	repo.
		EXPECT().
		Atualizar(mock.Anything, v).
		Return(nil)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		appveiculo.NewInativarVeiculoUseCase(repo),
		nil,
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	engine.PATCH("/v1/veiculos/:id/inativar", h.Inativar)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/v1/veiculos/1/inativar",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_ConsultarPorID_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		BuscarPorID(mock.Anything, uint64(1)).
		Return(veiculoExistente(t), nil)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewConsultarVeiculoPorIDUseCase(repo),
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	engine.GET("/v1/veiculos/:id", h.ConsultarPorID)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos/1",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpveiculo.VeiculoResponse

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, uint64(1), resp.ID)
}

func TestHandler_ConsultarPorID_VeiculoNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		BuscarPorID(mock.Anything, uint64(999)).
		Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewConsultarVeiculoPorIDUseCase(repo),
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	engine.GET("/v1/veiculos/:id", h.ConsultarPorID)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos/999",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_ConsultarPorID_ErroInternoDoUseCase_Retorna500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		BuscarPorID(mock.Anything, uint64(1)).
		Return(nil, errors.New("conexao recusada"))

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewConsultarVeiculoPorIDUseCase(repo),
		nil,
		nil,
		nil,
	)

	engine := gin.New()
	engine.GET("/v1/veiculos/:id", h.ConsultarPorID)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos/1",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_ConsultarPorPlaca_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		BuscarPorPlaca(
			mock.Anything,
			mock.AnythingOfType("veiculo.Placa"),
		).
		Return(veiculoExistente(t), nil)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewConsultarVeiculoPorPlacaUseCase(repo),
		nil,
		nil,
	)

	engine := gin.New()
	engine.GET("/v1/veiculos/placa/:placa", h.ConsultarPorPlaca)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos/placa/ABC1D23",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpveiculo.VeiculoResponse

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "ABC1D23", resp.Placa)
}

func TestHandler_ConsultarPorPlaca_PlacaInvalida_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewConsultarVeiculoPorPlacaUseCase(repo),
		nil,
		nil,
	)

	engine := gin.New()
	engine.GET("/v1/veiculos/placa/:placa", h.ConsultarPorPlaca)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos/placa/INVALIDA",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_ConsultarPorPlaca_VeiculoNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		BuscarPorPlaca(
			mock.Anything,
			mock.AnythingOfType("veiculo.Placa"),
		).
		Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewConsultarVeiculoPorPlacaUseCase(repo),
		nil,
		nil,
	)

	engine := gin.New()
	engine.GET("/v1/veiculos/placa/:placa", h.ConsultarPorPlaca)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos/placa/ZZZ9Z99",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Listar_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		Listar(mock.Anything, mock.AnythingOfType("query.Params")).
		Return(
			query.Page[*domainveiculo.Veiculo]{
				Items: []*domainveiculo.Veiculo{
					veiculoExistente(t),
				},
				Total:      1,
				Page:       1,
				PageSize:   20,
				TotalPages: 1,
				Order:      "id",
				Direction:  query.DirectionASC,
			},
			nil,
		)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewListarVeiculosUseCase(repo),
		novoQueryParser(),
	)

	engine := gin.New()
	engine.GET("/v1/veiculos", h.Listar)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpveiculo.ListarVeiculosResponse

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.Len(t, resp.Items, 1)
	require.Equal(t, int64(1), resp.Total)
	require.Equal(t, 1, resp.Page)
	require.Equal(t, 20, resp.PageSize)
	require.Equal(t, 1, resp.TotalPages)
	require.Equal(t, "id", resp.Order)
	require.Equal(t, "ASC", resp.Direction)

	require.Equal(t, uint64(1), resp.Items[0].ID)
	require.Equal(t, "ABC1D23", resp.Items[0].Placa)
	require.Equal(t, "Fiat", resp.Items[0].Marca)
	require.Equal(t, "Uno", resp.Items[0].Modelo)
}

func TestHandler_Listar_ListaVazia_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		Listar(mock.Anything, mock.AnythingOfType("query.Params")).
		Return(
			query.Page[*domainveiculo.Veiculo]{
				Items:      []*domainveiculo.Veiculo{},
				Total:      0,
				Page:       1,
				PageSize:   20,
				TotalPages: 0,
				Order:      "id",
				Direction:  query.DirectionASC,
			},
			nil,
		)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewListarVeiculosUseCase(repo),
		novoQueryParser(),
	)

	engine := gin.New()
	engine.GET("/v1/veiculos", h.Listar)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpveiculo.ListarVeiculosResponse

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.NotNil(t, resp.Items)
	require.Empty(t, resp.Items)
	require.Equal(t, int64(0), resp.Total)
	require.Equal(t, 1, resp.Page)
	require.Equal(t, 20, resp.PageSize)
	require.Zero(t, resp.TotalPages)
}

func TestHandler_Listar_ComPaginacaoEOrdenacao_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		Listar(
			mock.Anything,
			mock.MatchedBy(func(params query.Params) bool {
				return params.Page == 2 &&
					params.Order == "ano" &&
					params.Direction == query.DirectionDESC
			}),
		).
		Return(
			query.Page[*domainveiculo.Veiculo]{
				Items:      []*domainveiculo.Veiculo{},
				Total:      25,
				Page:       2,
				PageSize:   20,
				TotalPages: 2,
				Order:      "ano",
				Direction:  query.DirectionDESC,
			},
			nil,
		)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewListarVeiculosUseCase(repo),
		novoQueryParser(),
	)

	engine := gin.New()
	engine.GET("/v1/veiculos", h.Listar)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos?page=2&order=ano&direction=DESC",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpveiculo.ListarVeiculosResponse

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.Equal(t, int64(25), resp.Total)
	require.Equal(t, 2, resp.Page)
	require.Equal(t, 20, resp.PageSize)
	require.Equal(t, 2, resp.TotalPages)
	require.Equal(t, "ano", resp.Order)
	require.Equal(t, "DESC", resp.Direction)
}

func TestHandler_Listar_ComFiltros_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		Listar(
			mock.Anything,
			mock.MatchedBy(func(params query.Params) bool {
				return possuiFiltro(params, "marca", "Fiat") &&
					possuiFiltro(params, "ativo", "true")
			}),
		).
		Return(
			query.Page[*domainveiculo.Veiculo]{
				Items: []*domainveiculo.Veiculo{
					veiculoExistente(t),
				},
				Total:      1,
				Page:       1,
				PageSize:   20,
				TotalPages: 1,
				Order:      "id",
				Direction:  query.DirectionASC,
			},
			nil,
		)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewListarVeiculosUseCase(repo),
		novoQueryParser(),
	)

	engine := gin.New()
	engine.GET("/v1/veiculos", h.Listar)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos?marca=Fiat&ativo=true",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpveiculo.ListarVeiculosResponse

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Items, 1)
}

func TestHandler_Listar_ErroDoUseCase_Retorna500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	repo.
		EXPECT().
		Listar(mock.Anything, mock.AnythingOfType("query.Params")).
		Return(
			query.Page[*domainveiculo.Veiculo]{},
			errors.New("erro ao consultar banco"),
		)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewListarVeiculosUseCase(repo),
		novoQueryParser(),
	)

	engine := gin.New()
	engine.GET("/v1/veiculos", h.Listar)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_Listar_QueryInvalida_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := mocks.NewRepository(t)

	h := httpveiculo.NewHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		appveiculo.NewListarVeiculosUseCase(repo),
		novoQueryParser(),
	)

	engine := gin.New()
	engine.GET("/v1/veiculos", h.Listar)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/veiculos?page=abc",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func possuiFiltro(
	params query.Params,
	field string,
	value string,
) bool {
	for _, filter := range params.Filters {
		if filter.Field == field &&
			filter.Value == value {
			return true
		}
	}

	return false
}
