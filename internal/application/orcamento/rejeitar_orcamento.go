package orcamento

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/application/notificacao"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type RejeitarOrcamentoUseCase struct {
	ordemServicoRepository domainordemservico.OrdemServicoRepository
	notificador            *notificacao.NotificadorStatus
}

func NewRejeitarOrcamentoUseCase(
	repository domainordemservico.OrdemServicoRepository,
	notificador *notificacao.NotificadorStatus,
) *RejeitarOrcamentoUseCase {
	return &RejeitarOrcamentoUseCase{ordemServicoRepository: repository, notificador: notificador}
}

func (uc *RejeitarOrcamentoUseCase) Executar(ctx context.Context, input RejeitarOrcamentoInput) (FluxoOrcamentoOutput, error) {
	os, err := uc.ordemServicoRepository.BuscarPorID(ctx, input.OrdemServicoID)
	if err != nil {
		if errors.Is(err, domainordemservico.ErrOrdemServicoNaoEncontrada) {
			return FluxoOrcamentoOutput{}, err
		}
		return FluxoOrcamentoOutput{}, shared.NewInternalError("erro ao buscar ordem de serviço", err)
	}

	if input.ClienteID == 0 || os.ClienteID() != input.ClienteID {
		return FluxoOrcamentoOutput{}, shared.NewForbiddenError("acesso à ordem de serviço não permitido")
	}

	statusAnterior := os.Status()
	if err := os.RejeitarOrcamento(input.Motivo); err != nil {
		return FluxoOrcamentoOutput{}, err
	}
	if err := uc.ordemServicoRepository.Atualizar(ctx, os); err != nil {
		return FluxoOrcamentoOutput{}, shared.NewInternalError("erro ao atualizar ordem de serviço", err)
	}

	uc.notificador.Notificar(ctx, os, statusAnterior)

	return toFluxoOutput(os), nil
}
