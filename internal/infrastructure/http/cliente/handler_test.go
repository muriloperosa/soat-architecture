package cliente

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/stretchr/testify/require"
)

type criarClienteUseCaseMock struct {
	executar func(ctx context.Context, input app.CriarClienteInput) (app.ClienteOutput, error)
}

func (m *criarClienteUseCaseMock) Executar(
	ctx context.Context,
	input app.CriarClienteInput,
) (app.ClienteOutput, error) {
	return m.executar(ctx, input)
}

type atualizarClienteUseCaseMock struct {
	executar func(ctx context.Context, input app.AtualizarClienteInput) (app.ClienteOutput, error)
}

func (m *atualizarClienteUseCaseMock) Executar(
	ctx context.Context,
	input app.AtualizarClienteInput,
) (app.ClienteOutput, error) {
	return m.executar(ctx, input)
}

type consultarClientePorIDUseCaseMock struct {
	executar func(ctx context.Context, id uint64) (app.ClienteOutput, error)
}

func (m *consultarClientePorIDUseCaseMock) Executar(
	ctx context.Context,
	id uint64,
) (app.ClienteOutput, error) {
	return m.executar(ctx, id)
}

type consultarClientePorDocumentoUseCaseMock struct {
	executar func(ctx context.Context, documento string) (app.ClienteOutput, error)
}

func (m *consultarClientePorDocumentoUseCaseMock) Executar(
	ctx context.Context,
	documento string,
) (app.ClienteOutput, error) {
	return m.executar(ctx, documento)
}

type ativarClienteUseCaseMock struct {
	executar func(ctx context.Context, id uint64) (app.ClienteOutput, error)
}

func (m *ativarClienteUseCaseMock) Executar(
	ctx context.Context,
	id uint64,
) (app.ClienteOutput, error) {
	return m.executar(ctx, id)
}

type inativarClienteUseCaseMock struct {
	executar func(ctx context.Context, id uint64) (app.ClienteOutput, error)
}

func (m *inativarClienteUseCaseMock) Executar(
	ctx context.Context,
	id uint64,
) (app.ClienteOutput, error) {
	return m.executar(ctx, id)
}

type alterarSenhaUseCaseMock struct {
	executar func(ctx context.Context, input app.AlterarSenhaInput) (app.ClienteOutput, error)
}

func (m *alterarSenhaUseCaseMock) Executar(
	ctx context.Context,
	input app.AlterarSenhaInput,
) (app.ClienteOutput, error) {
	return m.executar(ctx, input)
}

func clienteOutputValido() app.ClienteOutput {
	return app.ClienteOutput{
		ID:                 1,
		Documento:          "52998224725",
		Tipo:               domain.TipoPessoaFisica,
		Nome:               "João Da Silva",
		Email:              "joao@email.com",
		Telefone:           "44999991234",
		Ativo:              true,
		RequerAlterarSenha: true,
	}
}

func requestJSON(t *testing.T, method string, path string, body any) *http.Request {
	t.Helper()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(method, path, bytes.NewReader(data))

	req.Header.Set("Content-Type", "application/json")

	return req
}

func TestNewHandler(t *testing.T) {
	criar := &criarClienteUseCaseMock{}
	atualizar := &atualizarClienteUseCaseMock{}
	consultarPorID := &consultarClientePorIDUseCaseMock{}
	consultarPorDocumento := &consultarClientePorDocumentoUseCaseMock{}
	ativar := &ativarClienteUseCaseMock{}
	inativar := &inativarClienteUseCaseMock{}
	alterarSenha := &alterarSenhaUseCaseMock{}

	handler := NewHandler(
		criar,
		atualizar,
		consultarPorID,
		consultarPorDocumento,
		ativar,
		inativar,
		alterarSenha,
	)

	require.NotNil(t, handler)
	require.Equal(t, criar, handler.criar)
	require.Equal(t, atualizar, handler.atualizar)
	require.Equal(t, consultarPorID, handler.consultarPorID)
	require.Equal(t, consultarPorDocumento, handler.consultarPorDocumento)
	require.Equal(t, ativar, handler.ativar)
	require.Equal(t, inativar, handler.inativar)
	require.Equal(t, alterarSenha, handler.alterarSenha)
}

func TestHandlerCriarComSucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &criarClienteUseCaseMock{
		executar: func(ctx context.Context, input app.CriarClienteInput) (app.ClienteOutput, error) {
			require.Equal(t, "529.982.247-25", input.Documento)
			require.Equal(t, domain.TipoPessoaFisica, input.Tipo)
			require.Equal(t, "João da Silva", input.Nome)
			require.Equal(t, "joao@email.com", input.Email)
			require.Equal(t, "(44) 99999-1234", input.Telefone)
			require.Equal(t, "senha123", input.Senha)

			return clienteOutputValido(), nil
		},
	}

	handler := &Handler{criar: useCase}

	router := gin.New()
	router.POST("/v1/clientes", handler.Criar)

	req := requestJSON(
		t,
		http.MethodPost,
		"/v1/clientes",
		CriarClienteRequest{
			Documento:  "529.982.247-25",
			TipoPessoa: "PF",
			Nome:       "João da Silva",
			Email:      "joao@email.com",
			Telefone:   "(44) 99999-1234",
			Senha:      "senha123",
		},
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusCreated, recorder.Code)
}

func TestHandlerCriarComBodyInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}

	router := gin.New()
	router.POST("/v1/clientes", handler.Criar)

	req := httptest.NewRequest(http.MethodPost, "/v1/clientes", bytes.NewBufferString("{"))

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlerCriarComErroDoUseCase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &criarClienteUseCaseMock{
		executar: func(ctx context.Context, input app.CriarClienteInput) (app.ClienteOutput, error) {
			return app.ClienteOutput{}, domain.ErrClienteJaCadastrado
		},
	}

	handler := &Handler{criar: useCase}

	router := gin.New()
	router.POST("/v1/clientes", handler.Criar)

	req := requestJSON(
		t,
		http.MethodPost,
		"/v1/clientes",
		CriarClienteRequest{
			Documento:  "529.982.247-25",
			TipoPessoa: "PF",
			Nome:       "João da Silva",
			Email:      "joao@email.com",
			Telefone:   "(44) 99999-1234",
			Senha:      "senha123",
		},
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusConflict, recorder.Code)
}

func TestHandlerAtualizarComSucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &atualizarClienteUseCaseMock{
		executar: func(
			ctx context.Context,
			input app.AtualizarClienteInput,
		) (app.ClienteOutput, error) {
			require.Equal(t, uint64(1), input.ID)
			require.Equal(t, "Maria da Silva", input.Nome)
			require.Equal(t, "maria@email.com", input.Email)
			require.Equal(t, "(44) 3031-1234", input.Telefone)

			output := clienteOutputValido()
			output.Nome = "Maria Da Silva"
			output.Email = "maria@email.com"
			output.Telefone = "4430311234"

			return output, nil
		},
	}

	handler := &Handler{atualizar: useCase}

	router := gin.New()
	router.PUT("/v1/clientes/:id", handler.Atualizar)

	req := requestJSON(
		t,
		http.MethodPut,
		"/v1/clientes/1",
		AtualizarClienteRequest{
			Nome:     "Maria da Silva",
			Email:    "maria@email.com",
			Telefone: "(44) 3031-1234",
		},
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestHandlerAtualizarComIDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}

	router := gin.New()
	router.PUT("/v1/clientes/:id", handler.Atualizar)

	req := requestJSON(t, http.MethodPut, "/v1/clientes/abc", AtualizarClienteRequest{})

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlerAtualizarComBodyInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}

	router := gin.New()
	router.PUT("/v1/clientes/:id", handler.Atualizar)

	req := httptest.NewRequest(http.MethodPut, "/v1/clientes/1", bytes.NewBufferString("{"))

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlerAtualizarComErroDoUseCase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &atualizarClienteUseCaseMock{
		executar: func(
			ctx context.Context,
			input app.AtualizarClienteInput,
		) (app.ClienteOutput, error) {
			return app.ClienteOutput{}, domain.ErrClienteNaoEncontrado
		},
	}

	handler := &Handler{atualizar: useCase}

	router := gin.New()
	router.PUT("/v1/clientes/:id", handler.Atualizar)

	req := requestJSON(
		t,
		http.MethodPut,
		"/v1/clientes/999",
		AtualizarClienteRequest{
			Nome:     "Maria da Silva",
			Email:    "maria@email.com",
			Telefone: "(44) 3031-1234",
		},
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestHandlerBuscarPorIDComSucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &consultarClientePorIDUseCaseMock{
		executar: func(ctx context.Context, id uint64) (app.ClienteOutput, error) {
			require.Equal(t, uint64(1), id)

			return clienteOutputValido(), nil
		},
	}

	handler := &Handler{consultarPorID: useCase}

	router := gin.New()
	router.GET("/v1/clientes/:id", handler.BuscarPorID)

	req := httptest.NewRequest(http.MethodGet, "/v1/clientes/1", nil)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestHandlerBuscarPorIDComIDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}

	router := gin.New()
	router.GET("/v1/clientes/:id", handler.BuscarPorID)

	req := httptest.NewRequest(http.MethodGet, "/v1/clientes/abc", nil)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlerBuscarPorIDComErroDoUseCase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &consultarClientePorIDUseCaseMock{
		executar: func(ctx context.Context, id uint64) (app.ClienteOutput, error) {
			return app.ClienteOutput{}, domain.ErrClienteNaoEncontrado
		},
	}

	handler := &Handler{consultarPorID: useCase}

	router := gin.New()
	router.GET("/v1/clientes/:id", handler.BuscarPorID)

	req := httptest.NewRequest(http.MethodGet, "/v1/clientes/999", nil)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestHandlerBuscarPorDocumentoComSucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &consultarClientePorDocumentoUseCaseMock{
		executar: func(ctx context.Context, documento string) (app.ClienteOutput, error) {
			require.Equal(t, "52998224725", documento)

			return clienteOutputValido(), nil
		},
	}

	handler := &Handler{consultarPorDocumento: useCase}

	router := gin.New()
	router.GET("/v1/clientes/documento/:documento", handler.BuscarPorDocumento)

	req := httptest.NewRequest(http.MethodGet, "/v1/clientes/documento/52998224725", nil)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestHandlerBuscarPorDocumentoComErroDoUseCase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &consultarClientePorDocumentoUseCaseMock{
		executar: func(ctx context.Context, documento string) (app.ClienteOutput, error) {
			return app.ClienteOutput{}, domain.ErrClienteNaoEncontrado
		},
	}

	handler := &Handler{consultarPorDocumento: useCase}

	router := gin.New()
	router.GET("/v1/clientes/documento/:documento", handler.BuscarPorDocumento)

	req := httptest.NewRequest(http.MethodGet, "/v1/clientes/documento/00000000000", nil)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestHandlerAtivarComSucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &ativarClienteUseCaseMock{
		executar: func(ctx context.Context, id uint64) (app.ClienteOutput, error) {
			require.Equal(t, uint64(1), id)

			return clienteOutputValido(), nil
		},
	}

	handler := &Handler{ativar: useCase}

	router := gin.New()
	router.PATCH("/v1/clientes/:id/ativar", handler.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/clientes/1/ativar", nil)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Empty(t, recorder.Body.String())
}

func TestHandlerAtivarComIDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}

	router := gin.New()
	router.PATCH("/v1/clientes/:id/ativar", handler.Ativar)

	req := httptest.NewRequest(http.MethodPatch, "/v1/clientes/abc/ativar", nil)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlerAtivarComErroDoUseCase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &ativarClienteUseCaseMock{
		executar: func(ctx context.Context, id uint64) (app.ClienteOutput, error) {
			return app.ClienteOutput{}, domain.ErrClienteNaoEncontrado
		},
	}

	handler := &Handler{
		ativar: useCase,
	}

	router := gin.New()
	router.PATCH("/v1/clientes/:id/ativar", handler.Ativar)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/v1/clientes/999/ativar",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestHandlerInativarComSucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &inativarClienteUseCaseMock{
		executar: func(
			ctx context.Context,
			id uint64,
		) (app.ClienteOutput, error) {
			require.Equal(t, uint64(1), id)

			return clienteOutputValido(), nil
		},
	}

	handler := &Handler{
		inativar: useCase,
	}

	router := gin.New()
	router.PATCH("/v1/clientes/:id/inativar", handler.Inativar)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/v1/clientes/1/inativar",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Empty(t, recorder.Body.String())
}

func TestHandlerInativarComIDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}

	router := gin.New()
	router.PATCH("/v1/clientes/:id/inativar", handler.Inativar)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/v1/clientes/abc/inativar",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlerInativarComErroDoUseCase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &inativarClienteUseCaseMock{
		executar: func(
			ctx context.Context,
			id uint64,
		) (app.ClienteOutput, error) {
			return app.ClienteOutput{}, domain.ErrClienteNaoEncontrado
		},
	}

	handler := &Handler{
		inativar: useCase,
	}

	router := gin.New()
	router.PATCH("/v1/clientes/:id/inativar", handler.Inativar)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/v1/clientes/999/inativar",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestHandlerAlterarSenhaComSucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &alterarSenhaUseCaseMock{
		executar: func(
			ctx context.Context,
			input app.AlterarSenhaInput,
		) (app.ClienteOutput, error) {
			require.Equal(t, uint64(1), input.ClienteID)
			require.Equal(t, "novaSenha123", input.SenhaNova)

			return clienteOutputValido(), nil
		},
	}

	handler := &Handler{
		alterarSenha: useCase,
	}

	router := gin.New()
	router.PATCH("/v1/clientes/:id/senha", handler.AlterarSenha)

	req := requestJSON(
		t,
		http.MethodPatch,
		"/v1/clientes/1/senha",
		AlterarSenhaRequest{
			SenhaNova: "novaSenha123",
		},
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Empty(t, recorder.Body.String())
}

func TestHandlerAlterarSenhaComIDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}

	router := gin.New()
	router.PATCH("/v1/clientes/:id/senha", handler.AlterarSenha)

	req := requestJSON(
		t,
		http.MethodPatch,
		"/v1/clientes/abc/senha",
		AlterarSenhaRequest{
			SenhaNova: "novaSenha123",
		},
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlerAlterarSenhaComBodyInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}

	router := gin.New()
	router.PATCH("/v1/clientes/:id/senha", handler.AlterarSenha)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/v1/clientes/1/senha",
		bytes.NewBufferString("{"),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlerAlterarSenhaComErroDoUseCase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	useCase := &alterarSenhaUseCaseMock{
		executar: func(
			ctx context.Context,
			input app.AlterarSenhaInput,
		) (app.ClienteOutput, error) {
			return app.ClienteOutput{}, domain.ErrClienteNaoEncontrado
		},
	}

	handler := &Handler{
		alterarSenha: useCase,
	}

	router := gin.New()
	router.PATCH("/v1/clientes/:id/senha", handler.AlterarSenha)

	req := requestJSON(
		t,
		http.MethodPatch,
		"/v1/clientes/999/senha",
		AlterarSenhaRequest{
			SenhaNova: "novaSenha123",
		},
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}
