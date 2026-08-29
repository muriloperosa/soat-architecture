package peca

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
)

// ConsultarDisponibilidadeUseCase calcula quanto de uma peça está
// disponível pra nova reserva: estoque físico menos a soma já reservada
// em Ordens de Serviço (docs/database/DATABASE.md).
type ConsultarDisponibilidadeUseCase struct {
	repository        domain.Repository
	reservaRepository domainreservapeca.Repository
}

func NewConsultarDisponibilidadeUseCase(repository domain.Repository, reservaRepository domainreservapeca.Repository) *ConsultarDisponibilidadeUseCase {
	return &ConsultarDisponibilidadeUseCase{repository: repository, reservaRepository: reservaRepository}
}

func (uc *ConsultarDisponibilidadeUseCase) Executar(ctx context.Context, pecaID uint64) (DisponibilidadePecaOutput, error) {
	peca, err := uc.repository.BuscarPorID(ctx, pecaID)
	if err != nil {
		return DisponibilidadePecaOutput{}, err
	}

	reservada, err := uc.reservaRepository.SomarQuantidadeReservada(ctx, pecaID)
	if err != nil {
		return DisponibilidadePecaOutput{}, err
	}

	return DisponibilidadePecaOutput{
		PecaID:               peca.ID(),
		QuantidadeEmEstoque:  peca.QuantidadeEmEstoque(),
		QuantidadeReservada:  reservada,
		QuantidadeDisponivel: peca.QuantidadeEmEstoque() - reservada,
	}, nil
}
