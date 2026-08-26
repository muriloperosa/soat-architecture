package veiculo

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
)

type ConsultarVeiculoPorPlacaUseCase struct {
	repository domain.Repository
}

func NewConsultarVeiculoPorPlacaUseCase(repository domain.Repository) *ConsultarVeiculoPorPlacaUseCase {
	return &ConsultarVeiculoPorPlacaUseCase{repository: repository}
}

func (uc *ConsultarVeiculoPorPlacaUseCase) Executar(ctx context.Context, placa string) (VeiculoOutput, error) {
	placaVO, err := domain.NewPlaca(placa)
	if err != nil {
		return VeiculoOutput{}, err
	}

	veiculo, err := uc.repository.BuscarPorPlaca(ctx, placaVO)
	if err != nil {
		return VeiculoOutput{}, err
	}

	return toOutput(veiculo), nil
}
