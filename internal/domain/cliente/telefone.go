package cliente

import (
	"fmt"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared/texts"
)

const (
	PhoneNumberLength       = 10
	MobilePhoneNumberLength = 11
)

type Telefone struct {
	valor string
}

func NewTelefone(valor string) (Telefone, error) {
	valor = texts.OnlyNumbers(valor)

	if valor == "" {
		return Telefone{}, ErrTelefoneObrigatorio
	}

	if !validarTelefone(valor) {
		return Telefone{}, ErrTelefoneInvalido
	}

	return Telefone{valor: valor}, nil
}

func (t Telefone) String() string {
	return t.valor
}

func (t Telefone) Formatado() string {
	switch len(t.valor) {
	case PhoneNumberLength:
		return fmt.Sprintf("(%s) %s-%s", t.valor[0:2], t.valor[2:6], t.valor[6:10])

	case MobilePhoneNumberLength:
		return fmt.Sprintf("(%s) %s-%s", t.valor[0:2], t.valor[2:7], t.valor[7:11])

	default:
		return t.valor
	}
}

func validarTelefone(valor string) bool {
	if len(valor) != PhoneNumberLength && len(valor) != MobilePhoneNumberLength {
		return false
	}

	if valor[0] == '0' {
		return false
	}

	if len(valor) == MobilePhoneNumberLength {
		return validarCelular(valor)
	}

	return validarTelefoneFixo(valor)
}

func validarCelular(valor string) bool {
	return valor[2] == '9'
}

func validarTelefoneFixo(valor string) bool {
	primeiroDigito := valor[2]

	return primeiroDigito >= '2' && primeiroDigito <= '5'
}
