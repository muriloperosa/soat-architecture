package cliente

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

var (
	ErrDocumentoObrigatorio = shared.NewValidationError("documento é obrigatório")
	ErrTipoPessoaInvalido   = shared.NewValidationError("tipo de pessoa inválido")
	ErrCPFInvalido          = shared.NewValidationError("CPF inválido")
	ErrCNPJInvalido         = shared.NewValidationError("CNPJ inválido")
	ErrTelefoneObrigatorio  = shared.NewValidationError("telefone é obrigatório")
	ErrTelefoneInvalido     = shared.NewValidationError("telefone inválido")
	ErrNomeObrigatorio      = shared.NewValidationError("nome é obrigatório")
	ErrEmailObrigatorio     = shared.NewValidationError("email é obrigatório")
	ErrSenhaObrigatoria     = shared.NewValidationError("senha é obrigatória")
)
