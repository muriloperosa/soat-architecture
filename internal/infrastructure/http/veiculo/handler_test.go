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
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainveiculo "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/domain/veiculo/mocks"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	httpveiculo "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/veiculo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func comSubjectAutenticado(engine *gin.Engine) {
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "1", Tipo: domainauth.TipoInterno, Papel: shared.PapelAdmin})
		c.Next()
	})
}

func TestHandler_Cadastrar_RequestValido_Retorna201(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorPlaca(mock.Anything, mock.AnythingOfType("veiculo.Placa")).
		Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)
	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*veiculo.Veiculo")).
		Run(func(ctx context.Context, v *domainveiculo.Veiculo) { v.AtribuirID(1) }).
		Return(nil)

	h := httpveiculo.NewHandler(appveiculo.NewCadastrarVeiculoUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	comSubjectAutenticado(engine)
	engine.POST("/v1/veiculos", h.Cadastrar)

	body, _ := json.Marshal(httpveiculo.CadastrarVeiculoRequest{Placa: "ABC1D23", Marca: "Fiat", Modelo: "Uno", QuilometragemAtual: 15000, Ano: 2020, Cor: "Prata"})
	req := httptest.NewRequest(http.MethodPost, "/v1/veiculos", bytes.NewReader(body))
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

	h := httpveiculo.NewHandler(appveiculo.NewCadastrarVeiculoUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	comSubjectAutenticado(engine)
	engine.POST("/v1/veiculos", h.Cadastrar)

	req := httptest.NewRequest(http.MethodPost, "/v1/veiculos", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Cadastrar_ErroDeValidacaoDoDominio_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)

	h := httpveiculo.NewHandler(appveiculo.NewCadastrarVeiculoUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	comSubjectAutenticado(engine)
	engine.POST("/v1/veiculos", h.Cadastrar)

	body, _ := json.Marshal(httpveiculo.CadastrarVeiculoRequest{Placa: "PLACAINVALIDA", Marca: "Fiat", Modelo: "Uno", QuilometragemAtual: 15000, Ano: 2020, Cor: "Prata"})
	req := httptest.NewRequest(http.MethodPost, "/v1/veiculos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Cadastrar_PlacaJaCadastrada_Retorna409(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorPlaca(mock.Anything, mock.AnythingOfType("veiculo.Placa")).
		Return(veiculoExistente(t), nil)

	h := httpveiculo.NewHandler(appveiculo.NewCadastrarVeiculoUseCase(repo), nil, nil, nil, nil, nil)
	engine := gin.New()
	comSubjectAutenticado(engine)
	engine.POST("/v1/veiculos", h.Cadastrar)

	body, _ := json.Marshal(httpveiculo.CadastrarVeiculoRequest{Placa: "ABC1D23", Marca: "Fiat", Modelo: "Uno", QuilometragemAtual: 15000, Ano: 2020, Cor: "Prata"})
	req := httptest.NewRequest(http.MethodPost, "/v1/veiculos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
}

func veiculoExistente(t *testing.T) *domainveiculo.Veiculo {
	t.Helper()

	v, err := domainveiculo.NewVeiculo("ABC1D23", "Fiat", "Uno", 15000, 2020, "Prata", 1)
	require.NoError(t, err)
	v.AtribuirID(1)

	return v
}

func TestHandler_Atualizar_RequestValido_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(veiculoExistente(t), nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*veiculo.Veiculo")).Return(nil)

	h := httpveiculo.NewHandler(nil, appveiculo.NewAtualizarVeiculoUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/veiculos/:id", h.Atualizar)

	body, _ := json.Marshal(httpveiculo.AtualizarVeiculoRequest{Marca: "Volkswagen", Modelo: "Gol", Cor: "Preto"})
	req := httptest.NewRequest(http.MethodPut, "/v1/veiculos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httpveiculo.VeiculoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "Volkswagen", resp.Marca)
}

func TestHandler_Atualizar_VeiculoNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)

	h := httpveiculo.NewHandler(nil, appveiculo.NewAtualizarVeiculoUseCase(repo), nil, nil, nil, nil)
	engine := gin.New()
	engine.PUT("/v1/veiculos/:id", h.Atualizar)

	body, _ := json.Marshal(httpveiculo.AtualizarVeiculoRequest{Marca: "Volkswagen", Modelo: "Gol", Cor: "Preto"})
	req := httptest.NewRequest(http.MethodPut, "/v1/veiculos/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Ativar_ComSucesso_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	v := veiculoExistente(t)
	v.Inativar()
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(v, nil)
	repo.EXPECT().Atualizar(mock.Anything, v).Return(nil)

	h := httpveiculo.NewHandler(nil, nil, appveiculo.NewAtivarVeiculoUseCase(repo), nil, nil, nil)
	engine := gin.New()
	engine.PATCH("/v1/veiculos/:id/ativar", h.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/veiculos/1/ativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_Ativar_VeiculoNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)

	h := httpveiculo.NewHandler(nil, nil, appveiculo.NewAtivarVeiculoUseCase(repo), nil, nil, nil)
	engine := gin.New()
	engine.PATCH("/v1/veiculos/:id/ativar", h.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/veiculos/999/ativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_Inativar_ComSucesso_Retorna204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	v := veiculoExistente(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(v, nil)
	repo.EXPECT().Atualizar(mock.Anything, v).Return(nil)

	h := httpveiculo.NewHandler(nil, nil, nil, appveiculo.NewInativarVeiculoUseCase(repo), nil, nil)
	engine := gin.New()
	engine.PATCH("/v1/veiculos/:id/inativar", h.Inativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/veiculos/1/inativar", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_ConsultarPorID_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(veiculoExistente(t), nil)

	h := httpveiculo.NewHandler(nil, nil, nil, nil, appveiculo.NewConsultarVeiculoPorIDUseCase(repo), nil)
	engine := gin.New()
	engine.GET("/v1/veiculos/:id", h.ConsultarPorID)

	req := httptest.NewRequest(http.MethodGet, "/v1/veiculos/1", nil)
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
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)

	h := httpveiculo.NewHandler(nil, nil, nil, nil, appveiculo.NewConsultarVeiculoPorIDUseCase(repo), nil)
	engine := gin.New()
	engine.GET("/v1/veiculos/:id", h.ConsultarPorID)

	req := httptest.NewRequest(http.MethodGet, "/v1/veiculos/999", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_ConsultarPorID_ErroInternoDoUseCase_Retorna500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	h := httpveiculo.NewHandler(nil, nil, nil, nil, appveiculo.NewConsultarVeiculoPorIDUseCase(repo), nil)
	engine := gin.New()
	engine.GET("/v1/veiculos/:id", h.ConsultarPorID)

	req := httptest.NewRequest(http.MethodGet, "/v1/veiculos/1", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_ConsultarPorPlaca_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorPlaca(mock.Anything, mock.AnythingOfType("veiculo.Placa")).Return(veiculoExistente(t), nil)

	h := httpveiculo.NewHandler(nil, nil, nil, nil, nil, appveiculo.NewConsultarVeiculoPorPlacaUseCase(repo))
	engine := gin.New()
	engine.GET("/v1/veiculos/placa/:placa", h.ConsultarPorPlaca)

	req := httptest.NewRequest(http.MethodGet, "/v1/veiculos/placa/ABC1D23", nil)
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

	h := httpveiculo.NewHandler(nil, nil, nil, nil, nil, appveiculo.NewConsultarVeiculoPorPlacaUseCase(repo))
	engine := gin.New()
	engine.GET("/v1/veiculos/placa/:placa", h.ConsultarPorPlaca)

	req := httptest.NewRequest(http.MethodGet, "/v1/veiculos/placa/INVALIDA", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_ConsultarPorPlaca_VeiculoNaoEncontrado_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorPlaca(mock.Anything, mock.AnythingOfType("veiculo.Placa")).Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)

	h := httpveiculo.NewHandler(nil, nil, nil, nil, nil, appveiculo.NewConsultarVeiculoPorPlacaUseCase(repo))
	engine := gin.New()
	engine.GET("/v1/veiculos/placa/:placa", h.ConsultarPorPlaca)

	req := httptest.NewRequest(http.MethodGet, "/v1/veiculos/placa/ZZZ9Z99", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
