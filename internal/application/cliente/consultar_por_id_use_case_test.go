package cliente

import (
	"context"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	"github.com/stretchr/testify/require"
)

func TestConsultarPorIDExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsultarPorID(repository)

	cliente, err := domain.NewCliente(
		"529.982.247-25",
		domain.TipoPessoaFisica,
		"João da Silva",
		"joao@email.com",
		"(44) 99999-1234",
		"senha123",
	)
	require.NoError(t, err)

	cliente.DefinirID(1)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(&cliente, nil).
		Once()

	resultado, err := useCase.Executar(ctx, 1)

	require.NoError(t, err)
	require.Same(t, &cliente, resultado)
}
