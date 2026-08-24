package veiculo

import domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"

type CadastrarVeiculoInput struct {
	Placa              string
	Marca              string
	Modelo             string
	QuilometragemAtual uint32
	Ano                uint16
	Cor                string
	CriadoPor          uint64
}

type AtualizarVeiculoInput struct {
	ID     uint64
	Marca  string
	Modelo string
	Cor    string
}

type VeiculoOutput struct {
	ID                 uint64
	Placa              string
	Marca              string
	Modelo             string
	QuilometragemAtual uint32
	Ano                uint16
	Cor                string
	CriadoPor          uint64
	Ativo              bool
}

func toOutput(v *domain.Veiculo) VeiculoOutput {
	return VeiculoOutput{
		ID:                 v.ID(),
		Placa:              v.Placa().String(),
		Marca:              v.Marca(),
		Modelo:             v.Modelo(),
		QuilometragemAtual: v.QuilometragemAtual(),
		Ano:                v.Ano(),
		Cor:                v.Cor().String(),
		CriadoPor:          v.CriadoPor(),
		Ativo:              v.Ativo(),
	}
}
