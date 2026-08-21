package peca

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
)

// ReporEstoqueUseCase adiciona quantidade ao estoque físico de uma peça
// (ex. entrada de compra/fornecedor).
type ReporEstoqueUseCase struct {
	repository domain.Repository
}

func NewReporEstoqueUseCase(repository domain.Repository) *ReporEstoqueUseCase {
	return &ReporEstoqueUseCase{repository: repository}
}

func (uc *ReporEstoqueUseCase) Executar(ctx context.Context, input ReporEstoqueInput) (PecaOutput, error) {
	peca, err := uc.repository.BuscarPorID(ctx, input.PecaID)
	if err != nil {
		return PecaOutput{}, err
	}

	if err := peca.Repor(input.Quantidade); err != nil {
		return PecaOutput{}, err
	}

	if err := uc.repository.Atualizar(ctx, peca); err != nil {
		return PecaOutput{}, err
	}

	return toOutput(peca), nil
}
