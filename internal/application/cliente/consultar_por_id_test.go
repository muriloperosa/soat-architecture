package cliente

import (
	"context"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewConsultarClientePorIDUseCase(t *testing.T) {
	repository := mocks.NewClienteRepository(t)

	useCase := NewConsultarClientePorIDUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestConsultarPorIDUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewClienteRepository(t)
	useCase := NewConsultarClientePorIDUseCase(repository)

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

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(&cliente, nil).
		Once()

	resultado, err := useCase.Executar(ctx, 1)

	require.NoError(t, err)

	require.Equal(t, uint64(1), resultado.ID)
	require.Equal(t, "52998224725", resultado.Documento)
	require.Equal(t, domain.TipoPessoaFisica, resultado.Tipo)
	require.Equal(t, "João Da Silva", resultado.Nome)
	require.Equal(t, "joao@email.com", resultado.Email)
	require.Equal(t, "44999991234", resultado.Telefone)
	require.True(t, resultado.Ativo)
	require.True(t, resultado.RequerAlterarSenha)
}

func TestConsultarPorIDUseCaseExecutarDeveRetornarErro(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewClienteRepository(t)
	useCase := NewConsultarClientePorIDUseCase(repository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrClienteNaoEncontrado).
		Once()

	resultado, err := useCase.Executar(ctx, 999)

	require.Equal(t, ClienteOutput{}, resultado)
	require.ErrorIs(t, err, domain.ErrClienteNaoEncontrado)
}
