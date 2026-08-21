package peca

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	"github.com/stretchr/testify/require"
)

func novaPecaValida(t *testing.T) *domain.Peca {
	t.Helper()

	p, err := domain.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	p.AtribuirID(1)

	return p
}

func TestNewInativarPecaUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewInativarPecaUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestInativarPecaUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewInativarPecaUseCase(repository)

	p := novaPecaValida(t)
	require.True(t, p.Ativo())

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

	output, err := useCase.Executar(ctx, 1)

	require.NoError(t, err)
	require.False(t, output.Ativo)
	require.False(t, p.Ativo())
}

func TestInativarPecaUseCaseExecutarDeveRetornarErroAoBuscarPeca(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewInativarPecaUseCase(repository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrPecaNaoEncontrada).
		Once()

	output, err := useCase.Executar(ctx, 999)

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
}

func TestInativarPecaUseCaseExecutarDeveRetornarErroAoAtualizarRepository(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewInativarPecaUseCase(repository)

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

	output, err := useCase.Executar(ctx, 1)

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
	require.False(t, p.Ativo())
}
