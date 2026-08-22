package peca

import (
	"context"

	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// LiberarReservaPecaUseCase libera quantidade de uma reserva existente
// (ex. serviço cancelado, peça removida da OS antes do consumo). Se a
// reserva chegar a zero, o registro é removido — reservas_pecas não
// mantém linhas com quantidade zero (CHECK da migration). Roda dentro de
// uma transação com a reserva travada, pra liberações concorrentes na
// mesma reserva não se sobrescreverem (lost update).
type LiberarReservaPecaUseCase struct {
	reservaRepository domainreservapeca.Repository
	transactionRunner shared.TransactionRunner
}

func NewLiberarReservaPecaUseCase(
	reservaRepository domainreservapeca.Repository,
	transactionRunner shared.TransactionRunner,
) *LiberarReservaPecaUseCase {
	return &LiberarReservaPecaUseCase{
		reservaRepository: reservaRepository,
		transactionRunner: transactionRunner,
	}
}

func (uc *LiberarReservaPecaUseCase) Executar(ctx context.Context, input LiberarReservaPecaInput) (ReservaPecaOutput, error) {
	if input.Quantidade <= 0 {
		return ReservaPecaOutput{}, domainreservapeca.ErrQuantidadeInvalida
	}

	var output ReservaPecaOutput

	err := uc.transactionRunner.Executar(ctx, func(ctx context.Context) error {
		reserva, err := uc.reservaRepository.BuscarPorOrdemEPecaComBloqueio(ctx, input.OrdemServicoID, input.PecaID)
		if err != nil {
			return err
		}

		if err := reserva.Reduzir(input.Quantidade); err != nil {
			return err
		}

		if reserva.Quantidade() == 0 {
			if err := uc.reservaRepository.Remover(ctx, input.OrdemServicoID, input.PecaID); err != nil {
				return err
			}

			output = toReservaOutput(reserva)

			return nil
		}

		if err := uc.reservaRepository.Atualizar(ctx, reserva); err != nil {
			return err
		}

		output = toReservaOutput(reserva)

		return nil
	})
	if err != nil {
		return ReservaPecaOutput{}, err
	}

	return output, nil
}
