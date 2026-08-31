package orcamento

import (
	"context"
	"errors"

	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type RemoverPecaOrcamentoUseCase struct {
	orcamentoRepository    domainorcamento.OrcamentoRepository
	ordemServicoRepository domainordemservico.OrdemServicoRepository
}

func NewRemoverPecaOrcamentoUseCase(
	orcamentoRepository domainorcamento.OrcamentoRepository,
	ordemServicoRepository domainordemservico.OrdemServicoRepository,
) *RemoverPecaOrcamentoUseCase {
	return &RemoverPecaOrcamentoUseCase{orcamentoRepository: orcamentoRepository, ordemServicoRepository: ordemServicoRepository}
}

func (uc *RemoverPecaOrcamentoUseCase) Executar(ctx context.Context, input RemoverPecaOrcamentoInput) (OrcamentoOutput, error) {
	if err := validarOrcamentoEditavel(ctx, uc.ordemServicoRepository, input.OrdemServicoID); err != nil {
		return OrcamentoOutput{}, err
	}

	orcamento, err := uc.orcamentoRepository.BuscarPorOrdemServicoID(ctx, input.OrdemServicoID)
	if err != nil {
		if errors.Is(err, domainorcamento.ErrOrcamentoNaoEncontrado) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar orçamento", err)
	}

	if err := orcamento.RemoverItemPeca(input.ItemPecaID); err != nil {
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
