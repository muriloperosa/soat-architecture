package peca

import (
	"context"
	"errors"
	"testing"

	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListarPecasUseCase_Executar_ComSucesso(t *testing.T) {
	repository := mocks.NewRepository(t)
	useCase := NewListarPecasUseCase(repository)
	ctx := context.Background()

	entity, err := domain.NewPeca(
		"Pastilha de Freio",
		"Bosch",
		"Pastilha dianteira",
		149.90,
		25,
		5,
		1,
	)
	require.NoError(t, err)

	entity.AtribuirID(1)

	input := ListarPecasInput{
		ParamsInput: appquery.ParamsInput{
			Page:      2,
			Order:     "nome",
			Direction: "DESC",
			Filters: []appquery.FilterInput{
				{
					Field:    "nome",
					Operator: "auto",
					Value:    "Teste",
				},
			},
		},
	}

	expectedParams := query.Params{
		Page:      2,
		Order:     "nome",
		Direction: query.DirectionDESC,
		Filters: []query.Filter{
			{
				Field:    "nome",
				Operator: query.OperatorAuto,
				Value:    "Teste",
			},
		},
	}

	repository.
		EXPECT().
		Listar(ctx, expectedParams).
		Return(query.Page[*domain.Peca]{
			Items:      []*domain.Peca{entity},
			Total:      42,
			Page:       2,
			PageSize:   20,
			TotalPages: 3,
			Order:      "nome",
			Direction:  query.DirectionDESC,
		}, nil).
		Once()

	out, err := useCase.Executar(ctx, input)

	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	require.Equal(t, int64(42), out.Total)
	require.Equal(t, 2, out.Page)
	require.Equal(t, 20, out.PageSize)
	require.Equal(t, 3, out.TotalPages)
	require.Equal(t, "nome", out.Order)
	require.Equal(t, "DESC", out.Direction)
}

func TestListarPecasUseCase_Executar_ListaVazia(t *testing.T) {
	repository := mocks.NewRepository(t)
	useCase := NewListarPecasUseCase(repository)
	ctx := context.Background()

	input := ListarPecasInput{
		ParamsInput: appquery.ParamsInput{
			Page: 1,
		},
	}

	repository.
		EXPECT().
		Listar(ctx, query.Params{
			Page: 1,
		}).
		Return(query.Page[*domain.Peca]{
			Items:      []*domain.Peca{},
			Total:      0,
			Page:       1,
			PageSize:   20,
			TotalPages: 0,
			Order:      "id",
			Direction:  query.DirectionASC,
		}, nil).
		Once()

	out, err := useCase.Executar(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, out.Items)
	require.Empty(t, out.Items)
	require.Equal(t, 1, out.Page)
	require.Equal(t, 20, out.PageSize)
	require.Zero(t, out.TotalPages)
}

func TestListarPecasUseCase_Executar_PreservaAppError(t *testing.T) {
	repository := mocks.NewRepository(t)
	appErr := shared.NewValidationError("filtro inválido")

	repository.
		EXPECT().
		Listar(mock.Anything, mock.Anything).
		Return(query.Page[*domain.Peca]{}, appErr)

	_, err := NewListarPecasUseCase(repository).Executar(
		context.Background(),
		ListarPecasInput{},
	)

	require.ErrorIs(t, err, appErr)
}

func TestListarPecasUseCase_Executar_ErroDeInfraestrutura(t *testing.T) {
	repository := mocks.NewRepository(t)

	repository.
		EXPECT().
		Listar(mock.Anything, mock.Anything).
		Return(
			query.Page[*domain.Peca]{},
			errors.New("banco indisponível"),
		)

	_, err := NewListarPecasUseCase(repository).Executar(
		context.Background(),
		ListarPecasInput{},
	)

	var appErr *shared.AppError

	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
