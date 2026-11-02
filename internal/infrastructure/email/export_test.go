package email

import "crypto/x509"

// WithRootCAs permite que os testes confiem no certificado do servidor SMTP falso.
func (s *SMTPSender) WithRootCAs(roots *x509.CertPool) *SMTPSender {
	s.rootCAs = roots
	return s
}
