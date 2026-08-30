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

func TestConsultarOrdemServicoPorNumeroUseCase_ComSucesso(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	os := ordemServicoConsultaValida(t)
	repository.EXPECT().BuscarPorNumero(mock.Anything, os.Numero().String()).Return(os, nil)

	out, err := app.NewConsultarOrdemServicoPorNumeroUseCase(repository).Executar(context.Background(), os.Numero().String())
	require.NoError(t, err)
	require.Equal(t, os.ID(), out.ID)
	require.Equal(t, os.Numero().String(), out.Numero)
}

func TestConsultarOrdemServicoPorNumeroUseCase_NaoEncontrada(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	repository.EXPECT().BuscarPorNumero(mock.Anything, "OS-INEXISTENTE").Return(nil, domain.ErrOrdemServicoNaoEncontrada)

	_, err := app.NewConsultarOrdemServicoPorNumeroUseCase(repository).Executar(context.Background(), "OS-INEXISTENTE")
	require.ErrorIs(t, err, domain.ErrOrdemServicoNaoEncontrada)
}
