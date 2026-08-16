package shared

import (
	"net/mail"
	"strings"
)

var (
	ErrEmailObrigatorio = NewValidationError("email é obrigatório")
	ErrEmailInvalido    = NewValidationError("email inválido")
)

// Email é o VO de endereço de email, sempre normalizado (trim + lowercase)
// antes de validar o formato.
type Email struct {
	valor string
}

// NewEmail normaliza valor e valida o formato (RFC 5322 via net/mail).
func NewEmail(valor string) (Email, error) {
	valor = strings.ToLower(strings.TrimSpace(valor))
	if valor == "" {
		return Email{}, ErrEmailObrigatorio
	}
	if _, err := mail.ParseAddress(valor); err != nil {
		return Email{}, ErrEmailInvalido
	}
	return Email{valor: valor}, nil
}

// String retorna o valor de email como string
func (e Email) String() string {
	return e.valor
}
