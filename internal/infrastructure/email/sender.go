package email

import (
	"context"
	"fmt"
	"strings"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// NewSender escolhe o provedor de e-mail. O padrão é log, para que testes e
// o ambiente local não precisem de credencial AWS. ses só entra quando
// EMAIL_PROVIDER=ses e SES_FROM_ADDRESS está definido.
func NewSender(ctx context.Context, provider, region, from string) (shared.EmailSender, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "log":
		return NewLogSender(), nil
	case "ses":
		return NewSESSender(ctx, region, from)
	default:
		return nil, fmt.Errorf("EMAIL_PROVIDER inválido: %s", provider)
	}
}
