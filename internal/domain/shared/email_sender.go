package shared

import "context"

// EmailSender é a porta de saída para envio de e-mail. A implementação real
// (SMTP, provedor externo etc.) fica na camada de infraestrutura; use cases
// dependem só desta interface, nunca de detalhe de transporte.
type EmailSender interface {
	Enviar(ctx context.Context, destinatario, assunto, corpo string) error
}
