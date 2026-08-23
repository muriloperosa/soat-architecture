package veiculo

import (
	"regexp"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared/texts"
)

var (
	ErrCorObrigatoria = shared.NewValidationError("cor é obrigatória")
	ErrCorInvalida    = shared.NewValidationError("cor inválida")

	corRegex = regexp.MustCompile(`^\p{L}+( \p{L}+)*$`)
)

type Cor struct {
	valor string
}

func NewCor(valor string) (Cor, error) {
	valor = texts.NormalizeUcFirst(valor)

	if valor == "" {
		return Cor{}, ErrCorObrigatoria
	}

	if !corRegex.MatchString(valor) {
		return Cor{}, ErrCorInvalida
	}

	return Cor{valor: valor}, nil
}

func (c Cor) String() string {
	return c.valor
}
