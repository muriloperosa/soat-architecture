package veiculo

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
)

type ConsultarVeiculoPorIDUseCase struct {
	repository domain.Repository
}

func NewConsultarVeiculoPorIDUseCase(repository domain.Repository) *ConsultarVeiculoPorIDUseCase {
	return &ConsultarVeiculoPorIDUseCase{repository: repository}
}

func (uc *ConsultarVeiculoPorIDUseCase) Executar(ctx context.Context, id uint64) (VeiculoOutput, error) {
	veiculo, err := uc.repository.BuscarPorID(ctx, id)
	if err != nil {
		return VeiculoOutput{}, err
	}

	return toOutput(veiculo), nil
}
