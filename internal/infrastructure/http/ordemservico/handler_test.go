package ordemservico_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	clientemocks "github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainveiculo "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	veiculomocks "github.com/muriloperosa/soat-architecture/internal/domain/veiculo/mocks"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlerAbrirRetorna201(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ordemRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	veiculoRepo := veiculomocks.NewRepository(t)

	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(clienteValido(t), nil)
	veiculoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(20)).Return(veiculoValido(t), nil)
	ordemRepo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*ordemservico.OrdemServico")).
		Run(func(_ context.Context, os *domainordemservico.OrdemServico) { os.AtribuirID(99) }).
		Return(nil)

	handler := httpordemservico.NewHandler(app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo), nil, nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "30", Tipo: domainauth.TipoInterno, Papel: shared.PapelAdmin})
		c.Next()
	})
	router.POST("/v1/ordens-servico", handler.Abrir)

	body, err := json.Marshal(httpordemservico.AbrirOrdemServicoRequest{
		ClienteID:            10,
		VeiculoID:            20,
		QuilometragemEntrada: 52_300,
		Observacoes:          "Ruído no motor",
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	var response httpordemservico.OrdemServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, uint64(99), response.ID)
	require.Equal(t, "RECEBIDA", response.Status)
	require.Equal(t, uint32(52_300), response.QuilometragemEntrada)
}

func TestHandlerAbrirClienteInexistenteRetorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ordemRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	veiculoRepo := veiculomocks.NewRepository(t)
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domaincliente.ErrClienteNaoEncontrado)

	handler := httpordemservico.NewHandler(app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo), nil, nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "30", Tipo: domainauth.TipoInterno})
		c.Next()
	})
	router.POST("/v1/ordens-servico", handler.Abrir)

	body := []byte(`{"cliente_id":999,"veiculo_id":20,"quilometragem_entrada":100}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandlerAbrirVeiculoInexistenteRetorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ordemRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	veiculoRepo := veiculomocks.NewRepository(t)
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(clienteValido(t), nil)
	veiculoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)

	handler := httpordemservico.NewHandler(app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo), nil, nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "30", Tipo: domainauth.TipoInterno})
		c.Next()
	})
	router.POST("/v1/ordens-servico", handler.Abrir)

	body := []byte(`{"cliente_id":10,"veiculo_id":999,"quilometragem_entrada":100}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/ordens-servico", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func clienteValido(t *testing.T) *domaincliente.Cliente {
	t.Helper()
	cliente, err := domaincliente.NewCliente("52998224725", domaincliente.TipoPessoaFisica, "Maria Silva", "maria@email.com", "11999998888", "senha123", 30)
	require.NoError(t, err)
	cliente.DefinirID(10)
	return &cliente
}

func veiculoValido(t *testing.T) *domainveiculo.Veiculo {
	t.Helper()
	veiculo, err := domainveiculo.NewVeiculo("ABC1D23", "Fiat", "Uno", 52_000, 2020, "Prata", 30)
	require.NoError(t, err)
	veiculo.AtribuirID(20)
	return veiculo
}

func TestHandlerIniciarDiagnosticoRetorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebidaHTTP(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).Return(nil)

	handler := httpordemservico.NewHandler(nil, app.NewIniciarDiagnosticoUseCase(repository), nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "30", Tipo: domainauth.TipoInterno, Papel: shared.PapelMecanico})
		c.Next()
	})
	router.PATCH("/v1/ordens-servico/:id/iniciar-diagnostico", handler.IniciarDiagnostico)

	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/iniciar-diagnostico", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var response httpordemservico.OrdemServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, uint64(42), response.ID)
	require.Equal(t, domainordemservico.StatusEmDiagnostico.String(), response.Status)
	require.Equal(t, uint64(30), os.HistoricoStatus()[1].AlteradoPor())
}

func TestHandlerInformarDiagnosticoRetorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebidaHTTP(t)
	require.NoError(t, os.IniciarDiagnostico(30))
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).Return(nil)

	handler := httpordemservico.NewHandler(nil, nil, app.NewInformarDiagnosticoUseCase(repository), nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.PUT("/v1/ordens-servico/:id/diagnostico", handler.InformarDiagnostico)
	body := []byte(`{"diagnostico":"Falha na bomba de combustível"}`)
	req := httptest.NewRequest(http.MethodPut, "/v1/ordens-servico/42/diagnostico", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var response httpordemservico.OrdemServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, "Falha na bomba de combustível", response.Diagnostico)
}

func TestHandlerInformarDiagnosticoVazioRetorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebidaHTTP(t)
	require.NoError(t, os.IniciarDiagnostico(30))
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	handler := httpordemservico.NewHandler(nil, nil, app.NewInformarDiagnosticoUseCase(repository), nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.PUT("/v1/ordens-servico/:id/diagnostico", handler.InformarDiagnostico)
	req := httptest.NewRequest(http.MethodPut, "/v1/ordens-servico/42/diagnostico", bytes.NewReader([]byte(`{"diagnostico":"   "}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerIniciarExecucaoRetorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoAprovadaHTTP(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).Return(nil)

	handler := httpordemservico.NewHandler(nil, nil, nil, app.NewIniciarExecucaoUseCase(repository), nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "30", Tipo: domainauth.TipoInterno, Papel: shared.PapelMecanico})
		c.Next()
	})
	router.PATCH("/v1/ordens-servico/:id/iniciar-execucao", handler.IniciarExecucao)

	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/iniciar-execucao", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var response httpordemservico.OrdemServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, uint64(42), response.ID)
	require.Equal(t, domainordemservico.StatusEmExecucao.String(), response.Status)
	require.Equal(t, uint64(30), os.HistoricoStatus()[2].AlteradoPor())
}

func TestHandlerIniciarExecucaoTransicaoInvalidaRetorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebidaHTTP(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	handler := httpordemservico.NewHandler(nil, nil, nil, app.NewIniciarExecucaoUseCase(repository), nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "30", Tipo: domainauth.TipoInterno})
		c.Next()
	})
	router.PATCH("/v1/ordens-servico/:id/iniciar-execucao", handler.IniciarExecucao)

	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/iniciar-execucao", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func ordemServicoAprovadaHTTP(t *testing.T) *domainordemservico.OrdemServico {
	t.Helper()
	numero, err := domainordemservico.NewNumeroOrdemServico("OS-20260827-a1b2c3d4e5f6")
	require.NoError(t, err)

	cadastro := time.Date(2026, time.August, 1, 9, 0, 0, 0, time.UTC)
	atualizacao := cadastro.Add(3 * time.Hour)
	historico := []domainordemservico.HistoricoStatus{
		domainordemservico.ReidratarHistoricoStatus(1, 42, domainordemservico.StatusRecebida, cadastro, 30, ""),
		domainordemservico.ReidratarHistoricoStatus(2, 42, domainordemservico.StatusAprovada, atualizacao, 30, ""),
	}

	return domainordemservico.ReidratarOrdemServico(
		42,
		numero,
		10,
		20,
		52_300,
		domainordemservico.StatusAprovada,
		"",
		"",
		30,
		historico,
		cadastro,
		atualizacao,
	)
}

func TestHandlerEntregarRetorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoFinalizadaHTTP(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).Return(nil)

	handler := httpordemservico.NewHandler(nil, nil, nil, nil, app.NewEntregarOrdemServicoUseCase(repository), nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "30", Tipo: domainauth.TipoInterno, Papel: shared.PapelAtendente})
		c.Next()
	})
	router.PATCH("/v1/ordens-servico/:id/entregar", handler.Entregar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/entregar", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var response httpordemservico.OrdemServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, uint64(42), response.ID)
	require.Equal(t, domainordemservico.StatusEntregue.String(), response.Status)
	require.Equal(t, uint64(30), os.HistoricoStatus()[2].AlteradoPor())
}

func TestHandlerEntregarTransicaoInvalidaRetorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebidaHTTP(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	handler := httpordemservico.NewHandler(nil, nil, nil, nil, app.NewEntregarOrdemServicoUseCase(repository), nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "30", Tipo: domainauth.TipoInterno})
		c.Next()
	})
	router.PATCH("/v1/ordens-servico/:id/entregar", handler.Entregar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/ordens-servico/42/entregar", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func ordemServicoFinalizadaHTTP(t *testing.T) *domainordemservico.OrdemServico {
	t.Helper()
	numero, err := domainordemservico.NewNumeroOrdemServico("OS-20260827-a1b2c3d4e5f6")
	require.NoError(t, err)

	cadastro := time.Date(2026, time.August, 1, 9, 0, 0, 0, time.UTC)
	atualizacao := cadastro.Add(3 * time.Hour)
	historico := []domainordemservico.HistoricoStatus{
		domainordemservico.ReidratarHistoricoStatus(1, 42, domainordemservico.StatusRecebida, cadastro, 30, ""),
		domainordemservico.ReidratarHistoricoStatus(2, 42, domainordemservico.StatusFinalizada, atualizacao, 30, ""),
	}

	return domainordemservico.ReidratarOrdemServico(
		42,
		numero,
		10,
		20,
		52_300,
		domainordemservico.StatusFinalizada,
		"",
		"",
		30,
		historico,
		cadastro,
		atualizacao,
	)
}

func ordemServicoRecebidaHTTP(t *testing.T) *domainordemservico.OrdemServico {
	t.Helper()
	os, err := domainordemservico.NewOrdemServico("OS-20260827-a1b2c3d4e5f6", 10, 20, 52_300, "", "", 30)
	require.NoError(t, err)
	os.AtribuirID(42)
	return os
}

func TestHandlerListarOrdensServicoComPaginacaoEFiltros(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebidaHTTP(t)
	params := domainquery.Params{
		Page: 2, Order: "data_cadastro", Direction: domainquery.DirectionDESC,
		Filters: []domainquery.Filter{
			{Field: "cliente_id", Operator: domainquery.OperatorAuto, Value: "10"},
			{Field: "status", Operator: domainquery.OperatorAuto, Value: "RECEBIDA"},
		},
	}
	repository.EXPECT().Listar(mock.Anything, params).Return(domainquery.Page[*domainordemservico.OrdemServico]{
		Items: []*domainordemservico.OrdemServico{os}, Total: 21, Page: 2, PageSize: 20, TotalPages: 2,
		Order: "data_cadastro", Direction: domainquery.DirectionDESC,
	}, nil)

	handler := httpordemservico.NewHandler(nil, nil, nil, nil, nil, nil, nil, app.NewListarOrdensServicoUseCase(repository), httpquery.NewParser())
	router := gin.New()
	router.GET("/v1/ordens-servico", handler.Listar)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/ordens-servico?page=2&order=data_cadastro&direction=desc&status=RECEBIDA&cliente_id=10", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	var response httpordemservico.ListarOrdensServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Len(t, response.Items, 1)
	require.Equal(t, int64(21), response.Total)
	require.Equal(t, 2, response.Page)
	require.Equal(t, 20, response.PageSize)
	require.Equal(t, 2, response.TotalPages)
	require.Equal(t, "data_cadastro", response.Order)
	require.Equal(t, "DESC", response.Direction)
}

func TestHandlerBuscarOrdemServicoPorIDRetorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebidaHTTP(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	handler := httpordemservico.NewHandler(nil, nil, nil, nil, nil, app.NewConsultarOrdemServicoPorIDUseCase(repository), nil, nil, nil)
	router := gin.New()
	router.GET("/v1/ordens-servico/:id", handler.BuscarPorID)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/ordens-servico/42", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	var response httpordemservico.OrdemServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, uint64(42), response.ID)
	require.NotEmpty(t, response.HistoricoStatus)
}

func TestHandlerBuscarOrdemServicoPorNumeroRetorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebidaHTTP(t)
	repository.EXPECT().BuscarPorNumero(mock.Anything, os.Numero().String()).Return(os, nil)

	handler := httpordemservico.NewHandler(nil, nil, nil, nil, nil, nil, app.NewConsultarOrdemServicoPorNumeroUseCase(repository), nil, nil)
	router := gin.New()
	router.GET("/v1/ordens-servico/numero/:numero", handler.BuscarPorNumero)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/ordens-servico/numero/"+os.Numero().String(), nil))

	require.Equal(t, http.StatusOK, rec.Code)
	var response httpordemservico.OrdemServicoResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, os.Numero().String(), response.Numero)
}
