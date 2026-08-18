package cliente

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

type InativarClienteUseCase struct {
	repository domain.Repository
}

func NewInativarClienteUseCase(repository domain.Repository) *InativarClienteUseCase {
	return &InativarClienteUseCase{repository: repository}
}

func (uc *InativarClienteUseCase) Executar(ctx context.Context,id uint64) (ClienteOutput, error) {
	cliente, err := uc.repository.BuscarPorID(ctx, id)
	if err != nil {
		return ClienteOutput{}, err
	}

	cliente.Inativar()

	if err := uc.repository.Atualizar(ctx, cliente); err != nil {
		return ClienteOutput{}, err
	}

	return toOutput(cliente), nil
}
