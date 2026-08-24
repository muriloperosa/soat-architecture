package veiculo

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
)

type AtivarVeiculoUseCase struct {
	repository domain.Repository
}

func NewAtivarVeiculoUseCase(repository domain.Repository) *AtivarVeiculoUseCase {
	return &AtivarVeiculoUseCase{repository: repository}
}

func (uc *AtivarVeiculoUseCase) Executar(ctx context.Context, id uint64) (VeiculoOutput, error) {
	veiculo, err := uc.repository.BuscarPorID(ctx, id)
	if err != nil {
		return VeiculoOutput{}, err
	}

	veiculo.Ativar()

	if err := uc.repository.Atualizar(ctx, veiculo); err != nil {
		return VeiculoOutput{}, err
	}

	return toOutput(veiculo), nil
}
