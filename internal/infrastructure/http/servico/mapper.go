package servico

import (
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	appservico "github.com/muriloperosa/soat-architecture/internal/application/servico"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
)

// toCriarInput converte o DTO HTTP de criação pro DTO de entrada do
// CriarServicoUseCase. criadoPor vem do subject do JWT, não do corpo.
func toCriarInput(req CriarServicoRequest, criadoPor uint64) appservico.CriarServicoInput {
	return appservico.CriarServicoInput{
		Nome:                 req.Nome,
		Descricao:            req.Descricao,
		PrecoBase:            *req.PrecoBase,
		TempoEstimadoMinutos: req.TempoEstimadoMinutos,
		CriadoPor:            criadoPor,
	}
}

// toAtualizarInput converte o DTO HTTP de atualização pro DTO de entrada do
// AtualizarServicoUseCase. id vem do path param, não do corpo.
func toAtualizarInput(id uint64, req AtualizarServicoRequest) appservico.AtualizarServicoInput {
	return appservico.AtualizarServicoInput{
		ID:                   id,
		Nome:                 req.Nome,
		Descricao:            req.Descricao,
		PrecoBase:            *req.PrecoBase,
		TempoEstimadoMinutos: req.TempoEstimadoMinutos,
	}
}

// toServicoResponse converte o DTO de saída dos use cases pra resposta HTTP.
func toServicoResponse(out appservico.ServicoOutput) ServicoResponse {
	return ServicoResponse{
		ID:                   out.ID,
		Nome:                 out.Nome,
		Descricao:            out.Descricao,
		PrecoBase:            out.PrecoBase,
		TempoEstimadoMinutos: out.TempoEstimadoMinutos,
		CriadoPor:            out.CriadoPor,
		Ativo:                out.Ativo,
	}
}

func toListarInput(params httpquery.Params) appservico.ListarServicosInput {
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

	return appservico.ListarServicosInput{
		ParamsInput: appquery.ParamsInput{
			Page:      params.Page,
			Order:     params.Order,
			Direction: params.Direction,
			Filters:   filters,
		},
	}
}

func toListResponse(output appservico.ListarServicosOutput) ListarServicosResponse {
	items := make([]ServicoResponse, 0, len(output.Items))

	for _, item := range output.Items {
		items = append(items, toServicoResponse(item))
	}

	return ListarServicosResponse{
		Items:      items,
		Total:      output.Total,
		Page:       output.Page,
		PageSize:   output.PageSize,
		TotalPages: output.TotalPages,
		Order:      output.Order,
		Direction:  output.Direction,
	}
}
