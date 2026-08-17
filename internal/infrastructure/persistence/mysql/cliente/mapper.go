package cliente

import domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"

func toModel(cliente domain.Cliente) ClienteModel {
	return ClienteModel{
		ID:              cliente.ID(),
		Documento:       cliente.Documento().String(),
		Tipo:            string(cliente.Tipo()),
		Nome:            cliente.Nome(),
		Email:           cliente.Email(),
		Senha:           cliente.Senha(),
		Telefone:        cliente.Telefone().String(),
		Ativo:           cliente.Ativo(),
		DataCadastro:    cliente.DataCadastro(),
		DataAtualizacao: cliente.DataAtualizacao(),
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
		model.Ativo,
		model.DataCadastro,
		model.DataAtualizacao,
	)
}
