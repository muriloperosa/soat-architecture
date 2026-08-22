package reservapeca

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

var (
	ErrOrdemServicoObrigatoria      = shared.NewValidationError("ordem de serviço é obrigatória")
	ErrPecaObrigatoria              = shared.NewValidationError("peça é obrigatória")
	ErrQuantidadeInvalida           = shared.NewValidationError("quantidade deve ser maior que zero")
	ErrQuantidadeSuperiorAReservada = shared.NewValidationError("quantidade a liberar não pode ser maior que a quantidade reservada")
	ErrReservaNaoEncontrada         = shared.NewNotFoundError("reserva não encontrada")
)
