package notificacao

import (
	"context"
	"fmt"
	"log"

	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// NotificadorStatus avisa o cliente por e-mail depois que a ordem de serviço
// já foi persistida. Falhas são registradas e não desfazem a transição.
// O envio é síncrono e usa o contexto da requisição; latência do provedor
// afeta a resposta e cancelamento pode interromper o envio. Uma fila/Outbox
// pode desacoplar esse efeito quando a garantia de entrega for necessária.
type NotificadorStatus struct {
	clientes domaincliente.ClienteRepository
	emails   shared.EmailSender
}

func NewStatus(clientes domaincliente.ClienteRepository, emails shared.EmailSender) *NotificadorStatus {
	return &NotificadorStatus{clientes: clientes, emails: emails}
}

func (n *NotificadorStatus) Notificar(
	ctx context.Context,
	os *domainordemservico.OrdemServico,
	statusAnterior domainordemservico.StatusOrdemServico,
) {
	if n == nil || n.emails == nil || os == nil {
		return
	}

	cliente, err := n.clientes.BuscarPorID(ctx, os.ClienteID())
	if err != nil {
		log.Printf("notificação de status da OS %s: falha ao buscar cliente: %v", os.Numero().String(), err)
		return
	}

	anterior := string(statusAnterior)
	if anterior == "" {
		anterior = "—"
	}

	assunto := fmt.Sprintf("Ordem de Serviço %s: %s", os.Numero().String(), os.Status())
	corpo := fmt.Sprintf(
		"Olá, %s!\n\nA ordem de serviço %s mudou de status.\nStatus anterior: %s\nStatus atual: %s\n",
		cliente.Nome(),
		os.Numero().String(),
		anterior,
		os.Status(),
	)

	if err := n.emails.Enviar(ctx, cliente.Email().String(), assunto, corpo); err != nil {
		log.Printf("notificação de status da OS %s: falha ao enviar e-mail: %v", os.Numero().String(), err)
	}
}
