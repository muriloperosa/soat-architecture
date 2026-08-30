package servico

import (
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
)

// CriarServicoInput é o DTO de entrada do CriarServicoUseCase.
type CriarServicoInput struct {
	Nome                 string
	Descricao            string
	PrecoBase            float64
	TempoEstimadoMinutos int
	CriadoPor            uint64
}

// AtualizarServicoInput é o DTO de entrada do AtualizarServicoUseCase.
type AtualizarServicoInput struct {
	ID                   uint64
	Nome                 string
	Descricao            string
	PrecoBase            float64
	TempoEstimadoMinutos int
}

// ServicoOutput é o DTO de saída comum aos use cases de catálogo de serviço.
type ServicoOutput struct {
	ID                   uint64
	Nome                 string
	Descricao            string
	PrecoBase            float64
	TempoEstimadoMinutos int
	CriadoPor            uint64
	Ativo                bool
}

func toOutput(s *domainservico.Servico) ServicoOutput {
	return ServicoOutput{
		ID:                   s.ID(),
		Nome:                 s.Nome(),
		Descricao:            s.Descricao(),
		PrecoBase:            s.PrecoBase(),
		TempoEstimadoMinutos: s.TempoEstimado().Minutos(),
		CriadoPor:            s.CriadoPor(),
		Ativo:                s.Ativo(),
	}
}

func toOutputList(servicos []*domainservico.Servico) []ServicoOutput {
	out := make([]ServicoOutput, 0, len(servicos))
	for _, s := range servicos {
		out = append(out, toOutput(s))
	}
	return out
}

// ListarServicosInput é o contrato de entrada do caso de uso de listagem.
type ListarServicosInput struct {
	appquery.ParamsInput
}

// ListarServicosOutput é o contrato de saída do caso de uso de listagem.
type ListarServicosOutput struct {
	Items      []ServicoOutput
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
	Order      string
	Direction  string
}
