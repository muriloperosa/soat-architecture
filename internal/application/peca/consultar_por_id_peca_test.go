package peca

import (
	"context"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewConsultarPecaPorIDUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewConsultarPecaPorIDUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestConsultarPecaPorIDUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsultarPecaPorIDUseCase(repository)

	p, err := domain.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)
	p.AtribuirID(1)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(p, nil).
		Once()

	resultado, err := useCase.Executar(ctx, 1)

	require.NoError(t, err)
	require.Equal(t, uint64(1), resultado.ID)
	require.Equal(t, "Peca 1", resultado.Nome)
}

func TestConsultarPecaPorIDUseCaseExecutarDeveRetornarErro(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsultarPecaPorIDUseCase(repository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrPecaNaoEncontrada).
		Once()

	resultado, err := useCase.Executar(ctx, 999)

	require.Equal(t, PecaOutput{}, resultado)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
}
