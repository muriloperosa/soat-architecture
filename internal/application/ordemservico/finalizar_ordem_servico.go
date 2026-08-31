package ordemservico

import (
	"context"
	"errors"
	"sort"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// FinalizarOrdemServicoUseCase conclui uma OS em execução: consome o estoque
// físico das reservas, remove as reservas e transiciona para FINALIZADA.
// Tudo ocorre na mesma transação; qualquer erro desfaz o consumo.
type FinalizarOrdemServicoUseCase struct {
	repository        domain.OrdemServicoRepository
	pecaRepository    domainpeca.Repository
	reservaRepository domainreservapeca.Repository
	transactionRunner shared.TransactionRunner
}

func NewFinalizarOrdemServicoUseCase(
	repository domain.OrdemServicoRepository,
	pecaRepository domainpeca.Repository,
	reservaRepository domainreservapeca.Repository,
	transactionRunner shared.TransactionRunner,
) *FinalizarOrdemServicoUseCase {
	return &FinalizarOrdemServicoUseCase{
		repository:        repository,
		pecaRepository:    pecaRepository,
		reservaRepository: reservaRepository,
		transactionRunner: transactionRunner,
	}
}

func (uc *FinalizarOrdemServicoUseCase) Executar(
	ctx context.Context,
	input FinalizarOrdemServicoInput,
) (OrdemServicoOutput, error) {
	var output OrdemServicoOutput

	err := uc.transactionRunner.Executar(ctx, func(ctx context.Context) error {
		os, err := uc.repository.BuscarPorIDComBloqueio(ctx, input.OrdemServicoID)
		if err != nil {
			if errors.Is(err, domain.ErrOrdemServicoNaoEncontrada) {
				return err
			}
			return shared.NewInternalError("erro ao buscar ordem de serviço", err)
		}
		if os == nil {
			return domain.ErrOrdemServicoNaoEncontrada
		}

		if err := os.ValidarTransicaoPara(domain.StatusFinalizada); err != nil {
			return err
		}

		if err := uc.consumirReservas(ctx, os.ID()); err != nil {
			return err
		}

		if err := os.Finalizar(input.UsuarioID); err != nil {
			return err
		}

		if err := uc.repository.Atualizar(ctx, os); err != nil {
			return shared.NewInternalError("erro ao finalizar ordem de serviço", err)
		}

		output = toOutput(os)
		return nil
	})
	if err != nil {
		return OrdemServicoOutput{}, err
	}

	return output, nil
}

func (uc *FinalizarOrdemServicoUseCase) consumirReservas(ctx context.Context, ordemServicoID uint64) error {
	reservas, err := uc.reservaRepository.BuscarPorOrdemServico(ctx, ordemServicoID)
	if err != nil {
		return shared.NewInternalError("erro ao carregar reservas da ordem de serviço", err)
	}

	sort.Slice(reservas, func(i, j int) bool {
		return reservas[i].PecaID() < reservas[j].PecaID()
	})

	for _, reserva := range reservas {
		peca, err := uc.pecaRepository.BuscarPorIDComBloqueio(ctx, reserva.PecaID())
		if err != nil {
			return err
		}

		if err := peca.Consumir(reserva.Quantidade()); err != nil {
			return err
		}

		if err := uc.pecaRepository.Atualizar(ctx, peca); err != nil {
			if errors.Is(err, domainpeca.ErrPecaNaoEncontrada) {
				return err
			}
			return shared.NewInternalError("erro ao consumir estoque da peça", err)
		}

		if err := uc.reservaRepository.Remover(ctx, ordemServicoID, reserva.PecaID()); err != nil {
			if errors.Is(err, domainreservapeca.ErrReservaNaoEncontrada) {
				return err
			}
			return shared.NewInternalError("erro ao remover reserva de peça", err)
		}
	}

	return nil
}
