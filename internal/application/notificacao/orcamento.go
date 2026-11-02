package notificacao

import (
	"context"
	"fmt"
	"log"
	"strings"

	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// NotificadorOrcamento envia ao cliente os detalhes do orçamento para aprovação.
// É independente da notificação de status: o cliente precisa de serviços, peças,
// quantidades e valores para decidir. O envio também é síncrono, no contexto da
// requisição, e uma falha aqui não desfaz a transição já persistida.
type NotificadorOrcamento struct {
	clientes domaincliente.ClienteRepository
	emails   shared.EmailSender
}

// NewOrcamento cria um notificador para o envio detalhado de orçamentos.
func NewOrcamento(clientes domaincliente.ClienteRepository, emails shared.EmailSender) *NotificadorOrcamento {
	return &NotificadorOrcamento{clientes: clientes, emails: emails}
}

func (n *NotificadorOrcamento) Notificar(
	ctx context.Context,
	os *domainordemservico.OrdemServico,
	orcamento *domainorcamento.Orcamento,
	atualizado bool,
) {
	if n == nil || n.emails == nil || os == nil || orcamento == nil {
		return
	}

	cliente, err := n.clientes.BuscarPorID(ctx, os.ClienteID())
	if err != nil {
		log.Printf("notificação de orçamento da OS %s: falha ao buscar cliente: %v", os.Numero().String(), err)
		return
	}

	prefixo := "Orçamento da Ordem de Serviço"
	if atualizado {
		prefixo = "Orçamento atualizado da Ordem de Serviço"
	}
	assunto := fmt.Sprintf("%s %s", prefixo, os.Numero().String())
	corpo := montarCorpoEmail(os.Numero().String(), cliente.Nome(), orcamento)
	if err := n.emails.Enviar(ctx, cliente.Email().String(), assunto, corpo); err != nil {
		log.Printf("notificação de orçamento da OS %s: falha ao enviar e-mail: %v", os.Numero().String(), err)
	}
}

func montarCorpoEmail(numeroOS, nomeCliente string, o *domainorcamento.Orcamento) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Olá, %s!\n\n", nomeCliente)
	fmt.Fprintf(&b, "Segue o orçamento da Ordem de Serviço %s:\n\n", numeroOS)

	if len(o.ItensServico()) > 0 {
		b.WriteString("Serviços:\n")
		for _, item := range o.ItensServico() {
			fmt.Fprintf(&b, "- Serviço #%d | quantidade: %d | valor unitário: R$ %.2f | subtotal: R$ %.2f\n",
				item.ServicoID(), item.Quantidade(), item.Valor(), item.CalcularSubtotal())
		}
		b.WriteString("\n")
	}

	if len(o.ItensPeca()) > 0 {
		b.WriteString("Peças:\n")
		for _, item := range o.ItensPeca() {
			fmt.Fprintf(&b, "- %s | quantidade: %d | valor unitário: R$ %.2f | subtotal: R$ %.2f\n",
				item.Descricao(), item.Quantidade(), item.Valor(), item.CalcularSubtotal())
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "Valor total dos serviços: R$ %.2f\n", o.ValorItemServicos())
	fmt.Fprintf(&b, "Valor total das peças: R$ %.2f\n", o.ValorItemPecas())
	fmt.Fprintf(&b, "Valor total do orçamento: R$ %.2f\n", o.ValorTotal())

	if o.Observacoes() != "" {
		fmt.Fprintf(&b, "\nObservações: %s\n", o.Observacoes())
	}

	return b.String()
}
