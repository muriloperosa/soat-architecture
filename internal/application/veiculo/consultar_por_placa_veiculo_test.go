package veiculo

import (
	"context"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/domain/veiculo/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewConsultarVeiculoPorPlacaUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewConsultarVeiculoPorPlacaUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestConsultarVeiculoPorPlacaUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsultarVeiculoPorPlacaUseCase(repository)

	v, err := domain.NewVeiculo("ABC1D23", "Fiat", "Uno", 15000, 2020, "Prata", 1)
	require.NoError(t, err)
	v.AtribuirID(1)

	repository.
		EXPECT().
		BuscarPorPlaca(ctx, placaBuscadaMatcher("ABC1D23")).
		Return(v, nil).
		Once()

	resultado, err := useCase.Executar(ctx, "abc-1d23")

	require.NoError(t, err)
	require.Equal(t, uint64(1), resultado.ID)
	require.Equal(t, "ABC1D23", resultado.Placa)
}

func TestConsultarVeiculoPorPlacaUseCaseExecutarDeveRetornarErroDePlacaInvalida(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsultarVeiculoPorPlacaUseCase(repository)

	resultado, err := useCase.Executar(ctx, "INVALIDA")

	require.Equal(t, VeiculoOutput{}, resultado)
	require.ErrorIs(t, err, domain.ErrPlacaInvalida)
}

func TestConsultarVeiculoPorPlacaUseCaseExecutarDeveRetornarErroNaoEncontrado(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsultarVeiculoPorPlacaUseCase(repository)

	repository.
		EXPECT().
		BuscarPorPlaca(ctx, placaBuscadaMatcher("ZZZ9Z99")).
		Return(nil, domain.ErrVeiculoNaoEncontrado).
		Once()

	resultado, err := useCase.Executar(ctx, "ZZZ9Z99")

	require.Equal(t, VeiculoOutput{}, resultado)
	require.ErrorIs(t, err, domain.ErrVeiculoNaoEncontrado)
}
