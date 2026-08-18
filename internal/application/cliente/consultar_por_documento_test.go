package cliente

import (
	"context"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewConsultarPorDocumento(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewConsultarPorDocumento(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestConsultarPorDocumentoExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsultarPorDocumento(repository)

	cliente, err := domain.NewCliente(
		"529.982.247-25",
		domain.TipoPessoaFisica,
		"João da Silva",
		"joao@email.com",
		"(44) 99999-1234",
		"senha123",
	)
	require.NoError(t, err)

	documento := "52998224725"

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, documento).
		Return(&cliente, nil).
		Once()

	resultado, err := useCase.Executar(ctx, documento)

	require.NoError(t, err)
	require.Same(t, &cliente, resultado)
}

func TestConsultarPorDocumentoExecutarDeveRetornarErro(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewConsultarPorDocumento(repository)

	documento := "00000000000"

	repository.
		EXPECT().
		BuscarPorDocumento(ctx, documento).
		Return(nil, domain.ErrClienteNaoEncontrado).
		Once()

	resultado, err := useCase.Executar(ctx, documento)

	require.Nil(t, resultado)
	require.ErrorIs(t, err, domain.ErrClienteNaoEncontrado)
}
