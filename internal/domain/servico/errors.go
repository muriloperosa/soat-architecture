package servico

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

var (
	ErrNomeObrigatorio      = shared.NewValidationError("nome é obrigatório")
	ErrDescricaoObrigatoria = shared.NewValidationError("descrição é obrigatória")
	ErrPrecoInvalido        = shared.NewValidationError("preço base não pode ser negativo")
	ErrCriadoPorObrigatorio = shared.NewValidationError("criado por é obrigatório")
	ErrServicoNaoEncontrado = shared.NewNotFoundError("serviço não encontrado")
)
