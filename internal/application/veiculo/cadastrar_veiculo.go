package veiculo

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
)

type CadastrarVeiculoUseCase struct {
	repository domain.Repository
}

func NewCadastrarVeiculoUseCase(repository domain.Repository) *CadastrarVeiculoUseCase {
	return &CadastrarVeiculoUseCase{repository: repository}
}

func (uc *CadastrarVeiculoUseCase) Executar(ctx context.Context, input CadastrarVeiculoInput) (VeiculoOutput, error) {
	placa, err := domain.NewPlaca(input.Placa)
	if err != nil {
		return VeiculoOutput{}, err
	}

	existente, err := uc.repository.BuscarPorPlaca(ctx, placa)
	if err != nil && !errors.Is(err, domain.ErrVeiculoNaoEncontrado) {
		return VeiculoOutput{}, err
	}
	if existente != nil {
		return VeiculoOutput{}, domain.ErrPlacaJaCadastrada
	}

	veiculo, err := domain.NewVeiculo(
		input.Placa,
		input.Marca,
		input.Modelo,
		input.QuilometragemAtual,
		input.Ano,
		input.Cor,
		input.CriadoPor,
	)
	if err != nil {
		return VeiculoOutput{}, err
	}

	if err := uc.repository.Salvar(ctx, veiculo); err != nil {
		return VeiculoOutput{}, err
	}

	return toOutput(veiculo), nil
}
