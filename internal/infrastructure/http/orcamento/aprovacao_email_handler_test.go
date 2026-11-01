package orcamento_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/application/notificacao"
	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	orcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	pecamocks "github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	reservamocks "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httporcamento "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/orcamento"
	"github.com/muriloperosa/soat-architecture/test/helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAprovacaoEmailHandler_ConfirmarApenasExibeFormulario(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := &approvalTokenHandlerStub{}
	handler := httporcamento.NewAprovacaoEmailHandler(tokens, "https://oficina.example", nil, nil)
	engine := gin.New()
	engine.GET("/v1/orcamentos/aprovacao", handler.Confirmar)

	request := httptest.NewRequest(http.MethodGet, "/v1/orcamentos/aprovacao?token=token-valido&acao=REJEITAR", nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Header().Get("Content-Type"), "text/html")
	require.Contains(t, response.Body.String(), "name=\"token\" value=\"token-valido\"")
	require.Contains(t, response.Body.String(), "name=\"motivo\"")
	require.Contains(t, response.Body.String(), "https://oficina.example/v1/orcamentos/aprovacao/decisao")
	require.Equal(t, 1, tokens.validations)
}

func TestAprovacaoEmailHandler_ConfirmarRejeitaTokenInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := &approvalTokenHandlerStub{err: shared.NewUnauthorizedError("invalid")}
	handler := httporcamento.NewAprovacaoEmailHandler(tokens, "https://oficina.example", nil, nil)
	engine := gin.New()
	engine.GET("/v1/orcamentos/aprovacao", handler.Confirmar)

	request := httptest.NewRequest(http.MethodGet, "/v1/orcamentos/aprovacao?token=inv%C3%A1lido&acao=APROVAR", nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	require.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestAprovacaoEmailHandler_ConfirmarRejeitaAcaoInvalidaSemValidarToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := &approvalTokenHandlerStub{}
	handler := httporcamento.NewAprovacaoEmailHandler(tokens, "https://oficina.example", nil, nil)
	engine := gin.New()
	engine.GET("/v1/orcamentos/aprovacao", handler.Confirmar)

	request := httptest.NewRequest(http.MethodGet, "/v1/orcamentos/aprovacao?token=token-valido&acao=IGNORAR", nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Zero(t, tokens.validations)
}

func TestAprovacaoEmailHandler_DecidirRejeitaTokenInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := &approvalTokenHandlerStub{err: shared.NewUnauthorizedError("invalid")}
	handler := httporcamento.NewAprovacaoEmailHandler(tokens, "https://oficina.example", nil, nil)
	engine := gin.New()
	engine.POST("/v1/orcamentos/aprovacao/decisao", handler.Decidir)

	response := postApprovalDecision(t, engine, url.Values{"token": {"token-invalido"}, "decisao": {"APROVAR"}})

	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.Equal(t, 1, tokens.validations)
}

func TestAprovacaoEmailHandler_DecidirRejeitaDecisaoInvalida(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := &approvalTokenHandlerStub{claims: shared.ApprovalTokenClaims{OrdemServicoID: 71, ClienteID: 92}}
	handler := httporcamento.NewAprovacaoEmailHandler(tokens, "https://oficina.example", nil, nil)
	engine := gin.New()
	engine.POST("/v1/orcamentos/aprovacao/decisao", handler.Decidir)

	response := postApprovalDecision(t, engine, url.Values{"token": {"token-valido"}, "decisao": {"IGNORAR"}})

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Equal(t, 1, tokens.validations)
}

func TestAprovacaoEmailHandler_DecidirRejeicaoSemMotivoNaoAlteraOS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := &approvalTokenHandlerStub{claims: shared.ApprovalTokenClaims{OrdemServicoID: 71, ClienteID: 92}}
	os := ordemAguardandoAprovacaoHTTP(t, 92)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(71)).Return(os, nil)
	rejeitar := app.NewRejeitarOrcamentoUseCase(osRepo, notificacao.NewStatus(nil, nil))
	handler := httporcamento.NewAprovacaoEmailHandler(tokens, "https://oficina.example", nil, rejeitar)
	engine := gin.New()
	engine.POST("/v1/orcamentos/aprovacao/decisao", handler.Decidir)

	response := postApprovalDecision(t, engine, url.Values{"token": {"token-valido"}, "decisao": {"REJEITAR"}})

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Equal(t, domainordemservico.StatusAguardandoAprovacao, os.Status())
}

func TestAprovacaoEmailHandler_DecidirDerivaIdentificadoresDoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := &approvalTokenHandlerStub{claims: shared.ApprovalTokenClaims{OrdemServicoID: 71, ClienteID: 92}}
	os := ordemAguardandoAprovacaoHTTP(t, 92)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	orcamento, err := domainorcamento.NewOrcamento(71, "", 2)
	require.NoError(t, err)
	require.NoError(t, orcamento.AdicionarItemServico(1, 1, 100, 60))
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(71)).Return(os, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, os).Return(nil)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(71)).Return(orcamento, nil)
	reservaRepo.EXPECT().BuscarPorOrdemServico(mock.Anything, uint64(71)).Return([]*domainreservapeca.ReservaPeca{}, nil)
	aprovar := app.NewAprovarOrcamentoUseCase(
		osRepo,
		orcamentoRepo,
		pecamocks.NewRepository(t),
		reservaRepo,
		&helpers.TransactionRunnerMock{},
		notificacao.NewStatus(nil, nil),
	)
	handler := httporcamento.NewAprovacaoEmailHandler(tokens, "https://oficina.example", aprovar, nil)
	engine := gin.New()
	engine.POST("/v1/orcamentos/aprovacao/decisao", handler.Decidir)

	form := url.Values{"token": {"token-valido"}, "decisao": {"APROVAR"}, "ordem_servico_id": {"1"}, "cliente_id": {"2"}}
	request := httptest.NewRequest(http.MethodPost, "/v1/orcamentos/aprovacao/decisao", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, domainordemservico.StatusAprovada, os.Status())
	require.Contains(t, response.Body.String(), "Orçamento aprovado")
}

func TestAprovacaoEmailHandler_DecidirPassaMotivoParaUseCaseDeRejeicao(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := &approvalTokenHandlerStub{claims: shared.ApprovalTokenClaims{OrdemServicoID: 71, ClienteID: 92}}
	os := ordemAguardandoAprovacaoHTTP(t, 92)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(71)).Return(os, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, os).Return(nil)
	rejeitar := app.NewRejeitarOrcamentoUseCase(osRepo, notificacao.NewStatus(nil, nil))
	handler := httporcamento.NewAprovacaoEmailHandler(tokens, "https://oficina.example", nil, rejeitar)
	engine := gin.New()
	engine.POST("/v1/orcamentos/aprovacao/decisao", handler.Decidir)

	form := url.Values{"token": {"token-valido"}, "decisao": {"REJEITAR"}, "motivo": {"Preciso revisar o valor"}, "ordem_servico_id": {"1"}, "cliente_id": {"2"}}
	request := httptest.NewRequest(http.MethodPost, "/v1/orcamentos/aprovacao/decisao", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, domainordemservico.StatusRejeitada, os.Status())
	require.Contains(t, response.Body.String(), "Orçamento recusado")
}

func ordemAguardandoAprovacaoHTTP(t *testing.T, clienteID uint64) *domainordemservico.OrdemServico {
	t.Helper()
	os, err := domainordemservico.NewOrdemServico("OS-20261101-a1b2c3d4e5f6", clienteID, 20, 1000, "", "", 1)
	require.NoError(t, err)
	os.AtribuirID(71)
	require.NoError(t, os.IniciarDiagnostico(2))
	require.NoError(t, os.InformarDiagnostico("Diagnóstico realizado"))
	require.NoError(t, os.EnviarParaAprovacao(2))
	return os
}

func postApprovalDecision(t *testing.T, engine *gin.Engine, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/v1/orcamentos/aprovacao/decisao", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response
}

type approvalTokenHandlerStub struct {
	claims      shared.ApprovalTokenClaims
	err         error
	validations int
}

func (s *approvalTokenHandlerStub) GerarToken(uint64, uint64) (string, error) { return "", nil }

func (s *approvalTokenHandlerStub) ValidarToken(string) (shared.ApprovalTokenClaims, error) {
	s.validations++
	return s.claims, s.err
}
