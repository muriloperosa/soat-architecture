package orcamento

import (
	"context"
	"errors"

	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type AprovarOrcamentoUseCase struct {
	ordemServicoRepository domainordemservico.OrdemServicoRepository
}

func NewAprovarOrcamentoUseCase(repository domainordemservico.OrdemServicoRepository) *AprovarOrcamentoUseCase {
	return &AprovarOrcamentoUseCase{ordemServicoRepository: repository}
}

func (uc *AprovarOrcamentoUseCase) Executar(ctx context.Context, input AprovarOrcamentoInput) (FluxoOrcamentoOutput, error) {
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

	if err := os.AprovarOrcamento(); err != nil {
		return FluxoOrcamentoOutput{}, err
	}
	if err := uc.ordemServicoRepository.Atualizar(ctx, os); err != nil {
		return FluxoOrcamentoOutput{}, shared.NewInternalError("erro ao atualizar ordem de serviço", err)
	}

	return toFluxoOutput(os), nil
}
