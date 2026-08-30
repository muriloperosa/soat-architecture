package ordemservico

import (
	"context"
	"errors"

	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type ListarOrdensServicoUseCase struct {
	repository domain.OrdemServicoRepository
}

func NewListarOrdensServicoUseCase(repository domain.OrdemServicoRepository) *ListarOrdensServicoUseCase {
	return &ListarOrdensServicoUseCase{repository: repository}
}

func (uc *ListarOrdensServicoUseCase) Executar(
	ctx context.Context,
	input ListarOrdensServicoInput,
) (ListarOrdensServicoOutput, error) {
	params := appquery.ToDomainParams(input.ParamsInput)

	page, err := uc.repository.Listar(ctx, params)
	if err != nil {
		var appErr *shared.AppError
		if errors.As(err, &appErr) {
			return ListarOrdensServicoOutput{}, err
		}

		return ListarOrdensServicoOutput{}, shared.NewInternalError("erro ao listar ordens de serviço", err)
	}

	items := make([]OrdemServicoResumoOutput, 0, len(page.Items))
	for _, entity := range page.Items {
		items = append(items, toResumoOutput(entity))
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
