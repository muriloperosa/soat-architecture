package orcamento

import (
	"context"
	"errors"

	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// AdicionarServicoOrcamentoUseCase inclui um serviço do catálogo no
// orçamento, copiando o preço e o tempo estimado vigentes no momento da
// inclusão para preservar o histórico do orçamento.
type AdicionarServicoOrcamentoUseCase struct {
	orcamentoRepository domainorcamento.OrcamentoRepository
	servicoRepository   domainservico.ServicoRepository
}

func NewAdicionarServicoOrcamentoUseCase(
	orcamentoRepository domainorcamento.OrcamentoRepository,
	servicoRepository domainservico.ServicoRepository,
) *AdicionarServicoOrcamentoUseCase {
	return &AdicionarServicoOrcamentoUseCase{
		orcamentoRepository: orcamentoRepository,
		servicoRepository:   servicoRepository,
	}
}

func (uc *AdicionarServicoOrcamentoUseCase) Executar(ctx context.Context, input AdicionarServicoOrcamentoInput) (OrcamentoOutput, error) {
	orcamento, err := uc.orcamentoRepository.BuscarPorOrdemServicoID(ctx, input.OrdemServicoID)
	if err != nil {
		if errors.Is(err, domainorcamento.ErrOrcamentoNaoEncontrado) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar orçamento", err)
	}

	servico, err := uc.servicoRepository.BuscarPorID(ctx, input.ServicoID)
	if err != nil {
		if errors.Is(err, domainservico.ErrServicoNaoEncontrado) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar serviço", err)
	}
	if !servico.Ativo() {
		return OrcamentoOutput{}, domainorcamento.ErrServicoInativo
	}

	if err := orcamento.AdicionarItemServico(
		servico.ID(),
		input.Quantidade,
		servico.PrecoBase(),
		servico.TempoEstimado().Minutos(),
	); err != nil {
		return OrcamentoOutput{}, err
	}

	if err := uc.orcamentoRepository.Atualizar(ctx, orcamento); err != nil {
		return OrcamentoOutput{}, shared.NewInternalError("erro ao atualizar orçamento", err)
	}

	atualizado, err := uc.orcamentoRepository.BuscarPorOrdemServicoID(ctx, input.OrdemServicoID)
	if err != nil {
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar orçamento atualizado", err)
	}

	return toOutput(atualizado), nil
}
