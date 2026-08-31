package peca

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
)

type ConsultarPecaPorIDUseCase struct {
	repository        domain.Repository
	reservaRepository domainreservapeca.Repository
}

func NewConsultarPecaPorIDUseCase(repository domain.Repository, reservaRepository domainreservapeca.Repository) *ConsultarPecaPorIDUseCase {
	return &ConsultarPecaPorIDUseCase{repository: repository, reservaRepository: reservaRepository}
}

func (uc *ConsultarPecaPorIDUseCase) Executar(ctx context.Context, id uint64) (PecaOutput, error) {
	peca, err := uc.repository.BuscarPorID(ctx, id)
	if err != nil {
		return PecaOutput{}, err
	}

	reservada, err := uc.reservaRepository.SomarQuantidadeReservada(ctx, id)
	if err != nil {
		return PecaOutput{}, err
	}

	return toOutputComReserva(peca, reservada), nil
}
