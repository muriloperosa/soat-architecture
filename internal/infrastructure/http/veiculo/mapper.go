package veiculo

import (
	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
)

func toCadastrarInput(criadoPor uint64, req CadastrarVeiculoRequest) appveiculo.CadastrarVeiculoInput {
	return appveiculo.CadastrarVeiculoInput{
		Placa:              req.Placa,
		Marca:              req.Marca,
		Modelo:             req.Modelo,
		QuilometragemAtual: req.QuilometragemAtual,
		Ano:                req.Ano,
		Cor:                req.Cor,
		CriadoPor:          criadoPor,
	}
}

func toAtualizarInput(id uint64, req AtualizarVeiculoRequest) appveiculo.AtualizarVeiculoInput {
	return appveiculo.AtualizarVeiculoInput{
		ID:                 id,
		Marca:              req.Marca,
		Modelo:             req.Modelo,
		Cor:                req.Cor,
		QuilometragemAtual: req.QuilometragemAtual,
	}
}

func toVeiculoResponse(out appveiculo.VeiculoOutput) VeiculoResponse {
	return VeiculoResponse{
		ID:                 out.ID,
		Placa:              out.Placa,
		Marca:              out.Marca,
		Modelo:             out.Modelo,
		QuilometragemAtual: out.QuilometragemAtual,
		Ano:                out.Ano,
		Cor:                out.Cor,
		CriadoPor:          out.CriadoPor,
		Ativo:              out.Ativo,
	}
}

func toListResponse(page query.Page[appveiculo.VeiculoOutput]) ListarVeiculosResponse {
	items := make([]VeiculoResponse, 0, len(page.Items))

	for _, item := range page.Items {
		items = append(items, toVeiculoResponse(item))
	}

	return ListarVeiculosResponse{
		Items:     items,
		Total:     page.Total,
		Offset:    page.Offset,
		Limit:     page.Limit,
		Order:     page.Order,
		Direction: string(page.Direction),
	}
}
