package veiculo

import (
	"context"

	"github.com/muriloperosa/soat-architecture/internal/domain/query"
)

type Repository interface {
	Salvar(ctx context.Context, veiculo *Veiculo) error
	BuscarPorID(ctx context.Context, id uint64) (*Veiculo, error)
	BuscarPorPlaca(ctx context.Context, placa Placa) (*Veiculo, error)
	Atualizar(ctx context.Context, veiculo *Veiculo) error
	Listar(ctx context.Context, params query.Params) (query.Page[*Veiculo], error)
}
