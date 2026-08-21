package peca

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewConsumirEstoqueUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewConsumirEstoqueUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestConsumirEstoqueUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsumirEstoqueUseCase(repository)

	p := novaPecaValida(t)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(p, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(ctx, p).
		Return(nil).
		Once()

	output, err := useCase.Executar(ctx, ConsumirEstoqueInput{PecaID: 1, Quantidade: 3})

	require.NoError(t, err)
	require.Equal(t, 7, output.QuantidadeEmEstoque)
	require.Equal(t, 7, p.QuantidadeEmEstoque())
}

func TestConsumirEstoqueUseCaseExecutarDeveRetornarErroAoBuscarPeca(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsumirEstoqueUseCase(repository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrPecaNaoEncontrada).
		Once()

	output, err := useCase.Executar(ctx, ConsumirEstoqueInput{PecaID: 999, Quantidade: 3})

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
}

func TestConsumirEstoqueUseCaseExecutarDeveRetornarErroDeEstoqueInsuficiente(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsumirEstoqueUseCase(repository)

	p := novaPecaValida(t)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(p, nil).
		Once()

	output, err := useCase.Executar(ctx, ConsumirEstoqueInput{PecaID: 1, Quantidade: 6})

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrEstoqueInsuficiente)
	require.Equal(t, 10, p.QuantidadeEmEstoque())
}

func TestConsumirEstoqueUseCaseExecutarDeveRetornarErroAoAtualizarRepository(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsumirEstoqueUseCase(repository)

	p := novaPecaValida(t)

	erroRepository := errors.New("erro ao atualizar peca")

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(p, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(ctx, p).
		Return(erroRepository).
		Once()

	output, err := useCase.Executar(ctx, ConsumirEstoqueInput{PecaID: 1, Quantidade: 3})

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
}
