package orcamento

import (
	"context"
	"errors"
	"testing"

	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestValidarOrcamentoEditavel_RecusaStatusImutavel(t *testing.T) {
	os, err := domainordemservico.NewOrdemServico("OS-20260830-a1b2c3d4e5f6", 10, 20, 100, "", "", 1)
	require.NoError(t, err)

	repo := ordemservicomocks.NewOrdemServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	err = validarOrcamentoEditavel(context.Background(), repo, 42)

	require.ErrorIs(t, err, domainorcamento.ErrOrcamentoImutavel)
}

func TestValidarOrcamentoEditavel_PropagaErrosDoRepositorio(t *testing.T) {
	t.Run("ordem inexistente", func(t *testing.T) {
		repo := ordemservicomocks.NewOrdemServicoRepository(t)
		repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(nil, domainordemservico.ErrOrdemServicoNaoEncontrada)
		err := validarOrcamentoEditavel(context.Background(), repo, 42)
		require.ErrorIs(t, err, domainordemservico.ErrOrdemServicoNaoEncontrada)
	})

	t.Run("erro inesperado", func(t *testing.T) {
		cause := errors.New("falha no banco")
		repo := ordemservicomocks.NewOrdemServicoRepository(t)
		repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(nil, cause)
		err := validarOrcamentoEditavel(context.Background(), repo, 42)
		require.ErrorIs(t, err, cause)
	})
}
