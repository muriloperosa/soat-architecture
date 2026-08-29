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

func criarClienteInputValido() CriarClienteInput {
	return CriarClienteInput{
		Documento: "529.982.247-25",
		Tipo:      string(domain.TipoPessoaFisica),
		Nome:      "João da Silva",
		Email:     "joao@email.com",
		Telefone:  "(44) 99999-1234",
		Senha:     "senha123",
		CriadoPor: 1,
	}
}

func clienteValidoMatcher() interface{} {
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
			cliente.RequerAlterarSenha() &&
			cliente.CriadoPor() == 1
	})
}

func TestNewCriarClienteUseCase(t *testing.T) {
	repository := mocks.NewClienteRepository(t)

	useCase := NewCriarClienteUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestCriarClienteUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewClienteRepository(t)
	useCase := NewCriarClienteUseCase(repository)

	input := criarClienteInputValido()

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, input.Documento).
		Return(nil, domain.ErrClienteNaoEncontrado).
		Once()

	repository.
		EXPECT().
		Salvar(ctx, clienteValidoMatcher()).
		Run(func(ctx context.Context, cliente *domain.Cliente) {
			cliente.DefinirID(1)
		}).
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
	require.True(t, output.RequerAlterarSenha)
	require.Equal(t, uint64(1), output.CriadoPor)
}

func TestCriarClienteUseCaseExecutarDeveRetornarErroAoBuscarCliente(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewClienteRepository(t)
	useCase := NewCriarClienteUseCase(repository)

	input := criarClienteInputValido()

	erroRepository := errors.New("erro ao consultar banco")

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, input.Documento).
		Return(nil, erroRepository).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
}

func TestCriarClienteUseCaseExecutarDeveRetornarErroQuandoClienteJaExistir(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewClienteRepository(t)
	useCase := NewCriarClienteUseCase(repository)

	input := criarClienteInputValido()

	existente, err := domain.NewCliente(
		input.Documento,
		domain.TipoPessoa(input.Tipo),
		input.Nome,
		input.Email,
		input.Telefone,
		input.Senha,
		1,
	)
	require.NoError(t, err)

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, input.Documento).
		Return(&existente, nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, domain.ErrClienteJaCadastrado)
}

func TestCriarClienteUseCaseExecutarDeveRetornarErroQuandoClienteForInvalido(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewClienteRepository(t)
	useCase := NewCriarClienteUseCase(repository)

	input := criarClienteInputValido()
	input.Nome = ""

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, input.Documento).
		Return(nil, domain.ErrClienteNaoEncontrado).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, domain.ErrNomeObrigatorio)
}

func TestCriarClienteUseCaseExecutarDeveRetornarErroAoSalvar(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewClienteRepository(t)
	useCase := NewCriarClienteUseCase(repository)

	input := criarClienteInputValido()

	erroRepository := errors.New("erro ao salvar cliente")

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, input.Documento).
		Return(nil, domain.ErrClienteNaoEncontrado).
		Once()

	repository.
		EXPECT().
		Salvar(ctx, clienteValidoMatcher()).
		Return(erroRepository).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, ClienteOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
}
