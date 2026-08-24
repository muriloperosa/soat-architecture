package veiculo

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
)

type AtualizarVeiculoUseCase struct {
	repository domain.Repository
}

func NewAtualizarVeiculoUseCase(repository domain.Repository) *AtualizarVeiculoUseCase {
	return &AtualizarVeiculoUseCase{repository: repository}
}

func (uc *AtualizarVeiculoUseCase) Executar(ctx context.Context, input AtualizarVeiculoInput) (VeiculoOutput, error) {
	veiculo, err := uc.repository.BuscarPorID(ctx, input.ID)
	if err != nil {
		return VeiculoOutput{}, err
	}

	if err := veiculo.Atualizar(input.Marca, input.Modelo, input.Cor); err != nil {
		return VeiculoOutput{}, err
	}

	if err := veiculo.AtualizarQuilometragem(input.QuilometragemAtual); err != nil {
		return VeiculoOutput{}, err
	}

	if err := uc.repository.Atualizar(ctx, veiculo); err != nil {
		return VeiculoOutput{}, err
	}

	return toOutput(veiculo), nil
}
