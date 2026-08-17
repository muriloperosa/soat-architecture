package cliente

import "context"

type Repository interface {
	Salvar(ctx context.Context, cliente Cliente) error
	BuscarPorID(ctx context.Context, id uint64) (Cliente, error)
	BuscarPorDocumento(ctx context.Context, documento Documento) (Cliente, error)
	BuscarPorEmail(ctx context.Context, email string) (Cliente, error)
	Atualizar(ctx context.Context, cliente Cliente) error
}
