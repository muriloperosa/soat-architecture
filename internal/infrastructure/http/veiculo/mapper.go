package veiculo

import (
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
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

func toListarInput(params httpquery.Params) appveiculo.ListarVeiculosInput {
	var filters []appquery.FilterInput

	if len(params.Filters) > 0 {
		filters = make([]appquery.FilterInput, 0, len(params.Filters))

		for _, filter := range params.Filters {
			filters = append(filters, appquery.FilterInput{
				Field:    filter.Field,
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		}
	}

	return appveiculo.ListarVeiculosInput{
		ParamsInput: appquery.ParamsInput{
			Page:      params.Page,
			Order:     params.Order,
			Direction: params.Direction,
			Filters:   filters,
		},
	}
}

func toListResponse(output appveiculo.ListarVeiculosOutput) ListarVeiculosResponse {
	items := make([]VeiculoResponse, 0, len(output.Items))

	for _, item := range output.Items {
		items = append(items, toVeiculoResponse(item))
	}

	return ListarVeiculosResponse{
		Items:      items,
		Total:      output.Total,
		Page:       output.Page,
		PageSize:   output.PageSize,
		TotalPages: output.TotalPages,
		Order:      output.Order,
		Direction:  output.Direction,
	}
}
