package peca

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	"github.com/stretchr/testify/require"
)

func pecaInativaValida(t *testing.T) *domain.Peca {
	t.Helper()

	p, err := domain.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	p.AtribuirID(1)
	p.Inativar()

	require.False(t, p.Ativo())

	return p
}

func TestNewAtivarPecaUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewAtivarPecaUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestAtivarPecaUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtivarPecaUseCase(repository)

	p := pecaInativaValida(t)

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
	require.True(t, output.Ativo)
	require.True(t, p.Ativo())
}

func TestAtivarPecaUseCaseExecutarDeveRetornarErroAoBuscarPeca(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtivarPecaUseCase(repository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrPecaNaoEncontrada).
		Once()

	output, err := useCase.Executar(ctx, 999)

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
}

func TestAtivarPecaUseCaseExecutarDeveRetornarErroAoAtualizarRepository(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtivarPecaUseCase(repository)

	p := pecaInativaValida(t)

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
	require.True(t, p.Ativo())
}
