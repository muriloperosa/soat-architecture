package cliente

import "errors"

var (
	ErrDocumentoObrigatorio = errors.New("documento é obrigatório")
	ErrTipoPessoaInvalido   = errors.New("tipo de pessoa inválido")
	ErrCPFInvalido          = errors.New("CPF inválido")
	ErrCNPJInvalido         = errors.New("CNPJ inválido")
	ErrTelefoneObrigatorio  = errors.New("telefone é obrigatório")
	ErrTelefoneInvalido     = errors.New("telefone inválido")
)
