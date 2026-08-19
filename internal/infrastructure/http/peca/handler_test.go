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
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
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

func TestHandler_Cadastrar_RequestValido_Retorna201(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*peca.Peca")).
		Run(func(ctx context.Context, p *domainpeca.Peca) { p.AtribuirID(1) }).
		Return(nil)

	h := httppeca.NewHandler(apppeca.NewCadastrarPecaUseCase(repo), nil, nil, nil, nil, nil)
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

	h := httppeca.NewHandler(apppeca.NewCadastrarPecaUseCase(repo), nil, nil, nil, nil, nil)
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

	h := httppeca.NewHandler(apppeca.NewCadastrarPecaUseCase(repo), nil, nil, nil, nil, nil)
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

	h := httppeca.NewHandler(nil, apppeca.NewAtualizarPecaUseCase(repo), nil, nil, nil, nil)
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

	h := httppeca.NewHandler(nil, apppeca.NewAtualizarPecaUseCase(repo), nil, nil, nil, nil)
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

	h := httppeca.NewHandler(nil, nil, apppeca.NewAtivarPecaUseCase(repo), nil, nil, nil)
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

	h := httppeca.NewHandler(nil, nil, apppeca.NewAtivarPecaUseCase(repo), nil, nil, nil)
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

	h := httppeca.NewHandler(nil, nil, nil, apppeca.NewInativarPecaUseCase(repo), nil, nil)
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
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(pecaExistente(t), nil)

	h := httppeca.NewHandler(nil, nil, nil, nil, apppeca.NewConsultarPecaPorIDUseCase(repo), nil)
	engine := gin.New()
	engine.GET("/v1/pecas/:id", h.ConsultarPorID)

	req := httptest.NewRequest(http.MethodGet, "/v1/pecas/1", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp httppeca.PecaResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, uint64(1), resp.ID)
}

func TestHandler_ConsultarPorID_PecaNaoEncontrada_Retorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainpeca.ErrPecaNaoEncontrada)

	h := httppeca.NewHandler(nil, nil, nil, nil, apppeca.NewConsultarPecaPorIDUseCase(repo), nil)
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

	h := httppeca.NewHandler(nil, nil, nil, nil, nil, apppeca.NewReporEstoqueUseCase(repo))
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

	h := httppeca.NewHandler(nil, nil, nil, nil, nil, apppeca.NewReporEstoqueUseCase(repo))
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

	h := httppeca.NewHandler(nil, nil, nil, nil, nil, apppeca.NewReporEstoqueUseCase(repo))
	engine := gin.New()
	engine.PATCH("/v1/pecas/:id/repor-estoque", h.ReporEstoque)

	body, _ := json.Marshal(httppeca.ReporEstoqueRequest{Quantidade: 10})
	req := httptest.NewRequest(http.MethodPatch, "/v1/pecas/1/repor-estoque", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
