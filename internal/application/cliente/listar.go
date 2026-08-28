package cliente

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type ListarClientesUseCase struct {
	repository domain.ClienteRepository
}

func NewListarClientesUseCase(repository domain.ClienteRepository) *ListarClientesUseCase {
	return &ListarClientesUseCase{repository: repository}
}

func (uc *ListarClientesUseCase) Executar(
	ctx context.Context,
	params query.Params,
) (query.Page[ClienteOutput], error) {
	page, err := uc.repository.Listar(ctx, params)
	if err != nil {
		var appErr *shared.AppError
		if errors.As(err, &appErr) {
			return query.Page[ClienteOutput]{}, err
		}
		return query.Page[ClienteOutput]{}, shared.NewInternalError("erro ao listar clientes", err)
	}

	items := make([]ClienteOutput, 0, len(page.Items))
	for _, cliente := range page.Items {
		items = append(items, toOutput(cliente))
	}

	return query.Page[ClienteOutput]{
		Items:     items,
		Total:     page.Total,
		Offset:    page.Offset,
		Limit:     page.Limit,
		Order:     page.Order,
		Direction: page.Direction,
	}, nil
}
