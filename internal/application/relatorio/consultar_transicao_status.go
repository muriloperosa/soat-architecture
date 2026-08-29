package relatorio

import (
	"context"

	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/relatorio"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type ConsultarTransicaoStatusUseCase struct {
	repository domain.RelatorioTransicaoStatusRepository
}

func NewConsultarTransicaoStatusUseCase(repository domain.RelatorioTransicaoStatusRepository) *ConsultarTransicaoStatusUseCase {
	return &ConsultarTransicaoStatusUseCase{repository: repository}
}

func (uc *ConsultarTransicaoStatusUseCase) Executar(
	ctx context.Context,
	input ConsultarTransicaoStatusInput,
) (ConsultarTransicaoStatusOutput, error) {
	deStatus, err := domainordemservico.NewStatusOrdemServico(input.DeStatus)
	if err != nil {
		return ConsultarTransicaoStatusOutput{}, err
	}

	paraStatus, err := domainordemservico.NewStatusOrdemServico(input.ParaStatus)
	if err != nil {
		return ConsultarTransicaoStatusOutput{}, err
	}

	if deStatus == paraStatus {
		return ConsultarTransicaoStatusOutput{}, domain.ErrTransicaoStatusIguais
	}

	if !deStatus.ExisteCaminhoValido(paraStatus) {
		return ConsultarTransicaoStatusOutput{}, domain.ErrTransicaoStatusSemCaminho
	}

	periodo, err := domain.NewPeriodo(input.DataInicio, input.DataFim)
	if err != nil {
		return ConsultarTransicaoStatusOutput{}, err
	}

	resultado, err := uc.repository.CalcularTransicaoStatus(ctx, domain.CalcularTransicaoStatusParams{
		FromStatus: deStatus,
		ToStatus:   paraStatus,
		Periodo:    periodo,
	})
	if err != nil {
		return ConsultarTransicaoStatusOutput{}, shared.NewInternalError("erro ao calcular relatório de transição de status", err)
	}

	return ConsultarTransicaoStatusOutput{
		TotalOrdensServico: resultado.TotalOrdens,
		TempoMedio:         resultado.DuracaoMedia,
		TempoMinimo:        resultado.DuracaoMinima,
		TempoMaximo:        resultado.DuracaoMaxima,
	}, nil
}
