package orcamento

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

var (
	ErrOrdemServicoObrigatoria  = shared.NewValidationError("ordem de serviço é obrigatória")
	ErrCriadoPorObrigatorio     = shared.NewValidationError("usuário responsável pela criação é obrigatório")
	ErrServicoObrigatorio       = shared.NewValidationError("serviço é obrigatório")
	ErrServicoInativo           = shared.NewValidationError("serviço inativo não pode ser incluído no orçamento")
	ErrPecaObrigatoria          = shared.NewValidationError("peça é obrigatória")
	ErrDescricaoObrigatoria     = shared.NewValidationError("descrição é obrigatória")
	ErrQuantidadeInvalida       = shared.NewValidationError("quantidade deve ser maior que zero")
	ErrValorInvalido            = shared.NewValidationError("valor não pode ser negativo")
	ErrItemServicoNaoEncontrado = shared.NewNotFoundError("item de serviço não encontrado no orçamento")
	ErrItemPecaNaoEncontrado    = shared.NewNotFoundError("item de peça não encontrado no orçamento")
	ErrOrcamentoNaoEncontrado   = shared.NewNotFoundError("orçamento não encontrado")
	ErrOrcamentoJaExiste        = shared.NewConflictError("ordem de serviço já possui um orçamento")
)
