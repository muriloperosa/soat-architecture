package cliente

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

type AlterarSenhaClienteUseCase struct {
	repository domain.ClienteRepository
}

func NewAlterarSenhaClienteUseCase(repository domain.ClienteRepository) *AlterarSenhaClienteUseCase {
	return &AlterarSenhaClienteUseCase{repository: repository}
}

func (uc *AlterarSenhaClienteUseCase) Executar(
	ctx context.Context,
	input AlterarSenhaInput,
) (ClienteOutput, error) {
	cliente, err := uc.repository.BuscarPorID(ctx, input.ClienteID)
	if err != nil {
		return ClienteOutput{}, err
	}

	if err := cliente.AlterarSenha(input.SenhaNova); err != nil {
		return ClienteOutput{}, err
	}

	if err := uc.repository.Atualizar(ctx, cliente); err != nil {
		return ClienteOutput{}, err
	}

	return toOutput(cliente), nil
}
