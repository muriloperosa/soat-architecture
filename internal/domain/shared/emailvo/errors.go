package emailvo

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

var (
	ErrEmailInvalido    = shared.NewValidationError("email inválido")
	ErrEmailObrigatorio = shared.NewValidationError("email obrigatório")
)
