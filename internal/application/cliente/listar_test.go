package cliente

import (
	"context"
	"errors"
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListarClientesUseCase_Executar_ComSucesso(t *testing.T) {
	repository := mocks.NewClienteRepository(t)
	useCase := NewListarClientesUseCase(repository)
	ctx := context.Background()

	cliente, err := domain.NewCliente("52998224725", domain.TipoPessoaFisica, "Maria Silva", "maria@email.com", "11999998888", "senha123", 7)
	require.NoError(t, err)
	cliente.DefinirID(1)

	input := ListarClientesInput{
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
		Page: 2, Order: "nome", Direction: query.DirectionDESC,
		Filters: []query.Filter{{Field: "nome", Operator: query.OperatorAuto, Value: "Teste"}},
	}

	repository.EXPECT().Listar(ctx, expectedParams).Return(query.Page[*domain.Cliente]{
		Items: []*domain.Cliente{&cliente}, Total: 42, Page: 2, PageSize: 20, TotalPages: 3,
		Order: "nome", Direction: query.DirectionDESC,
	}, nil).Once()

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

func TestListarClientesUseCase_Executar_ListaVazia(t *testing.T) {
	repository := mocks.NewClienteRepository(t)
	useCase := NewListarClientesUseCase(repository)
	ctx := context.Background()
	input := ListarClientesInput{appquery.ParamsInput{
		Page: 1,
	}}

	repository.EXPECT().Listar(ctx, query.Params{Page: 1}).Return(query.Page[*domain.Cliente]{
		Items: []*domain.Cliente{}, Total: 0, Page: 1, PageSize: 20, TotalPages: 0,
		Order: "id", Direction: query.DirectionASC,
	}, nil).Once()

	out, err := useCase.Executar(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, out.Items)
	require.Empty(t, out.Items)
	require.Equal(t, 1, out.Page)
	require.Equal(t, 20, out.PageSize)
	require.Zero(t, out.TotalPages)
}

func TestListarClientesUseCase_Executar_PreservaAppError(t *testing.T) {
	repository := mocks.NewClienteRepository(t)
	appErr := shared.NewValidationError("filtro inválido")
	repository.EXPECT().Listar(mock.Anything, mock.Anything).Return(query.Page[*domain.Cliente]{}, appErr)

	_, err := NewListarClientesUseCase(repository).Executar(context.Background(), ListarClientesInput{})
	require.ErrorIs(t, err, appErr)
}

func TestListarClientesUseCase_Executar_ErroDeInfraestrutura(t *testing.T) {
	repository := mocks.NewClienteRepository(t)
	repository.EXPECT().Listar(mock.Anything, mock.Anything).Return(query.Page[*domain.Cliente]{}, errors.New("banco indisponível"))

	_, err := NewListarClientesUseCase(repository).Executar(context.Background(), ListarClientesInput{})
	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
