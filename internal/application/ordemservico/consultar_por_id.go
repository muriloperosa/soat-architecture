package ordemservico

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
)

type ConsultarOrdemServicoPorIDUseCase struct {
	repository domain.OrdemServicoRepository
}

func NewConsultarOrdemServicoPorIDUseCase(repository domain.OrdemServicoRepository) *ConsultarOrdemServicoPorIDUseCase {
	return &ConsultarOrdemServicoPorIDUseCase{repository: repository}
}

func (uc *ConsultarOrdemServicoPorIDUseCase) Executar(ctx context.Context, id uint64) (OrdemServicoOutput, error) {
	ordemServico, err := uc.repository.BuscarPorID(ctx, id)
	if err != nil {
		return OrdemServicoOutput{}, err
	}

	return toOutput(ordemServico), nil
}
