package veiculo

import (
	"context"
	"errors"
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
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
	input ListarVeiculosInput,
) (ListarVeiculosOutput, error) {
	params := appquery.ToDomainParams(input.ParamsInput)

	page, err := uc.repository.Listar(ctx, params)
	if err != nil {
		var appErr *shared.AppError
		if errors.As(err, &appErr) {
			return ListarVeiculosOutput{}, err
		}
		return ListarVeiculosOutput{}, shared.NewInternalError("erro ao listar veículos", err)
	}

	items := make([]VeiculoOutput, 0, len(page.Items))
	for _, entity := range page.Items {
		items = append(items, toOutput(entity))
	}

	return ListarVeiculosOutput{
		Items:      items,
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
		Order:      page.Order,
		Direction:  string(page.Direction),
	}, nil
}
