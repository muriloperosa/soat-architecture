package peca

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
)

type ConsultarPecaPorIDUseCase struct {
	repository domain.Repository
}

func NewConsultarPecaPorIDUseCase(repository domain.Repository) *ConsultarPecaPorIDUseCase {
	return &ConsultarPecaPorIDUseCase{repository: repository}
}

func (uc *ConsultarPecaPorIDUseCase) Executar(ctx context.Context, id uint64) (PecaOutput, error) {
	peca, err := uc.repository.BuscarPorID(ctx, id)
	if err != nil {
		return PecaOutput{}, err
	}

	return toOutput(peca), nil
}
