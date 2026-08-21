package cliente

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	appcliente "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func clienteValido(t *testing.T, id uint64) *domaincliente.Cliente {
	t.Helper()
	cliente, err := domaincliente.NewCliente("529.982.247-25", domaincliente.TipoPessoaFisica, "João da Silva", "joao@email.com", "(44) 99999-1234", "senha123", uint64(1))
	require.NoError(t, err)
	cliente.DefinirID(id)
	return &cliente
}

func requestJSON(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	data, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(method, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func withSubject(subject string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: subject})
		c.Next()
	}
}

func TestNewHandler(t *testing.T) {
	repo := mocks.NewClienteRepository(t)
	criar := appcliente.NewCriarClienteUseCase(repo)
	atualizar := appcliente.NewAtualizarClienteUseCase(repo)
	consultarPorID := appcliente.NewConsultarClientePorIDUseCase(repo)
	consultarPorDocumento := appcliente.NewConsultarClientePorDocumentoUseCase(repo)
	ativar := appcliente.NewAtivarClienteUseCase(repo)
	inativar := appcliente.NewInativarClienteUseCase(repo)
	alterarSenha := appcliente.NewAlterarSenhaClienteUseCase(repo)
	handler := NewHandler(criar, atualizar, consultarPorID, consultarPorDocumento, ativar, inativar, alterarSenha)
	require.Same(t, criar, handler.criar)
	require.Same(t, atualizar, handler.atualizar)
	require.Same(t, consultarPorID, handler.consultarPorID)
	require.Same(t, consultarPorDocumento, handler.consultarPorDocumento)
	require.Same(t, ativar, handler.ativar)
	require.Same(t, inativar, handler.inativar)
	require.Same(t, alterarSenha, handler.alterarSenha)
}

func TestHandlerCriarComSucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewClienteRepository(t)
	repo.EXPECT().BuscarPorDocumento(mock.Anything, "529.982.247-25").Return(nil, domaincliente.ErrClienteNaoEncontrado)
	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*cliente.Cliente")).Run(func(_ context.Context, cliente *domaincliente.Cliente) {
		require.Equal(t, uint64(7), cliente.CriadoPor())
		cliente.DefinirID(1)
	}).Return(nil)
	handler := NewHandler(appcliente.NewCriarClienteUseCase(repo), nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.POST("/v1/clientes", withSubject("7"), handler.Criar)
	req := requestJSON(t, http.MethodPost, "/v1/clientes", CriarClienteRequest{Documento: "529.982.247-25", TipoPessoa: "PF", Nome: "João da Silva", Email: "joao@email.com", Telefone: "(44) 99999-1234", Senha: "senha123"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusCreated, recorder.Code)
	var response ClienteResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, uint64(1), response.ID)
	require.Equal(t, uint64(7), response.CriadoPor)
}

func TestHandlerCriarComBodyInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.POST("/v1/clientes", withSubject("1"), handler.Criar)
	req := httptest.NewRequest(http.MethodPost, "/v1/clientes", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlerCriarClienteDuplicado(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewClienteRepository(t)
	repo.EXPECT().BuscarPorDocumento(mock.Anything, "529.982.247-25").Return(clienteValido(t, 1), nil)
	handler := NewHandler(appcliente.NewCriarClienteUseCase(repo), nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.POST("/v1/clientes", withSubject("1"), handler.Criar)
	req := requestJSON(t, http.MethodPost, "/v1/clientes", CriarClienteRequest{Documento: "529.982.247-25", TipoPessoa: "PF", Nome: "João da Silva", Email: "joao@email.com", Telefone: "(44) 99999-1234", Senha: "senha123"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusConflict, recorder.Code)
}

func TestHandlerAtualizarComSucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewClienteRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(clienteValido(t, 1), nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*cliente.Cliente")).Return(nil)
	handler := NewHandler(nil, appcliente.NewAtualizarClienteUseCase(repo), nil, nil, nil, nil, nil)
	router := gin.New()
	router.PUT("/v1/clientes/:id", handler.Atualizar)
	req := requestJSON(t, http.MethodPut, "/v1/clientes/1", AtualizarClienteRequest{Nome: "João Souza", Email: "joao.souza@email.com", Telefone: "11999998888"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestHandlerAtualizarComIDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.PUT("/v1/clientes/:id", handler.Atualizar)
	req := requestJSON(t, http.MethodPut, "/v1/clientes/invalido", AtualizarClienteRequest{Nome: "João Souza", Email: "joao.souza@email.com", Telefone: "11999998888"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlerBuscarPorID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewClienteRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(clienteValido(t, 1), nil)
	handler := NewHandler(nil, nil, appcliente.NewConsultarClientePorIDUseCase(repo), nil, nil, nil, nil)
	router := gin.New()
	router.GET("/v1/clientes/:id", handler.BuscarPorID)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/clientes/1", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestHandlerBuscarPorDocumento(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewClienteRepository(t)
	repo.EXPECT().BuscarPorDocumento(mock.Anything, "52998224725").Return(clienteValido(t, 1), nil)
	handler := NewHandler(nil, nil, nil, appcliente.NewConsultarClientePorDocumentoUseCase(repo), nil, nil, nil)
	router := gin.New()
	router.GET("/v1/clientes/documento/:documento", handler.BuscarPorDocumento)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/clientes/documento/52998224725", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestHandlerAtivar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cliente := clienteValido(t, 1)
	cliente.Inativar()
	repo := mocks.NewClienteRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(cliente, nil)
	repo.EXPECT().Atualizar(mock.Anything, cliente).Return(nil)
	handler := NewHandler(nil, nil, nil, nil, appcliente.NewAtivarClienteUseCase(repo), nil, nil)
	router := gin.New()
	router.PATCH("/v1/clientes/:id/ativar", handler.Ativar)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPatch, "/v1/clientes/1/ativar", nil))
	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestHandlerInativar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cliente := clienteValido(t, 1)
	repo := mocks.NewClienteRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(cliente, nil)
	repo.EXPECT().Atualizar(mock.Anything, cliente).Return(nil)
	handler := NewHandler(nil, nil, nil, nil, nil, appcliente.NewInativarClienteUseCase(repo), nil)
	router := gin.New()
	router.PATCH("/v1/clientes/:id/inativar", handler.Inativar)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPatch, "/v1/clientes/1/inativar", nil))
	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestHandlerAlterarSenhaDoClienteLogado(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cliente := clienteValido(t, 1)
	repo := mocks.NewClienteRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(cliente, nil)
	repo.EXPECT().Atualizar(mock.Anything, cliente).Return(nil)
	handler := NewHandler(nil, nil, nil, nil, nil, nil, appcliente.NewAlterarSenhaClienteUseCase(repo))
	router := gin.New()
	router.PUT("/v1/clientes/me/senha", withSubject("1"), handler.AlterarSenha)
	req := requestJSON(t, http.MethodPut, "/v1/clientes/me/senha", AlterarSenhaRequest{SenhaNova: "novaSenha123"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.False(t, cliente.RequerAlterarSenha())
}

func TestHandlerRetorna404QuandoClienteNaoExiste(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewClienteRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domaincliente.ErrClienteNaoEncontrado)
	handler := NewHandler(nil, nil, appcliente.NewConsultarClientePorIDUseCase(repo), nil, nil, nil, nil)
	router := gin.New()
	router.GET("/v1/clientes/:id", handler.BuscarPorID)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/clientes/99", nil))
	require.Equal(t, http.StatusNotFound, recorder.Code)
}
