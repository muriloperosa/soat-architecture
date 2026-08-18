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
	}
}
