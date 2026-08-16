package emailvo

import (
	"net/mail"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared/texts"
)

type Email struct {
	valor string
}

func NewEmail(valor string) (Email, error) {
	valor = texts.NormalizeSpaces(valor)
	valor = texts.NormalizeLower(valor)

	if valor == "" {
		return Email{}, ErrEmailObrigatorio
	}

	if err := validateEmail(valor); err != nil {
		return Email{}, err
	}

	return Email{valor: valor}, nil
}

func (e Email) String() string {
	return e.valor
}

func validateEmail(valor string) error {
	address, err := mail.ParseAddress(valor)
	if err != nil {
		return ErrEmailInvalido
	}

	if address.Address != valor {
		return ErrEmailInvalido
	}

	return nil
}
