package ordemservico

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type InformarDiagnosticoUseCase struct {
	repository domain.OrdemServicoRepository
}

func NewInformarDiagnosticoUseCase(repository domain.OrdemServicoRepository) *InformarDiagnosticoUseCase {
	return &InformarDiagnosticoUseCase{repository: repository}
}

func (uc *InformarDiagnosticoUseCase) Executar(
	ctx context.Context,
	input InformarDiagnosticoInput,
) (OrdemServicoOutput, error) {
	os, err := uc.repository.BuscarPorID(ctx, input.OrdemServicoID)
	if err != nil {
		if errors.Is(err, domain.ErrOrdemServicoNaoEncontrada) {
			return OrdemServicoOutput{}, err
		}
		return OrdemServicoOutput{}, shared.NewInternalError("erro ao buscar ordem de serviço", err)
	}
	if os == nil {
		return OrdemServicoOutput{}, domain.ErrOrdemServicoNaoEncontrada
	}

	if err := os.InformarDiagnostico(input.Diagnostico); err != nil {
		return OrdemServicoOutput{}, err
	}

	if err := uc.repository.Atualizar(ctx, os); err != nil {
		return OrdemServicoOutput{}, shared.NewInternalError("erro ao informar diagnóstico da ordem de serviço", err)
	}

	return toOutput(os), nil
}
