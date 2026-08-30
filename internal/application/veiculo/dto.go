package veiculo

import (
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
)

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
	ID                 uint64
	Marca              string
	Modelo             string
	Cor                string
	QuilometragemAtual uint32
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

// ListarVeiculosInput é o contrato de entrada do caso de uso de listagem.
type ListarVeiculosInput struct {
	appquery.ParamsInput
}

// ListarVeiculosOutput é o contrato de saída do caso de uso de listagem.
type ListarVeiculosOutput struct {
	Items      []VeiculoOutput
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
	Order      string
	Direction  string
}
