package ordemservico

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
)

type ConsultarOrdemServicoPorNumeroUseCase struct {
	repository domain.OrdemServicoRepository
}

func NewConsultarOrdemServicoPorNumeroUseCase(repository domain.OrdemServicoRepository) *ConsultarOrdemServicoPorNumeroUseCase {
	return &ConsultarOrdemServicoPorNumeroUseCase{repository: repository}
}

func (uc *ConsultarOrdemServicoPorNumeroUseCase) Executar(ctx context.Context, numero string) (OrdemServicoOutput, error) {
	ordemServico, err := uc.repository.BuscarPorNumero(ctx, numero)
	if err != nil {
		return OrdemServicoOutput{}, err
	}

	return toOutput(ordemServico), nil
}
