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

func TestConsultarOrdemServicoPorIDUseCase_InternoComSucesso(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	os := ordemServicoConsultaValida(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	out, err := app.NewConsultarOrdemServicoPorIDUseCase(repository).Executar(
		context.Background(),
		42,
		30,
		domainauth.TipoInterno,
	)

	require.NoError(t, err)
	require.Equal(t, uint64(42), out.ID)
	require.Equal(t, os.Numero().String(), out.Numero)
	require.Len(t, out.HistoricoStatus, 1)
	require.Equal(t, domain.StatusRecebida.String(), out.HistoricoStatus[0].Status)
}

func TestConsultarOrdemServicoPorIDUseCase_ClienteDaOSComSucesso(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	os := ordemServicoConsultaValida(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	out, err := app.NewConsultarOrdemServicoPorIDUseCase(repository).Executar(
		context.Background(),
		42,
		os.ClienteID(),
		domainauth.TipoCliente,
	)

	require.NoError(t, err)
	require.Equal(t, os.ID(), out.ID)
	require.Equal(t, os.ClienteID(), out.ClienteID)
}

func TestConsultarOrdemServicoPorIDUseCase_ClienteDeOutraOSRetornaForbidden(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	os := ordemServicoConsultaValida(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	_, err := app.NewConsultarOrdemServicoPorIDUseCase(repository).Executar(
		context.Background(),
		42,
		999,
		domainauth.TipoCliente,
	)

	require.Error(t, err)
	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindForbidden, appErr.Kind)
}

func TestConsultarOrdemServicoPorIDUseCase_NaoEncontrada(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domain.ErrOrdemServicoNaoEncontrada)

	_, err := app.NewConsultarOrdemServicoPorIDUseCase(repository).Executar(
		context.Background(),
		999,
		30,
		domainauth.TipoInterno,
	)

	require.ErrorIs(t, err, domain.ErrOrdemServicoNaoEncontrada)
}
