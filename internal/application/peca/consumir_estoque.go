package peca

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
)

// ConsumirEstoqueUseCase baixa o estoque físico de uma peça (ex. aplicada
// numa Ordem de Serviço). Não cuida de reserva — isso é responsabilidade do
// agregado ReservaPeca, ainda não implementado.
type ConsumirEstoqueUseCase struct {
	repository domain.Repository
}

func NewConsumirEstoqueUseCase(repository domain.Repository) *ConsumirEstoqueUseCase {
	return &ConsumirEstoqueUseCase{repository: repository}
}

func (uc *ConsumirEstoqueUseCase) Executar(ctx context.Context, input ConsumirEstoqueInput) (PecaOutput, error) {
	peca, err := uc.repository.BuscarPorID(ctx, input.PecaID)
	if err != nil {
		return PecaOutput{}, err
	}

	if err := peca.Consumir(input.Quantidade); err != nil {
		return PecaOutput{}, err
	}

	if err := uc.repository.Atualizar(ctx, peca); err != nil {
		return PecaOutput{}, err
	}

	return toOutput(peca), nil
}
