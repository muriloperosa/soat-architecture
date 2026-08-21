package cliente

import domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"

// CriarClienteInput é o DTO de entrada do CriarClienteUseCase.
type CriarClienteInput struct {
	Documento string
	Tipo      domain.TipoPessoa
	Nome      string
	Email     string
	Telefone  string
	Senha     string
	CriadoPor uint64
}

// AtualizarClienteInput é o DTO de entrada do AtualizarClienteUseCase.
type AtualizarClienteInput struct {
	ID       uint64
	Nome     string
	Email    string
	Telefone string
}

// AlterarSenhaInput é o DTO de entrada do AlterarSenhaUseCase.
type AlterarSenhaInput struct {
	ClienteID uint64
	SenhaNova string
}

// ClienteOutput é o DTO de saída comum aos use cases de gestão/consulta de cliente.
type ClienteOutput struct {
	ID                 uint64
	Documento          string
	Tipo               domain.TipoPessoa
	Nome               string
	Email              string
	Telefone           string
	Ativo              bool
	RequerAlterarSenha bool
	CriadoPor          uint64
}

func toOutput(c *domain.Cliente) ClienteOutput {
	return ClienteOutput{
		ID:                 c.ID(),
		Documento:          c.Documento().String(),
		Tipo:               c.Tipo(),
		Nome:               c.Nome(),
		Email:              c.Email().String(),
		Telefone:           c.Telefone().String(),
		Ativo:              c.Ativo(),
		RequerAlterarSenha: c.RequerAlterarSenha(),
		CriadoPor:          c.CriadoPor(),
	}
}
