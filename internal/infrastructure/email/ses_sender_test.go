package email_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/email"
	"github.com/stretchr/testify/require"
)

type fakeSES struct {
	input *sesv2.SendEmailInput
	err   error
}

func (f *fakeSES) SendEmail(_ context.Context, params *sesv2.SendEmailInput, _ ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	f.input = params
	if f.err != nil {
		return nil, f.err
	}
	return &sesv2.SendEmailOutput{}, nil
}

func TestSESSender_EnviarMontaMensagemTexto(t *testing.T) {
	fake := &fakeSES{}
	sender := email.NewSESSenderWithClient(fake, "oficina@exemplo.com")

	err := sender.Enviar(context.Background(), "cliente@exemplo.com", "Ordem de Serviço OS-1: RECEBIDA", "Corpo")

	require.NoError(t, err)
	require.Equal(t, "oficina@exemplo.com", *fake.input.FromEmailAddress)
	require.Equal(t, []string{"cliente@exemplo.com"}, fake.input.Destination.ToAddresses)
	require.Equal(t, "Ordem de Serviço OS-1: RECEBIDA", *fake.input.Content.Simple.Subject.Data)
	require.Equal(t, "Corpo", *fake.input.Content.Simple.Body.Text.Data)
	require.Equal(t, "UTF-8", *fake.input.Content.Simple.Subject.Charset)
}

func TestSESSender_EnviarPropagaErroDoCliente(t *testing.T) {
	fake := &fakeSES{err: errors.New("throttling")}
	sender := email.NewSESSenderWithClient(fake, "oficina@exemplo.com")

	err := sender.Enviar(context.Background(), "cliente@exemplo.com", "assunto", "corpo")

	require.ErrorContains(t, err, "enviar e-mail via SES")
	require.ErrorContains(t, err, "throttling")
}

func TestNewSender_PadraoLogNaoExigeAWS(t *testing.T) {
	sender, err := email.NewSender(context.Background(), "", "", "")

	require.NoError(t, err)
	require.NoError(t, sender.Enviar(context.Background(), "a@b.com", "s", "c"))
}

func TestNewSESSender_RegiaoVaziaUsaPadrao(t *testing.T) {
	sender, err := email.NewSESSender(context.Background(), "  ", "oficina@exemplo.com")

	require.NoError(t, err)
	require.NotNil(t, sender)
}

func TestNewSender_SESSemRemetenteFalha(t *testing.T) {
	sender, err := email.NewSender(context.Background(), "ses", "us-east-1", "  ")

	require.Nil(t, sender)
	require.ErrorContains(t, err, "SES_FROM_ADDRESS")
}

func TestNewSender_ProvedorDesconhecido(t *testing.T) {
	sender, err := email.NewSender(context.Background(), "smtp", "us-east-1", "a@b.com")

	require.Nil(t, sender)
	require.ErrorContains(t, err, "EMAIL_PROVIDER inválido")
}
