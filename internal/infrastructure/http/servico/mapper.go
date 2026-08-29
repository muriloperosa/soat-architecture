package servico

import (
	appservico "github.com/muriloperosa/soat-architecture/internal/application/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
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

func toListResponse(page query.Page[appservico.ServicoOutput]) ListarServicosResponse {
	items := make([]ServicoResponse, 0, len(page.Items))

	for _, item := range page.Items {
		items = append(items, toServicoResponse(item))
	}

	return ListarServicosResponse{
		Items:     items,
		Total:     page.Total,
		Offset:    page.Offset,
		Limit:     page.Limit,
		Order:     page.Order,
		Direction: string(page.Direction),
	}
}
