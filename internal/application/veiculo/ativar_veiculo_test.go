package veiculo

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/domain/veiculo/mocks"
	"github.com/stretchr/testify/require"
)

func veiculoInativoValido(t *testing.T) *domain.Veiculo {
	t.Helper()

	v, err := domain.NewVeiculo("ABC1D23", "Fiat", "Uno", 15000, 2020, "Prata", 1)
	require.NoError(t, err)

	v.AtribuirID(1)
	v.Inativar()

	require.False(t, v.Ativo())

	return v
}

func TestNewAtivarVeiculoUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewAtivarVeiculoUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestAtivarVeiculoUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtivarVeiculoUseCase(repository)

	v := veiculoInativoValido(t)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(v, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(ctx, v).
		Return(nil).
		Once()

	output, err := useCase.Executar(ctx, 1)

	require.NoError(t, err)
	require.True(t, output.Ativo)
	require.True(t, v.Ativo())
}

func TestAtivarVeiculoUseCaseExecutarDeveRetornarErroAoBuscarVeiculo(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtivarVeiculoUseCase(repository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrVeiculoNaoEncontrado).
		Once()

	output, err := useCase.Executar(ctx, 999)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, domain.ErrVeiculoNaoEncontrado)
}

func TestAtivarVeiculoUseCaseExecutarDeveRetornarErroAoAtualizarRepository(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtivarVeiculoUseCase(repository)

	v := veiculoInativoValido(t)

	erroRepository := errors.New("erro ao atualizar veiculo")

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(v, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(ctx, v).
		Return(erroRepository).
		Once()

	output, err := useCase.Executar(ctx, 1)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
	require.True(t, v.Ativo())
}
