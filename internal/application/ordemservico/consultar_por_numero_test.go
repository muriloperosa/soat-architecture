package ordemservico_test

import (
	"context"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConsultarOrdemServicoPorNumeroUseCase_InternoComSucesso(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	os := ordemServicoConsultaValida(t)
	repository.EXPECT().BuscarPorNumero(mock.Anything, os.Numero().String()).Return(os, nil)

	out, err := app.NewConsultarOrdemServicoPorNumeroUseCase(repository).Executar(
		context.Background(),
		os.Numero().String(),
		30,
		domainauth.TipoInterno,
	)

	require.NoError(t, err)
	require.Equal(t, os.ID(), out.ID)
	require.Equal(t, os.Numero().String(), out.Numero)
}

func TestConsultarOrdemServicoPorNumeroUseCase_ClienteDaOSComSucesso(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	os := ordemServicoConsultaValida(t)
	repository.EXPECT().BuscarPorNumero(mock.Anything, os.Numero().String()).Return(os, nil)

	out, err := app.NewConsultarOrdemServicoPorNumeroUseCase(repository).Executar(
		context.Background(),
		os.Numero().String(),
		os.ClienteID(),
		domainauth.TipoCliente,
	)

	require.NoError(t, err)
	require.Equal(t, os.ClienteID(), out.ClienteID)
}

func TestConsultarOrdemServicoPorNumeroUseCase_ClienteDeOutraOSRetornaForbidden(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	os := ordemServicoConsultaValida(t)
	repository.EXPECT().BuscarPorNumero(mock.Anything, os.Numero().String()).Return(os, nil)

	_, err := app.NewConsultarOrdemServicoPorNumeroUseCase(repository).Executar(
		context.Background(),
		os.Numero().String(),
		999,
		domainauth.TipoCliente,
	)

	require.Error(t, err)
	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindForbidden, appErr.Kind)
}

func TestConsultarOrdemServicoPorNumeroUseCase_NaoEncontrada(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	repository.EXPECT().BuscarPorNumero(mock.Anything, "OS-INEXISTENTE").Return(nil, domain.ErrOrdemServicoNaoEncontrada)

	_, err := app.NewConsultarOrdemServicoPorNumeroUseCase(repository).Executar(
		context.Background(),
		"OS-INEXISTENTE",
		30,
		domainauth.TipoInterno,
	)

	require.ErrorIs(t, err, domain.ErrOrdemServicoNaoEncontrada)
}
