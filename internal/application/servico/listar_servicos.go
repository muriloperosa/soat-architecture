package servico

import (
	"context"
	"errors"

	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type ListarServicosUseCase struct {
	repository domain.ServicoRepository
}

func NewListarServicosUseCase(repository domain.ServicoRepository) *ListarServicosUseCase {
	return &ListarServicosUseCase{repository: repository}
}

func (uc *ListarServicosUseCase) Executar(
	ctx context.Context,
	input ListarServicosInput,
) (ListarServicosOutput, error) {
	params := appquery.ToDomainParams(input.ParamsInput)

	page, err := uc.repository.Listar(ctx, params)
	if err != nil {
		var appErr *shared.AppError
		if errors.As(err, &appErr) {
			return ListarServicosOutput{}, err
		}
		return ListarServicosOutput{}, shared.NewInternalError("erro ao listar serviços", err)
	}

	items := toOutputList(page.Items)

	return ListarServicosOutput{
		Items:      items,
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
		Order:      page.Order,
		Direction:  string(page.Direction),
	}, nil
}
