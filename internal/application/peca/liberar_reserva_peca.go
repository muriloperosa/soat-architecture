package peca

import (
	"context"

	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
)

// LiberarReservaPecaUseCase libera quantidade de uma reserva existente
// (ex. serviço cancelado, peça removida da OS antes do consumo). Se a
// reserva chegar a zero, o registro é removido — reservas_pecas não
// mantém linhas com quantidade zero (CHECK da migration).
type LiberarReservaPecaUseCase struct {
	reservaRepository domainreservapeca.Repository
}

func NewLiberarReservaPecaUseCase(reservaRepository domainreservapeca.Repository) *LiberarReservaPecaUseCase {
	return &LiberarReservaPecaUseCase{reservaRepository: reservaRepository}
}

func (uc *LiberarReservaPecaUseCase) Executar(ctx context.Context, input LiberarReservaPecaInput) (ReservaPecaOutput, error) {
	reserva, err := uc.reservaRepository.BuscarPorOrdemEPeca(ctx, input.OrdemServicoID, input.PecaID)
	if err != nil {
		return ReservaPecaOutput{}, err
	}

	if err := reserva.Reduzir(input.Quantidade); err != nil {
		return ReservaPecaOutput{}, err
	}

	if reserva.Quantidade() == 0 {
		if err := uc.reservaRepository.Remover(ctx, input.OrdemServicoID, input.PecaID); err != nil {
			return ReservaPecaOutput{}, err
		}

		return toReservaOutput(reserva), nil
	}

	if err := uc.reservaRepository.Atualizar(ctx, reserva); err != nil {
		return ReservaPecaOutput{}, err
	}

	return toReservaOutput(reserva), nil
}
