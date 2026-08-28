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

func cadastrarVeiculoInputValido() CadastrarVeiculoInput {
	return CadastrarVeiculoInput{
		Placa:              "ABC1D23",
		Marca:              "Fiat",
		Modelo:             "Uno",
		QuilometragemAtual: 15000,
		Ano:                2020,
		Cor:                "Prata",
		CriadoPor:          1,
	}
}

func veiculoValidoMatcher() interface{} {
	return mock.MatchedBy(func(v *domain.Veiculo) bool {
		if v == nil {
			return false
		}

		return v.Placa().String() == "ABC1D23" &&
			v.Marca() == "Fiat" &&
			v.Modelo() == "Uno" &&
			v.QuilometragemAtual() == 15000 &&
			v.Ano() == 2020 &&
			v.Cor().String() == "Prata" &&
			v.CriadoPor() == 1 &&
			v.Ativo()
	})
}

func placaBuscadaMatcher(valor string) interface{} {
	return mock.MatchedBy(func(p domain.Placa) bool {
		return p.String() == valor
	})
}

func TestNewCadastrarVeiculoUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewCadastrarVeiculoUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestCadastrarVeiculoUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrarVeiculoUseCase(repository)

	input := cadastrarVeiculoInputValido()

	repository.
		EXPECT().
		BuscarPorPlaca(ctx, placaBuscadaMatcher("ABC1D23")).
		Return(nil, domain.ErrVeiculoNaoEncontrado).
		Once()

	repository.
		EXPECT().
		Salvar(ctx, veiculoValidoMatcher()).
		Run(func(ctx context.Context, v *domain.Veiculo) {
			v.AtribuirID(1)
		}).
		Return(nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.NoError(t, err)

	require.Equal(t, uint64(1), output.ID)
	require.Equal(t, "ABC1D23", output.Placa)
	require.Equal(t, "Fiat", output.Marca)
	require.Equal(t, "Uno", output.Modelo)
	require.Equal(t, uint32(15000), output.QuilometragemAtual)
	require.Equal(t, uint16(2020), output.Ano)
	require.Equal(t, "Prata", output.Cor)
	require.True(t, output.Ativo)
}

func TestCadastrarVeiculoUseCaseExecutarDeveRetornarErroDePlacaInvalida(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrarVeiculoUseCase(repository)

	input := cadastrarVeiculoInputValido()
	input.Placa = "INVALIDA"

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, domain.ErrPlacaInvalida)
}

func TestCadastrarVeiculoUseCaseExecutarDeveRetornarErroDePlacaJaCadastrada(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrarVeiculoUseCase(repository)

	input := cadastrarVeiculoInputValido()

	existente, err := domain.NewVeiculo("ABC1D23", "Outra", "Outro", 0, 2018, "Verde", 1)
	require.NoError(t, err)
	existente.AtribuirID(99)

	repository.
		EXPECT().
		BuscarPorPlaca(ctx, placaBuscadaMatcher("ABC1D23")).
		Return(existente, nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, domain.ErrPlacaJaCadastrada)
}

func TestCadastrarVeiculoUseCaseExecutarDeveRetornarErroAoBuscarPorPlaca(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrarVeiculoUseCase(repository)

	input := cadastrarVeiculoInputValido()

	erroRepository := errors.New("erro ao consultar placa")

	repository.
		EXPECT().
		BuscarPorPlaca(ctx, placaBuscadaMatcher("ABC1D23")).
		Return(nil, erroRepository).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
}

func TestCadastrarVeiculoUseCaseExecutarDeveRetornarErroDeValidacao(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrarVeiculoUseCase(repository)

	input := cadastrarVeiculoInputValido()
	input.Marca = ""

	repository.
		EXPECT().
		BuscarPorPlaca(ctx, placaBuscadaMatcher("ABC1D23")).
		Return(nil, domain.ErrVeiculoNaoEncontrado).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, domain.ErrMarcaObrigatoria)
}

func TestCadastrarVeiculoUseCaseExecutarDeveRetornarErroAoSalvar(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewCadastrarVeiculoUseCase(repository)

	input := cadastrarVeiculoInputValido()

	erroRepository := errors.New("erro ao salvar veiculo")

	repository.
		EXPECT().
		BuscarPorPlaca(ctx, placaBuscadaMatcher("ABC1D23")).
		Return(nil, domain.ErrVeiculoNaoEncontrado).
		Once()

	repository.
		EXPECT().
		Salvar(ctx, veiculoValidoMatcher()).
		Return(erroRepository).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, VeiculoOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
}
