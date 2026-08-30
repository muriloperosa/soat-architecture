package cliente

import (
	"context"

	"github.com/muriloperosa/soat-architecture/internal/domain/query"
)

type ClienteRepository interface {
	Salvar(ctx context.Context, cliente *Cliente) error
	BuscarPorID(ctx context.Context, id uint64) (*Cliente, error)
	BuscarPorDocumento(ctx context.Context, documento string) (*Cliente, error)
	BuscarPorEmail(ctx context.Context, email string) (*Cliente, error)
	Listar(ctx context.Context, params query.Params) (query.Page[*Cliente], error)
	Atualizar(ctx context.Context, cliente *Cliente) error
}
