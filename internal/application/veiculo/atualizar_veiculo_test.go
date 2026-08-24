package veiculo

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/domain/veiculo/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func atualizarVeiculoValido(t *testing.T) *domain.Veiculo {
	t.Helper()

	v, err := domain.NewVeiculo("ABC1D23", "Fiat", "Uno", 15000, 2020, "Prata", 1)
	require.NoError(t, err)

	v.AtribuirID(1)

	return v
}

func TestNewAtualizarVeiculoUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewAtualizarVeiculoUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestAtualizarVeiculoUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarVeiculoUseCase(repository)

	v := atualizarVeiculoValido(t)

	input := AtualizarVeiculoInput{
		ID:                 1,
		Marca:              "Volkswagen",
		Modelo:             "Gol",
		Cor:                "Preto",
		QuilometragemAtual: 16000,
	}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(v, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(
			ctx,
			mock.MatchedBy(func(v *domain.Veiculo) bool {
				return v != nil &&
					v.ID() == 1 &&
					v.Marca() == "Volkswagen" &&
					v.Modelo() == "Gol" &&
					v.Cor().String() == "Preto" &&
					v.QuilometragemAtual() == 16000
			}),
		).
		Return(nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.NoError(t, err)

	require.Equal(t, uint64(1), output.ID)
	require.Equal(t, "Volkswagen", output.Marca)
	require.Equal(t, "Gol", output.Modelo)
	require.Equal(t, "Preto", output.Cor)
	require.Equal(t, "ABC1D23", output.Placa)
	require.Equal(t, uint32(16000), output.QuilometragemAtual)
}

func TestAtualizarVeiculoUseCaseExecutarDeveRetornarErroAoBuscarVeiculo(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarVeiculoUseCase(repository)

	input := AtualizarVeiculoInput{ID: 999, Marca: "Volkswagen", Modelo: "Gol", Cor: "Preto", QuilometragemAtual: 16000}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(nil, domain.ErrVeiculoNaoEncontrado).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, domain.ErrVeiculoNaoEncontrado)
}

func TestAtualizarVeiculoUseCaseExecutarDeveRetornarErroDeValidacao(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarVeiculoUseCase(repository)

	v := atualizarVeiculoValido(t)

	input := AtualizarVeiculoInput{ID: 1, Marca: "", Modelo: "Gol", Cor: "Preto", QuilometragemAtual: 16000}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(v, nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, domain.ErrMarcaObrigatoria)
}

func TestAtualizarVeiculoUseCaseExecutarDeveRetornarErroDeQuilometragemInvalida(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarVeiculoUseCase(repository)

	v := atualizarVeiculoValido(t)

	input := AtualizarVeiculoInput{ID: 1, Marca: "Volkswagen", Modelo: "Gol", Cor: "Preto", QuilometragemAtual: 14000}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(v, nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, domain.ErrQuilometragemInvalida)
	require.Equal(t, "Volkswagen", v.Marca())
	require.Equal(t, uint32(15000), v.QuilometragemAtual())
}

func TestAtualizarVeiculoUseCaseExecutarDeveRetornarErroAoAtualizarRepository(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarVeiculoUseCase(repository)

	v := atualizarVeiculoValido(t)

	input := AtualizarVeiculoInput{ID: 1, Marca: "Volkswagen", Modelo: "Gol", Cor: "Preto", QuilometragemAtual: 16000}

	erroRepository := errors.New("erro ao atualizar veiculo")

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(v, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(ctx, v).
		Return(erroRepository).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
}
