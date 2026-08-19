package cliente

import "context"

type ClienteRepository interface {
	Salvar(ctx context.Context, cliente *Cliente) error
	BuscarPorID(ctx context.Context, id uint64) (*Cliente, error)
	BuscarPorDocumento(ctx context.Context, documento string) (*Cliente, error)
	BuscarPorEmail(ctx context.Context, email string) (*Cliente, error)
	Atualizar(ctx context.Context, cliente *Cliente) error
}

// Repository é um alias de compatibilidade. Prefira ClienteRepository.
type Repository = ClienteRepository
