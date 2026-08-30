package servico

import (
	"context"
	"errors"
	"testing"

	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/servico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListarServicosUseCase_Executar_ComSucesso(t *testing.T) {
	repository := mocks.NewServicoRepository(t)
	useCase := NewListarServicosUseCase(repository)
	ctx := context.Background()

	entity, err := domain.NewServico(
		"Troca de óleo",
		"Troca completa",
		150,
		60,
		1,
	)
	require.NoError(t, err)

	entity.AtribuirID(1)

	input := ListarServicosInput{
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
		Return(
			query.Page[*domain.Servico]{
				Items:      []*domain.Servico{entity},
				Total:      42,
				Page:       2,
				PageSize:   20,
				TotalPages: 3,
				Order:      "nome",
				Direction:  query.DirectionDESC,
			},
			nil,
		).
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

func TestListarServicosUseCase_Executar_ListaVazia(t *testing.T) {
	repository := mocks.NewServicoRepository(t)
	useCase := NewListarServicosUseCase(repository)
	ctx := context.Background()

	input := ListarServicosInput{
		ParamsInput: appquery.ParamsInput{
			Page: 1,
		},
	}

	expectedParams := query.Params{
		Page: 1,
	}

	repository.
		EXPECT().
		Listar(ctx, expectedParams).
		Return(
			query.Page[*domain.Servico]{
				Items:      []*domain.Servico{},
				Total:      0,
				Page:       1,
				PageSize:   20,
				TotalPages: 0,
				Order:      "id",
				Direction:  query.DirectionASC,
			},
			nil,
		).
		Once()

	out, err := useCase.Executar(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, out.Items)
	require.Empty(t, out.Items)
	require.Equal(t, int64(0), out.Total)
	require.Equal(t, 1, out.Page)
	require.Equal(t, 20, out.PageSize)
	require.Zero(t, out.TotalPages)
	require.Equal(t, "id", out.Order)
	require.Equal(t, "ASC", out.Direction)
}

func TestListarServicosUseCase_Executar_PreservaAppError(t *testing.T) {
	repository := mocks.NewServicoRepository(t)
	useCase := NewListarServicosUseCase(repository)

	appErr := shared.NewValidationError("filtro inválido")

	repository.
		EXPECT().
		Listar(mock.Anything, mock.Anything).
		Return(
			query.Page[*domain.Servico]{},
			appErr,
		).
		Once()

	_, err := useCase.Executar(
		context.Background(),
		ListarServicosInput{},
	)

	require.ErrorIs(t, err, appErr)
}

func TestListarServicosUseCase_Executar_ErroDeInfraestrutura(t *testing.T) {
	repository := mocks.NewServicoRepository(t)
	useCase := NewListarServicosUseCase(repository)

	infraErr := errors.New("banco indisponível")

	repository.
		EXPECT().
		Listar(mock.Anything, mock.Anything).
		Return(
			query.Page[*domain.Servico]{},
			infraErr,
		).
		Once()

	_, err := useCase.Executar(
		context.Background(),
		ListarServicosInput{},
	)

	var appErr *shared.AppError

	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
