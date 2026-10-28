package email

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// emailAPI isola o cliente do SES para o teste não depender da rede.
type emailAPI interface {
	SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

// SESSender envia e-mail de texto pelo Amazon SES (API v2).
// No laboratório o SES fica em sandbox: só entrega para identidades verificadas.
type SESSender struct {
	client emailAPI
	from   string
}

func NewSESSender(ctx context.Context, region, from string) (*SESSender, error) {
	if strings.TrimSpace(from) == "" {
		return nil, errors.New("SES_FROM_ADDRESS é obrigatório")
	}
	if strings.TrimSpace(region) == "" {
		region = "us-east-1"
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("carregar configuração AWS: %w", err)
	}

	return &SESSender{client: sesv2.NewFromConfig(cfg), from: from}, nil
}

func NewSESSenderWithClient(client emailAPI, from string) *SESSender {
	return &SESSender{client: client, from: from}
}

func (s *SESSender) Enviar(ctx context.Context, destinatario, assunto, corpo string) error {
	_, err := s.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.from),
		Destination: &types.Destination{
			ToAddresses: []string{destinatario},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String(assunto), Charset: aws.String("UTF-8")},
				Body: &types.Body{
					Text: &types.Content{Data: aws.String(corpo), Charset: aws.String("UTF-8")},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("enviar e-mail via SES: %w", err)
	}
	return nil
}

var _ shared.EmailSender = (*SESSender)(nil)
