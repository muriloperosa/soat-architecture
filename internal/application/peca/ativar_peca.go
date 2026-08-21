package peca

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
)

type AtivarPecaUseCase struct {
	repository domain.Repository
}

func NewAtivarPecaUseCase(repository domain.Repository) *AtivarPecaUseCase {
	return &AtivarPecaUseCase{repository: repository}
}

func (uc *AtivarPecaUseCase) Executar(ctx context.Context, id uint64) (PecaOutput, error) {
	peca, err := uc.repository.BuscarPorID(ctx, id)
	if err != nil {
		return PecaOutput{}, err
	}

	peca.Ativar()

	if err := uc.repository.Atualizar(ctx, peca); err != nil {
		return PecaOutput{}, err
	}

	return toOutput(peca), nil
}
