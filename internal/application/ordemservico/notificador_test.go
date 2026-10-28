package ordemservico_test

import (
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/application/notificacao"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	clientemocks "github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	sharedmocks "github.com/muriloperosa/soat-architecture/internal/domain/shared/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func notificadorComEnvio(t *testing.T) *notificacao.NotificadorStatus {
	t.Helper()
	clientes := clientemocks.NewClienteRepository(t)
	emails := sharedmocks.NewEmailSender(t)
	cliente, err := domaincliente.NewCliente(
		"52998224725",
		domaincliente.TipoPessoaFisica,
		"Maria Silva",
		"maria@email.com",
		"11999998888",
		"senha123",
		1,
	)
	require.NoError(t, err)
	cliente.DefinirID(10)
	clientes.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(&cliente, nil)
	emails.EXPECT().Enviar(mock.Anything, "maria@email.com", mock.Anything, mock.Anything).Return(nil)
	return notificacao.NewStatus(clientes, emails)
}

func notificadorSemEnvio(t *testing.T) *notificacao.NotificadorStatus {
	t.Helper()
	return notificacao.NewStatus(clientemocks.NewClienteRepository(t), sharedmocks.NewEmailSender(t))
}
