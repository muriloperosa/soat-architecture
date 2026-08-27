package ordemservico_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	clientemocks "github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainveiculo "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	veiculomocks "github.com/muriloperosa/soat-architecture/internal/domain/veiculo/mocks"
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

	handler := httpordemservico.NewHandler(app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo))
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

	handler := httpordemservico.NewHandler(app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo))
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

	handler := httpordemservico.NewHandler(app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo))
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
