package cliente

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

type AtivarClienteUseCase struct {
	repository domain.Repository
}

func NewAtivarClienteUseCase(repository domain.Repository) *AtivarClienteUseCase {
	return &AtivarClienteUseCase{repository: repository}
}

func (uc *AtivarClienteUseCase) Executar(ctx context.Context, id uint64) (ClienteOutput, error) {
	cliente, err := uc.repository.BuscarPorID(ctx, id)
	if err != nil {
		return ClienteOutput{}, err
	}

	cliente.Ativar()

	if err := uc.repository.Atualizar(ctx, cliente); err != nil {
		return ClienteOutput{}, err
	}

	return toOutput(cliente), nil
}
