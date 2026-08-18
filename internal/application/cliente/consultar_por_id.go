package cliente

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

type ConsultarClientePorIDUseCase struct {
	repository domain.Repository
}

func NewConsultarClientePorIDUseCase(repository domain.Repository) *ConsultarClientePorIDUseCase {
	return &ConsultarClientePorIDUseCase{repository: repository}
}

func (uc *ConsultarClientePorIDUseCase) Executar(
	ctx context.Context,
	id uint64,
) (ClienteOutput, error) {
	cliente, err := uc.repository.BuscarPorID(ctx, id)
	if err != nil {
		return ClienteOutput{}, err
	}

	return toOutput(cliente), nil
}
