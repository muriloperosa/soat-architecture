package email_test

import (
	"context"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/email"
	"github.com/stretchr/testify/require"
)

func TestLogSender_Enviar_NaoRetornaErro(t *testing.T) {
	sender := email.NewLogSender()

	err := sender.Enviar(context.Background(), "cliente@exemplo.com", "Orçamento aprovado", "Corpo do e-mail")

	require.NoError(t, err)
}
