package peca

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewReporEstoqueUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewReporEstoqueUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestReporEstoqueUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewReporEstoqueUseCase(repository)

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

	output, err := useCase.Executar(ctx, ReporEstoqueInput{PecaID: 1, Quantidade: 5})

	require.NoError(t, err)
	require.Equal(t, 15, output.QuantidadeEmEstoque)
	require.Equal(t, 15, p.QuantidadeEmEstoque())
}

func TestReporEstoqueUseCaseExecutarDeveRetornarErroAoBuscarPeca(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewReporEstoqueUseCase(repository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrPecaNaoEncontrada).
		Once()

	output, err := useCase.Executar(ctx, ReporEstoqueInput{PecaID: 999, Quantidade: 5})

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
}

func TestReporEstoqueUseCaseExecutarDeveRetornarErroDeQuantidadeInvalida(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewReporEstoqueUseCase(repository)

	p := novaPecaValida(t)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(p, nil).
		Once()

	output, err := useCase.Executar(ctx, ReporEstoqueInput{PecaID: 1, Quantidade: 0})

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrQuantidadeInvalida)
}

func TestReporEstoqueUseCaseExecutarDeveRetornarErroAoAtualizarRepository(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewReporEstoqueUseCase(repository)

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

	output, err := useCase.Executar(ctx, ReporEstoqueInput{PecaID: 1, Quantidade: 5})

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
}
