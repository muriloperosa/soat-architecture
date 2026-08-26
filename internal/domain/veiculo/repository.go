package veiculo

import "context"

type Repository interface {
	Salvar(ctx context.Context, veiculo *Veiculo) error
	BuscarPorID(ctx context.Context, id uint64) (*Veiculo, error)
	BuscarPorPlaca(ctx context.Context, placa Placa) (*Veiculo, error)
	Atualizar(ctx context.Context, veiculo *Veiculo) error
}
