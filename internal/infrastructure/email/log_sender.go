// Package email contém implementações da porta shared.EmailSender.
package email

import (
	"context"
	"log"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// LogSender simula o envio de e-mail registrando destinatário, assunto e
// corpo no log da aplicação, sem falar com nenhum provedor SMTP real. Serve
// pra exercitar o fluxo de envio de orçamento enquanto não há integração
// com um provedor de e-mail definido; trocar por uma implementação real
// (SMTP, SES, SendGrid etc.) não exige mudança no use case, só na injeção.
type LogSender struct{}

var _ shared.EmailSender = (*LogSender)(nil)

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (s *LogSender) Enviar(_ context.Context, destinatario, assunto, corpo string) error {
	log.Printf("[email simulado] para=%s assunto=%q\n%s", destinatario, assunto, corpo)
	return nil
}
