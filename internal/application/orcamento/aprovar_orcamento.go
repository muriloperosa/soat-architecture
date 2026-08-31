package orcamento

import (
	"context"
	"errors"
	"sort"

	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type AprovarOrcamentoUseCase struct {
	ordemServicoRepository domainordemservico.OrdemServicoRepository
	orcamentoRepository    domainorcamento.OrcamentoRepository
	pecaRepository         domainpeca.Repository
	reservaRepository      domainreservapeca.Repository
	transactionRunner      shared.TransactionRunner
}

func NewAprovarOrcamentoUseCase(
	ordemServicoRepository domainordemservico.OrdemServicoRepository,
	orcamentoRepository domainorcamento.OrcamentoRepository,
	pecaRepository domainpeca.Repository,
	reservaRepository domainreservapeca.Repository,
	transactionRunner shared.TransactionRunner,
) *AprovarOrcamentoUseCase {
	return &AprovarOrcamentoUseCase{
		ordemServicoRepository: ordemServicoRepository,
		orcamentoRepository:    orcamentoRepository,
		pecaRepository:         pecaRepository,
		reservaRepository:      reservaRepository,
		transactionRunner:      transactionRunner,
	}
}

func (uc *AprovarOrcamentoUseCase) Executar(ctx context.Context, input AprovarOrcamentoInput) (FluxoOrcamentoOutput, error) {
	var os *domainordemservico.OrdemServico

	err := uc.transactionRunner.Executar(ctx, func(txCtx context.Context) error {
		var err error
		os, err = uc.ordemServicoRepository.BuscarPorID(txCtx, input.OrdemServicoID)
		if err != nil {
			return err
		}

		if input.ClienteID == 0 || os.ClienteID() != input.ClienteID {
			return shared.NewForbiddenError("acesso à ordem de serviço não permitido")
		}
		if err := os.ValidarTransicaoPara(domainordemservico.StatusAprovada); err != nil {
			return err
		}

		orcamento, err := uc.orcamentoRepository.BuscarPorOrdemServicoID(txCtx, input.OrdemServicoID)
		if err != nil {
			return err
		}
		if err := orcamento.ValidarParaEnvio(); err != nil {
			return err
		}

		// Agrupa itens repetidos da mesma peça. ReservaPeca é única por OS+peça.
		quantidades := make(map[uint64]int)
		for _, item := range orcamento.ItensPeca() {
			quantidades[item.PecaID()] += item.Quantidade()
		}

		// Se, por qualquer inconsistência anterior, a OS ainda possuir reservas,
		// incluímos essas peças no mesmo conjunto de locks. No fluxo normal de
		// reaprovação elas já terão sido removidas quando o orçamento foi alterado.
		reservasAtuais, err := uc.reservaRepository.BuscarPorOrdemServico(txCtx, input.OrdemServicoID)
		if err != nil {
			return err
		}

		pecasParaBloquear := make(map[uint64]struct{}, len(quantidades)+len(reservasAtuais))
		for pecaID := range quantidades {
			pecasParaBloquear[pecaID] = struct{}{}
		}
		for _, reserva := range reservasAtuais {
			pecasParaBloquear[reserva.PecaID()] = struct{}{}
		}

		// Todas as peças são bloqueadas antes de remover ou calcular reservas.
		// A ordem determinística reduz risco de deadlock entre aprovações concorrentes.
		pecaIDsBloqueados := make([]uint64, 0, len(pecasParaBloquear))
		pecasBloqueadas := make(map[uint64]*domainpeca.Peca, len(pecasParaBloquear))
		for pecaID := range pecasParaBloquear {
			pecaIDsBloqueados = append(pecaIDsBloqueados, pecaID)
		}
		sort.Slice(pecaIDsBloqueados, func(i, j int) bool { return pecaIDsBloqueados[i] < pecaIDsBloqueados[j] })

		for _, pecaID := range pecaIDsBloqueados {
			peca, err := uc.pecaRepository.BuscarPorIDComBloqueio(txCtx, pecaID)
			if err != nil {
				return err
			}
			pecasBloqueadas[pecaID] = peca
		}

		for _, reserva := range reservasAtuais {
			if err := uc.reservaRepository.Remover(txCtx, input.OrdemServicoID, reserva.PecaID()); err != nil {
				return err
			}
		}

		pecaIDs := make([]uint64, 0, len(quantidades))
		for pecaID := range quantidades {
			pecaIDs = append(pecaIDs, pecaID)
		}
		sort.Slice(pecaIDs, func(i, j int) bool { return pecaIDs[i] < pecaIDs[j] })

		for _, pecaID := range pecaIDs {
			peca := pecasBloqueadas[pecaID]

			reservada, err := uc.reservaRepository.SomarQuantidadeReservada(txCtx, pecaID)
			if err != nil {
				return err
			}

			quantidade := quantidades[pecaID]
			if !peca.PodeReservar(reservada, quantidade) {
				return domainpeca.ErrQuantidadeIndisponivelParaReserva
			}

			reserva, err := domainreservapeca.NewReservaPeca(input.OrdemServicoID, pecaID, quantidade)
			if err != nil {
				return err
			}
			if err := uc.reservaRepository.Salvar(txCtx, reserva); err != nil {
				return err
			}
		}

		if err := os.AprovarOrcamento(); err != nil {
			return err
		}
		return uc.ordemServicoRepository.Atualizar(txCtx, os)
	})
	if err != nil {
		if errors.Is(err, domainordemservico.ErrOrdemServicoNaoEncontrada) ||
			errors.Is(err, domainorcamento.ErrOrcamentoNaoEncontrado) ||
			errors.Is(err, domainpeca.ErrPecaNaoEncontrada) ||
			errors.Is(err, domainpeca.ErrQuantidadeIndisponivelParaReserva) ||
			errors.Is(err, domainorcamento.ErrOrcamentoVazio) {
			return FluxoOrcamentoOutput{}, err
		}
		var appErr *shared.AppError
		if errors.As(err, &appErr) {
			return FluxoOrcamentoOutput{}, err
		}
		return FluxoOrcamentoOutput{}, shared.NewInternalError("erro ao aprovar orçamento e reservar peças", err)
	}

	return toFluxoOutput(os), nil
}
