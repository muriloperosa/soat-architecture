package orcamento

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/application/notificacao"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// EnviarOrcamentoParaAprovacaoUseCase envia um orçamento válido para decisão
// do cliente. A aprovação/rejeição pertence ao status da OS; o orçamento não
// possui status próprio.
type EnviarOrcamentoParaAprovacaoUseCase struct {
	orcamentoRepository    domainorcamento.OrcamentoRepository
	ordemServicoRepository domainordemservico.OrdemServicoRepository
	notificadorStatus      *notificacao.NotificadorStatus
	notificadorOrcamento   *notificacao.NotificadorOrcamento
}

func NewEnviarOrcamentoParaAprovacaoUseCase(
	orcamentoRepository domainorcamento.OrcamentoRepository,
	ordemServicoRepository domainordemservico.OrdemServicoRepository,
	notificadorStatus *notificacao.NotificadorStatus,
	notificadorOrcamento *notificacao.NotificadorOrcamento,
) *EnviarOrcamentoParaAprovacaoUseCase {
	return &EnviarOrcamentoParaAprovacaoUseCase{
		orcamentoRepository:    orcamentoRepository,
		ordemServicoRepository: ordemServicoRepository,
		notificadorStatus:      notificadorStatus,
		notificadorOrcamento:   notificadorOrcamento,
	}
}

func (uc *EnviarOrcamentoParaAprovacaoUseCase) Executar(ctx context.Context, input EnviarOrcamentoParaAprovacaoInput) (OrcamentoOutput, error) {
	orcamento, err := uc.orcamentoRepository.BuscarPorOrdemServicoID(ctx, input.OrdemServicoID)
	if err != nil {
		if errors.Is(err, domainorcamento.ErrOrcamentoNaoEncontrado) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar orçamento", err)
	}

	os, err := uc.ordemServicoRepository.BuscarPorID(ctx, input.OrdemServicoID)
	if err != nil {
		if errors.Is(err, domainordemservico.ErrOrdemServicoNaoEncontrada) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar ordem de serviço", err)
	}

	// Valida a transição antes de validar o conteúdo para devolver o erro mais
	// representativo quando a OS estiver em um estado incompatível.
	if err := os.ValidarTransicaoPara(domainordemservico.StatusAguardandoAprovacao); err != nil {
		return OrcamentoOutput{}, err
	}
	if err := orcamento.ValidarParaEnvio(); err != nil {
		return OrcamentoOutput{}, err
	}
	statusAnterior := os.Status()
	if err := os.EnviarParaAprovacao(input.UsuarioID); err != nil {
		return OrcamentoOutput{}, err
	}

	if err := uc.ordemServicoRepository.Atualizar(ctx, os); err != nil {
		return OrcamentoOutput{}, shared.NewInternalError("erro ao atualizar ordem de serviço", err)
	}

	uc.notificadorStatus.Notificar(ctx, os, statusAnterior)
	uc.notificadorOrcamento.Notificar(ctx, os, orcamento, false)

	return toOutput(orcamento), nil
}
