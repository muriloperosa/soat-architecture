package cliente

import (
	"context"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

// CredenciaisAdapter projeta o agregado Cliente nos contratos de autenticação.
type CredenciaisAdapter struct {
	clientes domaincliente.ClienteRepository
}

func NewCredenciaisAdapter(clientes domaincliente.ClienteRepository) *CredenciaisAdapter {
	return &CredenciaisAdapter{clientes: clientes}
}

func (a *CredenciaisAdapter) BuscarPorEmail(ctx context.Context, email string) (*domainauth.Credencial, error) {
	cliente, err := a.clientes.BuscarPorEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return toCredencial(cliente), nil
}

func (a *CredenciaisAdapter) EstaAtivo(ctx context.Context, id uint64) (bool, error) {
	cliente, err := a.clientes.BuscarPorID(ctx, id)
	if err != nil {
		return false, err
	}
	return cliente.Ativo(), nil
}
