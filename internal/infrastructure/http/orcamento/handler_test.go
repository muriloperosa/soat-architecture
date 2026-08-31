package orcamento_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	clientemocks "github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	orcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	pecamocks "github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	servicomocks "github.com/muriloperosa/soat-architecture/internal/domain/servico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	sharedmocks "github.com/muriloperosa/soat-architecture/internal/domain/shared/mocks"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	httporcamento "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/orcamento"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func comClaims(c *gin.Context) {
	c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "30", Tipo: domainauth.TipoInterno, Papel: shared.PapelMecanico})
	c.Next()
}

// osExistenteHTTP devolve uma OS EM_DIAGNOSTICO, o status exigido pra gerar orçamento.
func osExistenteHTTP(t *testing.T) *domainordemservico.OrdemServico {
	t.Helper()
	os, err := domainordemservico.NewOrdemServico("OS-20260827-a1b2c3d4e5f6", 10, 20, 52_300, "", "", 30)
	require.NoError(t, err)
	os.AtribuirID(42)
	require.NoError(t, os.IniciarDiagnostico(30))
	return os
}

// ordemServicoRepoEditavelHTTP mantém os testes de handler focados no fluxo
// HTTP, fornecendo uma OS EM_DIAGNOSTICO para a política de edição.
func ordemServicoRepoEditavelHTTP(t *testing.T) *ordemservicomocks.OrdemServicoRepository {
	t.Helper()
	repo := ordemservicomocks.NewOrdemServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, id uint64) (*domainordemservico.OrdemServico, error) {
			o, err := domainordemservico.NewOrdemServico("OS-20260827-a1b2c3d4e5f6", 10, 20, 52_300, "", "", 30)
			if err != nil {
				return nil, err
			}
			o.AtribuirID(id)
			if err := o.IniciarDiagnostico(30); err != nil {
				return nil, err
			}
			return o, nil
		},
	)
	return repo
}

func TestHandlerGerarRetorna201(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(osExistenteHTTP(t), nil)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(nil, domainorcamento.ErrOrcamentoNaoEncontrado)
	orcamentoRepo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).
		Run(func(_ context.Context, o *domainorcamento.Orcamento) { o.AtribuirID(1) }).
		Return(nil)

	handler := httporcamento.NewHandler(app.NewGerarOrcamentoUseCase(orcamentoRepo, osRepo), nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(comClaims)
	router.POST("/v1/ordens-servico/:id/orcamento", handler.Gerar)

	body, err := json.Marshal(httporcamento.GerarOrcamentoRequest{Observacoes: "Aguardando aprovação"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico/42/orcamento", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	var response httporcamento.OrcamentoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, uint64(1), response.ID)
	require.Equal(t, uint64(42), response.OrdemServicoID)
}

func TestHandlerGerarOrcamentoJaExisteRetorna409(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(osExistenteHTTP(t), nil)
	existente, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(existente, nil)

	handler := httporcamento.NewHandler(app.NewGerarOrcamentoUseCase(orcamentoRepo, osRepo), nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(comClaims)
	router.POST("/v1/ordens-servico/:id/orcamento", handler.Gerar)

	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico/42/orcamento", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestHandlerGerarOSNaoEmDiagnosticoRetorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)

	os, err := domainordemservico.NewOrdemServico("OS-20260827-a1b2c3d4e5f6", 10, 20, 52_300, "", "", 30)
	require.NoError(t, err)
	os.AtribuirID(42)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	handler := httporcamento.NewHandler(app.NewGerarOrcamentoUseCase(orcamentoRepo, osRepo), nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(comClaims)
	router.POST("/v1/ordens-servico/:id/orcamento", handler.Gerar)

	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico/42/orcamento", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerAdicionarServicoRetorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	servicoRepo := servicomocks.NewServicoRepository(t)

	vazio, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	vazio.AtribuirID(1)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(vazio, nil).Once()

	servico, err := domainservico.NewServico("Troca de óleo", "Troca de óleo do motor", 100.0, 60, 30)
	require.NoError(t, err)
	servico.AtribuirID(5)
	servicoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(5)).Return(servico, nil)

	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).Return(nil)

	comItem, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	comItem.AtribuirID(1)
	require.NoError(t, comItem.AdicionarItemServico(5, 2, 100.0, 60))
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(comItem, nil).Once()

	handler := httporcamento.NewHandler(nil, app.NewAdicionarServicoOrcamentoUseCase(orcamentoRepo, servicoRepo, ordemServicoRepoEditavelHTTP(t)), nil, nil, nil, nil, nil)
	router := gin.New()
	router.POST("/v1/ordens-servico/:id/orcamento/itens-servico", handler.AdicionarServico)

	body, err := json.Marshal(httporcamento.AdicionarServicoOrcamentoRequest{ServicoID: 5, Quantidade: 2})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico/42/orcamento/itens-servico", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var response httporcamento.OrcamentoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, 200.0, response.ValorTotal)
	require.Len(t, response.ItensServico, 1)
}

func TestHandlerAdicionarServicoInexistenteRetorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	servicoRepo := servicomocks.NewServicoRepository(t)

	vazio, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(vazio, nil)
	servicoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	handler := httporcamento.NewHandler(nil, app.NewAdicionarServicoOrcamentoUseCase(orcamentoRepo, servicoRepo, ordemServicoRepoEditavelHTTP(t)), nil, nil, nil, nil, nil)
	router := gin.New()
	router.POST("/v1/ordens-servico/:id/orcamento/itens-servico", handler.AdicionarServico)

	body := []byte(`{"servico_id":999,"quantidade":1}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico/42/orcamento/itens-servico", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandlerAdicionarServicoQuantidadeInvalidaRetorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	servicoRepo := servicomocks.NewServicoRepository(t)

	vazio, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(vazio, nil)

	servico, err := domainservico.NewServico("Troca de óleo", "Troca de óleo do motor", 100.0, 60, 30)
	require.NoError(t, err)
	servico.AtribuirID(5)
	servicoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(5)).Return(servico, nil)

	handler := httporcamento.NewHandler(nil, app.NewAdicionarServicoOrcamentoUseCase(orcamentoRepo, servicoRepo, ordemServicoRepoEditavelHTTP(t)), nil, nil, nil, nil, nil)
	router := gin.New()
	router.POST("/v1/ordens-servico/:id/orcamento/itens-servico", handler.AdicionarServico)

	body := []byte(`{"servico_id":5,"quantidade":-1}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico/42/orcamento/itens-servico", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func pecaExistenteHTTP(t *testing.T) *domainpeca.Peca {
	t.Helper()
	peca, err := domainpeca.NewPeca("Filtro de óleo", "Marca X", "Filtro de óleo do motor", 50.0, 10, 2, 30)
	require.NoError(t, err)
	peca.AtribuirID(9)
	return peca
}

func TestHandlerAdicionarPecaRetorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)

	vazio, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	vazio.AtribuirID(1)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(vazio, nil).Once()

	pecaRepo.EXPECT().BuscarPorID(mock.Anything, uint64(9)).Return(pecaExistenteHTTP(t), nil)

	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).Return(nil)

	comItem, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	comItem.AtribuirID(1)
	require.NoError(t, comItem.AdicionarItemPeca(9, "Filtro de óleo do motor", 3, 50.0))
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(comItem, nil).Once()

	handler := httporcamento.NewHandler(nil, nil, app.NewAdicionarPecaOrcamentoUseCase(orcamentoRepo, pecaRepo, ordemServicoRepoEditavelHTTP(t)), nil, nil, nil, nil)
	router := gin.New()
	router.POST("/v1/ordens-servico/:id/orcamento/itens-peca", handler.AdicionarPeca)

	body, err := json.Marshal(httporcamento.AdicionarPecaOrcamentoRequest{PecaID: 9, Quantidade: 3})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico/42/orcamento/itens-peca", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var response httporcamento.OrcamentoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, 150.0, response.ValorTotal)
	require.Len(t, response.ItensPeca, 1)
}

func TestHandlerAdicionarPecaInexistenteRetorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)

	vazio, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(vazio, nil)
	pecaRepo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainpeca.ErrPecaNaoEncontrada)

	handler := httporcamento.NewHandler(nil, nil, app.NewAdicionarPecaOrcamentoUseCase(orcamentoRepo, pecaRepo, ordemServicoRepoEditavelHTTP(t)), nil, nil, nil, nil)
	router := gin.New()
	router.POST("/v1/ordens-servico/:id/orcamento/itens-peca", handler.AdicionarPeca)

	body := []byte(`{"peca_id":999,"quantidade":1}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico/42/orcamento/itens-peca", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandlerAdicionarPecaQuantidadeInvalidaRetorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)

	vazio, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(vazio, nil)
	pecaRepo.EXPECT().BuscarPorID(mock.Anything, uint64(9)).Return(pecaExistenteHTTP(t), nil)

	handler := httporcamento.NewHandler(nil, nil, app.NewAdicionarPecaOrcamentoUseCase(orcamentoRepo, pecaRepo, ordemServicoRepoEditavelHTTP(t)), nil, nil, nil, nil)
	router := gin.New()
	router.POST("/v1/ordens-servico/:id/orcamento/itens-peca", handler.AdicionarPeca)

	body := []byte(`{"peca_id":9,"quantidade":-2}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico/42/orcamento/itens-peca", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerRemoverServicoRetorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)

	comItem, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	comItem.AtribuirID(1)
	require.NoError(t, comItem.AdicionarItemServico(5, 2, 100.0, 60))
	item := comItem.ItensServico()[0]
	reidratado := domainorcamento.ReidratarOrcamento(
		1, 42,
		[]domainorcamento.ItemServico{domainorcamento.ReidratarItemServico(9, 1, item.ServicoID(), item.Quantidade(), item.Valor(), item.TempoEstimado())},
		nil,
		comItem.ValorItemServicos(), comItem.ValorItemPecas(), comItem.ValorTotal(),
		"", 30, comItem.CriadoEm(), comItem.AtualizadoEm(),
	)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(reidratado, nil).Once()
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).Return(nil)

	vazio, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	vazio.AtribuirID(1)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(vazio, nil).Once()

	handler := httporcamento.NewHandler(nil, nil, nil, app.NewRemoverServicoOrcamentoUseCase(orcamentoRepo, ordemServicoRepoEditavelHTTP(t)), nil, nil, nil)
	router := gin.New()
	router.DELETE("/v1/ordens-servico/:id/orcamento/itens-servico/:itemId", handler.RemoverServico)

	req := httptest.NewRequest(http.MethodDelete, "/v1/ordens-servico/42/orcamento/itens-servico/9", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var response httporcamento.OrcamentoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Empty(t, response.ItensServico)
}

func TestHandlerRemoverServicoItemInexistenteRetorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)

	vazio, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(vazio, nil)

	handler := httporcamento.NewHandler(nil, nil, nil, app.NewRemoverServicoOrcamentoUseCase(orcamentoRepo, ordemServicoRepoEditavelHTTP(t)), nil, nil, nil)
	router := gin.New()
	router.DELETE("/v1/ordens-servico/:id/orcamento/itens-servico/:itemId", handler.RemoverServico)

	req := httptest.NewRequest(http.MethodDelete, "/v1/ordens-servico/42/orcamento/itens-servico/999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandlerRemoverPecaRetorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)

	comItem, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	comItem.AtribuirID(1)
	require.NoError(t, comItem.AdicionarItemPeca(9, "Filtro de óleo", 3, 50.0))
	item := comItem.ItensPeca()[0]
	reidratado := domainorcamento.ReidratarOrcamento(
		1, 42,
		nil,
		[]domainorcamento.ItemPeca{domainorcamento.ReidratarItemPeca(11, 1, item.PecaID(), item.Descricao(), item.Quantidade(), item.Valor())},
		comItem.ValorItemServicos(), comItem.ValorItemPecas(), comItem.ValorTotal(),
		"", 30, comItem.CriadoEm(), comItem.AtualizadoEm(),
	)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(reidratado, nil).Once()
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).Return(nil)

	vazio, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	vazio.AtribuirID(1)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(vazio, nil).Once()

	handler := httporcamento.NewHandler(nil, nil, nil, nil, app.NewRemoverPecaOrcamentoUseCase(orcamentoRepo, ordemServicoRepoEditavelHTTP(t)), nil, nil)
	router := gin.New()
	router.DELETE("/v1/ordens-servico/:id/orcamento/itens-peca/:itemId", handler.RemoverPeca)

	req := httptest.NewRequest(http.MethodDelete, "/v1/ordens-servico/42/orcamento/itens-peca/11", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var response httporcamento.OrcamentoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Empty(t, response.ItensPeca)
}

func TestHandlerRemoverPecaItemInexistenteRetorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)

	vazio, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(vazio, nil)

	handler := httporcamento.NewHandler(nil, nil, nil, nil, app.NewRemoverPecaOrcamentoUseCase(orcamentoRepo, ordemServicoRepoEditavelHTTP(t)), nil, nil)
	router := gin.New()
	router.DELETE("/v1/ordens-servico/:id/orcamento/itens-peca/:itemId", handler.RemoverPeca)

	req := httptest.NewRequest(http.MethodDelete, "/v1/ordens-servico/42/orcamento/itens-peca/999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func osEmDiagnosticoComDiagnosticoHTTP(t *testing.T) *domainordemservico.OrdemServico {
	t.Helper()
	os, err := domainordemservico.NewOrdemServico("OS-20260827-a1b2c3d4e5f6", 10, 20, 52_300, "", "", 30)
	require.NoError(t, err)
	os.AtribuirID(42)
	require.NoError(t, os.IniciarDiagnostico(30))
	require.NoError(t, os.InformarDiagnostico("Falha na bomba de combustível"))
	return os
}

func TestHandlerEnviarParaAprovacaoRetorna200EEnviaEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	o, err := domainorcamento.NewOrcamento(42, "", 30)
	require.NoError(t, err)
	require.NoError(t, o.AdicionarItemServico(5, 1, 100.0, 60))
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(o, nil)

	os := osEmDiagnosticoComDiagnosticoHTTP(t)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, os).Return(nil)

	cliente, err := domaincliente.NewCliente("52998224725", domaincliente.TipoPessoaFisica, "Maria Silva", "maria@email.com", "11999998888", "senha123", 3)
	require.NoError(t, err)
	cliente.DefinirID(10)
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(&cliente, nil)

	emailSender.EXPECT().Enviar(mock.Anything, "maria@email.com", mock.Anything, mock.Anything).Return(nil)

	handler := httporcamento.NewHandler(nil, nil, nil, nil, nil, nil, app.NewEnviarOrcamentoParaAprovacaoUseCase(orcamentoRepo, osRepo, clienteRepo, emailSender))
	router := gin.New()
	router.Use(comClaims)
	router.PATCH("/v1/ordens-servico/:id/orcamento/enviar-aprovacao", handler.EnviarParaAprovacao)

	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/orcamento/enviar-aprovacao", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, domainordemservico.StatusAguardandoAprovacao, os.Status())
}

func TestHandlerEnviarParaAprovacaoOrcamentoInexistenteRetorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(nil, domainorcamento.ErrOrcamentoNaoEncontrado)

	handler := httporcamento.NewHandler(nil, nil, nil, nil, nil, nil, app.NewEnviarOrcamentoParaAprovacaoUseCase(orcamentoRepo, osRepo, clienteRepo, emailSender))
	router := gin.New()
	router.Use(comClaims)
	router.PATCH("/v1/ordens-servico/:id/orcamento/enviar-aprovacao", handler.EnviarParaAprovacao)

	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/orcamento/enviar-aprovacao", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
