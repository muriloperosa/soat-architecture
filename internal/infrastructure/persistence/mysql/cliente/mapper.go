package cliente

import (
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

func toModel(cliente domain.Cliente) *Model {
	return &Model{
		ID:                 cliente.ID(),
		Documento:          cliente.Documento().String(),
		Tipo:               string(cliente.Tipo()),
		Nome:               cliente.Nome(),
		Email:              cliente.Email().String(),
		Senha:              cliente.Senha().String(),
		Telefone:           cliente.Telefone().String(),
		Ativo:              cliente.Ativo(),
		DataCadastro:       cliente.DataCadastro(),
		DataAtualizacao:    cliente.DataAtualizacao(),
		RequerAlterarSenha: cliente.RequerAlterarSenha(),
		CriadoPor:          cliente.CriadoPor(),
	}
}

func toEntity(model Model) (*domain.Cliente, error) {
	return domain.RestaurarCliente(
		model.ID,
		model.Documento,
		domain.TipoPessoa(model.Tipo),
		model.Nome,
		model.Email,
		model.Telefone,
		model.Senha,
		model.CriadoPor,
		model.RequerAlterarSenha,
		model.Ativo,
		model.DataCadastro,
		model.DataAtualizacao,
	)
}

func toCredencial(cliente *domain.Cliente) *domainauth.Credencial {
	return &domainauth.Credencial{
		ID:                 cliente.ID(),
		SenhaHash:          cliente.Senha().String(),
		Papel:              shared.PapelCliente,
		Ativo:              cliente.Ativo(),
		RequerAlterarSenha: cliente.RequerAlterarSenha(),
	}
}
