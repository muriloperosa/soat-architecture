package usuario

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

var (
	ErrNomeObrigatorio      = shared.NewValidationError("nome é obrigatório")
	ErrPapelInvalido        = shared.NewValidationError("papel inválido")
	ErrUsuarioNaoEncontrado = shared.NewNotFoundError("usuário não encontrado")
)
