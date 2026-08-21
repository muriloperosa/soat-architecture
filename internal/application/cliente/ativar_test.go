package cliente

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	"github.com/stretchr/testify/require"
)

func clienteInativoValido(t *testing.T) *domain.Cliente {
	t.Helper()

	cliente, err := domain.NewCliente(
		"529.982.247-25",
		domain.TipoPessoaFisica,
		"João da Silva",
		"joao@email.com",
		"(44) 99999-1234",
		"senha123",
		1,
	)
	require.NoError(t, err)

	cliente.DefinirID(1)
	cliente.Inativar()

	require.False(t, cliente.Ativo())

	return &cliente
}

func TestNewAtivarClienteUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewAtivarClienteUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestAtivarClienteUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtivarClienteUseCase(repository)

	cliente := clienteInativoValido(t)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(cliente, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(ctx, cliente).
		Return(nil).
		Once()

	output, err := useCase.Executar(ctx, 1)

	require.NoError(t, err)

	require.Equal(t, uint64(1), output.ID)
	require.Equal(t, "52998224725", output.Documento)
	require.Equal(t, domain.TipoPessoaFisica, output.Tipo)
	require.Equal(t, "João Da Silva", output.Nome)
	require.Equal(t, "joao@email.com", output.Email)
	require.Equal(t, "44999991234", output.Telefone)
	require.True(t, output.Ativo)
	require.True(t, output.RequerAlterarSenha)

	require.True(t, cliente.Ativo())
}

func TestAtivarClienteUseCaseExecutarDeveRetornarErroAoBuscarCliente(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtivarClienteUseCase(repository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrClienteNaoEncontrado).
		Once()

	output, err := useCase.Executar(ctx, 999)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, domain.ErrClienteNaoEncontrado)
}

func TestAtivarClienteUseCaseExecutarDeveRetornarErroAoAtualizarRepository(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtivarClienteUseCase(repository)

	cliente := clienteInativoValido(t)

	erroRepository := errors.New("erro ao atualizar cliente")

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(cliente, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(ctx, cliente).
		Return(erroRepository).
		Once()

	output, err := useCase.Executar(ctx, 1)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, erroRepository)

	// A entidade foi ativada em memória antes da falha de persistência.
	require.True(t, cliente.Ativo())
}
