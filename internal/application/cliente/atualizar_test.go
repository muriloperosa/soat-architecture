package cliente

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func atualizarClienteValido(t *testing.T) *domain.Cliente {
	t.Helper()

	cliente, err := domain.NewCliente(
		"529.982.247-25",
		domain.TipoPessoaFisica,
		"João da Silva",
		"joao@email.com",
		"(44) 99999-1234",
		"senha123",
	)
	require.NoError(t, err)

	cliente.DefinirID(1)

	return &cliente
}

func TestNewAtualizarClienteUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewAtualizarClienteUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestAtualizarClienteUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarClienteUseCase(repository)

	cliente := atualizarClienteValido(t)

	input := AtualizarClienteInput{
		ID:       1,
		Nome:     "  mARIA   da SILVA ",
		Email:    "MARIA@EMAIL.COM",
		Telefone: "(44) 3031-1234",
	}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(cliente, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(
			ctx,
			mock.MatchedBy(func(c *domain.Cliente) bool {
				return c != nil &&
					c.ID() == 1 &&
					c.Nome() == "Maria Da Silva" &&
					c.Email().String() == "maria@email.com" &&
					c.Telefone().String() == "4430311234"
			}),
		).
		Return(nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.NoError(t, err)

	require.Equal(t, uint64(1), output.ID)
	require.Equal(t, "52998224725", output.Documento)
	require.Equal(t, domain.TipoPessoaFisica, output.Tipo)
	require.Equal(t, "Maria Da Silva", output.Nome)
	require.Equal(t, "maria@email.com", output.Email)
	require.Equal(t, "4430311234", output.Telefone)
	require.True(t, output.Ativo)
	require.True(t, output.RequerAlterarSenha)
}

func TestAtualizarClienteUseCaseExecutarDeveRetornarErroAoBuscarCliente(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarClienteUseCase(repository)

	input := AtualizarClienteInput{
		ID:       999,
		Nome:     "Maria da Silva",
		Email:    "maria@email.com",
		Telefone: "(44) 3031-1234",
	}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(nil, domain.ErrClienteNaoEncontrado).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, domain.ErrClienteNaoEncontrado)
}

func TestAtualizarClienteUseCaseExecutarDeveRetornarErroDeValidacao(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarClienteUseCase(repository)

	cliente := atualizarClienteValido(t)

	input := AtualizarClienteInput{
		ID:       1,
		Nome:     "",
		Email:    "maria@email.com",
		Telefone: "(44) 3031-1234",
	}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(cliente, nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, domain.ErrNomeObrigatorio)
}

func TestAtualizarClienteUseCaseExecutarDeveRetornarErroAoAtualizarRepository(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarClienteUseCase(repository)

	cliente := atualizarClienteValido(t)

	input := AtualizarClienteInput{
		ID:       1,
		Nome:     "Maria da Silva",
		Email:    "maria@email.com",
		Telefone: "(44) 3031-1234",
	}

	erroRepository := errors.New("erro ao atualizar cliente")

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(cliente, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(ctx, cliente).
		Return(erroRepository).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
}
