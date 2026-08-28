package cliente

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func TestNewAlterarSenhaClienteUseCase(t *testing.T) {
	repository := mocks.NewClienteRepository(t)

	useCase := NewAlterarSenhaClienteUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestAlterarSenhaClienteUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewClienteRepository(t)
	useCase := NewAlterarSenhaClienteUseCase(repository)

	cliente := novoClienteValido(t)
	cliente.DefinirID(1)

	require.True(t, cliente.RequerAlterarSenha())
	require.True(t, cliente.Senha().Confere("senha123"))

	input := AlterarSenhaInput{
		ClienteID: 1,
		SenhaNova: "novaSenha123",
	}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ClienteID).
		Return(&cliente, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(ctx, &cliente).
		Return(nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.NoError(t, err)

	require.Equal(t, uint64(1), output.ID)
	require.Equal(t, "52998224725", output.Documento)
	require.Equal(t, domain.TipoPessoaFisica, output.Tipo)
	require.Equal(t, "João Da Silva", output.Nome)
	require.Equal(t, "joao@email.com", output.Email)
	require.Equal(t, "44999991234", output.Telefone)
	require.True(t, output.Ativo)

	// Depois da alteração, o cliente não precisa mais trocar a senha.
	require.False(t, output.RequerAlterarSenha)

	require.False(t, cliente.Senha().Confere("senha123"))
	require.True(t, cliente.Senha().Confere("novaSenha123"))
	require.False(t, cliente.RequerAlterarSenha())
}

func TestAlterarSenhaClienteUseCaseExecutarDeveRetornarErroAoBuscarCliente(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewClienteRepository(t)
	useCase := NewAlterarSenhaClienteUseCase(repository)

	input := AlterarSenhaInput{ClienteID: 999, SenhaNova: "novaSenha123"}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ClienteID).
		Return(nil, domain.ErrClienteNaoEncontrado).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, domain.ErrClienteNaoEncontrado)
}

func TestAlterarSenhaClienteUseCaseExecutarDeveRetornarErroParaSenhaFraca(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewClienteRepository(t)
	useCase := NewAlterarSenhaClienteUseCase(repository)

	cliente := novoClienteValido(t)
	cliente.DefinirID(1)

	input := AlterarSenhaInput{ClienteID: 1, SenhaNova: "123"}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ClienteID).
		Return(&cliente, nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, shared.ErrSenhaFraca)

	// A senha anterior deve permanecer válida.
	require.True(t, cliente.Senha().Confere("senha123"))

	// Como a alteração falhou, continua sendo necessário alterar a senha.
	require.True(t, cliente.RequerAlterarSenha())
}

func TestAlterarSenhaClienteUseCaseExecutarDeveRetornarErroAoAtualizarRepository(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewClienteRepository(t)
	useCase := NewAlterarSenhaClienteUseCase(repository)

	cliente := novoClienteValido(t)
	cliente.DefinirID(1)

	input := AlterarSenhaInput{ClienteID: 1, SenhaNova: "novaSenha123"}

	erroRepository := errors.New("erro ao atualizar cliente")

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ClienteID).
		Return(&cliente, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(ctx, &cliente).
		Return(erroRepository).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, erroRepository)

	// A entidade já foi alterada em memória antes da falha do repository.
	require.True(t, cliente.Senha().Confere("novaSenha123"))
	require.False(t, cliente.RequerAlterarSenha())
}
