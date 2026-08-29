package servico

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// ListarServicosUseCase devolve os serviços aplicando paginação,
// ordenação e filtros suportados pelo repositório.
type ListarServicosUseCase struct {
	repo domainservico.ServicoRepository
}

func NewListarServicosUseCase(
	repo domainservico.ServicoRepository,
) *ListarServicosUseCase {
	return &ListarServicosUseCase{
		repo: repo,
	}
}

func (uc *ListarServicosUseCase) Executar(
	ctx context.Context,
	params query.Params,
) (query.Page[ServicoOutput], error) {
	page, err := uc.repo.Listar(ctx, params)
	if err != nil {
		var appErr *shared.AppError

		if errors.As(err, &appErr) {
			return query.Page[ServicoOutput]{}, err
		}

		return query.Page[ServicoOutput]{},
			shared.NewInternalError("erro ao listar serviços", err)
	}

	items := make([]ServicoOutput, 0, len(page.Items))

	for _, servico := range page.Items {
		items = append(items, toOutput(servico))
	}

	return query.Page[ServicoOutput]{
		Items:     items,
		Total:     page.Total,
		Offset:    page.Offset,
		Limit:     page.Limit,
		Order:     page.Order,
		Direction: page.Direction,
	}, nil
}