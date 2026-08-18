package cliente

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

type AtualizarClienteUseCase struct {
	repository domain.Repository
}

func NewAtualizarClienteUseCase(repository domain.Repository) *AtualizarClienteUseCase {
	return &AtualizarClienteUseCase{repository: repository}
}

func (uc *AtualizarClienteUseCase) Executar(
	ctx context.Context,
	input AtualizarClienteInput,
) (ClienteOutput, error) {
	cliente, err := uc.repository.BuscarPorID(ctx,input.ID)
	if err != nil {
		return ClienteOutput{}, err
	}

	if err := cliente.Atualizar(input.Nome,input.Email,input.Telefone); err != nil {
		return ClienteOutput{}, err
	}

	if err := uc.repository.Atualizar(ctx,cliente); err != nil {
		return ClienteOutput{}, err
	}

	return toOutput(cliente), nil
}
