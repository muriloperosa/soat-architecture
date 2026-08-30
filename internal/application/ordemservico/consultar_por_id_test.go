package ordemservico_test

import (
	"context"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConsultarOrdemServicoPorIDUseCase_ComSucesso(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	os := ordemServicoConsultaValida(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	out, err := app.NewConsultarOrdemServicoPorIDUseCase(repository).Executar(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, uint64(42), out.ID)
	require.Equal(t, os.Numero().String(), out.Numero)
	require.Len(t, out.HistoricoStatus, 1)
	require.Equal(t, domain.StatusRecebida.String(), out.HistoricoStatus[0].Status)
}

func TestConsultarOrdemServicoPorIDUseCase_NaoEncontrada(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domain.ErrOrdemServicoNaoEncontrada)

	_, err := app.NewConsultarOrdemServicoPorIDUseCase(repository).Executar(context.Background(), 999)
	require.ErrorIs(t, err, domain.ErrOrdemServicoNaoEncontrada)
}
