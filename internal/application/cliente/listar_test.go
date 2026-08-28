package cliente

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListarClientesUseCase(t *testing.T) {
	repository := mocks.NewClienteRepository(t)
	cliente, err := domain.NewCliente(
		"52998224725", domain.TipoPessoaFisica, "Maria Silva",
		"maria@email.com", "11999998888", "senha123", 7,
	)
	require.NoError(t, err)
	cliente.DefinirID(1)

	params := query.Params{Offset: 0, Limit: 10, Order: "nome", Direction: query.DirectionASC}
	repository.EXPECT().Listar(mock.Anything, params).Return(query.Page[*domain.Cliente]{
		Items: []*domain.Cliente{&cliente}, Total: 1, Offset: 0, Limit: 10,
		Order: "nome", Direction: query.DirectionASC,
	}, nil)

	output, err := NewListarClientesUseCase(repository).Executar(context.Background(), params)

	require.NoError(t, err)
	require.Len(t, output.Items, 1)
	require.Equal(t, uint64(1), output.Items[0].ID)
	require.Equal(t, int64(1), output.Total)
	require.Equal(t, "nome", output.Order)
}

func TestListarClientesUseCasePreservaErroDeValidacao(t *testing.T) {
	repository := mocks.NewClienteRepository(t)
	validationErr := shared.NewValidationError("filtro inválido")
	repository.EXPECT().Listar(mock.Anything, mock.Anything).Return(
		query.Page[*domain.Cliente]{}, validationErr,
	)

	_, err := NewListarClientesUseCase(repository).Executar(context.Background(), query.Params{})

	require.ErrorIs(t, err, validationErr)
}

func TestListarClientesUseCaseConverteErroDeInfraestrutura(t *testing.T) {
	repository := mocks.NewClienteRepository(t)
	repository.EXPECT().Listar(mock.Anything, mock.Anything).Return(
		query.Page[*domain.Cliente]{}, errors.New("banco indisponível"),
	)

	_, err := NewListarClientesUseCase(repository).Executar(context.Background(), query.Params{})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
