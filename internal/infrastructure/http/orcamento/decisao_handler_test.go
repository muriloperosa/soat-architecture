package orcamento_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	orcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	pecamocks "github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	reservamocks "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	httporcamento "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/orcamento"
	"github.com/muriloperosa/soat-architecture/test/helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func osAguardandoAprovacaoHTTP(t *testing.T, clienteID uint64) *domainordemservico.OrdemServico {
	t.Helper()
	o, err := domainordemservico.NewOrdemServico("OS-20260830-a1b2c3d4e5f6", clienteID, 20, 1000, "", "", 1)
	require.NoError(t, err)
	o.AtribuirID(42)
	require.NoError(t, o.IniciarDiagnostico(2))
	require.NoError(t, o.InformarDiagnostico("Falha identificada"))
	require.NoError(t, o.EnviarParaAprovacao(2))
	return o
}

func orcamentoComServicoValidoHTTP(t *testing.T) *domainorcamento.Orcamento {
	t.Helper()
	o, err := domainorcamento.NewOrcamento(42, "", 2)
	require.NoError(t, err)
	o.AtribuirID(100)
	require.NoError(t, o.AdicionarItemServico(1, 1, 100, 60))
	return o
}

func TestDecisaoHandler_Aprovar_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	runner := &helpers.TransactionRunnerMock{}
	o := osAguardandoAprovacaoHTTP(t, 10)
	orcamento := orcamentoComServicoValidoHTTP(t)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamento, nil)
	reservaRepo.EXPECT().BuscarPorOrdemServico(mock.Anything, uint64(42)).Return([]*domainreservapeca.ReservaPeca{}, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, o).Return(nil)

	uc := app.NewAprovarOrcamentoUseCase(osRepo, orcamentoRepo, pecaRepo, reservaRepo, runner)
	h := httporcamento.NewDecisaoHandler(uc, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "10", Tipo: domainauth.TipoCliente, Papel: shared.PapelCliente})
		c.Next()
	})
	router.PATCH("/v1/ordens-servico/:id/orcamento/aprovar", h.Aprovar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/orcamento/aprovar", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httporcamento.FluxoOrcamentoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, uint64(42), resp.OrdemServicoID)
	require.Equal(t, "APROVADA", resp.Status)
}

func TestDecisaoHandler_Aprovar_ClienteDeOutraOS_Retorna403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	runner := &helpers.TransactionRunnerMock{}
	o := osAguardandoAprovacaoHTTP(t, 10)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)

	uc := app.NewAprovarOrcamentoUseCase(osRepo, orcamentoRepo, pecaRepo, reservaRepo, runner)
	h := httporcamento.NewDecisaoHandler(uc, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "999", Tipo: domainauth.TipoCliente, Papel: shared.PapelCliente})
		c.Next()
	})
	router.PATCH("/v1/ordens-servico/:id/orcamento/aprovar", h.Aprovar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/orcamento/aprovar", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestDecisaoHandler_Rejeitar_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := ordemservicomocks.NewOrdemServicoRepository(t)
	o := osAguardandoAprovacaoHTTP(t, 10)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)
	repo.EXPECT().Atualizar(mock.Anything, o).Return(nil)

	uc := app.NewRejeitarOrcamentoUseCase(repo)
	h := httporcamento.NewDecisaoHandler(nil, uc)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "10", Tipo: domainauth.TipoCliente, Papel: shared.PapelCliente})
		c.Next()
	})
	router.PATCH("/v1/ordens-servico/:id/orcamento/rejeitar", h.Rejeitar)

	body, _ := json.Marshal(httporcamento.RejeitarOrcamentoRequest{Motivo: "Valor acima do esperado"})
	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/orcamento/rejeitar", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httporcamento.FluxoOrcamentoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "REJEITADA", resp.Status)
}

func TestDecisaoHandler_Rejeitar_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := ordemservicomocks.NewOrdemServicoRepository(t)
	uc := app.NewRejeitarOrcamentoUseCase(repo)
	h := httporcamento.NewDecisaoHandler(nil, uc)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "10", Tipo: domainauth.TipoCliente, Papel: shared.PapelCliente})
		c.Next()
	})
	router.PATCH("/v1/ordens-servico/:id/orcamento/rejeitar", h.Rejeitar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/orcamento/rejeitar", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_AlterarQuantidadePeca_ComSucesso_Retorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	runner := &helpers.TransactionRunnerMock{}

	os, err := domainordemservico.NewOrdemServico("OS-20260830-a1b2c3d4e5f6", 10, 20, 1000, "", "", 1)
	require.NoError(t, err)
	os.AtribuirID(42)
	require.NoError(t, os.IniciarDiagnostico(2))

	orcamento := domainorcamento.ReidratarOrcamento(
		100, 42, nil,
		[]domainorcamento.ItemPeca{domainorcamento.ReidratarItemPeca(11, 100, 7, "Pastilha", 2, 50)},
		0, 100, 100, "", 2,
		os.DataCadastro(), os.DataCadastro(),
	)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamento, nil)
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, orcamento).Return(nil)

	uc := app.NewAlterarQuantidadePecaOrcamentoUseCase(orcamentoRepo, osRepo, nil, nil, nil, runner, nil)
	h := httporcamento.NewHandler(nil, nil, nil, nil, nil, uc, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "2", Tipo: domainauth.TipoInterno, Papel: shared.PapelMecanico})
		c.Next()
	})
	router.PATCH("/v1/ordens-servico/:id/orcamento/itens-peca/:itemId/quantidade", h.AlterarQuantidadePeca)

	body, _ := json.Marshal(httporcamento.AlterarQuantidadePecaOrcamentoRequest{Quantidade: 3})
	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/orcamento/itens-peca/11/quantidade", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httporcamento.OrcamentoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 3, resp.ItensPeca[0].Quantidade)
}

func TestHandler_AlterarQuantidadePeca_BodyInvalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	runner := &helpers.TransactionRunnerMock{}

	uc := app.NewAlterarQuantidadePecaOrcamentoUseCase(orcamentoRepo, osRepo, nil, nil, nil, runner, nil)
	h := httporcamento.NewHandler(nil, nil, nil, nil, nil, uc, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "2", Tipo: domainauth.TipoInterno, Papel: shared.PapelMecanico})
		c.Next()
	})
	router.PATCH("/v1/ordens-servico/:id/orcamento/itens-peca/:itemId/quantidade", h.AlterarQuantidadePeca)

	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/orcamento/itens-peca/11/quantidade", bytes.NewReader([]byte("{invalido")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
