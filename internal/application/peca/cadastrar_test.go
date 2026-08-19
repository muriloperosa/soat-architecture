package peca

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func cadastrarPecaInputValido() CadastrarPecaInput {
	return CadastrarPecaInput{
		Nome:                "Peca 1",
		Marca:               "Marca 1",
		Descricao:           "Descricao 1",
		Preco:               100.0,
		QuantidadeEmEstoque: 10,
		EstoqueMinimo:       5,
		CriadoPor:           1,
	}
}

func pecaValidaMatcher() interface{} {
	return mock.MatchedBy(func(p *domain.Peca) bool {
		if p == nil {
			return false
		}

		return p.Nome() == "Peca 1" &&
			p.Marca() == "Marca 1" &&
			p.Descricao() == "Descricao 1" &&
			p.Preco() == 100.0 &&
			p.QuantidadeEmEstoque() == 10 &&
			p.EstoqueMinimo() == 5 &&
			p.CriadoPor() == 1 &&
			p.Ativo()
	})
}

func TestNewCadastrarPecaUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewCadastrarPecaUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestCadastrarPecaUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrarPecaUseCase(repository)

	input := cadastrarPecaInputValido()

	repository.
		EXPECT().
		Salvar(ctx, pecaValidaMatcher()).
		Run(func(ctx context.Context, p *domain.Peca) {
			p.AtribuirID(1)
		}).
		Return(nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.NoError(t, err)

	require.Equal(t, uint64(1), output.ID)
	require.Equal(t, "Peca 1", output.Nome)
	require.Equal(t, "Marca 1", output.Marca)
	require.Equal(t, "Descricao 1", output.Descricao)
	require.Equal(t, 100.0, output.Preco)
	require.Equal(t, 10, output.QuantidadeEmEstoque)
	require.Equal(t, 5, output.EstoqueMinimo)
	require.True(t, output.Ativo)
}

func TestCadastrarPecaUseCaseExecutarDeveRetornarErroDeValidacao(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrarPecaUseCase(repository)

	input := cadastrarPecaInputValido()
	input.Nome = ""

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrNomeObrigatorio)
}

func TestCadastrarPecaUseCaseExecutarDeveRetornarErroAoSalvar(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrarPecaUseCase(repository)

	input := cadastrarPecaInputValido()

	erroRepository := errors.New("erro ao salvar peca")

	repository.
		EXPECT().
		Salvar(ctx, pecaValidaMatcher()).
		Return(erroRepository).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
}
