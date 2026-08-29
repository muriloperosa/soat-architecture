package peca

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

var (
	ErrNomeObrigatorio                   = shared.NewValidationError("nome é obrigatório")
	ErrMarcaObrigatoria                  = shared.NewValidationError("marca é obrigatória")
	ErrDescricaoObrigatoria              = shared.NewValidationError("descrição é obrigatória")
	ErrPrecoInvalido                     = shared.NewValidationError("preço não pode ser negativo")
	ErrQuantidadeEmEstoqueInvalida       = shared.NewValidationError("quantidade em estoque não pode ser negativa")
	ErrEstoqueMinimoInvalido             = shared.NewValidationError("estoque mínimo não pode ser negativo")
	ErrCriadoPorObrigatorio              = shared.NewValidationError("usuário responsável pelo cadastro é obrigatório")
	ErrQuantidadeInvalida                = shared.NewValidationError("quantidade deve ser maior que zero")
	ErrEstoqueInsuficiente               = shared.NewValidationError("operação deixaria o estoque abaixo do mínimo")
	ErrQuantidadeIndisponivelParaReserva = shared.NewValidationError("quantidade solicitada não está disponível para reserva")
	ErrPecaNaoEncontrada                 = shared.NewNotFoundError("peça não encontrada")
)
