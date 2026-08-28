package veiculo

import (
	"context"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/domain/veiculo/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewConsultarVeiculoPorIDUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewConsultarVeiculoPorIDUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestConsultarVeiculoPorIDUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsultarVeiculoPorIDUseCase(repository)

	v, err := domain.NewVeiculo("ABC1D23", "Fiat", "Uno", 15000, 2020, "Prata", 1)
	require.NoError(t, err)
	v.AtribuirID(1)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(v, nil).
		Once()

	resultado, err := useCase.Executar(ctx, 1)

	require.NoError(t, err)
	require.Equal(t, uint64(1), resultado.ID)
	require.Equal(t, "ABC1D23", resultado.Placa)
}

func TestConsultarVeiculoPorIDUseCaseExecutarDeveRetornarErro(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsultarVeiculoPorIDUseCase(repository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrVeiculoNaoEncontrado).
		Once()

	resultado, err := useCase.Executar(ctx, 999)

	require.Equal(t, VeiculoOutput{}, resultado)
	require.ErrorIs(t, err, domain.ErrVeiculoNaoEncontrado)
}
