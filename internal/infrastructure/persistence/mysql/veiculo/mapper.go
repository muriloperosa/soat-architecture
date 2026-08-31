package veiculo

import domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"

func toModel(veiculo *domain.Veiculo) *Model {
	return &Model{
		ID:                 veiculo.ID(),
		Placa:              veiculo.Placa().String(),
		Marca:              veiculo.Marca(),
		Modelo:             veiculo.Modelo(),
		QuilometragemAtual: veiculo.QuilometragemAtual(),
		Ano:                veiculo.Ano(),
		Cor:                veiculo.Cor().String(),
		CriadoPor:          veiculo.CriadoPor(),
		Ativo:              veiculo.Ativo(),
		DataCadastro:       veiculo.DataCadastro(),
		DataAtualizacao:    veiculo.DataAtualizacao(),
	}
}

func toDomain(model Model) *domain.Veiculo {
	placaVO, _ := domain.NewPlaca(model.Placa)
	corVO, _ := domain.NewCor(model.Cor)

	return domain.RestaurarVeiculo(
		model.ID,
		placaVO,
		model.Marca,
		model.Modelo,
		model.QuilometragemAtual,
		model.Ano,
		corVO,
		model.CriadoPor,
		model.Ativo,
		model.DataCadastro,
		model.DataAtualizacao,
	)
}
