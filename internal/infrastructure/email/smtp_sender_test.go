package email_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"mime/quotedprintable"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/email"
	"github.com/stretchr/testify/require"
)

// fakeSMTPServer simula um servidor SMTP com STARTTLS e AUTH PLAIN em
// 127.0.0.1, sem depender de rede externa.
type fakeSMTPServer struct {
	semStartTLS bool
	falhaAuth   bool
	falhaRcpt   bool

	port  int
	roots *x509.CertPool
	got   chan smtpCapturado
}

type smtpCapturado struct {
	from, to, data string
	usouTLS        bool
}

func novoCertificadoLocal(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "127.0.0.1"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, roots
}

func (f *fakeSMTPServer) iniciar(t *testing.T) {
	t.Helper()
	cert, roots := novoCertificadoLocal(t)
	f.roots = roots
	f.got = make(chan smtpCapturado, 1)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	f.port = listener.Addr().(*net.TCPAddr).Port

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		f.atender(conn, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12})
	}()
}

func (f *fakeSMTPServer) atender(conn net.Conn, tlsConfig *tls.Config) {
	var capturado smtpCapturado
	text := textproto.NewConn(conn)
	_ = text.PrintfLine("220 fake ESMTP")
	for {
		line, err := text.ReadLine()
		if err != nil {
			return
		}
		verb := strings.ToUpper(strings.SplitN(line, " ", 2)[0])
		switch {
		case verb == "EHLO" && capturado.usouTLS:
			_ = text.PrintfLine("250-fake\r\n250 AUTH PLAIN")
		case verb == "EHLO" && f.semStartTLS:
			_ = text.PrintfLine("250 fake")
		case verb == "EHLO":
			_ = text.PrintfLine("250-fake\r\n250 STARTTLS")
		case verb == "STARTTLS":
			_ = text.PrintfLine("220 pronto")
			tlsConn := tls.Server(conn, tlsConfig)
			if tlsConn.Handshake() != nil {
				return
			}
			conn = tlsConn
			text = textproto.NewConn(tlsConn)
			capturado.usouTLS = true
		case verb == "AUTH" && f.falhaAuth:
			_ = text.PrintfLine("535 credenciais inválidas")
		case verb == "AUTH":
			_ = text.PrintfLine("235 ok")
		case verb == "MAIL":
			capturado.from = line
			_ = text.PrintfLine("250 ok")
		case verb == "RCPT" && f.falhaRcpt:
			_ = text.PrintfLine("550 destinatário recusado")
		case verb == "RCPT":
			capturado.to = line
			_ = text.PrintfLine("250 ok")
		case verb == "DATA":
			_ = text.PrintfLine("354 envie")
			data, err := text.ReadDotBytes()
			if err != nil {
				return
			}
			capturado.data = string(data)
			_ = text.PrintfLine("250 ok")
		case verb == "QUIT":
			_ = text.PrintfLine("221 tchau")
			f.got <- capturado
			return
		default:
			_ = text.PrintfLine("250 ok")
		}
	}
}

func (f *fakeSMTPServer) sender(t *testing.T) *email.SMTPSender {
	t.Helper()
	sender, err := email.NewSMTPSender(email.SMTPConfig{
		Host:     "127.0.0.1",
		Port:     f.port,
		Username: "oficina@exemplo.com",
		Password: "senha-de-app",
		From:     "Oficina <oficina@exemplo.com>",
	})
	require.NoError(t, err)
	return sender.WithRootCAs(f.roots)
}

func TestSMTPSender_EnviarUsaSTARTTLSEAutenticacao(t *testing.T) {
	server := &fakeSMTPServer{}
	server.iniciar(t)

	err := server.sender(t).Enviar(context.Background(), "cliente@exemplo.com", "Ordem de Serviço OS-1: RECEBIDA", "Olá, ação concluída.")

	require.NoError(t, err)
	got := <-server.got
	require.True(t, got.usouTLS)
	require.Equal(t, "MAIL FROM:<oficina@exemplo.com>", got.from)
	require.Equal(t, "RCPT TO:<cliente@exemplo.com>", got.to)

	header, body, found := strings.Cut(got.data, "\n\n")
	require.True(t, found)
	require.Contains(t, header, `From: "Oficina" <oficina@exemplo.com>`)
	require.Contains(t, header, "To: <cliente@exemplo.com>")
	require.Contains(t, header, "Subject: =?UTF-8?q?")
	require.Contains(t, header, "Content-Type: text/plain; charset=UTF-8")
	decoded, err := readQuotedPrintable(body)
	require.NoError(t, err)
	require.Equal(t, "Olá, ação concluída.", decoded)
}

func readQuotedPrintable(body string) (string, error) {
	decoded, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(body)))
	return strings.TrimRight(string(decoded), "\n"), err
}

func TestSMTPSender_EnviarFalhasDoServidor(t *testing.T) {
	cases := []struct {
		name   string
		server *fakeSMTPServer
		erro   string
	}{
		{"sem STARTTLS", &fakeSMTPServer{semStartTLS: true}, "não oferece STARTTLS"},
		{"autenticação recusada", &fakeSMTPServer{falhaAuth: true}, "autenticar no servidor SMTP"},
		{"destinatário recusado", &fakeSMTPServer{falhaRcpt: true}, "definir destinatário SMTP"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.server.iniciar(t)

			err := tc.server.sender(t).Enviar(context.Background(), "cliente@exemplo.com", "Assunto", "Corpo")

			require.ErrorContains(t, err, tc.erro)
		})
	}
}

func TestSMTPSender_EnviarRecusaCertificadoNaoConfiavel(t *testing.T) {
	server := &fakeSMTPServer{}
	server.iniciar(t)
	sender := server.sender(t).WithRootCAs(x509.NewCertPool())

	err := sender.Enviar(context.Background(), "cliente@exemplo.com", "Assunto", "Corpo")

	require.ErrorContains(t, err, "iniciar TLS no SMTP")
}

func TestSMTPSender_EnviarFalhaAoConectar(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	sender, err := email.NewSMTPSender(email.SMTPConfig{
		Host: "127.0.0.1", Port: port, Username: "u", Password: "p", From: "oficina@exemplo.com",
	})
	require.NoError(t, err)

	err = sender.Enviar(context.Background(), "cliente@exemplo.com", "Assunto", "Corpo")

	require.ErrorContains(t, err, "conectar ao servidor SMTP")
}

func TestSMTPSender_EnviarValidaEntradaAntesDeConectar(t *testing.T) {
	sender, err := email.NewSMTPSender(email.SMTPConfig{
		Host: "smtp.exemplo.com", Port: 587, Username: "u", Password: "p", From: "oficina@exemplo.com",
	})
	require.NoError(t, err)

	err = sender.Enviar(context.Background(), "não é e-mail", "Assunto", "Corpo")
	require.ErrorContains(t, err, "destinatário SMTP inválido")

	err = sender.Enviar(context.Background(), "cliente@exemplo.com", "Assunto\r\nBcc: x@exemplo.com", "Corpo")
	require.ErrorContains(t, err, "não pode conter quebras de linha")
}

func TestNewSMTPSender_ValidaConfiguracao(t *testing.T) {
	valida := email.SMTPConfig{Host: "smtp.gmail.com", Port: 587, Username: "u", Password: "p", From: "oficina@exemplo.com"}
	cases := []struct {
		name  string
		mudar func(*email.SMTPConfig)
		erro  string
	}{
		{"host vazio", func(c *email.SMTPConfig) { c.Host = " " }, "SMTP_HOST"},
		{"host com porta", func(c *email.SMTPConfig) { c.Host = "smtp.gmail.com:587" }, "SMTP_HOST"},
		{"porta zero", func(c *email.SMTPConfig) { c.Port = 0 }, "SMTP_PORT"},
		{"porta alta", func(c *email.SMTPConfig) { c.Port = 70000 }, "SMTP_PORT"},
		{"usuário vazio", func(c *email.SMTPConfig) { c.Username = "" }, "SMTP_USERNAME"},
		{"senha vazia", func(c *email.SMTPConfig) { c.Password = "  " }, "SMTP_PASSWORD"},
		{"remetente inválido", func(c *email.SMTPConfig) { c.From = "oficina" }, "SMTP_FROM_ADDRESS"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := valida
			tc.mudar(&config)

			sender, err := email.NewSMTPSender(config)

			require.Nil(t, sender)
			require.ErrorContains(t, err, tc.erro)
		})
	}
}

func TestNewSender_SMTP(t *testing.T) {
	sender, err := email.NewSender(context.Background(), email.SenderConfig{
		Provider: " SMTP ",
		SMTP:     email.SMTPConfig{Host: "smtp.gmail.com", Port: 587, Username: "u", Password: "p", From: "oficina@exemplo.com"},
	})

	require.NoError(t, err)
	require.IsType(t, &email.SMTPSender{}, sender)
}
