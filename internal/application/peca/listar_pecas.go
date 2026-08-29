package peca

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type ListarPecasUseCase struct {
	repository domain.Repository
}

func NewListarPecasUseCase(repository domain.Repository) *ListarPecasUseCase {
	return &ListarPecasUseCase{
		repository: repository,
	}
}

func (uc *ListarPecasUseCase) Executar(
	ctx context.Context,
	params query.Params,
) (query.Page[PecaOutput], error) {
	page, err := uc.repository.Listar(ctx, params)
	if err != nil {
		var appErr *shared.AppError

		if errors.As(err, &appErr) {
			return query.Page[PecaOutput]{}, err
		}

		return query.Page[PecaOutput]{}, shared.NewInternalError("erro ao listar peças", err)
	}

	items := make([]PecaOutput, 0, len(page.Items))

	for _, peca := range page.Items {
		items = append(items, toOutput(peca))
	}

	return query.Page[PecaOutput]{
		Items:     items,
		Total:     page.Total,
		Offset:    page.Offset,
		Limit:     page.Limit,
		Order:     page.Order,
		Direction: page.Direction,
	}, nil
}
