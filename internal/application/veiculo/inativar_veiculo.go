package veiculo

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
)

type InativarVeiculoUseCase struct {
	repository domain.Repository
}

func NewInativarVeiculoUseCase(repository domain.Repository) *InativarVeiculoUseCase {
	return &InativarVeiculoUseCase{repository: repository}
}

func (uc *InativarVeiculoUseCase) Executar(ctx context.Context, id uint64) (VeiculoOutput, error) {
	veiculo, err := uc.repository.BuscarPorID(ctx, id)
	if err != nil {
		return VeiculoOutput{}, err
	}

	veiculo.Inativar()

	if err := uc.repository.Atualizar(ctx, veiculo); err != nil {
		return VeiculoOutput{}, err
	}

	return toOutput(veiculo), nil
}
