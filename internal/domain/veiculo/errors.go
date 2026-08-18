package veiculo

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

var (
	ErrMarcaObrigatoria      = shared.NewValidationError("marca é obrigatória")
	ErrModeloObrigatorio     = shared.NewValidationError("modelo é obrigatório")
	ErrCorObrigatoria        = shared.NewValidationError("cor é obrigatória")
	ErrAnoInvalido           = shared.NewValidationError("ano inválido")
	ErrCriadoPorObrigatorio  = shared.NewValidationError("usuário responsável pelo cadastro é obrigatório")
	ErrQuilometragemInvalida = shared.NewValidationError("quilometragem não pode ser menor que a atual")
)
