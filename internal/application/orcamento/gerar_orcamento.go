package orcamento

import (
	"context"
	"errors"

	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// GerarOrcamentoUseCase cria o orçamento de uma Ordem de Serviço. Uma OS
// possui no máximo um orçamento.
type GerarOrcamentoUseCase struct {
	orcamentoRepository    domainorcamento.OrcamentoRepository
	ordemServicoRepository domainordemservico.OrdemServicoRepository
}

func NewGerarOrcamentoUseCase(
	orcamentoRepository domainorcamento.OrcamentoRepository,
	ordemServicoRepository domainordemservico.OrdemServicoRepository,
) *GerarOrcamentoUseCase {
	return &GerarOrcamentoUseCase{
		orcamentoRepository:    orcamentoRepository,
		ordemServicoRepository: ordemServicoRepository,
	}
}

func (uc *GerarOrcamentoUseCase) Executar(ctx context.Context, input GerarOrcamentoInput) (OrcamentoOutput, error) {
	os, err := uc.ordemServicoRepository.BuscarPorID(ctx, input.OrdemServicoID)
	if err != nil {
		if errors.Is(err, domainordemservico.ErrOrdemServicoNaoEncontrada) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar ordem de serviço", err)
	}
	if os == nil {
		return OrcamentoOutput{}, domainordemservico.ErrOrdemServicoNaoEncontrada
	}
	if os.Status() != domainordemservico.StatusEmDiagnostico {
		return OrcamentoOutput{}, domainorcamento.ErrOrdemServicoNaoEmDiagnostico
	}

	_, err = uc.orcamentoRepository.BuscarPorOrdemServicoID(ctx, input.OrdemServicoID)
	if err == nil {
		return OrcamentoOutput{}, domainorcamento.ErrOrcamentoJaExiste
	}
	if !errors.Is(err, domainorcamento.ErrOrcamentoNaoEncontrado) {
		return OrcamentoOutput{}, shared.NewInternalError("erro ao verificar orçamento existente", err)
	}

	novoOrcamento, err := domainorcamento.NewOrcamento(input.OrdemServicoID, input.Observacoes, input.UsuarioID)
	if err != nil {
		return OrcamentoOutput{}, err
	}

	if err := uc.orcamentoRepository.Salvar(ctx, novoOrcamento); err != nil {
		return OrcamentoOutput{}, shared.NewInternalError("erro ao salvar orçamento", err)
	}

	return toOutput(novoOrcamento), nil
}
