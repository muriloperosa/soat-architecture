package orcamento

import (
	"context"
	"errors"
	"sort"

	"github.com/muriloperosa/soat-architecture/internal/application/notificacao"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// AlterarQuantidadePecaOrcamentoUseCase altera a quantidade no orçamento.
// Se a OS já estava APROVADA, a aprovação anterior deixa de valer: as reservas
// atuais são removidas e a OS volta para AGUARDANDO_APROVACAO. A nova reserva
// só será criada quando o cliente aprovar novamente.
type AlterarQuantidadePecaOrcamentoUseCase struct {
	orcamentoRepository    domainorcamento.OrcamentoRepository
	ordemServicoRepository domainordemservico.OrdemServicoRepository
	reservaRepository      domainreservapeca.Repository
	pecaRepository         domainpeca.Repository
	transactionRunner      shared.TransactionRunner
	notificadorStatus      *notificacao.NotificadorStatus
	notificadorOrcamento   *notificacao.NotificadorOrcamento
}

func NewAlterarQuantidadePecaOrcamentoUseCase(
	orcamentoRepository domainorcamento.OrcamentoRepository,
	ordemServicoRepository domainordemservico.OrdemServicoRepository,
	reservaRepository domainreservapeca.Repository,
	pecaRepository domainpeca.Repository,
	transactionRunner shared.TransactionRunner,
	notificadorStatus *notificacao.NotificadorStatus,
	notificadorOrcamento *notificacao.NotificadorOrcamento,
) *AlterarQuantidadePecaOrcamentoUseCase {
	return &AlterarQuantidadePecaOrcamentoUseCase{
		orcamentoRepository:    orcamentoRepository,
		ordemServicoRepository: ordemServicoRepository,
		reservaRepository:      reservaRepository,
		pecaRepository:         pecaRepository,
		transactionRunner:      transactionRunner,
		notificadorStatus:      notificadorStatus,
		notificadorOrcamento:   notificadorOrcamento,
	}
}

func (uc *AlterarQuantidadePecaOrcamentoUseCase) Executar(
	ctx context.Context,
	input AlterarQuantidadePecaOrcamentoInput,
) (OrcamentoOutput, error) {
	var atualizado *domainorcamento.Orcamento
	var reenviado bool
	var osReenviada *domainordemservico.OrdemServico

	err := uc.transactionRunner.Executar(ctx, func(txCtx context.Context) error {
		os, err := uc.ordemServicoRepository.BuscarPorID(txCtx, input.OrdemServicoID)
		if err != nil {
			return err
		}

		switch os.Status() {
		case domainordemservico.StatusEmDiagnostico, domainordemservico.StatusRejeitada:
			// edição normal; o envio continua sendo uma ação explícita.
		case domainordemservico.StatusAprovada:
			reenviado = true
		default:
			return domainorcamento.ErrOrcamentoImutavel
		}

		orcamento, err := uc.orcamentoRepository.BuscarPorOrdemServicoID(txCtx, input.OrdemServicoID)
		if err != nil {
			return err
		}
		if err := orcamento.AlterarQuantidadeItemPeca(input.ItemPecaID, input.Quantidade); err != nil {
			return err
		}
		if err := orcamento.ValidarParaEnvio(); err != nil {
			return err
		}

		if reenviado {
			reservas, err := uc.reservaRepository.BuscarPorOrdemServico(txCtx, input.OrdemServicoID)
			if err != nil {
				return err
			}

			sort.Slice(reservas, func(i, j int) bool { return reservas[i].PecaID() < reservas[j].PecaID() })
			for _, reserva := range reservas {
				if _, err := uc.pecaRepository.BuscarPorIDComBloqueio(txCtx, reserva.PecaID()); err != nil {
					return err
				}
				if err := uc.reservaRepository.Remover(txCtx, input.OrdemServicoID, reserva.PecaID()); err != nil {
					return err
				}
			}

			if err := os.EnviarParaAprovacao(input.UsuarioID); err != nil {
				return err
			}
			if err := uc.ordemServicoRepository.Atualizar(txCtx, os); err != nil {
				return err
			}

			osReenviada = os
		}

		if err := uc.orcamentoRepository.Atualizar(txCtx, orcamento); err != nil {
			return err
		}

		atualizado = orcamento
		return nil
	})
	if err != nil {
		var appErr *shared.AppError
		if errors.As(err, &appErr) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao alterar quantidade da peça no orçamento", err)
	}

	if osReenviada != nil {
		uc.notificadorStatus.Notificar(ctx, osReenviada, domainordemservico.StatusAprovada)
		uc.notificadorOrcamento.Notificar(ctx, osReenviada, atualizado, true)
	}

	return toOutput(atualizado), nil
}
