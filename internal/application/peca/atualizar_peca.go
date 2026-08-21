package peca

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
)

type AtualizarPecaUseCase struct {
	repository domain.Repository
}

func NewAtualizarPecaUseCase(repository domain.Repository) *AtualizarPecaUseCase {
	return &AtualizarPecaUseCase{repository: repository}
}

func (uc *AtualizarPecaUseCase) Executar(ctx context.Context, input AtualizarPecaInput) (PecaOutput, error) {
	peca, err := uc.repository.BuscarPorID(ctx, input.ID)
	if err != nil {
		return PecaOutput{}, err
	}

	if err := peca.Atualizar(input.Nome, input.Marca, input.Descricao, input.Preco, input.EstoqueMinimo); err != nil {
		return PecaOutput{}, err
	}

	if err := uc.repository.Atualizar(ctx, peca); err != nil {
		return PecaOutput{}, err
	}

	return toOutput(peca), nil
}
