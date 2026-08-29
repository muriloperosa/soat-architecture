package relatorio

import (
	"time"

	app "github.com/muriloperosa/soat-architecture/internal/application/relatorio"
)

const (
	unidadeHoras    = "h"
	unidadeMinutos  = "m"
	unidadeSegundos = "s"
)

func toInput(request ConsultarTransicaoStatusRequest, dataInicio, dataFim time.Time) app.ConsultarTransicaoStatusInput {
	return app.ConsultarTransicaoStatusInput{
		DataInicio: dataInicio,
		DataFim:    dataFim,
		DeStatus:   request.FromStatus,
		ParaStatus: request.ToStatus,
	}
}

func toResponse(output app.ConsultarTransicaoStatusOutput, unidade string) TransicaoStatusResponse {
	return TransicaoStatusResponse{
		TotalOrdensServico: output.TotalOrdensServico,
		TempoMedio:         converterDuracao(output.TempoMedio, unidade),
		TempoMinimo:        converterDuracao(output.TempoMinimo, unidade),
		TempoMaximo:        converterDuracao(output.TempoMaximo, unidade),
		Unidade:            unidade,
	}
}

func converterDuracao(d time.Duration, unidade string) float64 {
	switch unidade {
	case unidadeMinutos:
		return d.Minutes()
	case unidadeSegundos:
		return d.Seconds()
	default:
		return d.Hours()
	}
}
