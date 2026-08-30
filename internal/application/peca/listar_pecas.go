package peca

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
)

type ListarPecasUseCase struct {
	repository domain.Repository
}

func NewListarPecasUseCase(repository domain.Repository) *ListarPecasUseCase {
	return &ListarPecasUseCase{repository: repository}
}

func (uc *ListarPecasUseCase) Executar(
	ctx context.Context,
	input ListarPecasInput,
) (ListarPecasOutput, error) {
	params := appquery.ToDomainParams(input.ParamsInput)

	page, err := uc.repository.Listar(ctx, params)
	if err != nil {
		var appErr *shared.AppError
		if errors.As(err, &appErr) {
			return ListarPecasOutput{}, err
		}
		return ListarPecasOutput{}, shared.NewInternalError("erro ao listar peças", err)
	}

	items := make([]PecaOutput, 0, len(page.Items))
	for _, entity := range page.Items {
		items = append(items, toOutput(entity))
	}

	return ListarPecasOutput{
		Items:      items,
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
		Order:      page.Order,
		Direction:  string(page.Direction),
	}, nil
}
