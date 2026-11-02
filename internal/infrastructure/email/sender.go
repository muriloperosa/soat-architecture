package email

import (
	"context"
	"fmt"
	"strings"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// SenderConfig reúne as configurações dos adapters de e-mail.
type SenderConfig struct {
	Provider       string
	AWSRegion      string
	SESFromAddress string
	SMTP           SMTPConfig
}

// NewSender escolhe o adapter de e-mail. O padrão é log, para que o processo
// local e os testes não dependam de um serviço externo.
func NewSender(ctx context.Context, config SenderConfig) (shared.EmailSender, error) {
	switch strings.ToLower(strings.TrimSpace(config.Provider)) {
	case "", "log":
		return NewLogSender(), nil
	case "ses":
		return NewSESSender(ctx, config.AWSRegion, config.SESFromAddress)
	case "smtp":
		return NewSMTPSender(config.SMTP)
	default:
		return nil, fmt.Errorf("EMAIL_PROVIDER inválido: %s", config.Provider)
	}
}
