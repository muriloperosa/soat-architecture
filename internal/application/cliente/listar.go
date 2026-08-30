package cliente

import (
	"context"
	"errors"

	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type ListarClientesUseCase struct {
	repository domain.ClienteRepository
}

func NewListarClientesUseCase(repository domain.ClienteRepository) *ListarClientesUseCase {
	return &ListarClientesUseCase{
		repository: repository,
	}
}

func (uc *ListarClientesUseCase) Executar(
	ctx context.Context,
	input ListarClientesInput,
) (ListarClientesOutput, error) {
	params := appquery.ToDomainParams(input.ParamsInput)

	page, err := uc.repository.Listar(ctx, params)
	if err != nil {
		var appErr *shared.AppError
		if errors.As(err, &appErr) {
			return ListarClientesOutput{}, err
		}

		return ListarClientesOutput{},
			shared.NewInternalError("erro ao listar clientes", err)
	}

	items := make([]ClienteOutput, 0, len(page.Items))

	for _, entity := range page.Items {
		items = append(items, toOutput(entity))
	}

	return ListarClientesOutput{
		Items:      items,
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
		Order:      page.Order,
		Direction:  string(page.Direction),
	}, nil
}
