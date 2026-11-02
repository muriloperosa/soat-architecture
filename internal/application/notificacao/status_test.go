package notificacao_test

import (
	"bytes"
	"context"
	"errors"
	"log"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/application/notificacao"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	clientemocks "github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	sharedmocks "github.com/muriloperosa/soat-architecture/internal/domain/shared/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNotificadorStatus_InformaStatusAnteriorEAtual(t *testing.T) {
	clientes := clientemocks.NewClienteRepository(t)
	emails := sharedmocks.NewEmailSender(t)
	os := ordemRecebida(t)

	cliente, err := domaincliente.NewCliente("52998224725", domaincliente.TipoPessoaFisica, "Maria Silva", "maria@email.com", "11999998888", "senha123", 1)
	require.NoError(t, err)
	cliente.DefinirID(10)
	clientes.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(&cliente, nil)

	var assunto, corpo string
	emails.EXPECT().Enviar(mock.Anything, "maria@email.com", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Run(func(_ context.Context, _, a, c string) {
			assunto = a
			corpo = c
		}).
		Return(nil)

	notificacao.NewStatus(clientes, emails).Notificar(context.Background(), os, "")

	require.Contains(t, assunto, "OS-20260827-a1b2c3d4e5f6")
	require.Contains(t, assunto, "RECEBIDA")
	require.Contains(t, corpo, "Maria Silva")
	require.Contains(t, corpo, "Status anterior: —")
	require.Contains(t, corpo, "Status atual: RECEBIDA")
}

func TestNotificadorStatus_FalhaDeEnvioNaoRetornaErro(t *testing.T) {
	clientes := clientemocks.NewClienteRepository(t)
	emails := sharedmocks.NewEmailSender(t)
	os := ordemRecebida(t)

	cliente, err := domaincliente.NewCliente("52998224725", domaincliente.TipoPessoaFisica, "Maria Silva", "maria@email.com", "11999998888", "senha123", 1)
	require.NoError(t, err)
	clientes.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(&cliente, nil)
	emails.EXPECT().Enviar(mock.Anything, "maria@email.com", mock.Anything, mock.Anything).Return(errors.New("ses indisponível"))

	var buf bytes.Buffer
	anterior := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(anterior) })

	notificacao.NewStatus(clientes, emails).Notificar(context.Background(), os, domainordemservico.StatusEmExecucao)

	require.Contains(t, buf.String(), "falha ao enviar e-mail")
}

func TestNotificadorStatus_ClienteInexistenteNaoRetornaErro(t *testing.T) {
	clientes := clientemocks.NewClienteRepository(t)
	emails := sharedmocks.NewEmailSender(t)
	clientes.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(nil, domaincliente.ErrClienteNaoEncontrado)

	var buf bytes.Buffer
	anterior := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(anterior) })

	notificacao.NewStatus(clientes, emails).Notificar(context.Background(), ordemRecebida(t), domainordemservico.StatusEmDiagnostico)

	require.Contains(t, buf.String(), "cliente não encontrado")
}

func TestNotificadorStatus_NilNaoEnvia(t *testing.T) {
	var n *notificacao.NotificadorStatus
	n.Notificar(context.Background(), ordemRecebida(t), "")
}

func ordemRecebida(t *testing.T) *domainordemservico.OrdemServico {
	t.Helper()
	os, err := domainordemservico.NewOrdemServico("OS-20260827-a1b2c3d4e5f6", 10, 20, 1000, "", "", 1)
	require.NoError(t, err)
	return os
}
