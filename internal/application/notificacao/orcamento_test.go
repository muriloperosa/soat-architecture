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
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	sharedmocks "github.com/muriloperosa/soat-architecture/internal/domain/shared/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNotificadorOrcamento_EnviaItensETotal(t *testing.T) {
	clientes := clientemocks.NewClienteRepository(t)
	emails := sharedmocks.NewEmailSender(t)
	os := ordemRecebida(t)
	orcamento := orcamentoComItens(t)

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

	notificacao.NewOrcamento(clientes, emails).Notificar(context.Background(), os, orcamento, false)

	require.Contains(t, assunto, "Orçamento da Ordem de Serviço OS-20260827-a1b2c3d4e5f6")
	require.NotContains(t, assunto, "atualizado")
	require.Contains(t, corpo, "Maria Silva")
	require.Contains(t, corpo, "Filtro de óleo | quantidade: 3 | valor unitário: R$ 50.00 | subtotal: R$ 150.00")
	require.Contains(t, corpo, "Valor total do orçamento: R$ 150.00")
}

func TestNotificadorOrcamento_ReenvioUsaAssuntoDeAtualizacao(t *testing.T) {
	clientes := clientemocks.NewClienteRepository(t)
	emails := sharedmocks.NewEmailSender(t)
	cliente, err := domaincliente.NewCliente("52998224725", domaincliente.TipoPessoaFisica, "Maria Silva", "maria@email.com", "11999998888", "senha123", 1)
	require.NoError(t, err)
	cliente.DefinirID(10)
	clientes.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(&cliente, nil)

	var assunto string
	emails.EXPECT().Enviar(mock.Anything, "maria@email.com", mock.AnythingOfType("string"), mock.Anything).
		Run(func(_ context.Context, _, a, _ string) { assunto = a }).
		Return(nil)

	notificacao.NewOrcamento(clientes, emails).Notificar(context.Background(), ordemRecebida(t), orcamentoComItens(t), true)

	require.Contains(t, assunto, "Orçamento atualizado da Ordem de Serviço")
}

func TestNotificadorOrcamento_ClienteInexistenteNaoEnvia(t *testing.T) {
	clientes := clientemocks.NewClienteRepository(t)
	emails := sharedmocks.NewEmailSender(t)
	clientes.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(nil, domaincliente.ErrClienteNaoEncontrado)

	var buf bytes.Buffer
	anterior := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(anterior) })

	notificacao.NewOrcamento(clientes, emails).Notificar(context.Background(), ordemRecebida(t), orcamentoComItens(t), false)

	require.Contains(t, buf.String(), "cliente não encontrado")
}

func TestNotificadorOrcamento_FalhaDeEnvioNaoInterrompeChamador(t *testing.T) {
	clientes := clientemocks.NewClienteRepository(t)
	emails := sharedmocks.NewEmailSender(t)
	cliente, err := domaincliente.NewCliente("52998224725", domaincliente.TipoPessoaFisica, "Maria Silva", "maria@email.com", "11999998888", "senha123", 1)
	require.NoError(t, err)
	cliente.DefinirID(10)
	clientes.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(&cliente, nil)
	emails.EXPECT().Enviar(mock.Anything, "maria@email.com", mock.Anything, mock.Anything).Return(errors.New("ses indisponível"))

	var buf bytes.Buffer
	anterior := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(anterior) })

	notificacao.NewOrcamento(clientes, emails).Notificar(context.Background(), ordemRecebida(t), orcamentoComItens(t), false)

	require.Contains(t, buf.String(), "falha ao enviar e-mail")
}

func orcamentoComItens(t *testing.T) *domainorcamento.Orcamento {
	t.Helper()
	o, err := domainorcamento.NewOrcamento(42, "", 1)
	require.NoError(t, err)
	require.NoError(t, o.AdicionarItemPeca(8, "Filtro de óleo", 3, 50))
	return o
}
