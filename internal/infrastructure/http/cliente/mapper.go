package cliente

import (
	app "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
)

func toCriarInput(criadoPor uint64, req CriarClienteRequest) app.CriarClienteInput {
	return app.CriarClienteInput{
		Documento: req.Documento,
		Tipo:      req.TipoPessoa,
		Nome:      req.Nome,
		Email:     req.Email,
		Telefone:  req.Telefone,
		Senha:     req.Senha,
		CriadoPor: criadoPor,
	}
}

func toAtualizarInput(id uint64, req AtualizarClienteRequest) app.AtualizarClienteInput {
	return app.AtualizarClienteInput{
		ID:       id,
		Nome:     req.Nome,
		Email:    req.Email,
		Telefone: req.Telefone,
	}
}

func (r AlterarSenhaRequest) toInput(id uint64) app.AlterarSenhaInput {
	return app.AlterarSenhaInput{
		ClienteID: id,
		SenhaNova: r.SenhaNova,
	}
}

func toResponse(output app.ClienteOutput) ClienteResponse {
	return ClienteResponse{
		ID:                 output.ID,
		Documento:          output.Documento,
		TipoPessoa:         string(output.Tipo),
		Nome:               output.Nome,
		Email:              output.Email,
		Telefone:           output.Telefone,
		Ativo:              output.Ativo,
		RequerAlterarSenha: output.RequerAlterarSenha,
		CriadoPor:          output.CriadoPor,
	}
}

func toListarInput(params httpquery.Params) app.ListarClientesInput {
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

	return app.ListarClientesInput{
		ParamsInput: appquery.ParamsInput{
			Page:      params.Page,
			Order:     params.Order,
			Direction: params.Direction,
			Filters:   filters,
		},
	}
}

func toListResponse(output app.ListarClientesOutput) ListarClientesResponse {
	items := make([]ClienteResponse, 0, len(output.Items))

	for _, item := range output.Items {
		items = append(items, toResponse(item))
	}

	return ListarClientesResponse{
		Items:      items,
		Total:      output.Total,
		Page:       output.Page,
		PageSize:   output.PageSize,
		TotalPages: output.TotalPages,
		Order:      output.Order,
		Direction:  output.Direction,
	}
}
