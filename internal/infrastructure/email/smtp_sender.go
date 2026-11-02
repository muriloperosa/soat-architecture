package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

const smtpTimeout = 30 * time.Second

// SMTPConfig configura uma conexão SMTP autenticada com STARTTLS.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// SMTPSender envia e-mails de texto usando SMTP e exige STARTTLS antes da
// autenticação para não transmitir credenciais em texto aberto.
type SMTPSender struct {
	host     string
	port     int
	username string
	password string
	from     *mail.Address
}

func NewSMTPSender(config SMTPConfig) (*SMTPSender, error) {
	host := strings.TrimSpace(config.Host)
	if host == "" || strings.ContainsAny(host, "/:@ \t\r\n") {
		return nil, errors.New("SMTP_HOST deve ser um hostname válido")
	}
	if config.Port < 1 || config.Port > 65535 {
		return nil, errors.New("SMTP_PORT deve estar entre 1 e 65535")
	}
	username := strings.TrimSpace(config.Username)
	if username == "" {
		return nil, errors.New("SMTP_USERNAME é obrigatório")
	}
	if strings.TrimSpace(config.Password) == "" {
		return nil, errors.New("SMTP_PASSWORD é obrigatório")
	}
	from, err := mail.ParseAddress(strings.TrimSpace(config.From))
	if err != nil {
		return nil, fmt.Errorf("SMTP_FROM_ADDRESS inválido: %w", err)
	}

	return &SMTPSender{
		host:     host,
		port:     config.Port,
		username: username,
		password: config.Password,
		from:     from,
	}, nil
}

func (s *SMTPSender) Enviar(ctx context.Context, destinatario, assunto, corpo string) error {
	to, err := mail.ParseAddress(strings.TrimSpace(destinatario))
	if err != nil {
		return fmt.Errorf("destinatário SMTP inválido: %w", err)
	}

	message, err := smtpMessage(s.from, to, assunto, corpo)
	if err != nil {
		return err
	}

	address := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	dialer := net.Dialer{Timeout: smtpTimeout}
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("conectar ao servidor SMTP: %w", err)
	}
	defer connection.Close()

	deadline := time.Now().Add(smtpTimeout)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}
	if err := connection.SetDeadline(deadline); err != nil {
		return fmt.Errorf("configurar timeout SMTP: %w", err)
	}
	stopCancellation := make(chan struct{})
	defer close(stopCancellation)
	go func() {
		select {
		case <-ctx.Done():
			_ = connection.Close()
		case <-stopCancellation:
		}
	}()

	client, err := smtp.NewClient(connection, s.host)
	if err != nil {
		return fmt.Errorf("iniciar sessão SMTP: %w", err)
	}
	defer client.Close()

	if supported, _ := client.Extension("STARTTLS"); !supported {
		return errors.New("servidor SMTP não oferece STARTTLS")
	}
	if err := client.StartTLS(&tls.Config{
		ServerName: s.host,
		MinVersion: tls.VersionTLS12,
	}); err != nil {
		return fmt.Errorf("iniciar TLS no SMTP: %w", err)
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("autenticar no servidor SMTP: %w", err)
	}
	if err := client.Mail(s.from.Address); err != nil {
		return fmt.Errorf("definir remetente SMTP: %w", err)
	}
	if err := client.Rcpt(to.Address); err != nil {
		return fmt.Errorf("definir destinatário SMTP: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("iniciar mensagem SMTP: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("escrever mensagem SMTP: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("finalizar mensagem SMTP: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("encerrar sessão SMTP: %w", err)
	}
	return nil
}

func smtpMessage(from, to *mail.Address, subject, body string) ([]byte, error) {
	if strings.ContainsAny(subject, "\r\n") {
		return nil, errors.New("assunto SMTP não pode conter quebras de linha")
	}

	var message bytes.Buffer
	headers := []string{
		"From: " + from.String(),
		"To: " + to.String(),
		"Subject: " + mime.QEncoding.Encode("UTF-8", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: quoted-printable",
	}
	for _, header := range headers {
		if strings.ContainsAny(header, "\r\n") {
			return nil, errors.New("cabeçalho SMTP contém quebra de linha inválida")
		}
		if _, err := message.WriteString(header + "\r\n"); err != nil {
			return nil, fmt.Errorf("montar cabeçalhos SMTP: %w", err)
		}
	}
	if _, err := message.WriteString("\r\n"); err != nil {
		return nil, fmt.Errorf("montar mensagem SMTP: %w", err)
	}
	encodedBody := quotedprintable.NewWriter(&message)
	if _, err := encodedBody.Write([]byte(body)); err != nil {
		return nil, fmt.Errorf("codificar corpo SMTP: %w", err)
	}
	if err := encodedBody.Close(); err != nil {
		return nil, fmt.Errorf("finalizar corpo SMTP: %w", err)
	}
	return message.Bytes(), nil
}

var _ shared.EmailSender = (*SMTPSender)(nil)
