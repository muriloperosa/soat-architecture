package peca

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// ReservarPecaUseCase reserva quantidade de uma peça pra uma Ordem de
// Serviço. Roda dentro de uma transação com a peça travada (BuscarPorID-
// ComBloqueio), pra reservas concorrentes não conseguirem, juntas,
// ultrapassar o estoque disponível.
type ReservarPecaUseCase struct {
	repository        domain.Repository
	reservaRepository domainreservapeca.Repository
	transactionRunner shared.TransactionRunner
}

func NewReservarPecaUseCase(
	repository domain.Repository,
	reservaRepository domainreservapeca.Repository,
	transactionRunner shared.TransactionRunner,
) *ReservarPecaUseCase {
	return &ReservarPecaUseCase{
		repository:        repository,
		reservaRepository: reservaRepository,
		transactionRunner: transactionRunner,
	}
}

func (uc *ReservarPecaUseCase) Executar(ctx context.Context, input ReservarPecaInput) (ReservaPecaOutput, error) {
	if input.Quantidade <= 0 {
		return ReservaPecaOutput{}, domainreservapeca.ErrQuantidadeInvalida
	}

	var output ReservaPecaOutput

	err := uc.transactionRunner.Executar(ctx, func(ctx context.Context) error {
		peca, err := uc.repository.BuscarPorIDComBloqueio(ctx, input.PecaID)
		if err != nil {
			return err
		}

		reservadaAtual, err := uc.reservaRepository.SomarQuantidadeReservada(ctx, input.PecaID)
		if err != nil {
			return err
		}

		if !peca.PodeReservar(reservadaAtual, input.Quantidade) {
			return domain.ErrQuantidadeIndisponivelParaReserva
		}

		existente, err := uc.reservaRepository.BuscarPorOrdemEPecaComBloqueio(ctx, input.OrdemServicoID, input.PecaID)
		if err != nil && !errors.Is(err, domainreservapeca.ErrReservaNaoEncontrada) {
			return err
		}

		if existente != nil {
			if err := existente.Aumentar(input.Quantidade); err != nil {
				return err
			}

			if err := uc.reservaRepository.Atualizar(ctx, existente); err != nil {
				return err
			}

			output = toReservaOutput(existente)

			return nil
		}

		nova, err := domainreservapeca.NewReservaPeca(input.OrdemServicoID, input.PecaID, input.Quantidade)
		if err != nil {
			return err
		}

		if err := uc.reservaRepository.Salvar(ctx, nova); err != nil {
			return err
		}

		output = toReservaOutput(nova)

		return nil
	})
	if err != nil {
		return ReservaPecaOutput{}, err
	}

	return output, nil
}
