package ordemservico

import (
	"context"
	"errors"

	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type ListarOrdensServicoUseCase struct {
	repository          domain.OrdemServicoRepository
	orcamentoRepository domainorcamento.OrcamentoRepository
}

func NewListarOrdensServicoUseCase(
	repository domain.OrdemServicoRepository,
	orcamentoRepository domainorcamento.OrcamentoRepository,
) *ListarOrdensServicoUseCase {
	return &ListarOrdensServicoUseCase{
		repository:          repository,
		orcamentoRepository: orcamentoRepository,
	}
}

func (uc *ListarOrdensServicoUseCase) Executar(
	ctx context.Context,
	input ListarOrdensServicoInput,
) (ListarOrdensServicoOutput, error) {
	params := appquery.ToDomainParams(input.ParamsInput)

	if input.TipoSolicitante == domainauth.TipoCliente {
		if input.SolicitanteID == 0 {
			return ListarOrdensServicoOutput{}, shared.NewUnauthorizedError("requisição não autorizada")
		}

		params.Filters = append(params.Filters, domainquery.Filter{
			Field:    "cliente_id",
			Operator: domainquery.OperatorEqual,
			Value:    formatUint(input.SolicitanteID),
		})
	}

	page, err := uc.repository.Listar(ctx, params)
	if err != nil {
		return ListarOrdensServicoOutput{}, mapListarErro("erro ao listar ordens de serviço", err)
	}

	orcamentosPorOS, err := uc.buscarOrcamentosDaPagina(ctx, page.Items)
	if err != nil {
		return ListarOrdensServicoOutput{}, mapListarErro("erro ao consultar orçamentos das ordens de serviço", err)
	}

	items := make([]OrdemServicoResumoOutput, 0, len(page.Items))
	for _, entity := range page.Items {
		item := toResumoOutput(entity)
		if orcamento, ok := orcamentosPorOS[entity.ID()]; ok {
			item.Orcamento = toOrcamentoResumoOutput(orcamento)
		}
		items = append(items, item)
	}

	return ListarOrdensServicoOutput{
		Items:      items,
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
		Order:      page.Order,
		Direction:  string(page.Direction),
	}, nil
}

func (uc *ListarOrdensServicoUseCase) buscarOrcamentosDaPagina(
	ctx context.Context,
	ordens []*domain.OrdemServico,
) (map[uint64]*domainorcamento.Orcamento, error) {
	resultado := make(map[uint64]*domainorcamento.Orcamento)
	if len(ordens) == 0 {
		return resultado, nil
	}

	ids := make([]uint64, 0, len(ordens))
	for _, ordem := range ordens {
		ids = append(ids, ordem.ID())
	}

	orcamentos, err := uc.orcamentoRepository.BuscarPorOrdensServicoIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	for _, orcamento := range orcamentos {
		resultado[orcamento.OrdemServicoID()] = orcamento
	}

	return resultado, nil
}

func mapListarErro(mensagem string, err error) error {
	if _, ok := errors.AsType[*shared.AppError](err); ok {
		return err
	}

	return shared.NewInternalError(mensagem, err)
}
