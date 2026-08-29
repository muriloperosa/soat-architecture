package veiculo

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
)

type ListarVeiculosUseCase struct {
	repository domain.Repository
}

func NewListarVeiculosUseCase(repository domain.Repository) *ListarVeiculosUseCase {
	return &ListarVeiculosUseCase{repository: repository}
}

func (uc *ListarVeiculosUseCase) Executar(
	ctx context.Context,
	params query.Params,
) (query.Page[VeiculoOutput], error) {
	page, err := uc.repository.Listar(ctx, params)
	if err != nil {
		var appErr *shared.AppError

		if errors.As(err, &appErr) {
			return query.Page[VeiculoOutput]{}, err
		}

		return query.Page[VeiculoOutput]{},
			shared.NewInternalError("erro ao listar veículos", err)
	}

	items := make([]VeiculoOutput, 0, len(page.Items))

	for _, veiculo := range page.Items {
		items = append(items, toOutput(veiculo))
	}

	return query.Page[VeiculoOutput]{
		Items:     items,
		Total:     page.Total,
		Offset:    page.Offset,
		Limit:     page.Limit,
		Order:     page.Order,
		Direction: page.Direction,
	}, nil
}
