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

func (uc *ConsultarOrdemServicoPorIDUseCase) Executar(ctx context.Context, input ConsultarOrdemServicoPorIDInput) (OrdemServicoOutput, error) {
	ordemServico, err := uc.repository.BuscarPorID(ctx, input.ID)
	if err != nil {
		return OrdemServicoOutput{}, err
	}

	if err = validarAcessoConsultaOrdemServico(ordemServico, input.SolicitanteID, input.TipoSolicitante); err != nil {
		return OrdemServicoOutput{}, err
	}

	return toOutput(ordemServico), nil
}
