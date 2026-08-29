package veiculo

import (
	"context"
	"errors"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/domain/veiculo/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func novoVeiculoParaListagem(t *testing.T) *domain.Veiculo {
	t.Helper()

	veiculo, err := domain.NewVeiculo("ABC1D23", "Fiat", "Uno", 15000, 2020, "Prata", 1)
	require.NoError(t, err)

	veiculo.AtribuirID(1)

	return veiculo
}

func TestListarVeiculosUseCase_Executar_RetornaPagina(t *testing.T) {
	repository := mocks.NewRepository(t)

	veiculo := novoVeiculoParaListagem(t)

	params := query.Params{
		Offset:    0,
		Limit:     20,
		Order:     "id",
		Direction: query.DirectionASC,
	}

	repository.
		EXPECT().
		Listar(mock.Anything, params).
		Return(query.Page[*domain.Veiculo]{
			Items: []*domain.Veiculo{
				veiculo,
			},
			Total:     1,
			Offset:    0,
			Limit:     20,
			Order:     "id",
			Direction: query.DirectionASC,
		}, nil)

	useCase := NewListarVeiculosUseCase(repository)

	page, err := useCase.Executar(context.Background(), params)

	require.NoError(t, err)

	require.Len(t, page.Items, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, 0, page.Offset)
	require.Equal(t, 20, page.Limit)
	require.Equal(t, "id", page.Order)
	require.Equal(t, query.DirectionASC, page.Direction)

	require.Equal(t, uint64(1), page.Items[0].ID)
	require.Equal(t, "ABC1D23", page.Items[0].Placa)
	require.Equal(t, "Fiat", page.Items[0].Marca)
	require.Equal(t, "Uno", page.Items[0].Modelo)
}

func TestListarVeiculosUseCase_Executar_ListaVaziaRetornaPaginaVazia(t *testing.T) {
	repository := mocks.NewRepository(t)

	params := query.Params{}

	repository.
		EXPECT().
		Listar(mock.Anything, params).
		Return(query.Page[*domain.Veiculo]{
			Items:     []*domain.Veiculo{},
			Total:     0,
			Offset:    0,
			Limit:     20,
			Order:     "id",
			Direction: query.DirectionASC,
		}, nil)

	useCase := NewListarVeiculosUseCase(repository)

	page, err := useCase.Executar(context.Background(), params)

	require.NoError(t, err)

	require.NotNil(t, page.Items)
	require.Empty(t, page.Items)

	require.Equal(t, int64(0), page.Total)
	require.Equal(t, 0, page.Offset)
	require.Equal(t, 20, page.Limit)
	require.Equal(t, "id", page.Order)
	require.Equal(t, query.DirectionASC, page.Direction)
}

func TestListarVeiculosUseCase_Executar_PropagaParametros(t *testing.T) {
	repository := mocks.NewRepository(t)

	params := query.Params{
		Offset:    10,
		Limit:     5,
		Order:     "ano",
		Direction: query.DirectionDESC,
		Filters: []query.Filter{
			{
				Field:    "marca",
				Operator: query.OperatorLike,
				Value:    "Fiat",
			},
		},
	}

	repository.
		EXPECT().
		Listar(mock.Anything, params).
		Return(query.Page[*domain.Veiculo]{
			Items:     []*domain.Veiculo{},
			Total:     0,
			Offset:    10,
			Limit:     5,
			Order:     "ano",
			Direction: query.DirectionDESC,
		}, nil)

	useCase := NewListarVeiculosUseCase(repository)

	page, err := useCase.Executar(context.Background(), params)

	require.NoError(t, err)

	require.Equal(t, 10, page.Offset)
	require.Equal(t, 5, page.Limit)
	require.Equal(t, "ano", page.Order)
	require.Equal(t, query.DirectionDESC, page.Direction)
}

func TestListarVeiculosUseCase_Executar_ErroDoBancoRetornaInternalError(t *testing.T) {
	repository := mocks.NewRepository(t)

	erroBanco := errors.New("erro ao consultar banco")

	params := query.Params{}

	repository.
		EXPECT().
		Listar(mock.Anything, params).
		Return(query.Page[*domain.Veiculo]{}, erroBanco)

	useCase := NewListarVeiculosUseCase(repository)

	page, err := useCase.Executar(context.Background(), params)

	require.Error(t, err)
	require.Empty(t, page.Items)

	var appErr *shared.AppError

	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "erro ao listar veículos", appErr.Message)
}

func TestListarVeiculosUseCase_Executar_AppErrorDoRepositoryPropagaErro(t *testing.T) {
	repository := mocks.NewRepository(t)

	params := query.Params{Limit: 101}

	erroValidacao := shared.NewValidationError("limite máximo permitido é 100")

	repository.
		EXPECT().
		Listar(mock.Anything, params).
		Return(query.Page[*domain.Veiculo]{}, erroValidacao)

	useCase := NewListarVeiculosUseCase(repository)

	page, err := useCase.Executar(context.Background(), params)

	require.Error(t, err)
	require.Empty(t, page.Items)

	require.ErrorIs(t, err, erroValidacao)
}
