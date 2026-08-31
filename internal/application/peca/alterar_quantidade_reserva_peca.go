package peca

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// AlterarQuantidadeReservaPecaUseCase define a quantidade total reservada por
// uma OS para uma peça. A peça é bloqueada com FOR UPDATE, a soma das reservas
// é recalculada dentro da mesma transação e a própria reserva atual é removida
// da soma antes de validar a nova quantidade.
type AlterarQuantidadeReservaPecaUseCase struct {
	repository        domain.Repository
	reservaRepository domainreservapeca.Repository
	transactionRunner shared.TransactionRunner
}

func NewAlterarQuantidadeReservaPecaUseCase(
	repository domain.Repository,
	reservaRepository domainreservapeca.Repository,
	transactionRunner shared.TransactionRunner,
) *AlterarQuantidadeReservaPecaUseCase {
	return &AlterarQuantidadeReservaPecaUseCase{
		repository:        repository,
		reservaRepository: reservaRepository,
		transactionRunner: transactionRunner,
	}
}

func (uc *AlterarQuantidadeReservaPecaUseCase) Executar(
	ctx context.Context,
	input AlterarQuantidadeReservaPecaInput,
) (ReservaPecaOutput, error) {
	if input.Quantidade <= 0 {
		return ReservaPecaOutput{}, domainreservapeca.ErrQuantidadeInvalida
	}

	var output ReservaPecaOutput

	err := uc.transactionRunner.Executar(ctx, func(ctx context.Context) error {
		peca, err := uc.repository.BuscarPorIDComBloqueio(ctx, input.PecaID)
		if err != nil {
			return err
		}

		reserva, err := uc.reservaRepository.BuscarPorOrdemEPecaComBloqueio(ctx, input.OrdemServicoID, input.PecaID)
		if err != nil {
			return err
		}

		totalReservado, err := uc.reservaRepository.SomarQuantidadeReservada(ctx, input.PecaID)
		if err != nil {
			return err
		}

		reservadoPorOutrasOS := totalReservado - reserva.Quantidade()
		if reservadoPorOutrasOS < 0 {
			reservadoPorOutrasOS = 0
		}

		if !peca.PodeReservar(reservadoPorOutrasOS, input.Quantidade) {
			return domain.ErrQuantidadeIndisponivelParaReserva
		}

		if err := reserva.AlterarQuantidade(input.Quantidade); err != nil {
			return err
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
