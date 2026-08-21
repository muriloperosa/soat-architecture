package peca

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
)

type InativarPecaUseCase struct {
	repository domain.Repository
}

func NewInativarPecaUseCase(repository domain.Repository) *InativarPecaUseCase {
	return &InativarPecaUseCase{repository: repository}
}

func (uc *InativarPecaUseCase) Executar(ctx context.Context, id uint64) (PecaOutput, error) {
	peca, err := uc.repository.BuscarPorID(ctx, id)
	if err != nil {
		return PecaOutput{}, err
	}

	peca.Inativar()

	if err := uc.repository.Atualizar(ctx, peca); err != nil {
		return PecaOutput{}, err
	}

	return toOutput(peca), nil
}
