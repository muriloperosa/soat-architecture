package veiculo

import (
	"regexp"
	"strings"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared/texts"
)

var (
	ErrPlacaObrigatoria = shared.NewValidationError("placa é obrigatória")
	ErrPlacaInvalida    = shared.NewValidationError("placa inválida")

	placaAntigaRegex   = regexp.MustCompile(`^[A-Z]{3}[0-9]{4}$`)
	placaMercosulRegex = regexp.MustCompile(`^[A-Z]{3}[0-9][A-Z][0-9]{2}$`)
)

type Placa struct {
	valor string
}

func NewPlaca(valor string) (Placa, error) {
	valor = strings.ReplaceAll(texts.NormalizeUpper(valor), "-", "")

	if valor == "" {
		return Placa{}, ErrPlacaObrigatoria
	}

	if !placaAntigaRegex.MatchString(valor) && !placaMercosulRegex.MatchString(valor) {
		return Placa{}, ErrPlacaInvalida
	}

	return Placa{valor: valor}, nil
}

func (p Placa) String() string {
	return p.valor
}
