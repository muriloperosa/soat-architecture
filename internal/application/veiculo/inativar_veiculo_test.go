package veiculo

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/domain/veiculo/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewInativarVeiculoUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewInativarVeiculoUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestInativarVeiculoUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewInativarVeiculoUseCase(repository)

	v := atualizarVeiculoValido(t)
	require.True(t, v.Ativo())

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
	require.False(t, output.Ativo)
	require.False(t, v.Ativo())
}

func TestInativarVeiculoUseCaseExecutarDeveRetornarErroAoBuscarVeiculo(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewInativarVeiculoUseCase(repository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrVeiculoNaoEncontrado).
		Once()

	output, err := useCase.Executar(ctx, 999)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, domain.ErrVeiculoNaoEncontrado)
}

func TestInativarVeiculoUseCaseExecutarDeveRetornarErroAoAtualizarRepository(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewInativarVeiculoUseCase(repository)

	v := atualizarVeiculoValido(t)

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
	require.False(t, v.Ativo())
}
