package cliente

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

type ConsultarPorID struct {
	repository domain.Repository
}

func NewConsultarPorID(repository domain.Repository) *ConsultarPorID {
	return &ConsultarPorID{repository: repository}
}

func (uc *ConsultarPorID) Executar(ctx context.Context, id uint64) (*domain.Cliente, error) {
	return uc.repository.BuscarPorID(ctx, id)
}
