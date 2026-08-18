package cliente

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

type CriarClienteUseCase struct {
	repository domain.Repository
}

func NewCriarClienteUseCase(repository domain.Repository) *CriarClienteUseCase {
	return &CriarClienteUseCase{repository: repository}
}

func (uc *CriarClienteUseCase) Executar(
	ctx context.Context,
	input CriarClienteInput,
) (ClienteOutput, error) {
	existente, err := uc.repository.BuscarPorDocumento(ctx, input.Documento)
	if err != nil && !errors.Is(err, domain.ErrClienteNaoEncontrado) {
		return ClienteOutput{}, err
	}

	if existente != nil {
		return ClienteOutput{}, domain.ErrClienteJaCadastrado
	}

	cliente, err := domain.NewCliente(
		input.Documento,
		input.Tipo,
		input.Nome,
		input.Email,
		input.Telefone,
		input.Senha,
	)
	if err != nil {
		return ClienteOutput{}, err
	}

	if err := uc.repository.Salvar(ctx, &cliente); err != nil {
		return ClienteOutput{}, err
	}

	return toOutput(&cliente), nil
}
