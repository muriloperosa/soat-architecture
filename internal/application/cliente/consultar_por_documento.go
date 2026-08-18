package cliente

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

type ConsultarPorDocumento struct {
	repository domain.Repository
}

func NewConsultarPorDocumento(repository domain.Repository) *ConsultarPorDocumento {
	return &ConsultarPorDocumento{repository: repository}
}

func (uc *ConsultarPorDocumento) Executar(
	ctx context.Context,
	documento string,
) (*domain.Cliente, error) {
	return uc.repository.BuscarPorDocumento(ctx,documento)
}
