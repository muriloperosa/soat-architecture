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

func cadastrarInputValido() CadastrarInput {
	return CadastrarInput{
		Documento: "529.982.247-25",
		Tipo:      domain.TipoPessoaFisica,
		Nome:      "João da Silva",
		Email:     "joao@email.com",
		Telefone:  "(44) 99999-1234",
		Senha:     "senha123",
	}
}

func mockCliente() interface{} {
	return mock.MatchedBy(func(cliente *domain.Cliente) bool {
		if cliente == nil {
			return false
		}

		return cliente.Documento().String() == "52998224725" &&
			cliente.Tipo() == domain.TipoPessoaFisica &&
			cliente.Nome() == "João Da Silva" &&
			cliente.Email().String() == "joao@email.com" &&
			cliente.Telefone().String() == "44999991234" &&
			cliente.Senha().Confere("senha123") &&
			cliente.Ativo() &&
			cliente.RequerAlterarSenha()
	})
}

func TestNewCadastrar(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewCadastrar(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestCadastrarExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrar(repository)

	input := cadastrarInputValido()

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, input.Documento).
		Return(nil, domain.ErrClienteNaoEncontrado).
		Once()

	repository.
		EXPECT().
		Salvar(ctx, mockCliente()).
		Return(nil).
		Once()

	cliente, err := useCase.Executar(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, cliente)

	require.Equal(t, "52998224725", cliente.Documento().String())
	require.Equal(t, domain.TipoPessoaFisica, cliente.Tipo())
	require.Equal(t, "João Da Silva", cliente.Nome())
	require.Equal(t, "joao@email.com", cliente.Email().String())
	require.Equal(t, "44999991234", cliente.Telefone().String())
	require.True(t, cliente.Senha().Confere("senha123"))
	require.True(t, cliente.Ativo())
	require.True(t, cliente.RequerAlterarSenha())
}

func TestCadastrarExecutarDeveRetornarErroAoBuscarCliente(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrar(repository)

	input := cadastrarInputValido()

	erroRepository := errors.New("erro ao consultar banco")

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, input.Documento).
		Return(nil, erroRepository).
		Once()

	cliente, err := useCase.Executar(ctx, input)

	require.Nil(t, cliente)
	require.ErrorIs(t, err, erroRepository)
}

func TestCadastrarExecutarDeveRetornarErroQuandoClienteJaExistir(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrar(repository)

	input := cadastrarInputValido()

	existente, err := domain.NewCliente(
		input.Documento,
		input.Tipo,
		input.Nome,
		input.Email,
		input.Telefone,
		input.Senha,
	)
	require.NoError(t, err)

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, input.Documento).
		Return(&existente, nil).
		Once()

	cliente, err := useCase.Executar(ctx, input)

	require.Nil(t, cliente)
	require.ErrorIs(t, err, domain.ErrClienteJaCadastrado)
}

func TestCadastrarExecutarDeveRetornarErroQuandoClienteForInvalido(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrar(repository)

	input := cadastrarInputValido()
	input.Nome = ""

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, input.Documento).
		Return(nil, domain.ErrClienteNaoEncontrado).
		Once()

	cliente, err := useCase.Executar(ctx, input)

	require.Nil(t, cliente)
	require.ErrorIs(t, err, domain.ErrNomeObrigatorio)
}

func TestCadastrarExecutarDeveRetornarErroAoSalvar(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrar(repository)

	input := cadastrarInputValido()

	erroRepository := errors.New("erro ao salvar cliente")

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, input.Documento).
		Return(nil, domain.ErrClienteNaoEncontrado).
		Once()

	repository.
		EXPECT().
		Salvar(ctx, mockCliente()).
		Return(erroRepository).
		Once()

	cliente, err := useCase.Executar(ctx, input)

	require.Nil(t, cliente)
	require.ErrorIs(t, err, erroRepository)
}
