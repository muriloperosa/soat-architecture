package peca

import (
	apppeca "github.com/muriloperosa/soat-architecture/internal/application/peca"
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
)

// toCadastrarInput converte o DTO HTTP de cadastro pro DTO de entrada do
// CadastrarPecaUseCase. criadoPor vem do subject do JWT, não do corpo da
// requisição.
func toCadastrarInput(criadoPor uint64, req CadastrarPecaRequest) apppeca.CadastrarPecaInput {
	return apppeca.CadastrarPecaInput{
		Nome:                req.Nome,
		Marca:               req.Marca,
		Descricao:           req.Descricao,
		Preco:               req.Preco,
		QuantidadeEmEstoque: req.QuantidadeEmEstoque,
		EstoqueMinimo:       req.EstoqueMinimo,
		CriadoPor:           criadoPor,
	}
}

// toAtualizarInput converte o DTO HTTP de atualização pro DTO de entrada do
// AtualizarPecaUseCase. id vem do path param, não do corpo da requisição.
func toAtualizarInput(id uint64, req AtualizarPecaRequest) apppeca.AtualizarPecaInput {
	return apppeca.AtualizarPecaInput{
		ID:            id,
		Nome:          req.Nome,
		Marca:         req.Marca,
		Descricao:     req.Descricao,
		Preco:         req.Preco,
		EstoqueMinimo: req.EstoqueMinimo,
	}
}

// toReporEstoqueInput converte o DTO HTTP de reposição pro DTO de entrada do
// ReporEstoqueUseCase. id vem do path param, não do corpo da requisição.
func toReporEstoqueInput(id uint64, req ReporEstoqueRequest) apppeca.ReporEstoqueInput {
	return apppeca.ReporEstoqueInput{
		PecaID:     id,
		Quantidade: req.Quantidade,
	}
}

// toPecaResponse converte o DTO de saída dos use cases de gestão/consulta
// pra resposta HTTP comum (criação/atualização/consulta de peça).
func toPecaResponse(out apppeca.PecaOutput) PecaResponse {
	return PecaResponse{
		ID:                  out.ID,
		Codigo:              out.Codigo,
		Nome:                out.Nome,
		Marca:               out.Marca,
		Descricao:           out.Descricao,
		Preco:               out.Preco,
		QuantidadeEmEstoque: out.QuantidadeEmEstoque,
		EstoqueMinimo:       out.EstoqueMinimo,
		CriadoPor:           out.CriadoPor,
		Ativo:               out.Ativo,
	}
}

func toListarInput(params httpquery.Params) apppeca.ListarPecasInput {
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

	return apppeca.ListarPecasInput{
		ParamsInput: appquery.ParamsInput{
			Page:      params.Page,
			Order:     params.Order,
			Direction: params.Direction,
			Filters:   filters,
		},
	}
}

func toListResponse(output apppeca.ListarPecasOutput) ListarPecasResponse {
	items := make([]PecaResponse, 0, len(output.Items))

	for _, item := range output.Items {
		items = append(items, toPecaResponse(item))
	}

	return ListarPecasResponse{
		Items:      items,
		Total:      output.Total,
		Page:       output.Page,
		PageSize:   output.PageSize,
		TotalPages: output.TotalPages,
		Order:      output.Order,
		Direction:  output.Direction,
	}
}
