package cliente

import domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"

func toModel(cliente domain.Cliente) *ClienteModel {
	return &ClienteModel{
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
	}
}

func toDomain(model ClienteModel) (*domain.Cliente, error) {
	return domain.ReidratarCliente(
		model.ID,
		model.Documento,
		domain.TipoPessoa(model.Tipo),
		model.Nome,
		model.Email,
		model.Telefone,
		model.Senha,
		model.RequerAlterarSenha,
		model.Ativo,
		model.DataCadastro,
		model.DataAtualizacao,
	)
}
